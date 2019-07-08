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

package gxsync

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	defaultTaskQNumber = 10
	defaultTaskQLen    = 128
)

/////////////////////////////////////////
// Task Pool Options
/////////////////////////////////////////

type TaskPoolOptions struct {
	tQLen      int // task queue length
	tQNumber   int // task queue number
	tQPoolSize int // task pool size
}

func (o *TaskPoolOptions) validate() {
	if o.tQPoolSize < 1 {
		panic(fmt.Sprintf("illegal pool size %d", o.tQPoolSize))
	}

	if o.tQLen < 1 {
		o.tQLen = defaultTaskQLen
	}

	if o.tQNumber < 1 {
		o.tQNumber = defaultTaskQNumber
	}

	if o.tQNumber > o.tQPoolSize {
		o.tQNumber = o.tQPoolSize
	}
}

type TaskPoolOption func(*TaskPoolOptions)

// @size is the task queue pool size
func WithTaskPoolTaskPoolSize(size int) TaskPoolOption {
	return func(o *TaskPoolOptions) {
		o.tQPoolSize = size
	}
}

// @length is the task queue length
func WithTaskPoolTaskQueueLength(length int) TaskPoolOption {
	return func(o *TaskPoolOptions) {
		o.tQLen = length
	}
}

// @number is the task queue number
func WithTaskPoolTaskQueueNumber(number int) TaskPoolOption {
	return func(o *TaskPoolOptions) {
		o.tQNumber = number
	}
}

/////////////////////////////////////////
// Task Pool
/////////////////////////////////////////

// task t
type task func()

// task pool: manage task ts
type TaskPool struct {
	TaskPoolOptions

	idx    uint32 // round robin index
	qArray []chan task
	wg     sync.WaitGroup

	once sync.Once
	done chan struct{}
}

// build a task pool
func NewTaskPool(opts ...TaskPoolOption) *TaskPool {
	var tOpts TaskPoolOptions
	for _, opt := range opts {
		opt(&tOpts)
	}

	tOpts.validate()

	p := &TaskPool{
		TaskPoolOptions: tOpts,
		qArray:          make([]chan task, tOpts.tQNumber),
		done:            make(chan struct{}),
	}

	for i := 0; i < p.tQNumber; i++ {
		p.qArray[i] = make(chan task, p.tQLen)
	}
	p.start()

	return p
}

// start task pool
func (p *TaskPool) start() {
	for i := 0; i < p.tQPoolSize; i++ {
		p.wg.Add(1)
		workerID := i
		q := p.qArray[workerID%p.tQNumber]
		go p.run(int(workerID), q)
	}
}

// worker
func (p *TaskPool) run(id int, q chan task) error {
	defer p.wg.Done()

	var (
		ok bool
		t  task
	)

	for {
		select {
		case <-p.done:
			if 0 < len(q) {
				return fmt.Errorf("task worker %d exit now while its task buffer length %d is greater than 0",
					id, len(q))
			}

			return nil

		case t, ok = <-q:
			if ok {
				t()
			}
		}
	}
}

// add task
func (p *TaskPool) AddTask(t task) {
	id := atomic.AddUint32(&p.idx, 1) % uint32(p.tQNumber)

	select {
	case <-p.done:
		return
	case p.qArray[id] <- t:
	}
}

// stop all tasks
func (p *TaskPool) stop() {
	select {
	case <-p.done:
		return
	default:
		p.once.Do(func() {
			close(p.done)
		})
	}
}

// check whether the session has been closed.
func (p *TaskPool) IsClosed() bool {
	select {
	case <-p.done:
		return true

	default:
		return false
	}
}

func (p *TaskPool) Close() {
	p.stop()
	p.wg.Wait()
	for i := range p.qArray {
		close(p.qArray[i])
	}
}
