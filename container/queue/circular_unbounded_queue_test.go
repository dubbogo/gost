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

package gxqueue

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestCircularUnboundedQueueWithoutGrowing(t *testing.T) {
	queue := NewCircularUnboundedQueue(10)

	queue.Reset()

	// write 1 element
	queue.Push(1)
	assert.Equal(t, 1, queue.Len())
	assert.Equal(t, 10, queue.Cap())
	// peek and pop
	assert.Equal(t, 1, queue.Peek())
	assert.Equal(t, 1, queue.Pop())
	// inspect len and cap
	assert.Equal(t, 0, queue.Len())
	assert.Equal(t, 10, queue.Cap())

	// write 8 elements
	for i := 0; i < 8; i++ {
		queue.Push(i)
	}
	assert.Equal(t, 8, queue.Len())
	assert.Equal(t, 10, queue.Cap())

	var v interface{}
	// pop 5 elements
	for i := 0; i < 5; i++ {
		v = queue.Pop()
		assert.Equal(t, i, v)
	}
	assert.Equal(t, 3, queue.Len())
	assert.Equal(t, 10, queue.Cap())

	// write 6 elements
	for i := 0; i < 6; i++ {
		queue.Push(i)
	}
	assert.Equal(t, 9, queue.Len())
	assert.Equal(t, 10, queue.Cap())
}

func TestBufferWithGrowing(t *testing.T) {
	// size < fastGrowThreshold
	queue := NewCircularUnboundedQueue(10)

	// write 11 elements
	for i := 0; i < 11; i++ {
		queue.Push(i)
	}

	assert.Equal(t, 11, queue.Len())
	assert.Equal(t, 20, queue.Cap())

	queue.Reset()
	assert.Equal(t, 0, queue.Len())
	assert.Equal(t, 10, queue.Cap())

	queue = NewCircularUnboundedQueue(fastGrowThreshold)

	// write fastGrowThreshold+1 elements
	for i := 0; i < fastGrowThreshold+1; i++ {
		queue.Push(i)
	}

	assert.Equal(t, fastGrowThreshold+1, queue.Len())
	assert.Equal(t, fastGrowThreshold+fastGrowThreshold/4, queue.Cap())

	queue.Reset()
	assert.Equal(t, 0, queue.Len())
	assert.Equal(t, fastGrowThreshold, queue.Cap())
}
