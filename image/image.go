package image

import (
	"bufio"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/cmplx"
	"os"
	"path"
	"sort"
)

var (
	basePath = "/Users/hoefel/Development/go/src/github.com/ahhoefel/snd/out"
)

func Colorize(v []float64, c []color.RGBA) {
	for i, a := range v {
		y := 0xff * a
		var r, b byte
		if y < 0 {
			b = byte(-y)
		} else {
			r = byte(y)
		}
		c[i] = color.RGBA{r, 0, b, 0xff}
	}
}

func ComplexColor(c complex128) color.RGBA {
	r, theta := cmplx.Polar(c) // * complex(1/math.Sqrt(2), 1/math.Sqrt(2)))
	if theta <= -2*math.Pi/3 {
		return color.RGBA{uint8(r * 0xff), uint8(r * 0xff * 3 * (theta + math.Pi) / math.Pi), 0x00, 0xff}
	}
	if theta <= -math.Pi/3 {
		return color.RGBA{uint8(r * 0xff * 3 * (-math.Pi/3 - theta) / math.Pi), uint8(r * 0xff), 0x00, 0xff}
	}
	if theta <= 0 {
		return color.RGBA{0x00, uint8(r * 0xff), uint8(r * 0xff * 3 * (theta + math.Pi/3) / math.Pi), 0xff}
	}
	if theta <= math.Pi/3 {
		return color.RGBA{0x00, uint8(r * 0xff * 3 * (math.Pi/3 - theta) / math.Pi), uint8(r * 0xff), 0xff}
	}
	if theta <= 2*math.Pi/3 {
		return color.RGBA{uint8(r * 0xff * 3 * (theta - math.Pi/3) / math.Pi), 0x00, uint8(r * 0xff), 0xff}
	}
	return color.RGBA{uint8(r * 0xff), 0x00, uint8(r * 0xff * 3 * (math.Pi - theta) / math.Pi), 0xff}
}

func Write(fileName string, img image.Image) error {
	f, err := os.Create(path.Join(basePath, fileName))
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	err = png.Encode(w, img)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

const imageWidth = 600

func WaveImage(w []float64) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, imageWidth, 200))
	for x := 0; x < imageWidth; x++ {
		for y := 0; y < 200; y++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0xff})
		}
	}
	scale := float64(len(w)) / imageWidth
	var buf []float64
	for x := 0; x < imageWidth; x++ {
		iMin, iMax := int(float64(x)*scale), int(float64(x+1)*scale)
		buf = buf[:0]
		for i := iMin; i < iMax; i++ {
			buf = append(buf, w[i])
		}
		sort.Float64s(buf)
		var i int
		for y := 0; y < 200; y++ {
			for i < len(buf) && 100*buf[i]+100 < float64(y) {
				i++
			}
			var c color.RGBA
			if y < 100 {
				b := uint8(0xff * i / len(buf))
				c = color.RGBA{b, b, b, 0xff}
			} else {
				b := byte(0xff * (len(buf) - i) / len(buf))
				c = color.RGBA{b, b, b, 0xff}
			}
			img.Set(x, y, c)
		}
	}
	return img
}
