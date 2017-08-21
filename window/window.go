// Copyright (c) 2017 Muges
// Copyright (c) 2012 Matt Jibson <matt.jibson@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

// Package window provides window functions for digital signal processing.
package window

import (
	"errors"
	"math"
)

// Hanning returns a periodic Hannning window of size n.
func Hanning(n int) []float64 {
	window := make([]float64, n)

	freq := 2 * math.Pi / float64(n)
	for k := 0; k < n; k++ {
		window[k] = 0.5 * (1 - math.Cos(freq*float64(k)))
	}

	return window
}

// Product returns the product of two windows.
//
// If one of the windows is equal to nil, the other will be returned. If both
// are equal to nil, nil will be returned.
func Product(window1 []float64, window2 []float64) ([]float64, error) {
	if window1 == nil {
		return window2, nil
	}
	if window2 == nil {
		return window1, nil
	}
	if len(window1) != len(window2) {
		return nil, errors.New("the two windows should have the same size")
	}

	product := make([]float64, len(window1))
	for i, v := range window1 {
		product[i] = v * window2[i]
	}

	return product, nil
}
