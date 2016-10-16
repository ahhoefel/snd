package video

import "os"

func (f *Frame) Write() error {
	var n int
	var err error
	b := f.b
	for len(b) > 0 && err == nil {
		n, err = f.file.Write(b)
		b = b[n:]
	}
	return err
}

func (f *Frame) Close() {
	f.file.Close()
}

type Frame struct {
	file   *os.File
	width  int
	height int
	enc    int
	b      []byte
}

func New(fileName string, width, height int) (*Frame, error) {
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	enc := 3
	return &Frame{file, width, height, enc, make([]byte, width*height*enc)}, nil
}

func (f *Frame) Contains(x, y int) bool {
	return 0 <= x && x < f.width && 0 <= y && y < f.height
}

func (f *Frame) WritePx(x, y int, a byte) {
	i := (x + f.width*y) * f.enc
	for j := 0; j < f.enc; j++ {
		f.b[i+j] = a
	}
}
