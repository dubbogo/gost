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
	"github.com/dubbogo/gost/container/queue"
)

type T interface{}

// UnboundedChan is a chan that could grow if the number of elements exceeds the capacity.
type UnboundedChan struct {
	in     chan T
	out   chan T
	queue *gxqueue.CircularUnboundedQueue
}

// NewUnboundedChan creates an instance of UnboundedChan.
func NewUnboundedChan(capacity int) *UnboundedChan {
	ch := &UnboundedChan{
		in:    make(chan T, capacity/3),
		out:   make(chan T, capacity/3),
		queue: gxqueue.NewCircularUnboundedQueue(capacity-2*(capacity/3)),
	}

	go ch.run()

	return ch
}

// In returns write-only chan
func (ch *UnboundedChan) In() chan<- T {
	return ch.in
}

// Out returns read-only chan
func (ch *UnboundedChan) Out() <-chan T {
	return ch.out
}

func (ch *UnboundedChan) Len() int {
	return len(ch.in) + len(ch.out) + ch.queue.Len()
}

func (ch *UnboundedChan) run() {
	defer func() {
		close(ch.out)
	}()

	for {
		val, ok := <-ch.in
		if !ok {
			// `ch.in` was closed and queue has no elements
			return
		}

		select {
		// data was written to `ch.out`
		case ch.out <- val:
			continue
		// `ch.out` is full, move the data to `ch.queue`
		default:
			ch.queue.Push(val)
		}

		for !ch.queue.IsEmpty() {
			select {
			case val, ok := <-ch.in:
				if !ok {
					ch.closeWait()
					return
				}
				ch.queue.Push(val)
			case ch.out <- ch.queue.Peek():
				ch.queue.Pop()
			}
		}
		ch.shrinkQueue()
	}
}

func (ch *UnboundedChan) shrinkQueue() {
	if ch.queue.IsEmpty() && ch.queue.Cap() > ch.queue.InitialSize() {
		ch.queue.Reset()
	}
}

func (ch *UnboundedChan) closeWait() {
	for !ch.queue.IsEmpty() {
		ch.out <- ch.queue.Pop()
	}
}
