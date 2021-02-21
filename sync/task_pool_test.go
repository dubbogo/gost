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

func newIOTask() (func(), *int64) {
	var cnt int64
	return func() {
		time.Sleep(700 * time.Microsecond)
	}, &cnt
}

func newCPUTask() (func(), *int64) {
	var cnt int64
	return func() {
		atomic.AddInt64(&cnt, int64(fib(22)))
	}, &cnt
}

func newRandomTask() (func(), *int64) {
	c := rand.Intn(4)
	tasks := []func(){
		func() { _ = fib(rand.Intn(20)) },
		func() { t, _ := newCountTask(); t() },
		func() { runtime.Gosched() },
		func() { time.Sleep(time.Duration(rand.Int63n(100)) * time.Microsecond) },
	}
	return tasks[c], nil
}

func TestTaskPoolSimple(t *testing.T) {
	numCPU := runtime.NumCPU()
	taskCnt := int64(numCPU * numCPU * 100)

	tp := NewTaskPoolSimple(1)

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
	tp.Close()

	cntValue := atomic.LoadInt64(cnt)
	if taskCnt != cntValue {
		t.Error("want ", taskCnt, " got ", cntValue)
	}
}

func BenchmarkTaskPoolSimple_CountTask(b *testing.B) {
	tp := NewTaskPoolSimple(runtime.NumCPU())

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
func BenchmarkTaskPoolSimple_CPUTask(b *testing.B) {
	tp := NewTaskPoolSimple(runtime.NumCPU())

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
func BenchmarkTaskPoolSimple_IOTask(b *testing.B) {
	tp := NewTaskPoolSimple(runtime.NumCPU())

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

func BenchmarkTaskPoolSimple_RandomTask(b *testing.B) {
	tp := NewTaskPoolSimple(runtime.NumCPU())

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

func TestTaskPool(t *testing.T) {
	numCPU := runtime.NumCPU()
	//taskCnt := int64(numCPU * numCPU * 100)

	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(1),
		WithTaskPoolTaskQueueNumber(1),
		WithTaskPoolTaskQueueLength(1),
	)

	//task, cnt := newCountTask()
	task, _ := newCountTask()

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
	tp.Close()

	//if taskCnt != atomic.LoadInt64(cnt) {
	//	//t.Error("want ", taskCnt, " got ", *cnt)
	//}
}

func BenchmarkTaskPool_CountTask(b *testing.B) {
	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(runtime.NumCPU()),
		WithTaskPoolTaskQueueNumber(runtime.NumCPU()),
		//WithTaskPoolTaskQueueLength(runtime.NumCPU()),
	)

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

	b.Run(`AddTaskBalance`, func(b *testing.B) {
		task, _ := newCountTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tp.AddTaskBalance(task)
			}
		})
	})

}

// cpu-intensive task
func BenchmarkTaskPool_CPUTask(b *testing.B) {
	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(runtime.NumCPU()),
		WithTaskPoolTaskQueueNumber(runtime.NumCPU()),
		//WithTaskPoolTaskQueueLength(runtime.NumCPU()),
	)

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

	b.Run(`AddTaskBalance`, func(b *testing.B) {
		task, _ := newCPUTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tp.AddTaskBalance(task)
			}
		})
	})

}

// IO-intensive task
func BenchmarkTaskPool_IOTask(b *testing.B) {
	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(runtime.NumCPU()),
		WithTaskPoolTaskQueueNumber(runtime.NumCPU()),
		//WithTaskPoolTaskQueueLength(runtime.NumCPU()),
	)

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

	b.Run(`AddTaskBalance`, func(b *testing.B) {
		task, _ := newIOTask()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tp.AddTaskBalance(task)
			}
		})
	})
}

func BenchmarkTaskPool_RandomTask(b *testing.B) {
	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(runtime.NumCPU()),
		WithTaskPoolTaskQueueNumber(runtime.NumCPU()),
		//WithTaskPoolTaskQueueLength(runtime.NumCPU()),
	)

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

	b.Run(`AddTaskBalance`, func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				task, _ := newRandomTask()
				tp.AddTaskBalance(task)
			}
		})
	})
}

func PrintMemUsage(t *testing.T, prefix string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("%s Alloc = %v MiB", prefix, bToMb(m.Alloc))
	t.Logf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	t.Logf("\tSys = %v MiB", bToMb(m.Sys))
	t.Logf("\tNumGC = %v\n", m.NumGC)
}

func elapsed(t *testing.T, what string) func() {
	start := time.Now()
	return func() {
		t.Logf("\n\t %s took %v\n", what, time.Since(start))
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

var n = 100000

func TestWithoutPool(t *testing.T) {
	PrintMemUsage(t, "Before")
	numG := runtime.NumGoroutine()
	defer elapsed(t, "TestWithoutPool")()
	var wg sync.WaitGroup
	task, _ := newIOTask()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			task()
			wg.Done()
		}()
	}
	t.Logf("TestWithoutPool took %v goroutines\n", runtime.NumGoroutine()-numG)
	wg.Wait()
	PrintMemUsage(t, "After")
}

func TestWithSimpledPoolUseAlways(t *testing.T) {
	PrintMemUsage(t, "Before")
	numG := runtime.NumGoroutine()
	defer elapsed(t, "TestWithSimplePool")()
	tp := NewTaskPoolSimple(1000)
	task, _ := newIOTask()
	for i := 0; i < n; i++ {
		tp.AddTaskAlways(task)
	}
	t.Logf("TestWithSimplePool took %v goroutines\n", runtime.NumGoroutine()-numG)
	tp.Close()
	PrintMemUsage(t, "After")
}

func TestWithSimplePool(t *testing.T) {
	PrintMemUsage(t, "Before")
	numG := runtime.NumGoroutine()
	defer elapsed(t, "TestWithSimplePool")()
	tp := NewTaskPoolSimple(1000)
	task, _ := newIOTask()
	for i := 0; i < n; i++ {
		tp.AddTask(task)
	}
	t.Logf("TestWithSimplePool took %v goroutines\n", runtime.NumGoroutine()-numG)
	tp.Close()
	PrintMemUsage(t, "After")
}

func TestWithPool(t *testing.T) {
	PrintMemUsage(t, "Before")
	numG := runtime.NumGoroutine()
	defer elapsed(t, "TestWithPool")()
	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(1000),
		WithTaskPoolTaskQueueNumber(2),
		//WithTaskPoolTaskQueueLength(runtime.NumCPU()),
	)
	task, _ := newIOTask()
	for i := 0; i < n; i++ {
		tp.AddTask(task)
	}
	t.Logf("TestWithPool took %v goroutines\n", runtime.NumGoroutine()-numG)
	tp.Close()
	PrintMemUsage(t, "After")
}

func TestWithPoolUseAlways(t *testing.T) {
	PrintMemUsage(t, "Before")
	numG := runtime.NumGoroutine()
	defer elapsed(t, "TestWithPoolUseAlways")()
	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(1000),
		WithTaskPoolTaskQueueNumber(10),
		//WithTaskPoolTaskQueueLength(runtime.NumCPU()),
	)
	task, _ := newIOTask()
	for i := 0; i < n; i++ {
		tp.AddTaskAlways(task)
	}
	t.Logf("TestWithPoolUseAlways took %v goroutines\n", runtime.NumGoroutine()-numG)
	tp.Close()
	PrintMemUsage(t, "After")
}

/*
goos: darwin
goarch: amd64
pkg: github.com/dubbogo/gost/sync                        执行次数             单次执行时间                单次执行内存消耗   单次执行内存分配次数
BenchmarkTaskPoolSimple_CountTask/AddTask-8              1693192               700 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPoolSimple_CountTask/AddTaskAlways-8        3262932               315 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPoolSimple_CPUTask/fib-8                      83479             14760 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPoolSimple_CPUTask/AddTask-8                  85956             14571 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPoolSimple_CPUTask/AddTaskAlways-8          1000000             17712 ns/op              19 B/op          0 allocs/op
BenchmarkTaskPoolSimple_IOTask/AddTask-8                   10000            107361 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPoolSimple_IOTask/AddTaskAlways-8           2772476               477 ns/op              79 B/op          1 allocs/op
BenchmarkTaskPoolSimple_RandomTask/AddTask-8              499417              2451 ns/op               6 B/op          0 allocs/op
BenchmarkTaskPoolSimple_RandomTask/AddTaskAlways-8       3307748               354 ns/op              21 B/op          0 allocs/op

BenchmarkTaskPool_CountTask/AddTask-8                    5367189               229 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_CountTask/AddTaskAlways-8              5438667               218 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_CountTask/AddTaskBalance-8             4765616               247 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_CPUTask/fib-8                            74749             17153 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_CPUTask/AddTask-8                        71020             18131 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_CPUTask/AddTaskAlways-8                 563931             17725 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_CPUTask/AddTaskBalance-8                204085             17720 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_IOTask/AddTask-8                         12427            106108 ns/op               0 B/op          0 allocs/op
BenchmarkTaskPool_IOTask/AddTaskAlways-8                 2607068               504 ns/op              81 B/op          1 allocs/op
BenchmarkTaskPool_IOTask/AddTaskBalance-8                2065213               580 ns/op              63 B/op          0 allocs/op
BenchmarkTaskPool_RandomTask/AddTask-8                    590595              2274 ns/op               6 B/op          0 allocs/op
BenchmarkTaskPool_RandomTask/AddTaskAlways-8             3565921               333 ns/op              21 B/op          0 allocs/op
BenchmarkTaskPool_RandomTask/AddTaskBalance-8            1487217               839 ns/op              17 B/op          0 allocs/op
PASS


*/
