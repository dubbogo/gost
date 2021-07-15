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

type T interface{}

const (
	fastGrowThreshold = 1024
)

// CircularUnboundedQueue is a circular structure and will grow automatically if it exceeds the capacity.
type CircularUnboundedQueue struct {
	data       []T
	head, tail int
	isize      int // initial size
}

func NewCircularUnboundedQueue(size int) *CircularUnboundedQueue {
	if size < 0 {
		panic("size should be greater than zero")
	}
	return &CircularUnboundedQueue{
		data:  make([]T, size+1),
		isize: size,
	}
}

func (q *CircularUnboundedQueue) IsEmpty() bool {
	return q.head == q.tail
}

func (q *CircularUnboundedQueue) Push(t T) {
	q.data[q.tail] = t

	q.tail = (q.tail + 1) % len(q.data)
	if q.tail == q.head {
		q.grow()
	}
}

func (q *CircularUnboundedQueue) Pop() T {
	if q.IsEmpty() {
		panic("queue has no element")
	}

	t := q.data[q.head]
	q.head = (q.head + 1) % len(q.data)

	return t
}

func (q *CircularUnboundedQueue) Peek() T {
	if q.IsEmpty() {
		panic("queue has no element")
	}
	return q.data[q.head]
}

func (q *CircularUnboundedQueue) Cap() int {
	return len(q.data) - 1
}

func (q *CircularUnboundedQueue) Len() int {
	head, tail := q.head, q.tail
	if head > tail {
		tail += len(q.data)
	}
	return tail - head
}

func (q *CircularUnboundedQueue) Reset() {
	q.data = make([]T, q.isize+1)
	q.head, q.tail = 0, 0
}

func (q *CircularUnboundedQueue) InitialSize() int {
	return q.isize
}

func (q *CircularUnboundedQueue) grow() {
	oldsize := len(q.data) - 1
	var newsize int
	if oldsize < fastGrowThreshold {
		newsize = oldsize * 2
	} else {
		newsize = oldsize + oldsize/4
	}

	newdata := make([]T, newsize+1)
	copy(newdata[0:], q.data[q.head:])
	copy(newdata[len(q.data)-q.head:], q.data[:q.head])

	q.data = newdata
	q.head, q.tail = 0, oldsize+1
}
