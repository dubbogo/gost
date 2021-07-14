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

package chanx

type T interface{}

// AdaptiveChan is a chan that could grow if the number of elements exceeds the capacity.
type AdaptiveChan struct {
	in     chan T
	out    chan T
	buffer *Buffer
}

// NewAdaptiveChan creates an instance of AdaptiveChan.
// incap: The capacity of the in chan
// outcap: The capacity of the out chan
// bufcap: The Capacity of the buffer
func NewAdaptiveChan(incap, outcap, bufcap int) *AdaptiveChan {
	ch := &AdaptiveChan{
		in:     make(chan T, incap),
		out:    make(chan T, outcap),
		buffer: NewBuffer(bufcap),
	}

	go ch.process()

	return ch
}

// In returns write-only chan
func (ch *AdaptiveChan) In() chan<- T {
	return ch.in
}

// Out returns read-only chan
func (ch *AdaptiveChan) Out() <-chan T {
	return ch.out
}

func (ch *AdaptiveChan) Len() int {
	return len(ch.in) + len(ch.out) + ch.buffer.Len()
}

func (ch *AdaptiveChan) BufLen() int {
	return ch.buffer.Len()
}

func (ch *AdaptiveChan) process() {
	defer close(ch.out)

loop:
	for {
		val, ok := <-ch.in
		if !ok { // `in` is closed
			break loop
		}

		select {
		case ch.out <- val: // `ch.out` is not full
			continue
		default:
		}

		// `ch.out` is full, write the value to buffer
		ch.buffer.Write(val)
		for !ch.buffer.IsEmpty() {
			select {
			case val, ok := <-ch.in:
				if !ok {
					break loop
				}
				ch.buffer.Write(val)
			case ch.out <- ch.buffer.Peek():
				ch.buffer.Pop()
				ch.tryToShrinkBuffer()
			}
		}
	}

	// waiting for out chan
	for !ch.buffer.IsEmpty() {
		ch.out <- ch.buffer.Pop()
	}
}

func (ch *AdaptiveChan) tryToShrinkBuffer() {
	if ch.buffer.IsEmpty() && ch.buffer.Cap() > ch.buffer.isize {
		ch.buffer.Reset()
	}
}
