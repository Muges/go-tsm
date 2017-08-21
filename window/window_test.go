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

package window_test

import (
	"fmt"
	"github.com/Muges/tsm/window"
	"github.com/stretchr/testify/assert"
	"testing"
)

type hanningTest struct {
	in  int
	out []float64
}

var hanningTests = []hanningTest{
	{1, []float64{0}},
	{2, []float64{0, 1}},
	{3, []float64{0, 0.75, 0.75}},
	{4, []float64{0, 0.5, 1, 0.5}},
	{8, []float64{0, 0.14644661, 0.5, 0.85355339, 1, 0.85355339, 0.5, 0.146446611}},
}

func TestHanning(t *testing.T) {
	assert := assert.New(t)

	for i, c := range hanningTests {
		out := window.Hanning(c.in)
		assert.InDeltaSlice(c.out, out, 0.000001, fmt.Sprintf("Hanning (%d)", i))
	}
}

type productTest struct {
	window1 []float64
	window2 []float64

	out []float64
	err bool
}

var productTests = []productTest{
	{nil, nil, nil, false},
	{[]float64{}, nil, []float64{}, false},
	{[]float64{2}, nil, []float64{2}, false},
	{[]float64{1, 2, 3}, nil, []float64{1, 2, 3}, false},

	{[]float64{1, 2, 3}, []float64{0, 2, 4}, []float64{0, 4, 12}, false},

	{[]float64{1, 2, 3}, []float64{0}, nil, true},
}

func TestProduct(t *testing.T) {
	assert := assert.New(t)

	for i, c := range productTests {
		out12, err12 := window.Product(c.window1, c.window2)
		out21, err21 := window.Product(c.window2, c.window1)

		if c.err {
			assert.Error(err12, fmt.Sprintf("Product (%d)", i))
			assert.Error(err21, fmt.Sprintf("Product (%d reversed)", i))
		} else if c.out == nil {
			assert.Equal(c.out, out12, fmt.Sprintf("Product (%d)", i))
			assert.Equal(c.out, out21, fmt.Sprintf("Product (%d reversed)", i))
		} else {
			assert.InDeltaSlice(c.out, out12, 0.000001, fmt.Sprintf("Product (%d)", i))
			assert.InDeltaSlice(c.out, out21, 0.000001, fmt.Sprintf("Product (%d reversed)", i))
		}
	}
}
