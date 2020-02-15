package gxsync

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
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

func BenchmarkTaskPool_CPUTask(b *testing.B) {
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
