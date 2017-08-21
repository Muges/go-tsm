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

// A Buffer is a representation of a multi-channel audio buffer.
type Buffer interface {
	// Channels returns the number of channels of the buffer.
	Channels() int

	// Len returns the number of samples of each channel of the buffer.
	Len() int

	// Sample returns the index-th sample of the channel-th channel of the
	// buffer.
	Sample(channel int, index int) float64

	// SetSample sets the index-th sample of the channel-c channel of the
	// buffer to value.
	SetSample(channel int, index int, value float64)

	// Slice returns a Buffer containing only the audio samples between from
	// (included) and to (excluded) for each channel. It is the equivalent of
	// buffer[from:to], if buffer is a mono-channel buffer represented by a
	// slice.
	Slice(from int, to int) Buffer
}

// A StereoBuffer is a representation of a stereo audio buffer which implements
// the buffer interface.
//
// If buffer is a StereoBuffer, the value at buffer[i][0] is the value of the
// left channel of the i-th sample, and buffer[i][1] is the value of the right
// channel of the i-th sample.
type StereoBuffer [][2]float64

// Channels returns the number of channels of the buffer (always 2 for a
// StereoBuffer).
func (s StereoBuffer) Channels() int {
	return 2
}

// Len returns the number of samples of each channel of the buffer.
func (s StereoBuffer) Len() int {
	return len(s)
}

// Sample returns the index-th sample of the channel-th channel of the buffer.
func (s StereoBuffer) Sample(channel int, index int) float64 {
	return s[index][channel]
}

// SetSample sets the index-th sample of the channel-c channel of the buffer to
// value.
func (s StereoBuffer) SetSample(channel int, index int, value float64) {
	s[index][channel] = value
}

// Slice returns a Buffer containing only the audio samples between from
// (included) and to (excluded) for each channel. It is the equivalent of
// buffer[from:to], if buffer is a mono-channel buffer represented by a slice.
func (s StereoBuffer) Slice(from int, to int) Buffer {
	return s[from:to]
}
