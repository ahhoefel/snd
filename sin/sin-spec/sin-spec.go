package main

import (
	"fmt"
	"math"
	"time"

	"github.com/ahhoefel/snd/fft/spec"
	"github.com/ahhoefel/snd/image"
	"github.com/gordonklaus/portaudio"
)

const sampleRate = 44100
const freq = 820

var step = 1 / float64(sampleRate)

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	s := newStream()
	defer s.Close()
	chk(s.Start())
	time.Sleep(2 * time.Second)
	chk(s.Stop())

	var x []float64
	var avg float32
	for _, y := range s.savedOutput {
		x = append(x, float64(y))
		if y > 0 {
			avg += y
		} else {
			avg += -y
		}

	}
	fmt.Printf("avg %f", avg/float32(len(x)))
	image.Write("spec.png", spec.Spec(x))
	image.Write("wave.png", image.WaveImage(x))
}

type stream struct {
	*portaudio.Stream
	freq        float64
	phase       float64
	savedOutput []float32
	index       int
}

func newStream() *stream {
	s := &stream{nil, freq, 0, nil, 0}
	var err error
	s.Stream, err = portaudio.OpenDefaultStream(0, 2, sampleRate, 0, s.processAudio)
	chk(err)
	return s
}

func (s *stream) processAudio(out [][]float32) {
	for i := range out[0] {
		v := float32(math.Sin(2 * math.Pi * s.freq * s.phase))
		s.phase += step
		s.index++
		if s.index%15 == 0 {
			//s.phase *= s.freq / (s.freq + 1)
			s.freq += 1
		}
		out[0][i] = v
		out[1][i] = v
		s.savedOutput = append(s.savedOutput, v)
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
