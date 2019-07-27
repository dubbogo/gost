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
	"errors"
	"sync"
)

// minQueueLen is smallest capacity that queue may have.
// Must be power of 2 for bitwise modulus: x % n == x & (n - 1).
const (
	defaultQueueLen = 16
)

var (
	ErrEmpty = errors.New("")
)

// Queue represents a single instance of the queue data structure.
type Queue struct {
	buf               []interface{}
	head, tail, count int
	lock              sync.Mutex
}

// New constructs and returns a new Queue.
func NewQueue() *Queue {
	return &Queue{
		buf: make([]interface{}, defaultQueueLen),
	}
}

// Length returns the number of elements currently stored in the queue.
func (q *Queue) Length() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.count
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (q *Queue) resize() {
	newBuf := make([]interface{}, q.count<<1)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

// Add puts an element on the end of the queue.
func (q *Queue) Add(elem interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.count == len(q.buf) {
		q.resize()
	}

	q.buf[q.tail] = elem
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
}

// Peek returns the element at the head of the queue. return ErrEmpty
// if the queue is empty.
func (q *Queue) Peek() (interface{}, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.count <= 0 {
		return nil, ErrEmpty
	}
	return q.buf[q.head], nil
}

// Remove removes element from the front of queue. If the
// queue is empty, return ErrEmpty
func (q *Queue) Remove() (interface{}, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.count <= 0 {
		return nil, ErrEmpty
	}
	ret := q.buf[q.head]
	q.buf[q.head] = nil
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if len(q.buf) > defaultQueueLen && (q.count<<2) == len(q.buf) {
		q.resize()
	}
	return ret, nil
}
