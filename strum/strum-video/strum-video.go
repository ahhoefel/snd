package main

import (
	"github.com/ahhoefel/snd/strum"
	"github.com/ahhoefel/snd/video"
)

func main() {
	f, err := video.New("strum.rgb", 100, 100)
	if err != nil {
		panic(err)
	}
	s := strum.New(98, 100, 40)
	for i := 0; i < 1000; i++ {
		step(f, s)
	}
	f.Close()
}

func step(f *video.Frame, s *strum.Strum) {
	for _, n := range s.Nodes {
		x, y := int(n.X), int(n.Y+50)
		if f.Contains(x, y) {
			f.WritePx(x, y, 0)
		}
	}
	s.Step()
	for _, n := range s.Nodes {
		x, y := int(n.X), int(n.Y+50)
		if f.Contains(x, y) {
			f.WritePx(x, y, 0xff)
		}
	}
	if err := f.Write(); err != nil {
		panic(err)
	}
}
