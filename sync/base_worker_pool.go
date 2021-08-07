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
	"runtime/debug"
	"sync"
)

import (
	"go.uber.org/atomic"
)

import (
	gxlog "github.com/dubbogo/gost/log"
)

type WorkerPoolConfig struct {
	NumWorkers int
	NumQueues  int
	QueueSize  int
	Logger     gxlog.Logger
}

// baseWorkerPool is a worker pool with multiple queues.
//
// The below picture shows baseWorkerPool architecture.
// Note that:
// - TaskQueueX is a channel with buffer, please refer to taskQueues.
// - Workers consume tasks in the dispatched queue only, please refer to dispatch(numWorkers).
// - taskId will be incremented by 1 after a task is enqueued.
// ┌───────┐  ┌───────┐  ┌───────┐                 ┌─────────────────────────┐
// │worker0│  │worker2│  │worker4│               ┌─┤ taskId % NumQueues == 0 │
// └───────┘  └───────┘  └───────┘               │ └─────────────────────────┘
//     │          │          │                   │
//     └───────consume───────┘                enqueue
//                ▼                             task    ╔══════════════════╗
//              ┌──┬──┬──┬──┬──┬──┬──┬──┬──┬──┐  │      ║ baseWorkerPool:  ║
//  TaskQueue0  │t0│t1│t2│t3│t4│t5│t6│t7│t8│t9│◀─┘      ║                  ║
//              ├──┼──┼──┼──┼──┼──┼──┼──┼──┼──┤         ║ *NumWorkers=6    ║
//  TaskQueue1  │t0│t1│t2│t3│t4│t5│t6│t7│t8│t9│◀┐       ║ *NumQueues=2     ║
//              └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┘ │       ║ *QueueSize=10    ║
//                ▲                          enqueue    ╚══════════════════╝
//     ┌───────consume───────┐                 task
//     │          │          │                  │
// ┌───────┐  ┌───────┐  ┌───────┐              │  ┌─────────────────────────┐
// │worker1│  │worker3│  │worker5│              └──│ taskId % NumQueues == 1 │
// └───────┘  └───────┘  └───────┘                 └─────────────────────────┘
type baseWorkerPool struct {
	logger gxlog.Logger

	taskId     uint32
	taskQueues []chan task

	numWorkers *atomic.Int32

	wg *sync.WaitGroup
}

func newBaseWorkerPool(config WorkerPoolConfig) *baseWorkerPool {
	if config.NumWorkers < 1 {
		config.NumWorkers = 1
	}
	if config.NumQueues < 1 {
		config.NumQueues = 1
	}
	if config.QueueSize < 0 {
		config.QueueSize = 0
	}

	taskQueues := make([]chan task, config.NumQueues)
	for i := range taskQueues {
		taskQueues[i] = make(chan task, config.QueueSize)
	}

	p := &baseWorkerPool{
		logger:     config.Logger,
		taskQueues: taskQueues,
		numWorkers: new(atomic.Int32),
		wg:         new(sync.WaitGroup),
	}

	initWg := new(sync.WaitGroup)
	initWg.Add(config.NumWorkers)
	p.dispatch(config.NumWorkers, initWg)

	initWg.Wait()

	return p
}

func (p *baseWorkerPool) dispatch(numWorkers int, wg *sync.WaitGroup) {
	for i := 0; i < numWorkers; i++ {
		p.newWorker(i%len(p.taskQueues), i, wg)
	}
}

func (p *baseWorkerPool) Submit(t task) error {
	panic("implement me")
}

func (p *baseWorkerPool) SubmitSync(t task) error {
	panic("implement me")
}

func (p *baseWorkerPool) Close() {
	if p.IsClosed() {
		return
	}

	for _, q := range p.taskQueues {
		close(q)
	}
	p.wg.Wait()
}

func (p *baseWorkerPool) IsClosed() bool {
	return p.NumWorkers() == 0
}

func (p *baseWorkerPool) NumWorkers() int32 {
	return p.numWorkers.Load()
}

func (p *baseWorkerPool) newWorker(chanId, workerId int, wg *sync.WaitGroup) {
	p.wg.Add(1)
	p.numWorkers.Add(1)
	go p.worker(chanId, workerId, wg)
}

func (p *baseWorkerPool) worker(chanId, workerId int, wg *sync.WaitGroup) {
	defer func() {
		if n := p.numWorkers.Add(-1); n < 0 {
			panic(fmt.Sprintf("numWorkers should be greater or equal to 0, but the value is %d", n))
		}
		p.wg.Done()
	}()

	if p.logger != nil {
		p.logger.Infof("worker #%d is started\n", workerId)
	}

	wg.Done()
	for {
		select {
		case t, ok := <-p.taskQueues[chanId]:
			if !ok {
				if p.logger != nil {
					p.logger.Infof("worker #%d is closed\n", workerId)
				}
				return
			}
			if t != nil {
				func() {
					// prevent from goroutine panic
					defer func() {
						if r := recover(); r != nil {
							if p.logger != nil {
								p.logger.Errorf("goroutine panic: %v\n%s\n", r, string(debug.Stack()))
							}
						}
					}()
					// execute task
					t()
				}()
			}
		}
	}
}
