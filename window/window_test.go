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

package window

import (
	"github.com/mjibson/go-dsp/dsputils"
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
	for _, v := range hanningTests {
		out := Hanning(v.in)
		if !dsputils.PrettyClose(out, v.out) {
			t.Error("error\ninput:", v.in, "\noutput:", out, "\nexpected:", v.out)
		}
	}
}
