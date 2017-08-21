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

// TODO : the first frames may need to be handled differently to avoid fade-in
// TODO : the last frames may need to be handled differently to avoid fade-out
// TODO : should functions return errors?
// TODO : process in put, not in receive?

// Package tsm implements several real-time time-scale modification methods,
// i.e. algorithms that change the playback speed of an audio signal without
// changing its pitch.
package tsm

import (
	"github.com/Muges/tsm/multichannel"
	"github.com/Muges/tsm/window"
	"github.com/pkg/errors"
	"io"
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
	bufferSize      int
	analysisWindow  []float64
	synthesisWindow []float64
	converter       Converter

	normalizeWindow []float64

	inBuffer        multichannel.CBuffer
	analysisFrame   multichannel.TSMBuffer
	outBuffer       multichannel.CBuffer
	normalizeBuffer multichannel.NormalizeBuffer
}

// New creates a new TSM object.
//
// channels is the number of channels of the signal that the TSM will process,
// bufferSize is the size of the input buffer, and should be larger than
// frameSize. Read the documentation of the TSM type above for an explanation
// of the other arguments.
func New(channels int, analysisHop int, synthesisHop int, frameSize int,
	bufferSize int, analysisWindow []float64, synthesisWindow []float64,
	converter Converter) (*TSM, error) {

	if frameSize > bufferSize {
		return nil, errors.New("bufferSize should be larger than frameSize")
	}
	if analysisHop > bufferSize {
		return nil, errors.New("bufferSize should be larger than analysisHop")
	}

	normalizeWindow, err := window.Product(analysisWindow, synthesisWindow)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create normalizeWindow")
	}

	return &TSM{
		analysisHop:     analysisHop,
		synthesisHop:    synthesisHop,
		frameSize:       frameSize,
		bufferSize:      bufferSize,
		analysisWindow:  analysisWindow,
		synthesisWindow: synthesisWindow,
		converter:       converter,

		normalizeWindow: normalizeWindow,

		inBuffer:        multichannel.NewCBuffer(channels, bufferSize),
		analysisFrame:   multichannel.NewTSMBuffer(channels, frameSize),
		outBuffer:       multichannel.NewCBuffer(channels, frameSize+synthesisHop),
		normalizeBuffer: multichannel.NewNormalizeBuffer(frameSize + synthesisHop),
	}, nil
}

// Flush writes the last output samples to the buffer, assuming that no samples
// will be added to the input.
//
// It returns an integer and an error. The error will be equal to io.EOF when
// the input buffer is empty. The integer will always be equal to the length of
// the buffer, except when the input buffer is empty.
func (t *TSM) Flush(buffer multichannel.Buffer) (int, error) {
	length := t.outBuffer.Read(buffer, 0)

	if t.outBuffer.Len() == 0 {
		return length, io.EOF
	}
	return length, nil
}

// InputBufferSize returns the size of the input buffer.
func (t *TSM) InputBufferSize() int {
	return t.bufferSize
}

// Put stores samples in a buffer for them to be processed later. Put will
// return an error if the length of buffer is larger than the result of
// RemainingInputSpace().
func (t *TSM) Put(buffer multichannel.Buffer) error {
	err := t.inBuffer.Write(buffer)
	if err != nil {
		return errors.Wrap(err, "unable to copy samples to input buffer")
	}
	return nil
}

// Receive writes the result of the Time-Scale Modification procedure to
// buffer, and returns the number of samples that were written per channels.
//
// It returns and integer and an error. The error will be equal to io.EOF when
// the input buffer is empty, in which case you should either use Put to add
// new samples, or use Flush if there are no more samples to add to the input.
// The integer will always be equal to the length of the buffer, except when
// the input buffer is empty.
func (t *TSM) Receive(buffer multichannel.Buffer) (int, error) {
	length := t.outBuffer.Read(buffer, 0)

	for length < buffer.Len() && t.inBuffer.Len() >= t.frameSize && t.inBuffer.Len() >= t.analysisHop {
		t.processFrame()

		n := t.outBuffer.Read(buffer, length)
		length += n
	}

	if t.inBuffer.Len() < t.frameSize || t.inBuffer.Len() < t.analysisHop {
		// There is not enough samples in the input buffer for them to be
		// processed
		return length, io.EOF
	}

	return length, nil
}

// process reads an analysis frame from the input buffer, process it, and writes the result to the output buffer.
func (t *TSM) processFrame() {
	// Generate analysis frame, and discard the input samples that won't be
	// needed anymore
	t.inBuffer.Peek(t.analysisFrame, 0)
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
	return t.inBuffer.RemainingSpace()
}
