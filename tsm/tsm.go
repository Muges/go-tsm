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

// Package tsm implements the skeleton of an analysis-synthesis based
// time-scale modification procedure.
package tsm

import (
	"github.com/Muges/go-tsm/multichannel"
	"github.com/Muges/go-tsm/window"
	"github.com/pkg/errors"
)

// A Converter is an object implementing the conversion of an analysis frame
// into a synthesis frame.
type Converter interface {
	// Convert converts an analysis frame into a synthesis frame.
	Convert(analysisFrame multichannel.TSMBuffer) (synthesisFrame multichannel.TSMBuffer)

	// Clear clears the state of the Converter, making it ready to be used on
	// another signal (or another part of a signal). It is automatically called
	// by the Flush, Clear and New methods of the TSM object.
	Clear()
}

// A Settings is a struct containing the settings for a TSM object. It is used
// for the creation of a new TSM
//
// Channels is the number of channels of the signal that the TSM will process.
// The other fields are parameters of the TSM algorithm that are explained
// below.
//
// The basic principle of the TSM is to first decompose the input signal into
// short overlapping frames, called the analysis frames. The frames have a
// fixed length FrameLength, and are separated by a distance AnalysisHop, as
// illustrated below.
//
//              <--------FrameLength--------><-AnalysisHop->
//    Frame 1:  [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//    Frame 2:                 [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//    Frame 3:                                [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//
// It then relocates the frames on the time axis by changing the distance
// between them (to SynthesisHop), as illustrated below.
//
//              <--------FrameLength--------><---SynthesisHop--->
//    Frame 1:  [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//    Frame 2:                      [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//    Frame 3:                                          [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//
// This changes the speed of the signal by the ratio AnalysisHop/SynthesisHop
// (for example, if the SynthesisHop is twice the AnalysisHop, the output
// signal will be half as fast as the input signal).
//
// However this simple method introduces artifacts to the signal. These
// artifacts can be reduced by modifying the analysis frames by various
// methods. The modified frames are called the synthesis frames. The conversion
// of the analysis frames into the synthesis frames is handled by the
// Converter.
//
// To further reduce the artifacts, window functions (the AnalysisWindow and
// the SynthesisWindow) can be applied to the analysis frames and the synthesis
// frames in order to smooth the signal.
//
// For more details on Time-Scale Modification procedures, I recommend reading
// "A Review of Time-Scale Modification of music Signals" by Jonathan Driedger
// and Meinard Müller (http://www.mdpi.com/2076-3417/6/2/57).
//
type Settings struct {
	Channels        int
	AnalysisHop     int
	SynthesisHop    int
	FrameLength     int
	AnalysisWindow  []float64
	SynthesisWindow []float64

	// Some TSM methods (such as WSOLA) may need to have access to samples
	// before and after the analysis frame.
	DeltaBefore int
	DeltaAfter  int

	Converter Converter
}

// A TSM is an object implementing a Time-Scale Modification procedure.
//
type TSM struct {
	s *Settings

	// When AnalysisHop is larger than FrameLength, some samples from the input
	// need to be skipped. skipInputSamples tracks how many samples should be
	// skipped before reading the analysis frame.
	skipInputSamples  int
	normalizeWindow   []float64
	skipOutputSamples int

	inBuffer        multichannel.CBuffer
	analysisFrame   multichannel.TSMBuffer
	outBuffer       multichannel.CBuffer
	normalizeBuffer multichannel.NormalizeBuffer
}

// New creates a new TSM object.
//
// New should only be used if you want to implement a new TSM procedure. If you
// just want to use an existing one, you should create the TSM object from one
// of the subpackages of this package.
func New(s Settings) (*TSM, error) {
	normalizeWindow, err := window.Product(s.AnalysisWindow, s.SynthesisWindow)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create normalizeWindow")
	}

	t := &TSM{
		s: &s,

		normalizeWindow: normalizeWindow,

		inBuffer:        multichannel.NewCBuffer(s.Channels, s.DeltaBefore+s.FrameLength+s.DeltaAfter),
		analysisFrame:   multichannel.NewTSMBuffer(s.Channels, s.DeltaBefore+s.FrameLength+s.DeltaAfter),
		outBuffer:       multichannel.NewCBuffer(s.Channels, s.FrameLength),
		normalizeBuffer: multichannel.NewNormalizeBuffer(s.FrameLength),
	}
	t.Clear()

	return t, nil
}

// Clear clears the state of the TSM object, making it ready to be used on
// another signal (or another part of a signal). It is automatically called by
// Flush.
func (t *TSM) Clear() {
	// Clear the buffers
	t.inBuffer.Remove(t.s.FrameLength)
	t.outBuffer.Remove(t.s.FrameLength)
	t.normalizeBuffer.Remove(t.s.FrameLength)

	// Left pad the input with half a frame of zeros, and ignore that half
	// frame in the output. This makes the output signal start in the middle of
	// a frame, which should be the peak of the window function.
	t.inBuffer.SetReadable(t.s.DeltaBefore + t.s.FrameLength/2)
	t.skipOutputSamples = t.s.FrameLength / 2

	t.s.Converter.Clear()
}

// Flush writes the last output samples to the buffer, assuming that no samples
// will be added to the input, and returns the number of samples that were
// written.
//
// The return value will always be equal to buffer.Len(), except when there is
// no more values to be written.
func (t *TSM) Flush(buffer multichannel.Buffer) int {
	expectedLength := buffer.Len()
	length := t.outBuffer.Read(buffer)

	if expectedLength < buffer.Len() {
		t.Clear()
	}

	return length
}

// Put reads samples from buffer and processes them. It returns the number of samples that were read.
//
// Ideally, the length of buffer should be equal to RemainingInputSpace(), but
// it is not required. If it is lower, the samples will be buffered but will
// not be processed. If it is larger, some samples from buffer will not be
// read.
func (t *TSM) Put(buffer multichannel.Buffer) int {
	n := 0
	if t.skipInputSamples >= buffer.Len() {
		// All the samples in the buffer have to be skipped
		n = buffer.Len()
	} else {
		n := t.skipInputSamples
		n += t.inBuffer.Write(buffer.Slice(t.skipInputSamples, buffer.Len()))
	}
	t.skipInputSamples -= n

	if t.inBuffer.RemainingSpace() == 0 && t.outBuffer.RemainingSpace() >= t.s.FrameLength {
		// The input buffer has enough data to process, and there is enough
		// space in the output buffer to put the result.
		t.processFrame()

		if t.skipOutputSamples > t.outBuffer.Len() {
			t.skipOutputSamples -= t.outBuffer.Len()
			t.outBuffer.Remove(t.outBuffer.Len())
		} else if t.skipOutputSamples > 0 {
			t.outBuffer.Remove(t.skipOutputSamples)
			t.skipOutputSamples = 0
		}

		t.skipInputSamples = t.s.AnalysisHop - t.s.FrameLength
		if t.skipInputSamples < 0 {
			t.skipInputSamples = 0
		}
	}

	return n
}

// Receive writes the result of the Time-Scale Modification procedure to
// buffer, and returns the number of samples that were written per channels.
//
// The return value will always be equal to buffer.Len(), except when there is
// no more values to be written. In this case, you should either call Put to
// provide more input samples, or Flush if there is no input samples remaining.
func (t *TSM) Receive(buffer multichannel.Buffer) int {
	return t.outBuffer.Read(buffer)
}

// process reads an analysis frame from the input buffer, process it, and writes the result to the output buffer.
func (t *TSM) processFrame() {
	// Generate analysis frame, and discard the input samples that won't be
	// needed anymore
	t.inBuffer.Peek(t.analysisFrame)
	t.inBuffer.Remove(t.s.AnalysisHop)

	if t.s.AnalysisWindow != nil {
		t.analysisFrame.ApplyWindow(t.s.AnalysisWindow)
	}

	// Convert the analysis frame into a synthesis frame
	synthesisFrame := t.s.Converter.Convert(t.analysisFrame)

	if t.s.SynthesisWindow != nil {
		synthesisFrame.ApplyWindow(t.s.SynthesisWindow)
	}

	// Overlap and add the synthesis frame in the output buffer
	t.outBuffer.Add(synthesisFrame)

	// The overlap and add step changes the volume of the signal. The
	// normalizeBuffer is used to keep track of "how much of the input
	// signal was added" to each part of the output buffer, allowing to
	// normalize it.
	t.normalizeBuffer.Add(t.normalizeWindow)

	// Normalize the samples that are ready to be written to the output
	// (the first synthesisHop ones)
	t.outBuffer.Divide(t.normalizeBuffer, t.s.SynthesisHop)
	t.normalizeBuffer.Remove(t.s.SynthesisHop)
	t.outBuffer.SetReadable(t.s.SynthesisHop)
}

// RemainingInputSpace returns the amount of space available in the input
// buffer, i.e. the number of samples that can be added to each channel of the
// buffer.
func (t *TSM) RemainingInputSpace() int {
	return t.skipInputSamples + t.inBuffer.RemainingSpace()
}

// SetSpeed changes the speed ratio.
func (t *TSM) SetSpeed(speed float64) {
	t.s.AnalysisHop = int(float64(t.s.SynthesisHop) * speed)

}
