package main

import (
	"github.com/gordonklaus/portaudio"
	"math"
	"time"
)

const sampleRate = 44100

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	s := newStream([]float64{220, 440, 660, 880}, []float64{8, 4, 2, 1})
	//s := newStream([]float64{220, 441, 881, 1761}, []float64{1, 0.5, .25, .125})
	defer s.Close()
	chk(s.Start())
	time.Sleep(5 * time.Second)
	chk(s.Stop())
}

type stream struct {
	*portaudio.Stream
	signals []*sine
}

func newStream(freqs []float64, weights []float64) *stream {
	var w float64
	for _, v := range weights {
		w += v
	}
	var signals []*sine
	for i, f := range freqs {
		signals = append(signals, &sine{f / sampleRate, 0, weights[i] / w})
	}
	s := &stream{nil, signals}
	var err error
	s.Stream, err = portaudio.OpenDefaultStream(0, 2, sampleRate, 0, s.processAudio)
	chk(err)
	return s
}

func (s *stream) processAudio(out [][]float32) {
	for i := range out[0] {
		var v float32
		for _, w := range s.signals {
			v += w.next()
		}
		out[0][i] = v
		out[1][i] = v
	}
}

type sine struct {
	step, phase, amp float64
}

func (s *sine) next() float32 {
	v := float32(s.amp * math.Sin(2*math.Pi*s.phase))
	_, s.phase = math.Modf(s.phase + s.step)
	return v
}

type stereoSine struct {
	*portaudio.Stream
	stepL, phaseL float64
	stepR, phaseR float64
}

func newStereoSine(freqL, freqR, sampleRate float64) *stereoSine {
	s := &stereoSine{nil, freqL / sampleRate, 0, freqR / sampleRate, 0}
	var err error
	s.Stream, err = portaudio.OpenDefaultStream(0, 2, sampleRate, 0, s.processAudio)
	chk(err)
	return s
}

func (g *stereoSine) processAudio(out [][]float32) {
	for i := range out[0] {
		out[0][i] = float32(math.Sin(2 * math.Pi * g.phaseL))
		_, g.phaseL = math.Modf(g.phaseL + g.stepL)
		out[1][i] = float32(math.Sin(2 * math.Pi * g.phaseR))
		_, g.phaseR = math.Modf(g.phaseR + g.stepR)
	}
}

func (g *stereoSine) processAudio2(out [][]float32) {
	for i := range out[0] {
		out[0][i] = 0.5*float32(math.Sin(2*math.Pi*g.phaseL)) + 0.5*float32(math.Sin(2*math.Pi*g.phaseR))
		out[1][i] = out[0][i]
		_, g.phaseL = math.Modf(g.phaseL + g.stepL)
		_, g.phaseR = math.Modf(g.phaseR + g.stepR)
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
