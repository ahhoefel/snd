package main

import (
	"fmt"
	"time"

	"github.com/ahhoefel/snd/sin"
	"github.com/gordonklaus/portaudio"
)

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	s := sin.New()
	s.Add(440, 1)
	defer s.Close()
	chk(s.Start())
	time.Sleep(3 * time.Second)
	chk(s.Stop())
	fmt.Println("done!")
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
