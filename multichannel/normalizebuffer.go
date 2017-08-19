// Copyright (c) 2017 Muges
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

package multichannel

// NormalizeBuffer is a mono-channel circular buffer, used to normalize audio
// buffers.
type NormalizeBuffer struct {
	data    []float64
	pointer int
}

// NewNormalizeBuffer returns a new NormalizeBuffer of length n.
func NewNormalizeBuffer(n int) NormalizeBuffer {
	return NormalizeBuffer{
		data: make([]float64, n),
	}
}

// Add adds a window element-wise to the buffer.
func (b *NormalizeBuffer) Add(window []float64) {
	if len(window) > len(b.data) {
		panic("the window should be smaller than the buffer")
	}

	for i, v := range window {
		b.data[(b.pointer+i)%len(b.data)] += v
	}
}

// Get returns the i-th value of the buffer.
func (b *NormalizeBuffer) Get(i int) float64 {
	if i < 0 || i > len(b.data) {
		panic("index out of bounds")
	}
	return b.data[(b.pointer+i)%len(b.data)]
}

// Remove removes the first n values of the buffer.
func (b *NormalizeBuffer) Remove(n int) {
	if len(b.data) != 0 {
		for i := 0; i < n; i++ {
			b.data[b.pointer%len(b.data)] = 0
			b.pointer = (b.pointer + 1) % len(b.data)
		}
	}
}
