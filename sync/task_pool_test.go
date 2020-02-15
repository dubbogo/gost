package gxsync

import (
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
	taskCnt := int64(numCPU * numCPU * 10)

	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(numCPU/2),
		WithTaskPoolTaskQueueNumber(numCPU),
		WithTaskPoolTaskQueueLength(1),
	)

	task, cnt := newCountTask()

	var wg sync.WaitGroup
	for i := 0; i < numCPU*numCPU; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 10; j++ {
				tp.AddTask(task)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	tp.Close()

	if taskCnt != *cnt {
		t.Error("want ", taskCnt, " got ", *cnt)
	}
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
