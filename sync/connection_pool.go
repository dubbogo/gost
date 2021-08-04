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
	perrors "github.com/pkg/errors"
)

import (
	"github.com/dubbogo/gost/log"
	gxruntime "github.com/dubbogo/gost/runtime"
)

func NewConnectionPool(maxWorkers, taskQueueSize int, logger gxlog.Logger) WorkerPool {
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	if taskQueueSize < 0 {
		taskQueueSize = 0
	}

	p := &ConnectionPool{
		logger: logger,
		maxWorkers:  maxWorkers,
		workerQueue: make(chan task),
		taskQueue:   make(chan task, taskQueueSize),
		done:        make(chan struct{}),
	}

	go p.dispatch()

	return p
}

type ConnectionPool struct {
	logger gxlog.Logger

	maxWorkers int

	workerQueue chan task
	taskQueue   chan task

	done chan struct{}
}

func (p *ConnectionPool) dispatch() {
	defer close(p.done)

	var workerCount int

loop:
	for {
		select {
		case t, ok := <-p.taskQueue:
			if !ok {
				break loop
			}
			// select a worker to execute the task
			select {
			case p.workerQueue <- t:
			default:
				if workerCount < p.maxWorkers {
					workerId := workerCount
					// number of workers not reaches the limitation
					gxruntime.GoSafely(nil, false, func() {
						newWorker(t, p.workerQueue, p.logger, workerId)
					}, nil)
					workerCount++
				} else {
					// blocked and waiting for a worker
					p.workerQueue <- t
				}
			}
		}
	}

	// waiting for the end of all tasks, and shutting down workers
	for workerCount > 0 {
		p.workerQueue <- nil
		workerCount--
	}

}

// Submit adds a task to queue asynchronously.
func (p *ConnectionPool) Submit(t task) error {
	if t == nil {
		return perrors.New("task shouldn't be nil")
	}
	select {
	case p.taskQueue <- t:
		return nil
	default:
		return PoolBusyErr
	}
}

// SubmitSync adds a task to queue synchronously.
func (p *ConnectionPool) SubmitSync(t task) error {
	if t == nil {
		return perrors.New("task shouldn't be nil")
	}

	done := make(chan struct{})
	fn := func() {
		t()
		close(done)
	}
	select {
	case p.taskQueue <- fn:
		<-done
		return nil
	default:
		return PoolBusyErr
	}
}

func (p *ConnectionPool) Close() {
	select {
	case <-p.done:
		return
	default:
	}

	close(p.taskQueue)
	<-p.done
}

func (p *ConnectionPool) IsClosed() bool {
	select {
	case <-p.done:
		return true
	default:
	}
	return false
}

func newWorker(t task, workerQueue chan task, logger gxlog.Logger, workerId int) {
	gxruntime.GoSafely(nil, false, t, nil)
	gxruntime.GoSafely(nil, false, func() {
		worker(workerQueue, logger, workerId)
	}, nil)

}

func worker(workerQueue chan task, logger gxlog.Logger, workerId int) {
	if logger != nil {
		logger.Debugf("worker #%d is started\n", workerId)
	}
	for task := range workerQueue {
		if task == nil {
			if logger != nil {
				logger.Debugf("worker #%d is closed\n", workerId)
			}
			return
		}
		task()
	}
}
