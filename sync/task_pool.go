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
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

import (
	gxruntime "github.com/dubbogo/gost/runtime"
)

// task t
type task func()

/////////////////////////////////////////
// Task Pool
/////////////////////////////////////////
type TaskPool interface {
	AddTask(t task) bool
	AddTaskAlways(t task) bool
}

type taskPool struct {
	work chan task
	sem  chan struct{}
}

func NewTaskPool(size int) TaskPool {
	if size < 1 {
		size = runtime.NumCPU() * 100
	}
	return &taskPool{
		work: make(chan task),
		sem:  make(chan struct{}, size),
	}
}

func (p *taskPool) AddTask(t task) (ok bool) {
	select {
	case p.work <- t:
	case p.sem <- struct{}{}:
		go p.worker(t)
	}
	return true
}

func (p *taskPool) AddTaskAlways(t task) (ok bool) {
	select {
	case p.work <- t:
		return
	default:
	}
	select {
	case p.work <- t:
	case p.sem <- struct{}{}:
		go p.worker(t)
	default:
		p.goSafely(t)
	}
	return true
}

func (p *taskPool) worker(t task) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "%s goroutine panic: %v\n%s\n",
				time.Now(), r, string(debug.Stack()))
		}
		<-p.sem
	}()
	for {
		t()
		t = <-p.work
	}
}

func (p *taskPool) goSafely(fn func()) {
	gxruntime.GoSafely(nil, false, fn, nil)
}
