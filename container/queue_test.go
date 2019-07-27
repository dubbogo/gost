/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package container

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestQueueSimple(t *testing.T) {
	q := NewQueue()

	for i := 0; i < defaultQueueLen; i++ {
		q.Add(i)
	}
	for i := 0; i < defaultQueueLen; i++ {
		v, _ := q.Peek()
		assert.Equal(t, i, v.(int))
		x, _ := q.Remove()
		assert.Equal(t, i, x)
	}
}

func TestQueueWrapping(t *testing.T) {
	q := NewQueue()

	for i := 0; i < defaultQueueLen; i++ {
		q.Add(i)
	}
	for i := 0; i < 3; i++ {
		q.Remove()
		q.Add(defaultQueueLen + i)
	}

	for i := 0; i < defaultQueueLen; i++ {
		v, _ := q.Peek()
		assert.Equal(t, i+3, v.(int))
		q.Remove()
	}
}

func TestQueueLength(t *testing.T) {
	q := NewQueue()

	assert.Equal(t, 0, q.Length(), "empty queue length should be 0")

	for i := 0; i < 1000; i++ {
		q.Add(i)
		assert.Equal(t, i+1, q.Length())
	}
	for i := 0; i < 1000; i++ {
		q.Remove()
		assert.Equal(t, 1000-i-1, q.Length())
	}
}

func TestQueuePeekOutOfRangePanics(t *testing.T) {
	q := NewQueue()

	_, err := q.Peek()
	assert.Equal(t, ErrEmpty, err)

	q.Add(1)
	_, err = q.Remove()
	assert.Nil(t, err)

	_, err = q.Peek()
	assert.Equal(t, ErrEmpty, err)
}

func TestQueueRemoveOutOfRangePanics(t *testing.T) {
	q := NewQueue()

	_, err := q.Remove()
	assert.Equal(t, ErrEmpty, err)

	q.Add(1)
	_, err = q.Remove()
	assert.Nil(t, err)

	_, err = q.Remove()
	assert.Equal(t, ErrEmpty, err)
}

// General warning: Go's benchmark utility (go test -bench .) increases the number of
// iterations until the benchmarks take a reasonable amount of time to run; memory usage
// is *NOT* considered. On my machine, these benchmarks hit around ~1GB before they've had
// enough, but if you have less than that available and start swapping, then all bets are off.

func BenchmarkQueueSerial(b *testing.B) {
	q := NewQueue()
	for i := 0; i < b.N; i++ {
		q.Add(nil)
	}
	for i := 0; i < b.N; i++ {
		q.Peek()
		q.Remove()
	}
}

func BenchmarkQueueTickTock(b *testing.B) {
	q := NewQueue()
	for i := 0; i < b.N; i++ {
		q.Add(nil)
		q.Peek()
		q.Remove()
	}
}
