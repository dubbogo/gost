package gxsync

import (
	"runtime"
	"sync/atomic"
	"testing"
)

func BenchmarkTaskPool_AddTask(b *testing.B) {
	tp := NewTaskPool(
		WithTaskPoolTaskPoolSize(runtime.NumCPU()),
		WithTaskPoolTaskQueueNumber(runtime.NumCPU()),
		// WithTaskPoolTaskQueueLength(1),
	)

	var cnt int64 = 0
	task := func() {
		atomic.AddInt64(&cnt, 1)
		// time.Sleep(1 * time.Microsecond)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tp.AddTask(task)
		}
	})
	// tp.Close()

	b.Log(runtime.NumCPU(), b.N, cnt)
}
