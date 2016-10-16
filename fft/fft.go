package fft

import (
	"fmt"
	"math"
	"math/cmplx"
)

func FFT(x []float64) {
	c := FFTComplex(x)
	for i, z := range c {
		x[i] = real(z)
	}
}

func FFTComplex(x []float64) []complex128 {
	k := intLog(len(x))
	if 1<<uint(k) != len(x) {
		panic(fmt.Sprintf("Only fft powers of two! %d.", len(x)))
	}
	var c []complex128
	for _, a := range x {
		c = append(c, complex(a, 0))
	}
	RecFFT(c)
	return c
}

func intLog(n int) int {
	p := 0
	m := 1
	for n > m {
		m = m << 1
		p++
	}
	return p
}

func iterFFT(x []complex128) []complex128 {
	target := make([]complex128, len(x))
	m := len(x)
	for n := 1; n <= len(x); n = n << 1 {
		for i := 0; i < m; i++ {
			k := i
			for j := 0; j < m-1; j++ {
				t := cmplx.Exp(complex(0, 2*math.Pi*float64(j)/float64(n)))
				target[k] = x[k] + t*x[k+n]
				target[k+n] = x[k] - t*x[k+n]
				k += m
			}
		}
		m = m >> 1
	}
	return x
}

func RecFFT(x []complex128) {
	if len(x) == 1 {
		return
	}
	e, o := evenOdd(x)
	RecFFT(e)
	RecFFT(o)
	for k := range e {
		c := cmplx.Exp(complex(0, -2*math.Pi*float64(k)/float64(len(x))))
		x[k], x[k+len(e)] = e[k]+c*o[k], e[k]-c*o[k]
	}
}

func evenOdd(x []complex128) (even []complex128, odd []complex128) {
	for i, z := range x {
		if i%2 == 0 {
			even = append(even, z)
		} else {
			odd = append(odd, z)
		}
	}
	return even, odd
}
