package main

import (
	"image"
	"math/cmplx"

	imgpkg "github.com/ahhoefel/snd/image"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, 201, 201))
	for x := -100; x <= 100; x++ {
		for y := -100; y <= 100; y++ {
			z := complex(float64(x)/100, float64(y)/100)
			if cmplx.Abs(z) > 1 {
				z = 0 // complex(1/cmplx.Abs(z), 0) * z
			}
			img.Set(x+100, -y+100, imgpkg.ComplexColor(z))
		}
	}
	imgpkg.Write("complexcolor.png", img)
}
