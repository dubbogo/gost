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
		p := NewConnectionPool(ConnectionPoolConfig{
			NumWorkers: 100,
			NumQueues:  runtime.NumCPU(),
			QueueSize:  10,
			Logger:     nil,
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
		p := NewConnectionPool(ConnectionPoolConfig{
			NumWorkers: 1,
			NumQueues:  1,
			QueueSize:  1,
			Logger:     nil,
		})
		_ = p.Submit(func() {
			time.Sleep(1 * time.Second)
		})

		err := p.Submit(func() {})
		assert.Equal(t, err, PoolBusyErr)

		time.Sleep(1 * time.Second)
		err = p.Submit(func() {})
		assert.Nil(t, err)

		p.Close()
	})

	t.Run("Close", func(t *testing.T) {
		p := NewConnectionPool(ConnectionPoolConfig{
			NumWorkers: runtime.NumCPU(),
			NumQueues:  runtime.NumCPU(),
			QueueSize:  100,
			Logger:     nil,
		})

		assert.Equal(t, runtime.NumCPU(), int(p.NumWorkers()))

		p.Close()
		assert.True(t, p.IsClosed())

		assert.Panics(t, func() {
			_ = p.Submit(func() {})
		})
	})
}

func BenchmarkConnectionPool(b *testing.B) {
	p := NewConnectionPool(ConnectionPoolConfig{
		NumWorkers: 100,
		NumQueues:  runtime.NumCPU(),
		QueueSize:  100,
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
