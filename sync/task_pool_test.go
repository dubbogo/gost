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
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newCountTask() (func(), *int64) {
	var cnt int64
	return func() {
		atomic.AddInt64(&cnt, 1)
	}, &cnt
}

func TestTaskPool(t *testing.T) {
	numCPU := runtime.NumCPU()
	taskCnt := int64(numCPU * numCPU * 100)

	tp := NewTaskPool(10)

	task, cnt := newCountTask()

	var wg sync.WaitGroup
	for i := 0; i < numCPU*numCPU; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				ok := tp.AddTask(task)
				if !ok {
					t.Log(j)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if taskCnt != *cnt {
		t.Error("want ", taskCnt, " got ", *cnt)
	}
}

func BenchmarkTaskPool_CountTask(b *testing.B) {
	tp := NewTaskPool(runtime.NumCPU())

	b.Run(`AddTask`, func(b *testing.B) {
		task, _ := newCountTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tp.AddTask(task)
			}
		})
	})

	b.Run(`AddTaskAlways`, func(b *testing.B) {
		task, _ := newCountTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tp.AddTaskAlways(task)
			}
		})
	})
}

func fib(n int) int {
	if n < 3 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}

// cpu-intensive task
func BenchmarkTaskPool_CPUTask(b *testing.B) {
	tp := NewTaskPool(runtime.NumCPU())

	newCPUTask := func() (func(), *int64) {
		var cnt int64
		return func() {
			atomic.AddInt64(&cnt, int64(fib(22)))
		}, &cnt
	}

	b.Run(`fib`, func(b *testing.B) {
		t, _ := newCPUTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				t()
			}
		})
	})

	b.Run(`AddTask`, func(b *testing.B) {
		task, _ := newCPUTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tp.AddTask(task)
			}
		})
	})

	b.Run(`AddTaskAlways`, func(b *testing.B) {
		task, _ := newCPUTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tp.AddTaskAlways(task)
			}
		})
	})
}

// IO-intensive task
func BenchmarkTaskPool_IOTask(b *testing.B) {
	tp := NewTaskPool(runtime.NumCPU())

	newIOTask := func() (func(), *int64) {
		var cnt int64
		return func() {
			time.Sleep(700 * time.Microsecond)
		}, &cnt
	}

	b.Run(`AddTask`, func(b *testing.B) {
		task, _ := newIOTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tp.AddTask(task)
			}
		})
	})

	b.Run(`AddTaskAlways`, func(b *testing.B) {
		task, _ := newIOTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tp.AddTaskAlways(task)
			}
		})
	})
}

func BenchmarkTaskPool_RandomTask(b *testing.B) {
	tp := NewTaskPool(runtime.NumCPU())

	newRandomTask := func() (func(), *int64) {
		c := rand.Intn(4)
		tasks := []func(){
			func() { _ = fib(rand.Intn(20)) },
			func() { t, _ := newCountTask(); t() },
			func() { runtime.Gosched() },
			func() { time.Sleep(time.Duration(rand.Int63n(100)) * time.Microsecond) },
		}
		return tasks[c], nil
	}

	b.Run(`AddTask`, func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				task, _ := newRandomTask()
				tp.AddTask(task)
			}
		})
	})

	b.Run(`AddTaskAlways`, func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				task, _ := newRandomTask()
				tp.AddTaskAlways(task)
			}
		})
	})
}

/*
goos: darwin
goarch: amd64
pkg: github.com/dubbogo/gost/sync
BenchmarkTaskPool_CountTask/AddTask-8            1724671               655 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_CountTask/AddTaskAlways-8      3102237               339 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_CPUTask/fib-8                    75745             16507 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_CPUTask/AddTask-8                65875             18167 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_CPUTask/AddTaskAlways-8         116119             18804 ns/op               1 B/op          0 allocs/op
BenchmarkTaskPool_IOTask/AddTask-8                 10000            103712 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_IOTask/AddTaskAlways-8         2034420               618 ns/op              87 B/op          1 allocs/op
BenchmarkTaskPool_RandomTask/AddTask-8            462364              2575 ns/op               6 B/op          0 allocs/op
BenchmarkTaskPool_RandomTask/AddTaskAlways-8     3025962               415 ns/op              21 B/op          0 allocs/op
PASS
*/
