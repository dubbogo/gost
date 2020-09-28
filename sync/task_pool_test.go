package gxsync

import (
	"github.com/stretchr/testify/assert"
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

func newCountTaskAndRespite() (func(), *int64) {
	var cnt int64
	return func() {
		atomic.AddInt64(&cnt, 1)
		time.Sleep(10 * time.Millisecond)
	}, &cnt
}

func TestTaskPool(t *testing.T) {
	numCPU := runtime.NumCPU()
	taskCnt := int64(numCPU * numCPU * 100)

	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(1),
		WithTaskPoolTaskQueueNumber(1),
		WithTaskPoolTaskQueueLength(1),
	)

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
	tp.Close(true)

	if taskCnt != *cnt {
		t.Error("want ", taskCnt, " got ", *cnt)
	}
}

func TestTaskPool_Close(t *testing.T) {
	numCPU := runtime.NumCPU()
	taskCnt := int64(numCPU * 100)

	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(1),
		WithTaskPoolTaskQueueNumber(5),
		WithTaskPoolTaskQueueLength(1),
	)

	task, cnt := newCountTaskAndRespite()

	var wg sync.WaitGroup
	for i := 0; i < numCPU; i++ {
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
	// close immediately, so taskCnt not equal *cnt
	tp.Close(true)

	assert.NotEqual(t, taskCnt, *cnt)
}

func TestTaskPool_CloseTillTaskComplete(t *testing.T) {
	numCPU := runtime.NumCPU()
	taskCnt := int64(numCPU * 100)

	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(1),
		WithTaskPoolTaskQueueNumber(5),
		WithTaskPoolTaskQueueLength(1),
	)

	task, cnt := newCountTaskAndRespite()

	var wg sync.WaitGroup
	for i := 0; i < numCPU; i++ {
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
	// wait till all task completed, so taskCnt should equal *cnt
	tp.Close(false)

	assert.Equal(t, taskCnt, *cnt)
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

func fib(n int) int {
	if n < 3 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}

// cpu-intensive task
func BenchmarkTaskPool_CPUTask(b *testing.B) {
	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(runtime.NumCPU()),
		WithTaskPoolTaskQueueNumber(runtime.NumCPU()),
		//WithTaskPoolTaskQueueLength(runtime.NumCPU()),
	)

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

	b.Run(`AddTaskBalance`, func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				task, _ := newRandomTask()
				tp.AddTaskBalance(task)
			}
		})
	})
}

/*

pkg: github.com/dubbogo/gost/sync
BenchmarkTaskPool_CountTask/AddTask-8         	 2872177	       380 ns/op	       0 B/op	       0 allocs/op
BenchmarkTaskPool_CountTask/AddTaskAlways-8   	 2769730	       455 ns/op	       1 B/op	       0 allocs/op
BenchmarkTaskPool_CountTask/AddTaskBalance-8  	 4630167	       248 ns/op	       0 B/op	       0 allocs/op
BenchmarkTaskPool_CPUTask/fib-8               	   73975	     16524 ns/op	       0 B/op	       0 allocs/op
BenchmarkTaskPool_CPUTask/AddTask-8           	   72525	     18160 ns/op	       0 B/op	       0 allocs/op
BenchmarkTaskPool_CPUTask/AddTaskAlways-8     	  606813	     16464 ns/op	      40 B/op	       0 allocs/op
BenchmarkTaskPool_CPUTask/AddTaskBalance-8    	  137926	     17646 ns/op	       0 B/op	       0 allocs/op
BenchmarkTaskPool_IOTask/AddTask-8            	   10000	    108520 ns/op	       0 B/op	       0 allocs/op
BenchmarkTaskPool_IOTask/AddTaskAlways-8      	 1000000	      1236 ns/op	      95 B/op	       1 allocs/op
BenchmarkTaskPool_IOTask/AddTaskBalance-8     	 1518144	       673 ns/op	      63 B/op	       0 allocs/op
BenchmarkTaskPool_RandomTask/AddTask-8        	  497055	      2517 ns/op	       6 B/op	       0 allocs/op
BenchmarkTaskPool_RandomTask/AddTaskAlways-8  	 2511391	       415 ns/op	      21 B/op	       0 allocs/op
BenchmarkTaskPool_RandomTask/AddTaskBalance-8 	 1381711	       868 ns/op	      17 B/op	       0 allocs/op
PASS

*/
