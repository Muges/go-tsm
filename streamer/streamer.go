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

// Package streamer provides the time-scale modification methods as Streamers,
// to be used with the beep library (https://github.com/faiface/beep)
package streamer

import (
	"github.com/Muges/tsm"
	"github.com/Muges/tsm/multichannel"
	"github.com/faiface/beep"
)

// A TSMStreamer is a beep.Streamer that changes the speed of a wrapped
// Streamer without changing its pitch.
type TSMStreamer struct {
	t             *tsm.TSM
	inputStreamer beep.Streamer
	buffer        multichannel.StereoBuffer
}

// New creates a new TSMSTreamer, which changes the speed of the inputStreamer
// using the TSM procedure t.
func New(t *tsm.TSM, inputStreamer beep.Streamer) TSMStreamer {
	return TSMStreamer{
		t:             t,
		inputStreamer: inputStreamer,
	}
}

// Stream copies at most len(samples) next audio samples to the samples slice.
func (s TSMStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	length := 0

	for length < len(samples) {
		// Read samples from input stream and transfer them to TSM
		nmax := s.t.RemainingInputSpace()
		if len(s.buffer) < nmax {
			// This should only happen once
			s.buffer = make([][2]float64, nmax)
		}
		n, ok := s.inputStreamer.Stream(s.buffer[:nmax])
		s.t.Put(s.buffer[:n])

		l := s.t.Receive(multichannel.StereoBuffer(samples[length:]))
		length += l

		if l == 0 && !ok {
			l = s.t.Flush(multichannel.StereoBuffer(samples[length:]))
			length += l

			if l == 0 {
				return length, false
			}
		}
	}

	return length, true
}

// Err propagates the wrapped Streamer's errors.
func (s TSMStreamer) Err() error {
	return s.inputStreamer.Err()
}
