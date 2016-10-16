package spec

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/cmplx"
	"sort"

	"github.com/ahhoefel/snd/fft"
	imgpkg "github.com/ahhoefel/snd/image"
)

const (
	singalFreq       = 44100
	downSampleFactor = 2
	subsampleFreq    = singalFreq / downSampleFactor
	frameWidth       = 1 << 10
	imageWidth       = 1 << 10
	startFreq        = 110
	pixelsPerOctave  = 160
	numOctaves       = 6
	imageHeight      = numOctaves * pixelsPerOctave
)

var (
	hannMultiplier []float64
)

func frameStep(samples int) int {
	return ((samples - frameWidth) / imageWidth)
}

func downSample(x []float64, factor int) []float64 {
	m := len(x) / factor
	var y []float64
	for i := 0; i < m; i++ {
		var v float64
		for j := 0; j < factor; j++ {
			v += x[factor*i+j]
		}
		y = append(y, v/float64(factor))
	}
	return y
}

func subFrame(from []float64, to []float64, start int) {
	copy(to, from[start:start+frameWidth])
	hannWindow(to)
}

func hannWindow(a []float64) {
	if hannMultiplier == nil {
		hannMultiplier = make([]float64, frameWidth)
		for i := range hannMultiplier {
			hannMultiplier[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(len(a)-1)))
		}
	}
	for i, c := range hannMultiplier {
		a[i] *= c
	}
}

func logSpec(x []float64, log bool) []float64 {
	var s []float64
	factor := math.Pow(2, 1/float64(pixelsPerOctave))
	f := float64(startFreq)
	c := float64(len(x)) / subsampleFreq
	if log {
		fmt.Printf("Factor: %f, Startfreq: %f, constant: %f\n", factor, f, c)
	}
	for i := 0; i < imageHeight; i++ {
		g := f * factor
		s = append(s, average(x[int(c*f):int(c*g)+1]))
		if log {
			fmt.Printf("i: %d, f: %f, avg: %f\n", i, f, s[i])
		}
		f = g
	}
	return s
}

func average(x []float64) float64 {
	var sum float64
	for _, v := range x {
		sum += v
	}
	return sum / float64(len(x))
}

func normalize(x []float64) []float64 {
	a := make([]float64, len(x))
	for i, e := range x {
		if e > 0 {
			a[i] = e
		} else {
			a[i] = -e
		}
	}
	sort.Float64s(a)
	v := a[int(float64(len(a))*0.99)]
	for i, e := range x {
		if e > v {
			a[i] = 1
		} else {
			a[i] = e / v
		}
	}
	return a
}

func normalizeComplex(c []complex128) {
	a := make([]float64, len(c))
	for i, z := range c {
		a[i] = cmplx.Abs(z)
	}
	sort.Float64s(a)
	v := a[int(float64(len(a))*0.99)]
	for i, z := range c {
		if m := cmplx.Abs(z); m > v {
			c[i] = complex(1/m, 0) * z
		} else {
			c[i] = complex(1/v, 0) * z
		}
	}
}

func Spec(x []float64) *image.RGBA {
	x = downSample(x, downSampleFactor)
	fs := frameStep(len(x))
	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	f := make([]float64, frameWidth)
	colors := make([]color.RGBA, imageHeight)
	for i := 0; i < imageWidth; i++ {
		subFrame(x, f, i*fs)
		if i == 0 {
			imgpkg.Write("prefft.png", imgpkg.WaveImage(f))
		}
		c := fft.FFTComplex(f)
		normalizeComplex(c)
		// if i == 0 {
		//  imgpkg.Write("fft.png", imgpkg.WaveImage(f))
		// }
		var v []float64
		//if i == 0 {
		//	v = logSpec(f, true)
		//} else {
		//		v = logSpec(f, false)
		//		}
		// if i == 0 {
		//   imgpkg.Write("logspec.png", imgpkg.WaveImage(v))
		// }
		imgpkg.Colorize(v, colors)
		for j, z := range c {
			img.Set(i, j, imgpkg.ComplexColor(z))
		}
	}
	return img
}
