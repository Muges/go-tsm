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

// Package ola implements the OLA (Overlap-Add) time-scale modification
// procedure.
//
// It should give good results for percusive signals.
package ola

import (
	"github.com/Muges/tsm"
	"github.com/Muges/tsm/multichannel"
	"github.com/Muges/tsm/window"
)

// An olaConverter implements the conversion of an analysis frame into a
// synthesis frame for the OLA (Overlap-Add) method.
type olaConverter struct{}

// Convert returns the analysisFrame without modifying it.
func (o olaConverter) Convert(analysisFrame multichannel.TSMBuffer) multichannel.TSMBuffer {
	return analysisFrame
}

// New returns a TSM implementing the OLA procedure.
//
// channels is the number of channels of the signal that the TSM will process,
// bufferSize is the size of the input buffer, and should be larger than
// frameSize. Read the documentation of the TSM type above for an explanation
// of the other arguments.
func New(channels int, analysisHop int, synthesisHop int, frameSize int,
	bufferSize int) (*tsm.TSM, error) {
	return tsm.New(channels, analysisHop, synthesisHop, frameSize, bufferSize,
		nil, window.Hanning(frameSize), olaConverter{})
}

// NewWithSpeed returns a TSM implementing the OLA procedure, modifying the
// sped of the input signal by the ratio speed.
//
// The arguments speed, synthesisHop, frameSize and bufferSize may be equal to
// zero, in which case they will be replaced by default values.
func NewWithSpeed(channels int, speed float64, synthesisHop int, frameSize int, bufferSize int) (*tsm.TSM, error) {
	if speed == 0 {
		speed = 1
	}
	if frameSize == 0 {
		frameSize = 256
	}
	if synthesisHop == 0 {
		synthesisHop = frameSize / 2
	}

	analysisHop := int(float64(synthesisHop) * speed)

	if bufferSize == 0 {
		bufferSize = frameSize + analysisHop
	}

	return New(channels, analysisHop, synthesisHop, frameSize, bufferSize)
}

// Default returns a TSM implementing the OLA procedure with sane default
// parameters.
func Default(channels int, speed float64) (*tsm.TSM, error) {
	return NewWithSpeed(channels, speed, 0, 0, 0)
}
