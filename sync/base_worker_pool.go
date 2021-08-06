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
	"sync/atomic"
)

import (
	gxlog "github.com/dubbogo/gost/log"
)

type baseWorkerPool struct {
	logger gxlog.Logger

	taskId     uint32
	taskQueues []chan task

	numWorkers int32

	wg *sync.WaitGroup
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
	return p.numWorkers == 0
}

func (p *baseWorkerPool) NumWorkers() int32 {
	return p.numWorkers
}

func (p *baseWorkerPool) newWorker(chanId, workerId int) {
	p.wg.Add(1)
	p.numWorkers++
	go p.worker(chanId, workerId)
}

func (p *baseWorkerPool) worker(chanId, workerId int) {
	defer func() {
		if n := atomic.AddInt32(&p.numWorkers, -1); n < 0 {
			panic(fmt.Sprintf("numWorkers should be greater or equal to 0, but the value is %d", n))
		}
		p.wg.Done()
	}()

	if p.logger != nil {
		p.logger.Debugf("worker #%d is started\n", workerId)
	}

	for {
		select {
		case t, ok := <-p.taskQueues[chanId]:
			if !ok {
				if p.logger != nil {
					p.logger.Debugf("worker #%d is closed\n", workerId)
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
