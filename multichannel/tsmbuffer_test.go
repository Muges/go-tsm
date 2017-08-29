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

package multichannel_test

import (
	"fmt"
	"github.com/Muges/go-tsm/multichannel"
	"github.com/stretchr/testify/assert"
	"testing"
)

type newBufferTest struct {
	channels int
	size     int

	panics bool
	out    multichannel.TSMBuffer
}

var newBufferTests = []newBufferTest{
	{0, 0, false, multichannel.TSMBuffer{}},
	{0, 3, false, multichannel.TSMBuffer{}},
	{2, 0, false, multichannel.TSMBuffer{{}, {}}},
	{2, 3, false, multichannel.TSMBuffer{{0, 0, 0}, {0, 0, 0}}},

	{-1, 1, true, nil},
	{2, -2, true, nil},
}

func TestNewTSMBuffer(t *testing.T) {
	assert := assert.New(t)

	for _, c := range newBufferTests {
		if c.panics {
			assert.Panics(func() {
				multichannel.NewTSMBuffer(c.channels, c.size)
			}, fmt.Sprintf("NewTSMBuffer(%d, %d)", c.channels, c.size))
		} else {
			var out multichannel.TSMBuffer

			ok := assert.NotPanics(func() {
				out = multichannel.NewTSMBuffer(c.channels, c.size)
			}, fmt.Sprintf("NewTSMBuffer(%d, %d)", c.channels, c.size))

			if ok {
				assert.Equal(c.out, out, fmt.Sprintf("NewTSMBuffer(%d, %d)", c.channels, c.size))
			}
		}
	}
}

type applyWindowTest struct {
	in     multichannel.TSMBuffer
	window []float64

	panics bool
	out    multichannel.TSMBuffer
}

var applyWindowTests = []applyWindowTest{
	{multichannel.TSMBuffer{}, []float64{}, false, multichannel.TSMBuffer{}},
	{multichannel.TSMBuffer{}, []float64{1, 1, 1, 1}, false, multichannel.TSMBuffer{}},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, []float64{1, 1, 1, 1}, false, multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, []float64{0, 1, 0, 1}, false, multichannel.TSMBuffer{{0, 2, 0, 4}, {0, 6, 0, 8}}},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, []float64{0, 0.5, 1, 0.5}, false, multichannel.TSMBuffer{{0, 1, 3, 2}, {0, 3, 7, 4}}},

	{multichannel.TSMBuffer{{1, 2, 3, 4}}, []float64{}, true, nil},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, []float64{1, 1, 1}, true, nil},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, []float64{1, 1, 1, 1, 1}, true, nil},
}

func TestApplyWindow(t *testing.T) {
	assert := assert.New(t)

	for i, c := range applyWindowTests {
		if c.panics {
			assert.Panics(func() {
				c.in.ApplyWindow(c.window)
			}, fmt.Sprintf("Buffer.ApplyWindow (%d)", i))
		} else {
			ok := assert.NotPanics(func() {
				c.in.ApplyWindow(c.window)
			}, fmt.Sprintf("Buffer.ApplyWindow (%d)", i))

			if ok {
				assert.Equal(c.out, c.in, fmt.Sprintf("Buffer.ApplyWindow (%d)", i))
			}
		}
	}
}

type channelTest struct {
	in      multichannel.TSMBuffer
	channel int

	panics bool
	out    []float64
}

var channelTests = []channelTest{
	{multichannel.TSMBuffer{{}}, 0, false, []float64{}},
	{multichannel.TSMBuffer{{0, 0}}, 0, false, []float64{0, 0}},
	{multichannel.TSMBuffer{{1, 2}, {3, 4}, {5, 6}}, 1, false, []float64{3, 4}},

	{multichannel.TSMBuffer{{}}, -1, true, nil},
	{multichannel.TSMBuffer{{0, 0}}, 1, true, nil},
	{multichannel.TSMBuffer{{1, 2}, {3, 4}, {5, 6}}, 5, true, nil},
}

func TestChannel(t *testing.T) {
	assert := assert.New(t)

	for i, c := range channelTests {
		if c.panics {
			assert.Panics(func() {
				c.in.Channel(c.channel)
			}, fmt.Sprintf("Buffer.Channel (%d)", i))
		} else {
			var out []float64

			ok := assert.NotPanics(func() {
				out = c.in.Channel(c.channel)
			}, fmt.Sprintf("Buffer.Channel (%d)", i))

			if ok {
				assert.Equal(c.out, out, fmt.Sprintf("Buffer.Channel (%d)", i))
			}
		}
	}
}

type channelsTest struct {
	in  multichannel.TSMBuffer
	out int
}

var channelsTests = []channelsTest{
	{multichannel.TSMBuffer{}, 0},
	{multichannel.TSMBuffer{{}}, 1},
	{multichannel.TSMBuffer{{0, 0}}, 1},
	{multichannel.TSMBuffer{{}, {}, {}}, 3},
}

func TestChannels(t *testing.T) {
	assert := assert.New(t)

	for i, c := range channelsTests {
		var out int

		ok := assert.NotPanics(func() {
			out = c.in.Channels()
		}, fmt.Sprintf("Buffer.Channels (%d)", i))

		if ok {
			assert.Equal(c.out, out, fmt.Sprintf("Buffer.Channels (%d)", i))
		}
	}
}

type lenTest struct {
	in  multichannel.TSMBuffer
	out int
}

var lenTests = []lenTest{
	{multichannel.TSMBuffer{}, 0},
	{multichannel.TSMBuffer{{}}, 0},
	{multichannel.TSMBuffer{{0, 0}}, 2},
	{multichannel.TSMBuffer{{}, {}, {}}, 0},
	{multichannel.TSMBuffer{{1, 2}, {3, 4}, {5, 6}}, 2},
}

func TestLen(t *testing.T) {
	assert := assert.New(t)

	for i, c := range lenTests {
		var out int

		ok := assert.NotPanics(func() {
			out = c.in.Len()
		}, fmt.Sprintf("Buffer.Len (%d)", i))

		if ok {
			assert.Equal(c.out, out, fmt.Sprintf("Buffer.Len (%d)", i))
		}
	}
}

type sampleTest struct {
	in      multichannel.TSMBuffer
	channel int
	index   int

	panics bool
	out    float64
}

var sampleTests = []sampleTest{
	{multichannel.TSMBuffer{{0}}, 0, 0, false, 0},
	{multichannel.TSMBuffer{{1, 2}}, 0, 0, false, 1},
	{multichannel.TSMBuffer{{1, 2}}, 0, 1, false, 2},
	{multichannel.TSMBuffer{{1, 2}, {3, 4}, {5, 6}}, 1, 0, false, 3},
	{multichannel.TSMBuffer{{1, 2}, {3, 4}, {5, 6}}, 2, 1, false, 6},

	{multichannel.TSMBuffer{{}}, 0, 0, true, 0},
	{multichannel.TSMBuffer{{}, {}}, 0, 0, true, 0},
	{multichannel.TSMBuffer{{1, 2}}, -1, 0, true, 0},
	{multichannel.TSMBuffer{{1, 2}}, 1, 0, true, 0},
	{multichannel.TSMBuffer{{1, 2}}, 0, -1, true, 0},
	{multichannel.TSMBuffer{{1, 2}}, 0, 2, true, 0},
	{multichannel.TSMBuffer{{1, 2}, {3, 4}, {5, 6}}, 3, 0, true, 0},
}

func TestSample(t *testing.T) {
	assert := assert.New(t)

	for i, c := range sampleTests {
		if c.panics {
			assert.Panics(func() {
				c.in.Sample(c.channel, c.index)
			}, fmt.Sprintf("Buffer.Sample (%d)", i))
		} else {
			var out float64

			ok := assert.NotPanics(func() {
				out = c.in.Sample(c.channel, c.index)
			}, fmt.Sprintf("Buffer.Sample (%d)", i))

			if ok {
				assert.Equal(c.out, out, fmt.Sprintf("Buffer.Sample (%d)", i))
			}
		}
	}
}

type setSampleTest struct {
	in      multichannel.TSMBuffer
	channel int
	index   int
	value   float64

	panics bool
	out    multichannel.TSMBuffer
}

var setSampleTests = []setSampleTest{
	{multichannel.TSMBuffer{{0}}, 0, 0, 1, false, multichannel.TSMBuffer{{1}}},
	{multichannel.TSMBuffer{{1, 2}}, 0, 0, 3, false, multichannel.TSMBuffer{{3, 2}}},
	{multichannel.TSMBuffer{{1, 2}}, 0, 1, 4, false, multichannel.TSMBuffer{{1, 4}}},
	{multichannel.TSMBuffer{{1, 2}, {3, 4}, {5, 6}}, 1, 0, 7, false, multichannel.TSMBuffer{{1, 2}, {7, 4}, {5, 6}}},
	{multichannel.TSMBuffer{{1, 2}, {3, 4}, {5, 6}}, 2, 1, 8, false, multichannel.TSMBuffer{{1, 2}, {3, 4}, {5, 8}}},

	{multichannel.TSMBuffer{{}}, 0, 0, 0, true, nil},
	{multichannel.TSMBuffer{{}, {}}, 0, 0, 0, true, nil},
	{multichannel.TSMBuffer{{1, 2}}, -1, 0, 0, true, nil},
	{multichannel.TSMBuffer{{1, 2}}, 1, 0, 0, true, nil},
	{multichannel.TSMBuffer{{1, 2}}, 0, -1, 0, true, nil},
	{multichannel.TSMBuffer{{1, 2}}, 0, 2, 0, true, nil},
	{multichannel.TSMBuffer{{1, 2}, {3, 4}, {5, 6}}, 3, 0, 0, true, nil},
}

func TestSetSample(t *testing.T) {
	assert := assert.New(t)

	for i, c := range setSampleTests {
		if c.panics {
			assert.Panics(func() {
				c.in.SetSample(c.channel, c.index, c.value)
			}, fmt.Sprintf("Buffer.SetSample (%d)", i))
		} else {
			ok := assert.NotPanics(func() {
				c.in.SetSample(c.channel, c.index, c.value)
			}, fmt.Sprintf("Buffer.SetSample (%d)", i))

			if ok {
				assert.Equal(c.out, c.in, fmt.Sprintf("Buffer.SetSample (%d)", i))
			}
		}
	}
}

type sliceTest struct {
	in   multichannel.TSMBuffer
	from int
	to   int

	panics bool
	out    multichannel.TSMBuffer
}

var sliceTests = []sliceTest{
	{multichannel.TSMBuffer{}, 0, 0, false, multichannel.TSMBuffer{}},

	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 0, 0, false, multichannel.TSMBuffer{{}, {}}},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 0, 3, false, multichannel.TSMBuffer{{1, 2, 3}, {5, 6, 7}}},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 0, 4, false, multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 1, 4, false, multichannel.TSMBuffer{{2, 3, 4}, {6, 7, 8}}},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 1, 3, false, multichannel.TSMBuffer{{2, 3}, {6, 7}}},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 2, 2, false, multichannel.TSMBuffer{{}, {}}},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 4, 4, false, multichannel.TSMBuffer{{}, {}}},

	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 0, -1, true, nil},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 0, 5, true, nil},

	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, -1, 4, true, nil},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 5, 4, true, nil},

	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, -3, -1, true, nil},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, -1, 0, true, nil},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, -1, 2, true, nil},

	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 5, 6, true, nil},
	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 3, 6, true, nil},

	{multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}}, 2, 0, true, nil},
}

func TestSlice(t *testing.T) {
	assert := assert.New(t)

	for i, c := range sliceTests {
		if c.panics {
			assert.Panics(func() {
				c.in.Slice(c.from, c.to)
			}, fmt.Sprintf("Buffer.Slice (%d)", i))
		} else {
			var out multichannel.Buffer

			ok := assert.NotPanics(func() {
				out = c.in.Slice(c.from, c.to)
			}, fmt.Sprintf("Buffer.Slice (%d)", i))

			if ok {
				assert.Equal(c.out, out, fmt.Sprintf("Buffer.Slice (%d)", i))
			}
		}
	}
}
