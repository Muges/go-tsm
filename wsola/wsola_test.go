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

package wsola

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type crossCorrelationTest struct {
	buffer1 []float64
	buffer2 []float64
	offset  int

	out float64
}

var crossCorrelationTests = []crossCorrelationTest{
	{[]float64{}, []float64{}, 0, 0},
	{[]float64{0, 0, 0}, []float64{0, 0, 0, 0}, 0, 0},
	{[]float64{0, 0, 0}, []float64{0, 0, 0, 0}, 1, 0},

	{[]float64{1}, []float64{0, 1, 2}, 0, 0},
	{[]float64{1}, []float64{0, 1, 2}, 1, 1},
	{[]float64{1}, []float64{0, 1, 2}, 2, 2},

	{[]float64{0, 0.5, -1}, []float64{1, 0, 0.5, -1, -0.5}, 0, -0.5},
	{[]float64{0, 0.5, -1}, []float64{1, 0, 0.5, -1, -0.5}, 1, 1.25},
	{[]float64{0, 0.5, -1}, []float64{1, 0, 0.5, -1, -0.5}, 2, 0},
}

func TestCrossCorrelation(t *testing.T) {
	assert := assert.New(t)

	for i, c := range crossCorrelationTests {
		out := crossCorrelation(c.buffer1, c.buffer2, c.offset)
		assert.Equal(c.out, out, fmt.Sprintf("CrossCorrelation (%d)", i))
	}
}
