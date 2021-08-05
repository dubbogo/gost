package gxsync

import (
	gxlog "github.com/dubbogo/gost/log"
	"runtime/debug"
	"sync"
)

type baseWorkerPool struct {
	logger gxlog.Logger

	taskId     uint32
	taskQueues []chan task

	wg   *sync.WaitGroup
	done chan struct{}
}

func (p *baseWorkerPool) Submit(t task) error {
	panic("implement me")
}

func (p *baseWorkerPool) SubmitSync(t task) error {
	panic("implement me")
}

func (p *baseWorkerPool) Close() {
	if p.IsClosed() {
		return
	}

	close(p.done)
	for _, q := range p.taskQueues {
		close(q)
	}
	p.wg.Wait()
}

func (p *baseWorkerPool) IsClosed() bool {
	select {
	case <-p.done:
		return true
	default:
	}
	return false
}

func newWorker(q chan task, logger gxlog.Logger, workerId int, wg *sync.WaitGroup) {
	defer wg.Done()

	if logger != nil {
		logger.Debugf("worker #%d is started\n", workerId)
	}
	for {
		select {
		case task, ok := <-q:
			if !ok {
				if logger != nil {
					logger.Debugf("worker #%d is closed\n", workerId)
				}
				return
			}
			if task != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							if logger != nil {
								logger.Errorf("goroutine panic: %v\n%s\n", r, string(debug.Stack()))
							}
						}
					}()
					task()
				}()
			}
		}
	}
}
