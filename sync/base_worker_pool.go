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
	gxlog "github.com/dubbogo/gost/log"
	"runtime/debug"
	"sync"
)

type baseWorkerPool struct {
	logger gxlog.Logger

	taskId     uint32
	taskQueues []chan task

	wg   *sync.WaitGroup
	done chan struct{}
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

	close(p.done)
	for _, q := range p.taskQueues {
		close(q)
	}
	p.wg.Wait()
}

func (p *baseWorkerPool) IsClosed() bool {
	select {
	case <-p.done:
		return true
	default:
	}
	return false
}

func newWorker(q chan task, logger gxlog.Logger, workerId int, wg *sync.WaitGroup) {
	defer wg.Done()

	if logger != nil {
		logger.Debugf("worker #%d is started\n", workerId)
	}
	for {
		select {
		case task, ok := <-q:
			if !ok {
				if logger != nil {
					logger.Debugf("worker #%d is closed\n", workerId)
				}
				return
			}
			if task != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							if logger != nil {
								logger.Errorf("goroutine panic: %v\n%s\n", r, string(debug.Stack()))
							}
						}
					}()
					task()
				}()
			}
		}
	}
}
