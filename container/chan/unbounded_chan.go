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
	"sync/atomic"
	"time"
)

import (
	"github.com/dubbogo/gost/container/queue"
)

// UnboundedChan is a chan that could grow if the number of elements exceeds the capacity.
// UnboundedChan is not thread-safe.
type UnboundedChan struct {
	in       chan interface{}
	out      chan interface{}
	queue    *gxqueue.CircularUnboundedQueue
	queueLen int32
	queueCap int32
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
	atomic.StoreInt32(&ch.queueCap, int32(ch.queue.Cap()))

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
func (ch *UnboundedChan) Len() int {
	// time.Sleep is required to ensure Len() returns the correct results
	time.Sleep(1 * time.Millisecond)
	return len(ch.in) + len(ch.out) + int(atomic.LoadInt32(&ch.queueLen))
}

// Cap returns the total capacity of chan.
// WARNING: DO NOT call Cap() when growing, it may cause data race.
func (ch *UnboundedChan) Cap() int {
	// time.Sleep is required to ensure Len() returns the correct results
	time.Sleep(1 * time.Millisecond)
	return cap(ch.in) + cap(ch.out) + int(atomic.LoadInt32(&ch.queueCap)) + 1
}

func (ch *UnboundedChan) run() {
	defer func() {
		close(ch.out)
	}()

	for {
		val, ok := <-ch.in
		atomic.AddInt32(&ch.queueLen, 1)
		if !ok { // `ch.in` was closed and queue has no elements
			return
		}

		select {
		case ch.out <- val: // data was written to `ch.out`
			atomic.AddInt32(&ch.queueLen, -1)
			continue
		default: // `ch.out` is full, move the data to `ch.queue`
			ch.queue.Push(val)
			atomic.StoreInt32(&ch.queueCap, int32(ch.queue.Cap()))
		}

		for !ch.queue.IsEmpty() {
			select {
			case val, ok := <-ch.in: // `ch.in` was closed
				if !ok {
					ch.closeWait()
					return
				}
				atomic.AddInt32(&ch.queueLen, 1)
				if ok = ch.queue.Push(val); !ok { // try to push the value into queue
					ch.block(val)
				}
				atomic.StoreInt32(&ch.queueCap, int32(ch.queue.Cap()))
			case ch.out <- ch.queue.Peek():
				atomic.AddInt32(&ch.queueLen, -1)
				ch.queue.Pop()
			}
		}
		if ch.queue.Cap() > ch.queue.InitialCap() {
			ch.queue.Reset()
			atomic.StoreInt32(&ch.queueCap, int32(ch.queue.Cap()))
		}
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
	select {
	case ch.out <- ch.queue.Peek():
		ch.queue.Pop()
		atomic.AddInt32(&ch.queueLen, -1)
		ch.queue.Push(val)
		// Here not needs `atomic.StoreInt32(&ch.queueCap, int32(ch.queue.Cap()))` due to capacity couldn't be larger.
	}
}
