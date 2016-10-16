package sin

import (
	"fmt"
	"math"

	"github.com/gordonklaus/portaudio"
)

const sampleRate = 44100

var step = 1 / float64(sampleRate)

type SinStream struct {
	*portaudio.Stream
	waves   map[float64]*sinState
	changes chan delta
}

type delta struct {
	add  bool
	freq float64
	vel  float64
}

type sinState struct {
	freq     float64
	phase    float64
	duration float64
	vel      float64
}

func New() *SinStream {
	s := &SinStream{waves: make(map[float64]*sinState), changes: make(chan delta, 100)}
	var err error
	s.Stream, err = portaudio.OpenDefaultStream(0, 2, sampleRate, 0, s.processAudio)
	chk(err)
	return s
}

func (s *SinStream) Add(freq, vel float64) {
	fmt.Printf("Add wave! freq: %f, vel: %f\n", freq, vel)
	s.changes <- delta{true, freq, vel}
}

func (s *SinStream) Remove(freq float64) {
	fmt.Printf("Remove wave! %f\n", freq)
	s.changes <- delta{false, freq, 0}
}

func (s *SinStream) processAudio(out [][]float32) {
	for i := range out[0] {
		out[0][i] = 0
		out[1][i] = 0
	}
	select {
	case d := <-s.changes:
		if d.add {
			fmt.Printf("Add delta recieved\n")
			s.waves[d.freq] = &sinState{d.freq, 0, 5, d.vel}
		} else {
			fmt.Printf("Remove delta recieved\n")
			delete(s.waves, d.freq)
		}
	default:
	}

	for i := range out[0] {
		for k, w := range s.waves {
			v := float32(w.vel * math.Sin(2*math.Pi*w.freq*w.phase))
			out[0][i] += v
			out[1][i] += v
			w.phase += step
			if w.phase > w.duration {
				s.waves[k] = nil
				fmt.Println("done wave!")
			}
		}
	}
}

func (s *SinStream) Image() []byte {
	b := make([]byte, 200*100*4)
	for x := 0; x < 200; x++ {
		for y := 0; y < 100; y++ {
			b[4*(x+200*y)] = 0x00
			b[4*(x+200*y)+1] = 0x00
			b[4*(x+200*y)+2] = 0x00
			b[4*(x+200*y)+3] = 0xff
		}
		y := int(50*math.Sin(2*math.Pi*(1/float64(200))*float64(x)) + 50)
		if y > 99 {
			y = 99
		}
		b[4*(x+200*y)] = 0xff
		b[4*(x+200*y)+1] = 0xff
		b[4*(x+200*y)+2] = 0xff
		b[4*(x+200*y)+3] = 0xff
	}
	return b
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
