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
	"github.com/Muges/go-tsm/multichannel"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmptyCBuffer(t *testing.T) {
	assert := assert.New(t)

	buffer := multichannel.NewCBuffer(2, 10)

	assert.Equal(10, buffer.RemainingSpace(), "Remaining space in a new CBuffer")
	assert.Equal(0, buffer.Len(), "Used space in a new CBuffer")

	samples := multichannel.NewTSMBuffer(2, 3)
	n := buffer.Peek(samples)
	assert.Equal(0, n, "Size of read on a new CBuffer")
	assert.Equal(multichannel.TSMBuffer{{0, 0, 0}, {0, 0, 0}}, samples, "Peek on a new CBuffer")
}

func TestFull(t *testing.T) {
	assert := assert.New(t)

	buffer := multichannel.NewCBuffer(2, 5)
	buffer.Write(multichannel.TSMBuffer{{0, 1, 2, 3, 4}, {5, 6, 7, 8, 9}})

	assert.Equal(0, buffer.RemainingSpace(), "Remaining space in a full CBuffer (1)")
	assert.Equal(5, buffer.Len(), "Used space in a full CBuffer (1)")

	samples := multichannel.NewTSMBuffer(2, 1)
	n := buffer.Peek(samples)
	buffer.Remove(n)

	buffer.Write(multichannel.TSMBuffer{{0}, {0}})

	assert.Equal(0, buffer.RemainingSpace(), "Remaining space in a full CBuffer (2)")
	assert.Equal(5, buffer.Len(), "Used space in a full CBuffer (2)")
}

func TestAddSetReadableDivide(t *testing.T) {
	assert := assert.New(t)

	buffer := multichannel.NewCBuffer(2, 10)
	buffer.SetReadable(2)
	buffer.Remove(2)
	buffer.Add(multichannel.TSMBuffer{{1, 2, 3}, {4, 5, 6}})
	buffer.Add(multichannel.TSMBuffer{{1, 2, 3}, {4, 5, 6}})

	normalizeBuffer := multichannel.NewNormalizeBuffer(3)
	normalizeBuffer.Remove(2)
	normalizeBuffer.Add([]float64{2, 2, 2})
	buffer.Divide(normalizeBuffer, 3)

	// Check that the buffer is still considered empty
	assert.Equal(10, buffer.RemainingSpace(), "Remaining space in a new CBuffer after Add")
	assert.Equal(0, buffer.Len(), "Used space in a new CBuffer after Add")

	samples := multichannel.NewTSMBuffer(2, 3)
	n := buffer.Peek(samples)
	assert.Equal(0, n, "Size of read on a new CBuffer after Add")
	assert.Equal(multichannel.TSMBuffer{{0, 0, 0}, {0, 0, 0}}, samples, "Peek on a new CBuffer after Add")

	buffer.SetReadable(3)
	assert.Equal(7, buffer.RemainingSpace(), "Remaining space in a new CBuffer after Add and SetReadable")
	assert.Equal(3, buffer.Len(), "Used space in a new CBuffer after Add and SetReadable")

	assert.Panics(func() {
		buffer.Add(multichannel.TSMBuffer{{1, 2, 3, 4, 5, 6, 7, 8}, {1, 2, 3, 4, 5, 6, 7, 8}})
	}, "Panic on Add")

	n = buffer.Peek(samples)
	assert.Equal(3, n, "Size of read on a new CBuffer after Add and SetReadable")
	assert.Equal(multichannel.TSMBuffer{{1, 2, 3}, {4, 5, 6}}, samples, "Peek on a new CBuffer after Add and SetReadable")
}

func TestWrite(t *testing.T) {
	assert := assert.New(t)

	buffer := multichannel.NewCBuffer(2, 5)
	buffer.Write(multichannel.TSMBuffer{{1, 2, 3}, {4, 5, 6}})

	assert.Equal(2, buffer.RemainingSpace(), "Remaining space in a new CBuffer after Write")
	assert.Equal(3, buffer.Len(), "Used space in a new CBuffer after Write")

	samples := multichannel.NewTSMBuffer(2, 3)
	n := buffer.Peek(samples)
	assert.Equal(3, n, "Size of read on a new CBuffer after Write")
	assert.Equal(multichannel.TSMBuffer{{1, 2, 3}, {4, 5, 6}}, samples, "Peek on a new CBuffer after Write")

	buffer.Remove(n)
	assert.Equal(5, buffer.RemainingSpace(), "Remaining space in a new CBuffer after Write and Read")
	assert.Equal(0, buffer.Len(), "Used space in a new CBuffer after Write and Peek")

	buffer.Write(multichannel.TSMBuffer{{1, 2, 3, 4}, {5, 6, 7, 8}})

	assert.Equal(1, buffer.RemainingSpace(), "Remaining space in a new CBuffer after Write and Peek and Write")
	assert.Equal(4, buffer.Len(), "Used space in a new CBuffer after Write and Peek and Write")

	n = buffer.Write(multichannel.TSMBuffer{{1, 2}, {3, 4}})
	assert.Equal(1, n, "Incomplete Write")
}
