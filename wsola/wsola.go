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

// Package wsola implements the WSOLA (Waveform Similariy-based Overlap-Add)
// time-scale modification procedure.
//
// WSOLA works in the same way as OLA, with the exception that it allows slight
// shift of the position of the analysis frames.
package wsola

import (
	"github.com/Muges/tsm/multichannel"
	"github.com/Muges/tsm/tsm"
	"github.com/Muges/tsm/window"
)

// A wsolaConverter implements the conversion of an analysis frame into a
// synthesis frame for the WSOLA (Waveform Similarity-based Overlap-Add)
// method.
type wsolaConverter struct {
	frameLength        int
	synthesisHop       int
	tolerance          int
	naturalProgression multichannel.TSMBuffer
}

// crossCorrelation returns the cross-correlation of buffer1 and
// buffer2[offset:offset+len(buffer1)].
func crossCorrelation(buffer1 []float64, buffer2 []float64, offset int) float64 {
	var result float64

	for i, v := range buffer1 {
		result += v * buffer2[offset+i]
	}

	return result
}

// maximizeCrossCorrelation returns the value delta of the interval [0,
// 2*tolerance] that maximizes crossCorrelation(buffer1, buffer2, delta)
func maximizeCrossCorrelation(buffer1 []float64, buffer2 []float64, tolerance int) int {
	var maxDelta int
	maxValue := crossCorrelation(buffer1, buffer2, 0)

	for delta := 1; delta < 2*tolerance; delta++ {
		value := crossCorrelation(buffer1, buffer2, delta)
		if value > maxValue {
			maxValue = value
			maxDelta = delta
		}
	}

	if maxValue == 0 {
		return tolerance
	}

	return maxDelta
}

// Convert creates the synthesis frame by taking the part of the analysis frame
// which aligns best with the natural progression of the signal.
func (c *wsolaConverter) Convert(analysisFrame multichannel.TSMBuffer) multichannel.TSMBuffer {
	synthesisFrame := make([][]float64, analysisFrame.Channels())

	for k := range analysisFrame {
		delta := maximizeCrossCorrelation(c.naturalProgression[k], analysisFrame[k], c.tolerance)

		copy(c.naturalProgression[k],
			analysisFrame[k][delta+c.synthesisHop:delta+c.synthesisHop+c.frameLength])

		synthesisFrame[k] = analysisFrame[k][delta : delta+c.frameLength]

	}

	return synthesisFrame
}

// Clear clears the state of the Converter, making it ready to be used on
// another signal (or another part of a signal). It is automatically called by
// the Flush, Clear and New methods of the TSM object.
func (c *wsolaConverter) Clear() {
	// Reset the natural progression
	for k := range c.naturalProgression {
		for i := range c.naturalProgression[k] {
			c.naturalProgression[k][i] = 0
		}
	}
}

// New returns a TSM implementing the WSOLA procedure.
//
// channels is the number of channels of the signal that the TSM will process.
// tolerance is the maximum number of samples that the analysis frame can be
// shifted.  Read the documentation of the tsm.Settings type for an explanation
// of the other arguments.
func New(channels int, analysisHop int, synthesisHop int, frameLength int, tolerance int) (*tsm.TSM, error) {
	converter := wsolaConverter{
		frameLength:        frameLength,
		synthesisHop:       synthesisHop,
		tolerance:          tolerance,
		naturalProgression: multichannel.NewTSMBuffer(channels, frameLength),
	}

	return tsm.New(tsm.Settings{
		Channels:        channels,
		AnalysisHop:     analysisHop,
		SynthesisHop:    synthesisHop,
		FrameLength:     frameLength,
		SynthesisWindow: window.Hanning(frameLength),

		DeltaBefore: tolerance,
		DeltaAfter:  tolerance + synthesisHop,

		Converter: &converter,
	})
}

// NewWithSpeed returns a TSM implementing the WSOLA procedure, modifying the
// speed of the input signal by the ratio speed.
//
// The arguments speed, synthesisHop, frameLength and tolerance may be strictly
// negative, in which case they will be replaced by default values.
func NewWithSpeed(channels int, speed float64, synthesisHop int, frameLength int, tolerance int) (*tsm.TSM, error) {
	if speed < 0 {
		speed = 1
	}
	if frameLength < 0 {
		frameLength = 1024
	}
	if synthesisHop < 0 {
		synthesisHop = frameLength / 2
	}
	if tolerance < 0 {
		tolerance = frameLength / 2
	}

	analysisHop := int(float64(synthesisHop) * speed)

	return New(channels, analysisHop, synthesisHop, frameLength, tolerance)
}

// Default returns a TSM implementing the WSOLA procedure with sane default
// parameters.
func Default(channels int, speed float64) (*tsm.TSM, error) {
	return NewWithSpeed(channels, speed, -1, -1, -1)
}
