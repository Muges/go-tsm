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

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type addTest struct {
	in     NormalizeBuffer
	window []float64

	panics bool
	out    NormalizeBuffer
}

var addTests = []addTest{
	{NewNormalizeBuffer(0), []float64{}, false, NormalizeBuffer{[]float64{}, 0}},
	{NewNormalizeBuffer(3), []float64{}, false, NormalizeBuffer{[]float64{0, 0, 0}, 0}},
	{NewNormalizeBuffer(3), []float64{1}, false, NormalizeBuffer{[]float64{1, 0, 0}, 0}},
	{NewNormalizeBuffer(3), []float64{1, 2, 3}, false, NormalizeBuffer{[]float64{1, 2, 3}, 0}},

	{NormalizeBuffer{[]float64{0, 0, 0}, 1}, []float64{}, false, NormalizeBuffer{[]float64{0, 0, 0}, 1}},
	{NormalizeBuffer{[]float64{0, 0, 0}, 1}, []float64{1}, false, NormalizeBuffer{[]float64{0, 1, 0}, 1}},
	{NormalizeBuffer{[]float64{0, 0, 0}, 1}, []float64{1, 2, 3}, false, NormalizeBuffer{[]float64{3, 1, 2}, 1}},

	{NewNormalizeBuffer(3), []float64{1, 2, 3, 4}, true, NewNormalizeBuffer(0)},
}

func TestAdd(t *testing.T) {
	assert := assert.New(t)

	for i, c := range addTests {
		if c.panics {
			assert.Panics(func() {
				c.in.Add(c.window)
			}, fmt.Sprintf("NormalizeBuffer.Add (%d)", i))
		} else {
			ok := assert.NotPanics(func() {
				c.in.Add(c.window)
			}, fmt.Sprintf("NormalizeBuffer.Add (%d)", i))

			if ok {
				assert.Equal(c.out, c.in, fmt.Sprintf("NormalizeBuffer.Add (%d)", i))
			}
		}
	}
}

type getTest struct {
	in    NormalizeBuffer
	index int

	panics bool
	out    float64
}

var getTests = []getTest{
	{NewNormalizeBuffer(1), 0, false, 0},
	{NormalizeBuffer{[]float64{1, 2, 3}, 0}, 1, false, 2},
	{NormalizeBuffer{[]float64{1, 2, 3}, 0}, 2, false, 3},
	{NormalizeBuffer{[]float64{1, 2, 3}, 1}, 1, false, 3},
	{NormalizeBuffer{[]float64{1, 2, 3}, 1}, 2, false, 1},

	{NewNormalizeBuffer(0), 0, true, 0},
	{NewNormalizeBuffer(2), -1, true, 0},
	{NewNormalizeBuffer(2), 3, true, 0},
	{NormalizeBuffer{[]float64{1, 2, 3}, 1}, -1, true, 0},
}

func TestGet(t *testing.T) {
	assert := assert.New(t)

	for i, c := range getTests {
		if c.panics {
			assert.Panics(func() {
				c.in.Get(c.index)
			}, fmt.Sprintf("NormalizeBuffer.Get (%d)", i))
		} else {
			var out float64
			ok := assert.NotPanics(func() {
				out = c.in.Get(c.index)
			}, fmt.Sprintf("NormalizeBuffer.Get (%d)", i))

			if ok {
				assert.Equal(c.out, out, fmt.Sprintf("NormalizeBuffer.Get (%d)", i))
			}
		}
	}
}

type removeTest struct {
	in     NormalizeBuffer
	length int

	out NormalizeBuffer
}

var removeTests = []removeTest{
	{NewNormalizeBuffer(0), 0, NormalizeBuffer{[]float64{}, 0}},
	{NewNormalizeBuffer(3), 2, NormalizeBuffer{[]float64{0, 0, 0}, 2}},

	{NormalizeBuffer{[]float64{1, 2, 3}, 0}, 0, NormalizeBuffer{[]float64{1, 2, 3}, 0}},
	{NormalizeBuffer{[]float64{1, 2, 3}, 1}, 0, NormalizeBuffer{[]float64{1, 2, 3}, 1}},
	{NormalizeBuffer{[]float64{1, 2, 3}, 0}, 2, NormalizeBuffer{[]float64{0, 0, 3}, 2}},
	{NormalizeBuffer{[]float64{1, 2, 3}, 1}, 2, NormalizeBuffer{[]float64{1, 0, 0}, 0}},
}

func TestRemove(t *testing.T) {
	assert := assert.New(t)

	for i, c := range removeTests {
		ok := assert.NotPanics(func() {
			c.in.Remove(c.length)
		}, fmt.Sprintf("NormalizeBuffer.Remove (%d)", i))

		if ok {
			assert.Equal(c.out, c.in, fmt.Sprintf("NormalizeBuffer.Remove (%d)", i))
		}
	}
}
