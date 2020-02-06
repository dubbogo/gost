package gxsync

import (
	"context"
	"math/rand"
)

type (
	Task        interface{}           // Task is a job for worker handle
	FuncTask    func()                // FuncTask functional task
	CtxFuncTask func(context.Context) // CtxFuncTask is function task with context
)

// TaskQueue accept new job
type TaskQueue interface {
	Push(Task)
	Pop() <-chan Task
}

// Runner todo
type Runner interface {
	Start()
	Wait() // wait all jobs finished
	Stop() // stop runner and discard rest jobs
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

var _ TaskQueue = &ChannelTaskQueue{}

type ChannelTaskQueue struct {
	Queue chan Task
	Size  int64
}

func NewChannelTaskQueue(bufferSize int64) *ChannelTaskQueue {
	return &ChannelTaskQueue{
		Queue: make(chan Task, bufferSize),
	}
}

func (c *ChannelTaskQueue) Push(t Task) {
	c.Queue <- t
}

func (c *ChannelTaskQueue) Pop() (t <-chan Task) {
	return c.Queue
}

var _ Worker = &GoWorker{}

type GoWorker struct {
	*ChannelTaskQueue
	sigStop chan struct{}
	//stopped chan struct{}
}

func (g *GoWorker) Start() {
	go func() {
	LOOP:
		for {
			select {
			case job := <-g.ChannelTaskQueue.Queue:
				job.(FuncTask)()
			case <-g.sigStop:
				break LOOP
			}
		}
		//g.stopped <- struct{}{}
	}()
}

func (g *GoWorker) Wait() {
	//<-g.stopped
}

func (g *GoWorker) Stop() {
	g.sigStop <- struct{}{}
}

func NewWorker(bufferSize int64) Worker {
	return &GoWorker{
		ChannelTaskQueue: NewChannelTaskQueue(bufferSize),
		sigStop:          nil,
	}
}

type ChannelWorkerManager struct {
	TaskQueue // todo 这个地方可能是瓶颈
	workers   []Worker
}

func NewChannelWorkerManager() *ChannelWorkerManager {
	return &ChannelWorkerManager{
		TaskQueue: NewChannelTaskQueue(10),
		workers:   make([]Worker, 10),
	}
}

func (c *ChannelWorkerManager) Start() {
	for i := range c.workers {
		worker := NewWorker(10)
		c.workers[i] = worker

		worker.Start()
	}
}

func (c *ChannelWorkerManager) dispatch() {
	go func() {
		for {
			var t Task
			select {
			case t = <-c.Pop():
			}

			// todo 选一个闲置的 worker
			i := rand.Intn(len(c.workers))
			c.workers[i].Push(t)
		}
	}()
}

func (c *ChannelWorkerManager) Wait() {
	for _, worker := range c.workers {
		worker.Wait()
	}
}

func (c *ChannelWorkerManager) Stop() {
	for _, worker := range c.workers {
		worker.Stop()
	}
}
