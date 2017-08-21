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

// A CBuffer is a fixed-size circular buffer used to store multi-channel audio
// data.
//
// A CBuffer is divided into two parts : a readable part, that can be read but
// cannot be modified, and an writable part that can be modified but cannot be
// read.
type CBuffer struct {
	data [][]float64
	size int

	readPointer int
	length      int
}

// NewCBuffer creates a new empty CBuffer, each channel containing at most size
// samples.
func NewCBuffer(channels int, size int) CBuffer {
	data := make([][]float64, channels)
	for k := range data {
		data[k] = make([]float64, size)
	}

	return CBuffer{
		data: data,
		size: size,
	}
}

// Add adds a buffer to the CBuffer element-wise.
//
// The buffer is added in the writable part of the CBuffer, but does not mark
// the samples as readable, allowing them to be modified again by the Add and
// Divide methods. SetReadable should be called to mark these samples as
// readable and to prevent them from being modified.
//
// Add will panic if the two buffer do not have the same number of channels or
// if there is not enough space in the writable part of the CBuffer.
//
//    c := multichannel.NewCBuffer(1, 4)
//    c.Add(multichannel.Buffer{{1, 2}}
//    buffer := multichannel.NewBuffer(1, 3)
//
//    fmt.Println(c.Len()) // prints 0
//    fmt.Println(c.Read(buffer)) // prints 0, and leaves buffer unchanged
//    fmt.Println(buffer) // prints [[0, 0, 0]]
//
//    c.SetReadable(2)
//    fmt.Println(c.Len()) // prints 2
//    fmt.Println(c.Read(buffer)) // prints 2
//    fmt.Println(buffer) // prints [[1, 2, 0]]
//
func (c *CBuffer) Add(buffer TSMBuffer) {
	if len(c.data) != len(buffer) {
		panic("the two buffers should have the same number of channels")
	}

	remainingSpace := c.RemainingSpace()

	for k := range c.data {
		if len(buffer[k]) > remainingSpace {
			panic("not enough space remaining in the circular buffer")
		}

		for i := range buffer[k] {
			c.data[k][(c.readPointer+c.length+i)%c.size] += buffer[k][i]
		}
	}
}

// Divide divides each channel of the CBuffer by the first n values of the
// NormalizeBuffer element-wise.
//
// The CBuffer is divided in its writable part, and the samples are not marked
// as readable, allowing them to be modified again by the Add and Divide
// methods. SetReadable should be called to mark these samples as readable and
// to prevent them from being modified.
//
// The values of the NormalizeBuffer that are lower than 0.0001 are ignored to
// avoid division by zero.
//
// Divide will panic if there is not enough space in the writable part of the
// CBuffer.
func (c *CBuffer) Divide(buffer NormalizeBuffer, n int) {
	const epsilon = 0.0001

	if n > c.RemainingSpace() {
		panic("not enough space remaining in the circular buffer")
	}

	for i := 0; i < n; i++ {
		v := buffer.Get(i)
		if v < -epsilon || v > epsilon {
			for k := range c.data {
				c.data[k][(c.readPointer+c.length+i)%c.size] /= v
			}
		}
	}
}

// Len returns the number of samples that each channel contains (i.e. the size
// of the readable part).
func (c *CBuffer) Len() int {
	return c.length
}

// Peek reads as many samples from the CBuffer as possible (min(c.Len(),
// buffer.Len())) without removing them from the CBuffer, writes them to the
// buffer, and returns the number of samples that were read.
//
// It panics if the two buffer do not have the same number of channels.
func (c *CBuffer) Peek(samples Buffer) int {
	if len(c.data) != samples.Channels() {
		panic("the two buffers should have the same number of channels")
	}

	n := samples.Len()
	if c.length < n {
		n = c.length
	}

	for k := range c.data {
		for i := 0; i < n; i++ {
			samples.SetSample(k, i, c.data[k][(c.readPointer+i)%c.size])
		}

		// TODO : optimize for Buffer
		//copy(buffer[k], c.data[k][c.offset:])
		//if c.size-c.offset < size {
		//	copy(buffer[k][c.size-c.offset:], c.data[k][:c.offset])
		//}
	}

	return n
}

// Read reads as many samples from the CBuffer as possible (min(c.Len(),
// buffer.Len()), removes them from the CBuffer, writes them to the buffer, and
// returns the number of samples that were read.
//
// It panics if the two buffer do not have the same number of channels.
func (c *CBuffer) Read(buffer Buffer) int {
	n := c.Peek(buffer)
	c.Remove(n)
	return n
}

// RemainingSpace returns the number of samples that can be added to each
// channel (i.e. the size of the writable part).
func (c *CBuffer) RemainingSpace() int {
	return c.size - c.length
}

// Remove removes the first n samples of the buffer, preventing them to be read
// again, and leaving more space for new samples to be written.
func (c *CBuffer) Remove(n int) {
	if n > c.length {
		// Remove everything
		n = c.length
	}

	for k := range c.data {
		for i := 0; i < n; i++ {
			c.data[k][(c.readPointer+i)%c.size] = 0
		}
	}

	c.readPointer = (c.readPointer + n) % c.size
	c.length -= n
}

// SetReadable sets the next n samples as readable.
//
// It panics if there is not enough space in the CBuffer.
func (c *CBuffer) SetReadable(n int) {
	if c.RemainingSpace() < n {
		panic("not enough space remaining in the circular buffer")
	}
	c.length += n
}

// Write writes all the data from the buffer to the CBuffer, and returns the
// number of samples that were written.
//
// It panics if the CBuffer and the buffer do not have the same number of
// channels.
func (c *CBuffer) Write(buffer Buffer) int {
	if len(c.data) != buffer.Channels() {
		panic("the two buffers should have the same number of channels")
	}

	n := buffer.Len()
	if c.RemainingSpace() < n {
		n = c.RemainingSpace()
	}

	for k := range c.data {
		for i := 0; i < n; i++ {
			c.data[k][(c.readPointer+c.length+i)%c.size] = buffer.Sample(k, i)
		}

		// TODO : optimize for Buffer
		//copy(buffer[k], c.data[k][c.offset:])
		//if c.size-c.offset < size {
		//	copy(buffer[k][c.size-c.offset:], c.data[k][:c.offset])
		//}
	}
	c.length += n

	return n
}
