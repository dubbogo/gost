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

const (
	blocked = 1 << iota
)

// UnboundedChan is a chan that could grow if the number of elements exceeds the capacity.
// UnboundedChan is not thread-safe.
type UnboundedChan struct {
	in     chan interface{}
	out    chan interface{}
	queue  *gxqueue.CircularUnboundedQueue
	status uint32
}

// NewUnboundedChan creates an instance of UnboundedChan.
func NewUnboundedChan(capacity int) *UnboundedChan {
	return NewUnboundedChanWithQuota(capacity, 0)
}

func NewUnboundedChanWithQuota(capacity, quota int) *UnboundedChan {
	if capacity <= 0 {
		panic("capacity should be greater than 0")
	}
	if quota < 0 {
		panic("quota should be greater or equal to 0")
	}
	if quota != 0 && capacity > quota {
		capacity = quota
	}

	var qquota int
	if quota == 0 {
		qquota = 0
	} else {
		qquota = quota - 2*(capacity/3)
	}

	ch := &UnboundedChan{
		in:    make(chan interface{}, capacity/3-1), // block() could store an extra value
		out:   make(chan interface{}, capacity/3),
		queue: gxqueue.NewCircularUnboundedQueueWithQuota(capacity-2*(capacity/3), qquota),
	}

	go ch.run()

	return ch
}

// In returns write-only chan
func (ch *UnboundedChan) In() chan<- interface{} {
	return ch.in
}

// Out returns read-only chan
func (ch *UnboundedChan) Out() <-chan interface{} {
	return ch.out
}

// Len returns the total length of chan.
// WARNING: DO NOT call Len() when growing, it may cause data race.
func (ch *UnboundedChan) Len() (l int) {
	l = len(ch.in) + len(ch.out) + ch.queue.Len()
	if ch.status&blocked == blocked {
		l++
	}
	return
}

// Cap returns the total capacity of chan.
// WARNING: DO NOT call Cap() when growing, it may cause data race.
func (ch *UnboundedChan) Cap() int {
	return cap(ch.in) + cap(ch.out) + ch.queue.Cap() + 1
}

func (ch *UnboundedChan) run() {
	defer func() {
		close(ch.out)
	}()

	for {
		val, ok := <-ch.in
		if !ok { // `ch.in` was closed and queue has no elements
			return
		}

		select {
		case ch.out <- val: // data was written to `ch.out`
			continue
		default: // `ch.out` is full, move the data to `ch.queue`
			ch.queue.Push(val)
		}

		for !ch.queue.IsEmpty() {
			select {
			case val, ok := <-ch.in: // `ch.in` was closed
				if !ok {
					ch.closeWait()
					return
				}
				if ok = ch.queue.Push(val); !ok { // try to push the value into queue
					ch.block(val)
				}
			case ch.out <- ch.queue.Peek():
				ch.queue.Pop()
			}
		}
		ch.shrinkQueue()
	}
}

func (ch *UnboundedChan) shrinkQueue() {
	if ch.queue.IsEmpty() && ch.queue.Cap() > ch.queue.InitialCap() {
		ch.queue.Reset()
	}
}

// closeWait waits for being empty of `ch.queue`
func (ch *UnboundedChan) closeWait() {
	for !ch.queue.IsEmpty() {
		ch.out <- ch.queue.Pop()
	}
}

// block waits for having an idle space on `ch.out`
func (ch *UnboundedChan) block(val interface{}) {
	defer func() {
		ch.status &^= blocked
	}()
	if ch.status&blocked == blocked {
		panic("reenter block() is not allowed")
	}
	ch.status |= blocked
	select {
	case ch.out <- ch.queue.Peek():
		ch.queue.Pop()
		ch.queue.Push(val)
	}
}
