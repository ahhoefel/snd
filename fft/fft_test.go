package fft_test

import (
	"math/cmplx"
	"testing"

	"github.com/ahhoefel/snd/fft"
)

func TestRecFFT(t *testing.T) {
	tests := []struct {
		in  []complex128
		out []complex128
	}{
		{in: []complex128{1, 0, 0, 0}, out: []complex128{1, 1, 1, 1}},
		{in: []complex128{0, 1, 0, 0}, out: []complex128{1, complex(0, -1), -1, complex(0, 1)}},
		{in: []complex128{0, 0, 1, 0}, out: []complex128{1, -1, 1, -1}},
		{in: []complex128{0, 0, 0, 1}, out: []complex128{1, complex(0, 1), -1, complex(0, -1)}},
	}
	for i, test := range tests {
		fft.RecFFT(test.in)
		approxEq := true
		for j, z := range test.in {
			if cmplx.Abs(z-test.out[j]) > 0.00001 {
				approxEq = false
			}
		}
		if !approxEq {
			t.Errorf("%d. RecFFT(x) = %v, expected %v", i, test.in, test.out)
		}
	}
}
