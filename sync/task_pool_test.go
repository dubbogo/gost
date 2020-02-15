package gxsync

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func TestTaskPool(t *testing.T) {
	numCPU := runtime.NumCPU()
	taskCnt := numCPU * numCPU * 1000

	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(numCPU),
		WithTaskPoolTaskQueueNumber(numCPU),
		WithTaskPoolTaskQueueLength(numCPU),
	)

	var cnt int64 = 0
	task := func() {
		atomic.AddInt64(&cnt, 1)
	}

	var wg sync.WaitGroup
	for i := 0; i < numCPU*numCPU; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 1000; i++ {
				tp.AddTask(task)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	tp.Close()

	if taskCnt != int(cnt) {
		t.Error("want ", taskCnt, " got ", cnt)
	}
}

func BenchmarkTaskPool_CPUTask(b *testing.B) {
	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(runtime.NumCPU()),
		WithTaskPoolTaskQueueNumber(runtime.NumCPU()),
		WithTaskPoolTaskQueueLength(runtime.NumCPU()),
	)

	var cnt int64 = 0
	task := func() {
		atomic.AddInt64(&cnt, 1)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tp.AddTask(task)
		}
	})
}
