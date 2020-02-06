package gxsync

import "context"

type Task interface{}
type FuncTask func()
type CtxFuncTask func(context.Context)
type StructTask struct{}

type TaskQueue interface {
	Push(Task)
	//Pop() Task
}

type Runner interface {
	Start()
	Wait()
	Stop()
}

type Worker interface {
	TaskQueue
	Runner
}

type WorkerManager interface {
	TaskQueue
	Runner
}

// PoolManager
type PoolManager struct{}

// PoolPreferManager is PoolManager with do task in new goroutine when pool is full.
type PoolPreferManager struct{}
