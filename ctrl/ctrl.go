package ctrl

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/ahhoefel/snd/constants"
	"github.com/ahhoefel/snd/wave"
	"github.com/gordonklaus/portaudio"

	"golang.org/x/net/websocket"
)

var debug bool

type Ctrl struct {
	audio        *portaudio.Stream
	waves        map[float64]wave.Wave
	Changes      chan delta
	buf          []float32
	fs           *wave.FlexSettings
	waveSelector int
	bank         knobBank
}

type delta struct {
	add  bool
	freq float64
	vel  float32
}

type command int

const (
	IMAGE command = iota
	TEXT
	MIDI
	KnobBank
	KnobValues
)

type knobBank int

const (
	waveBank knobBank = iota
	delayBank
	numKnobBanks
)

type message struct {
	Cmd  command
	Data []byte
}

func New() *Ctrl {
	c := &Ctrl{
		waves:   make(map[float64]wave.Wave),
		Changes: make(chan delta, 100),
		fs:      wave.NewFlexSettings(),
	}
	return c
}

func (c *Ctrl) Run() {
	fmt.Println("Starting audio.")
	portaudio.Initialize()
	defer portaudio.Terminate()
	var err error
	c.audio, err = portaudio.OpenDefaultStream(0, 2, constants.SampleRate, 0, c.processAudio)
	chk(err)
	chk(c.audio.Start())
	defer c.audio.Close()
	http.Handle("/echo", websocket.Handler(c.onMessage))
	fmt.Println("Listening to websocket.")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
	chk(c.audio.Stop())
	fmt.Println("Ctrl done running!")
}

func (c *Ctrl) onMessage(conn *websocket.Conn) {
	b := make([]byte, 100)
	fmt.Printf("%#v\n", conn.Config())
	sendMessage(conn, message{TEXT, []byte("Howdy Hey!")})
	sendMessage(conn, message{IMAGE, c.fs.Image(c.newWave(constants.SampleRate/200, 1))})
	for true {
		n, err := conn.Read(b)
		if err != nil {
			panic(err)
		}

		if n == 3 {
			x, y, z := b[0], b[1], b[2]
			if debug {
				fmt.Printf("Read %d bytes\n", n)
				fmt.Printf("%d, %d, %d\n", x, y, z)
			}
			freq := 440 * math.Pow(2, float64(int(y)-69)/12)
			vel := float32(z) / 0xff

			switch x {
			case 144: // Key down
				c.Changes <- delta{true, freq, vel}
			case 128: // Key up
				c.Changes <- delta{false, freq, 0}
			case 176: // Knob
				switch c.bank {
				case waveBank:
					switch y {
					case 1:
						c.fs.SetSpike(float64(z) / 127)
					case 2:
						c.fs.SetSkew(float64(z) / 127)
					case 3:
						c.fs.SetFlat(float64(z) / 127)
					case 5:
						c.fs.SetOvertones(z / 16)
					case 6:
						c.fs.SetOvertoneSpacing(float64(z) / 127)
					case 7:
						c.fs.SetOvertoneDecay(float32(z) / 127)
					}
				case delayBank:
					switch y {
					case 1:
						c.fs.SetDelays(z / 16)
					case 2:
						c.fs.SetDelayTime(float64(z) / 127)
					case 3:
						c.fs.SetDelayDecay(float32(z) / 127)
					}
				}
				sendMessage(conn, message{IMAGE, c.fs.Image(c.newWave(constants.SampleRate/200, 1))})
				valuesJson, err := c.fs.Json()
				if err == nil {
					if debug {
						fmt.Printf("Sending json: %v\n", string(valuesJson))
					}
					sendMessage(conn, message{KnobValues, valuesJson})
				} else {
					panic(err)
				}
			case 192: // Program pad.
				switch y {
				case 0:
					debug = !debug
					msg := fmt.Sprintf("Debug: %v", debug)
					sendMessage(conn, message{TEXT, []byte(msg)})
				case 1:
					c.waveSelector = (c.waveSelector + 1) % len(waveSelectors)
					msg := fmt.Sprintf("Wave selector: %#v", waveSelectors[c.waveSelector].name)
					sendMessage(conn, message{TEXT, []byte(msg)})
				case 2:
					c.bank = (c.bank + 1) % numKnobBanks
					msg := fmt.Sprintf("Knob bank: %#v", c.bank)
					sendMessage(conn, message{TEXT, []byte(msg)})
					sendMessage(conn, message{KnobBank, []byte{byte(c.bank)}})
				}
			}
		} else {
			panic("not three bytes!")
		}
	}
}

func sendMessage(conn *websocket.Conn, msg message) {
	m, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("Sending %q\n", m)
	for len(m) > 0 {
		n, err := conn.Write(m)
		//fmt.Printf("Sent %q\n", m[:n])
		if err != nil {
			panic(err)
		}
		m = m[n:]
	}
}

func (c *Ctrl) processAudio(out [][]float32) {
	n := len(out[0])
	c.prepareBuf(n)
	c.commitChanges()

	// Add to buffer
	for k, w := range c.waves {
		w.AddToBuffer(c.buf)
		if w.IsDone() {
			delete(c.waves, k)
		}
	}

	// Write to output
	for i, v := range c.buf {
		out[0][i] = v
		out[1][i] = v
	}
}

func (c *Ctrl) prepareBuf(n int) {
	if len(c.buf) < n {
		c.buf = make([]float32, n)
	} else {
		c.buf = c.buf[:n]
		for i := range c.buf {
			c.buf[i] = 0
		}
	}
}

func (c *Ctrl) commitChanges() {
	for {
		select {
		case d := <-c.Changes:
			if d.add {
				if debug {
					fmt.Printf("Add delta recieved\n")
				}
				c.waves[d.freq] = c.newWave(d.freq, d.vel)
			} else {
				if debug {
					fmt.Printf("Remove delta recieved\n")
				}
				delete(c.waves, d.freq)
			}
		default:
			return
		}
	}
}

func (c *Ctrl) newWave(freq float64, vel float32) wave.Wave {
	return waveSelectors[c.waveSelector].fn(c.fs, freq, vel)
}

type waveSelector struct {
	fn   func(fs *wave.FlexSettings, freq float64, vel float32) wave.Wave
	name string
}

var waveSelectors = []waveSelector{
	{echoWaveSelector, "echo wave selector"},
	{staticWaveSelector, "static wave selector"},
	{flexWaveSelector, "flex wave selector"},
}

func staticWaveSelector(fs *wave.FlexSettings, freq float64, vel float32) wave.Wave {
	return wave.NewStaticWave(fs, freq, 5, vel)
}

func flexWaveSelector(fs *wave.FlexSettings, freq float64, vel float32) wave.Wave {
	return &wave.FlexWave{fs, freq, 0, 5, float64(vel)}
}

func echoWaveSelector(fs *wave.FlexSettings, freq float64, vel float32) wave.Wave {
	return wave.NewEchoWave(fs, freq, 5, vel)
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
