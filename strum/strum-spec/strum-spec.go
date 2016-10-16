package main

import (
	"fmt"
	"time"

	"github.com/ahhoefel/snd/fft/spec"
	"github.com/ahhoefel/snd/image"
	"github.com/ahhoefel/snd/strum"
	"github.com/gordonklaus/portaudio"
)

const sampleRate = 44100

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	w := strum.New(800, 400, 40)
	s := newStream(w)
	defer s.Close()
	chk(s.Start())
	time.Sleep(5 * time.Second)
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
}

type stream struct {
	*portaudio.Stream
	wave        *strum.Strum
	maxPull     float64
	savedOutput []float32
}

func newStream(wave *strum.Strum) *stream {
	s := &stream{nil, wave, 0, nil}
	var err error
	s.Stream, err = portaudio.OpenDefaultStream(0, 2, sampleRate, 0, s.processAudio)
	chk(err)
	return s
}

func (s *stream) processAudio(out [][]float32) {
	for i := range out[0] {
		p := s.wave.Pull()
		if p > s.maxPull {
			s.maxPull = p
		}
		v := float32(p / s.maxPull)
		out[0][i] = v
		out[1][i] = v
		s.savedOutput = append(s.savedOutput, v)
		s.wave.Step()
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
