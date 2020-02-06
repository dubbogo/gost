package gxsync

import "context"

type (
	Task        interface{}           // Task is a job for worker handle
	FuncTask    func()                // FuncTask functional task
	CtxFuncTask func(context.Context) // CtxFuncTask is function task with context
)

// TaskQueue accept new job
type TaskQueue interface {
	Push(Task)
	//Pop() Task
}

// Runner todo
type Runner interface {
	Start()
	Wait()
	Stop()
}

// Worker can consume jobs one by one
type Worker interface {
	TaskQueue
	Runner
}

// WorkerManager accept jobs and dispatch to workers
type WorkerManager interface {
	TaskQueue
	Runner
}

// PoolManager todo
type PoolManager struct{}

// PoolPreferManager is PoolManager with do task in new goroutine when pool is full.
type PoolPreferManager struct{}
