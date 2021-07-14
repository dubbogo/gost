/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gxchan

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestBufferWithoutGrowing(t *testing.T) {
	isize := 10
	buffer := NewBuffer(isize)

	buffer.Reset()

	// no data
	v, ok := buffer.Read()
	assert.Nil(t, v)
	assert.False(t, ok)

	// write 1 element
	buffer.Write(1)
	assert.Equal(t, 1, buffer.Len())
	assert.Equal(t, 10, buffer.Cap())
	// peek and pop
	assert.Equal(t, 1, buffer.Peek())
	assert.Equal(t, 1, buffer.Pop())
	// inspect len and cap
	assert.Equal(t, 0, buffer.Len())
	assert.Equal(t, 10, buffer.Cap())

	// write 8 elements
	for i := 0; i < 8; i++ {
		buffer.Write(i)
	}
	assert.Equal(t, 8, buffer.Len())
	assert.Equal(t, 10, buffer.Cap())

	// pop 5 elements
	for i := 0; i < 5; i++ {
		v = buffer.Pop()
		assert.Equal(t, i, v)
	}
	assert.Equal(t, 3, buffer.Len())
	assert.Equal(t, 10, buffer.Cap())

	// write 6 elements
	for i := 0; i < 6; i++ {
		buffer.Write(i)
	}
	assert.Equal(t, 9, buffer.Len())
	assert.Equal(t, 10, buffer.Cap())
}

func TestBufferWithGrowing(t *testing.T) {
	// size < fastGrowThreshold
	buffer := NewBuffer(10)

	// write 11 elements
	for i := 0; i < 11; i++ {
		buffer.Write(i)
	}

	assert.Equal(t, 11, buffer.Len())
	assert.Equal(t, 20, buffer.Cap())

	buffer.Reset()
	assert.Equal(t, 0, buffer.Len())
	assert.Equal(t, 10, buffer.Cap())

	buffer = NewBuffer(fastGrowThreshold)

	// write fastGrowThreshold+1 elements
	for i := 0; i < fastGrowThreshold+1; i++ {
		buffer.Write(i)
	}

	assert.Equal(t, fastGrowThreshold+1, buffer.Len())
	assert.Equal(t, fastGrowThreshold+fastGrowThreshold/4, buffer.Cap())

	buffer.Reset()
	assert.Equal(t, 0, buffer.Len())
	assert.Equal(t, fastGrowThreshold, buffer.Cap())
}
