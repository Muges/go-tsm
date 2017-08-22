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

// TODO : the last frames may need to be handled differently to avoid fade-out
// TODO : should functions return errors?

// Package tsm implements several real-time time-scale modification methods,
// i.e. algorithms that change the playback speed of an audio signal without
// changing its pitch.
package tsm

import (
	"github.com/Muges/tsm/multichannel"
	"github.com/Muges/tsm/window"
	"github.com/pkg/errors"
)

// A Converter is an object implementing the conversion of an analysis frame
// into a synthesis frame.
type Converter interface {
	Convert(analysisFrame multichannel.TSMBuffer) (synthesisFrame multichannel.TSMBuffer)
}

// A TSM is an object implementing a Time-Scale Modification procedure.
//
// The basic principle of the TSM is to first decompose the input signal into
// short overlapping frames, called the analysis frames. The frames have a
// fixed length frameSize, and are separated by a distance analysisHop, as
// illustrated below.
//
//              <---------frameSize---------><-analysisHop->
//    Frame 1:  [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//    Frame 2:                 [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//    Frame 3:                                [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//
// It then relocates the frames on the time axis by changing the distance
// between them (to synthesisHop), as illustrated below.
//
//              <---------frameSize---------><---synthesisHop--->
//    Frame 1:  [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//    Frame 2:                      [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//    Frame 3:                                          [~~~~~~~~~~~~~~~~~~~~~~~~~~~]
//
// This changes the speed of the signal by the ratio analysisHop/synthesisHop
// (for example, if the synthesisHop is twice the analysisHop, the output
// signal will be half as fast as the input signal).
//
// However this simple method introduces artifacts to the signal. These
// artifacts can be reduced by modifying the analysis frames by various
// methods. The modified frames are called the synthesis frames. The conversion
// of the analysis frames into the synthesis frames is handled by the
// converter.
//
// To further reduce the artifacts, window functions (the analysisWindow and
// the synthesisWindow) can be applied to the analysis frames and the synthesis
// frames in order to smooth the signal.
//
// For more details on Time-Scale Modification procedures, I recommend reading
// "A Review of Time-Scale Modification of music Signals" by Jonathan Driedger
// and Meinard Müller (http://www.mdpi.com/2076-3417/6/2/57).
type TSM struct {
	analysisHop     int
	synthesisHop    int
	frameSize       int
	analysisWindow  []float64
	synthesisWindow []float64
	converter       Converter

	// When analysisHop is larger than frameSize, some samples from the input
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
// channels is the number of channels of the signal that the TSM will process.
// Read the documentation of the TSM type above for an explanation of the other
// arguments.
func New(channels int, analysisHop int, synthesisHop int, frameSize int, analysisWindow []float64, synthesisWindow []float64, converter Converter) (*TSM, error) {
	normalizeWindow, err := window.Product(analysisWindow, synthesisWindow)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create normalizeWindow")
	}

	t := &TSM{
		analysisHop:     analysisHop,
		synthesisHop:    synthesisHop,
		frameSize:       frameSize,
		analysisWindow:  analysisWindow,
		synthesisWindow: synthesisWindow,
		converter:       converter,

		normalizeWindow: normalizeWindow,

		inBuffer:        multichannel.NewCBuffer(channels, frameSize),
		analysisFrame:   multichannel.NewTSMBuffer(channels, frameSize),
		outBuffer:       multichannel.NewCBuffer(channels, frameSize),
		normalizeBuffer: multichannel.NewNormalizeBuffer(frameSize),
	}
	t.Clear()

	return t, nil
}

// Clear clears the state of the TSM object, making it ready to be used on
// another signal (or another part of a signal). It is automatically called by
// Flush.
func (t *TSM) Clear() {
	// Clear the buffers
	t.inBuffer.Remove(t.frameSize)
	t.outBuffer.Remove(t.frameSize)
	t.normalizeBuffer.Remove(t.frameSize)

	// Left pad the input with half a frame of zeros, and ignore that half
	// frame in the output. This makes the output signal start in the middle of
	// a frame, which should be the peak of the window function.
	t.inBuffer.SetReadable(t.frameSize / 2)
	t.skipOutputSamples = t.frameSize / 2
}

// Flush writes the last output samples to the buffer, assuming that no samples
// will be added to the input, and returns the number of samples that were
// written.
//
// The return value will always be equal to buffer.Len(), except when there is
// no more values to be written.
func (t *TSM) Flush(buffer multichannel.Buffer) int {
	length := t.outBuffer.Read(buffer)

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

	if t.inBuffer.Len() >= t.frameSize && t.outBuffer.RemainingSpace() >= t.frameSize {
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

		t.skipInputSamples = t.analysisHop - t.frameSize
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
	t.inBuffer.Remove(t.analysisHop)

	if t.analysisWindow != nil {
		t.analysisFrame.ApplyWindow(t.analysisWindow)
	}

	// Convert the analysis frame into a synthesis frame
	synthesisFrame := t.converter.Convert(t.analysisFrame)

	if t.synthesisWindow != nil {
		synthesisFrame.ApplyWindow(t.synthesisWindow)
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
	t.outBuffer.Divide(t.normalizeBuffer, t.synthesisHop)
	t.normalizeBuffer.Remove(t.synthesisHop)
	t.outBuffer.SetReadable(t.synthesisHop)
}

// RemainingInputSpace returns the amount of space available in the input
// buffer, i.e. the number of samples that can be added to each channel of the
// buffer.
func (t *TSM) RemainingInputSpace() int {
	return t.skipInputSamples + t.inBuffer.RemainingSpace()
}
