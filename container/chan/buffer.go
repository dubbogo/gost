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

const (
	fastGrowThreshold = 1024
)

// Buffer is a circular structure and will grow automatically if it exceeds the capacity.
type Buffer struct {
	data       []T
	head, tail int
	isize      int // initial size
}

func NewBuffer(size int) *Buffer {
	if size < 0 {
		panic("size should be greater than zero")
	}
	return &Buffer{
		data:  make([]T, size+1),
		isize: size,
	}
}

func (b *Buffer) Read() (T, bool) {
	if b.IsEmpty() {
		return nil, false
	}

	t := b.data[b.head]
	b.head = (b.head + 1) % len(b.data)
	return t, true
}

func (b *Buffer) Write(t T) {
	b.data[b.tail] = t

	b.tail = (b.tail + 1) % len(b.data)
	if b.tail == b.head {
		b.grow()
	}
}

func (b *Buffer) IsEmpty() bool {
	return b.head == b.tail
}

func (b *Buffer) Pop() (t T) {
	t, ok := b.Read()
	if !ok {
		panic("buffer has no element")
	}
	return
}

func (b *Buffer) Peek() T {
	if b.IsEmpty() {
		panic("buffer has no element")
	}
	return b.data[b.head]
}

func (b *Buffer) Cap() int {
	return len(b.data) - 1
}

func (b *Buffer) Len() int {
	head, tail := b.head, b.tail
	if head > tail {
		tail += len(b.data)
	}
	return tail - head
}

func (b *Buffer) Reset() {
	b.data = make([]T, b.isize+1)
	b.head, b.tail = 0, 0
}

func (b *Buffer) grow() {
	oldsize := len(b.data) - 1
	var newsize int
	if oldsize < fastGrowThreshold {
		newsize = oldsize * 2
	} else {
		newsize = oldsize + oldsize/4
	}

	newdata := make([]T, newsize+1)
	copy(newdata[0:], b.data[b.head:])
	copy(newdata[len(b.data)-b.head:], b.data[:b.head])

	b.data = newdata
	b.head, b.tail = 0, oldsize+1
}
