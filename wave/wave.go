package wave

// Notes to self on unit conversion.
// From seconds to steps:
// Given t seconds, s (steps) = t (sec) * constants.SampleRate (steps/sec).
// From frequence to wave length:
// Given f freq, wavelength (sec) = 1/f.
// Given f freq, wavelength (steps) = constants.SampleRate / f

import (
	"encoding/json"
	"math"

	"github.com/ahhoefel/snd/constants"
)

type FlexSettings struct {
	Spike          float64
	Skew           float64
	Flat           float64
	StaticBuffer   []float32
	Overtones      byte
	OvertoneSpacer float64
	OvertoneDecay  float32
	Delays         byte
	DelayTime      float64
	DelayDecay     float32
	StaticBuffer2  []float32
}

const staticBufferLength = 256

func NewFlexSettings() *FlexSettings {
	fs := &FlexSettings{
		Spike:          0.5,
		Skew:           0.5,
		Flat:           0.5,
		StaticBuffer:   make([]float32, staticBufferLength),
		Overtones:      0,
		OvertoneSpacer: 1.0,
		OvertoneDecay:  .25,
		Delays:         0,
		DelayTime:      0,
		DelayDecay:     0,
		StaticBuffer2:  make([]float32, staticBufferLength),
	}
	fs.updateStaticBuffer()
	return fs
}

func (s *FlexSettings) SetSpike(spike float64) {
	s.Spike = spike
	s.updateStaticBuffer()
}

func (s *FlexSettings) SetSkew(skew float64) {
	s.Skew = skew
	s.updateStaticBuffer()
}

func (s *FlexSettings) SetFlat(flat float64) {
	s.Flat = flat
	s.updateStaticBuffer()
}

func (s *FlexSettings) SetOvertones(overtones byte) {
	s.Overtones = overtones
	s.updateStaticBuffer()
}

func (s *FlexSettings) SetOvertoneSpacing(spacer float64) {
	s.OvertoneSpacer = spacer
	s.updateStaticBuffer()
}

func (s *FlexSettings) SetOvertoneDecay(decay float32) {
	s.OvertoneDecay = decay
	s.updateStaticBuffer()
}

func (s *FlexSettings) SetDelays(delays byte) {
	s.Delays = delays
	s.updateStaticBuffer()
}

func (s *FlexSettings) SetDelayTime(delay float64) {
	s.DelayTime = delay
	s.updateStaticBuffer()
}

func (s *FlexSettings) SetDelayDecay(delayDecay float32) {
	s.DelayDecay = delayDecay
	s.updateStaticBuffer()
}

func (s *FlexSettings) newUpdateStaticBuffer() {
	for x := range s.StaticBuffer {
		s.StaticBuffer[x] = 0
	}
	for x := range s.StaticBuffer {
		xf := float64(x)
		a := spikeHeight(xf/staticBufferLength, s.Skew)
		f := flatHeight(xf / staticBufferLength)
		c := math.Sin(2 * math.Pi * xf / float64(staticBufferLength))
		y := s.Flat*f + (1-s.Flat)*(s.Spike*a+(1-s.Spike)*c)
		s.StaticBuffer[x] = float32(y)
	}
}

func (s *FlexSettings) updateStaticBuffer() {
	for x := range s.StaticBuffer {
		s.StaticBuffer[x] = 0
	}
	vol := float32(1)
	for t := byte(0); t < s.Overtones+1; t++ {
		for x := range s.StaticBuffer {
			xf := math.Mod(float64(x)*(1/float64(staticBufferLength))*(float64(t)*s.OvertoneSpacer+1), 1)
			a := spikeHeight(xf, s.Skew)
			f := flatHeight(xf)
			c := math.Sin(2 * math.Pi * xf)
			y := s.Flat*f + (1-s.Flat)*(s.Spike*a+(1-s.Spike)*c)
			s.StaticBuffer[x] += vol * float32(y)
		}
		vol *= s.OvertoneDecay
	}
}

func (s *FlexSettings) Image(w Wave) []byte {
	h := make([]float32, 400)
	w.AddToBuffer(h)
	b := make([]byte, 400*100*4)
	for x := 0; x < 400; x++ {
		for y := 0; y < 100; y++ {
			b[4*(x+200*y)] = 0x00
			b[4*(x+200*y)+1] = 0x00
			b[4*(x+200*y)+2] = 0x00
			b[4*(x+200*y)+3] = 0xff
		}
		y := int(40*h[x] + 50)
		if y > 99 {
			y = 99
		}
		if y < 0 {
			y = 0
		}
		b[4*(x+200*y)] = 0xff
		b[4*(x+200*y)+1] = 0xff
		b[4*(x+200*y)+2] = 0xff
		b[4*(x+200*y)+3] = 0xff
	}
	return b
}

func (s *FlexSettings) Json() ([]byte, error) {
	waveform := []float64{s.Spike, s.Skew, s.Flat, 0, float64(s.Overtones), s.OvertoneSpacer, float64(s.OvertoneDecay), 0}
	decay := []float64{float64(s.Delays), s.DelayTime, float64(s.DelayDecay), 0, 0, 0, 0, 0}
	return json.Marshal([][]float64{waveform, decay})
}

func spikeHeight(x float64, skew float64) float64 {
	if x < 0.5*skew {
		return x * 2 / skew
	} else if x < 0.5 {
		return (0.5 - x) * 2 / (1 - skew)
	}
	x = x - 0.5
	if x < 0.5*skew {
		return -x * 2 / skew
	}
	return -(0.5 - x) * 2 / (1 - skew)
}

func flatHeight(x float64) float64 {
	if x < 0.5 {
		return 1
	}
	return -1
}

type FlexWave struct {
	Settings *FlexSettings
	Freq     float64
	Phase    float64
	Duration float64
	Vel      float64
}

func (w *FlexWave) AddToBuffer(buf []float32) {
	n := int((w.Duration - w.Phase) / constants.Step)
	if len(buf) < n {
		n = len(buf)
	}
	for i := 0; i < n; i++ {
		x := math.Mod(w.Phase, 1/w.Freq)
		a := spikeHeight(x*w.Freq, w.Settings.Skew)
		f := flatHeight(x * w.Freq)
		c := math.Sin(2 * math.Pi * w.Freq * w.Phase)
		buf[i] += float32(w.Vel * (w.Settings.Flat*f + (1-w.Settings.Flat)*(w.Settings.Spike*a+(1-w.Settings.Spike)*c)))
		w.Phase += constants.Step
	}
}

func (w FlexWave) IsDone() bool {
	return w.Phase > w.Duration
}

type Wave interface {
	AddToBuffer(buf []float32)
	IsDone() bool
}

type StaticWave struct {
	Settings *FlexSettings
	Phase    float64
	Step     float64
	Duration float64
	Vel      float32
}

func NewStaticWave(fs *FlexSettings, freq, duration float64, vel float32) *StaticWave {
	step := staticBufferLength * freq / constants.SampleRate
	return &StaticWave{fs, 0, step, duration * constants.SampleRate, vel}
}

func (w *StaticWave) AddToBuffer(buf []float32) {
	if w.Duration < 0 {
		w.Vel = w.Vel - 0.1
		if w.IsDone() {
			return
		}
	}
	i := 0
	if start := int(math.Ceil(-w.Phase / w.Step)); start >= 0 {
		i = start
	}
	if i >= len(buf) {
		w.Phase += w.Step * float64(len(buf))
	} else {
		w.Phase += w.Step * float64(i)
		if w.Phase < 0 {
			panic("NEGATIVE PHASE")
		}
	}
	for int(w.Phase) >= staticBufferLength {
		w.Phase -= staticBufferLength
	}

	for n := len(buf); i < n; i++ {
		buf[i] += w.Vel * w.Settings.StaticBuffer[int(w.Phase)]
		w.Duration -= w.Step
		w.Phase += w.Step
		for int(w.Phase) >= staticBufferLength {
			w.Phase -= staticBufferLength
		}
	}
}

func (w *StaticWave) IsDone() bool {
	return w.Vel < 0.05
}

type EchoWave struct {
	Settings *FlexSettings
	Waves    []*StaticWave
	Duration float64
	Step     float64
	Vel      float32
}

func NewEchoWave(fs *FlexSettings, freq, duration float64, vel float32) *EchoWave {
	var waves []*StaticWave
	step := staticBufferLength * freq / constants.SampleRate
	dur := duration * constants.SampleRate
	vol := vel
	for i := byte(0); i < fs.Delays; i++ {
		waves = append(waves, &StaticWave{fs, -fs.DelayTime * float64(i) * constants.SampleRate, step, dur, vol})
		vol *= fs.DelayDecay
	}
	return &EchoWave{fs, waves, fs.DelayTime*float64(fs.Delays)*constants.SampleRate + dur, step, vel}
}

func (w *EchoWave) AddToBuffer(buf []float32) {
	if w.Duration < 0 {
		w.Vel = w.Vel - 0.1
		if w.IsDone() {
			return
		}
	}
	for _, sw := range w.Waves {
		sw.AddToBuffer(buf)
	}
	w.Duration -= w.Step * float64(len(buf))
}

func (w *EchoWave) IsDone() bool {
	return w.Duration < 0
}
