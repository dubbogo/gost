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
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestConnectionPool(t *testing.T) {
	t.Run("Count", func(t *testing.T) {
		p := NewConnectionPool(WorkerPoolConfig{
			NumWorkers: 100,
			NumQueues:  runtime.NumCPU(),
			QueueSize:  10,
			Logger:     nil,
			Enable:     true,
		})
		var count int64
		wg := new(sync.WaitGroup)
		for i := 1; i <= 100; i++ {
			wg.Add(1)
			value := i
			err := p.Submit(func() {
				defer wg.Done()
				atomic.AddInt64(&count, int64(value))
			})
			assert.Nil(t, err)
		}
		wg.Wait()
		assert.Equal(t, int64(5050), count)

		p.Close()
	})

	t.Run("PoolBusyErr", func(t *testing.T) {
		p := NewConnectionPool(WorkerPoolConfig{
			NumWorkers: 1,
			NumQueues:  1,
			QueueSize:  0,
			Logger:     nil,
			Enable:     true,
		})

		wg := new(sync.WaitGroup)
		wg.Add(1)
		err := p.Submit(func() {
			wg.Wait()
		})
		assert.Nil(t, err)

		err = p.Submit(func() {})
		assert.Equal(t, PoolBusyErr, err)

		wg.Done()
		time.Sleep(100 * time.Millisecond)
		err = p.Submit(func() {})
		assert.Nil(t, err)

		p.Close()
	})

	t.Run("Close", func(t *testing.T) {
		p := NewConnectionPool(WorkerPoolConfig{
			NumWorkers: runtime.NumCPU(),
			NumQueues:  runtime.NumCPU(),
			QueueSize:  100,
			Enable:     true,
			Logger:     nil,
		})

		assert.Equal(t, runtime.NumCPU(), int(p.NumWorkers()))

		p.Close()
		assert.True(t, p.IsClosed())

		assert.Panics(t, func() {
			_ = p.Submit(func() {})
		})
	})

	t.Run("BorderCondition", func(t *testing.T) {
		p := NewConnectionPool(WorkerPoolConfig{
			NumWorkers: 0,
			NumQueues:  runtime.NumCPU(),
			Enable:     true,
			QueueSize:  100,
			Logger:     nil,
		})
		assert.Equal(t, 1, int(p.NumWorkers()))
		p.Close()

		p = NewConnectionPool(WorkerPoolConfig{
			NumWorkers: 1,
			NumQueues:  0,
			Enable:     true,
			QueueSize:  0,
			Logger:     nil,
		})
		err := p.Submit(func() {})
		assert.Nil(t, err)
		p.Close()

		p = NewConnectionPool(WorkerPoolConfig{
			NumWorkers: 1,
			NumQueues:  1,
			QueueSize:  -1,
			Logger:     nil,
			Enable:     true,
		})

		err = p.Submit(func() {})
		assert.Nil(t, err)
		p.Close()
	})

	t.Run("NilTask", func(t *testing.T) {
		p := NewConnectionPool(WorkerPoolConfig{
			NumWorkers: 1,
			NumQueues:  1,
			Enable:     true,
			QueueSize:  0,
			Logger:     nil,
		})

		err := p.Submit(nil)
		assert.NotNil(t, err)
		p.Close()
	})

	t.Run("CountTask", func(t *testing.T) {
		p := NewConnectionPool(WorkerPoolConfig{
			NumWorkers: runtime.NumCPU(),
			NumQueues:  runtime.NumCPU(),
			QueueSize:  10,
			Logger:     nil,
			Enable:     true,
		})

		task, v := newCountTask()
		wg := new(sync.WaitGroup)
		wg.Add(100)
		for i := 0; i < 100; i++ {
			if err := p.Submit(func() {
				defer wg.Done()
				task()
			}); err != nil {
				i--
			}
		}

		wg.Wait()
		assert.Equal(t, 100, int(*v))
		p.Close()
	})

	t.Run("CountTaskSync", func(t *testing.T) {
		p := NewConnectionPool(WorkerPoolConfig{
			NumWorkers: runtime.NumCPU(),
			NumQueues:  runtime.NumCPU(),
			QueueSize:  10,
			Logger:     nil,
			Enable:     true,
		})

		task, v := newCountTask()
		for i := 0; i < 100; i++ {
			err := p.SubmitSync(task)
			assert.Nil(t, err)
		}

		assert.Equal(t, 100, int(*v))
		p.Close()
	})
}

func BenchmarkConnectionPool(b *testing.B) {
	p := NewConnectionPool(WorkerPoolConfig{
		NumWorkers: 100,
		NumQueues:  runtime.NumCPU(),
		QueueSize:  100,
		Enable:     true,
		Logger:     nil,
	})

	b.Run("CountTask", func(b *testing.B) {
		task, _ := newCountTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = p.Submit(task)
			}
		})
	})

	b.Run("CPUTask", func(b *testing.B) {
		task, _ := newCPUTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = p.Submit(task)
			}
		})
	})

	b.Run("IOTask", func(b *testing.B) {
		task, _ := newIOTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = p.Submit(task)
			}
		})
	})

	b.Run("RandomTask", func(b *testing.B) {
		task, _ := newRandomTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = p.Submit(task)
			}
		})
	})
}
