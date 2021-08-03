package gxsync

import (
	gxruntime "github.com/dubbogo/gost/runtime"
	perrors "github.com/pkg/errors"
)

var (
	PoolBusyErr = perrors.New("pool is busy")
)

func NewWorkerPool(maxWorkers, taskQueueSize int) (p *WorkerPool) {
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	if taskQueueSize < 0 {
		taskQueueSize = 0
	}

	p = &WorkerPool{
		maxWorkers:  maxWorkers,
		workerQueue: make(chan task),
		taskQueue:   make(chan task, taskQueueSize),
		done:        make(chan struct{}),
	}

	gxruntime.GoSafely(nil, false, p.dispatch, nil)

	return
}

type WorkerPool struct {
	maxWorkers int

	workerQueue chan task
	taskQueue   chan task

	done chan struct{}
}

func (p *WorkerPool) dispatch() {
	defer close(p.done)

	var workerCount int

loop:
	for {
		select {
		case t, ok := <-p.taskQueue:
			if !ok {
				break loop
			}
			// select a worker to execute the task
			select {
			case p.workerQueue <-t:
			default:
				if workerCount < p.maxWorkers {
					// number of workers not reaches the limitation
					gxruntime.GoSafely(nil, false, func() {
						newWorker(t, p.workerQueue)
					}, nil)
					workerCount++
				} else {
					// blocked and waiting for a worker
					select {
					case p.workerQueue <-t:
					}
				}
			}
		}
	}

	// wait for the end of all tasks, and shutting down workers
	for workerCount > 0 {
		p.workerQueue <-nil
		workerCount--
	}

}

// Submit adds a task to queue asynchronously.
func (p *WorkerPool) Submit(t task) error {
	if t == nil {
		return perrors.New("task shouldn't be nil")
	}
	select {
	case p.taskQueue <- t:
		return nil
	default:
		return PoolBusyErr
	}
}

// SubmitSync adds a task to queue synchronously.
func (p *WorkerPool) SubmitSync(t task) error {
	if t == nil {
		return perrors.New("task shouldn't be nil")
	}

	done := make(chan struct{})
	fn := func() {
		t()
		close(done)
	}
	select {
	case p.taskQueue <- fn:
		<-done
		return nil
	default:
		return PoolBusyErr
	}
}

func (p *WorkerPool) Close() {
	select {
	case <-p.done:
		return
	default:
	}

	close(p.taskQueue)
	<-p.done
}

func (p *WorkerPool) IsClosed() bool {
	select {
	case <-p.done:
		return true
	default:
	}
	return false
}

func newWorker(t task, workerQueue chan task) {
	gxruntime.GoSafely(nil, false, t, nil)
	gxruntime.GoSafely(nil, false, func() {
		worker(workerQueue)
	}, nil)

}

func worker(workerQueue chan task) {
	for task := range workerQueue {
		if task == nil {
			return
		}
		task()
	}
}
