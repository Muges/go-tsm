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

// Package multichannel provides data structure used for manipulating
// multi-channel audio data.
package multichannel

// A TSMBuffer is a representation of a multi-channel audio buffer which
// implements the Buffer interface and is used internally by the tsm package.
//
// If buffer is a StereoBuffer, the value at buffer[c][i] is the value of the
// i-th sample of the c-th channel. Each channel should have the same number of
// samples. This representation makes it easy to work on each channel
// individually, which is often required for time-scale modifications.
type TSMBuffer [][]float64

// NewTSMBuffer creates a new TSMBuffer, each channel containing length
// samples.
func NewTSMBuffer(channels int, length int) TSMBuffer {
	buffer := make(TSMBuffer, channels)
	for k := range buffer {
		buffer[k] = make([]float64, length)
	}
	return buffer
}

// ApplyWindow applies a window to each channel of the buffer.
//
// A window is a slice of float64 (as returned by the functions of the packages
// tsm/windows and github.com/mjibson/go-dsp/window), and is applied by
// multiplying each channel by the window element-wise.
//
// ApplyWindow will panic if the buffer and the window have different lengths.
func (b TSMBuffer) ApplyWindow(window []float64) {
	if len(b) == 0 {
		return
	}

	if len(b[0]) != len(window) {
		panic("the buffer and the window should have the same size")
	}

	for k := range b {
		for i, v := range window {
			b[k][i] *= v
		}
	}
}

// Channel returns the channel-th channel of the buffer.
func (b TSMBuffer) Channel(channel int) []float64 {
	return b[channel]
}

// Channels returns the number of channels of the buffer.
func (b TSMBuffer) Channels() int {
	return len(b)
}

// Len returns the number of samples of each channel of the buffer.
func (b TSMBuffer) Len() int {
	if len(b) == 0 {
		return 0
	}
	return len(b[0])
}

// Sample returns the index-th sample of the channel-th channel.
func (b TSMBuffer) Sample(channel int, index int) float64 {
	return b[channel][index]
}

// SetSample sets the value of the index-th sample of the channel-th channel of
// the buffer to value.
func (b TSMBuffer) SetSample(channel int, index int, value float64) {
	b[channel][index] = value
}

// Slice returns a TSMBuffer containing only the audio samples between from
// (included) and to (excluded) for each channel. It is the equivalent of
// buffer[from:to], if buffer is a mono-channel buffer represented by a slice.
func (b TSMBuffer) Slice(from int, to int) TSMBuffer {
	slice := make(TSMBuffer, len(b))

	for k := range b {
		slice[k] = b[k][from:to]
	}

	return slice
}
