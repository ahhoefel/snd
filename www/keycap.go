package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/ahhoefel/snd/sin"
	"github.com/gordonklaus/portaudio"

	"golang.org/x/net/websocket"
)

var (
	ratios = []float64{
		1,
		256 / float64(243),
		9 / float64(8),
		32 / float64(27),
		81 / float64(64),
		4 / float64(3),
		729 / float64(519),
		1027 / float64(729),
		3 / float64(2),
		128 / float64(81),
		27 / float64(16),
		16 / float64(9),
		243 / float64(128),
	}
)

type piano struct {
	s *sin.SinStream
}

type command int

const (
	IMAGE command = iota
	TEXT
	MIDI
)

type message struct {
	Cmd  command
	Data []byte
}

func (p *piano) play(conn *websocket.Conn) {
	b := make([]byte, 100)
	fmt.Printf("%#v\n", conn.Config())
	sendMessage(conn, message{TEXT, []byte("Howdy Hey!")})
	sendMessage(conn, message{IMAGE, p.s.Image()})
	for true {
		n, err := conn.Read(b)
		if err != nil {
			panic(err)
		}
		if n == 1 {
			freq := 440 * ratios[int(b[0]-'a')]
			fmt.Printf("%c, %d, %f\n", b[0], int(b[0]-'a'), freq)
			//p.s.Add(freq, 1)
		} else {
			fmt.Printf("Read %d bytes\n", n)
			x, y, z := b[0], b[1], b[2]
			fmt.Printf("%d, %d, %d\n", x, y, z)
			freq := 440 * math.Pow(2, float64(int(y)-69)/12)
			vel := float64(z) / 0xff
			if x == 144 {
				p.s.Add(freq, vel)
			}
			if x == 128 {
				p.s.Remove(freq)
			}
		}
	}
}

func sendMessage(conn *websocket.Conn, msg message) {
	m, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Sending %q\n", m)
	for len(m) > 0 {
		n, err := conn.Write(m)
		fmt.Printf("Sent %q\n", m[:n])
		if err != nil {
			panic(err)
		}
		m = m[n:]
	}
}

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	p := &piano{sin.New()}
	defer p.s.Close()
	chk(p.s.Start())
	http.Handle("/echo", websocket.Handler(p.play))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
	chk(p.s.Stop())
	fmt.Println("done!")
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
