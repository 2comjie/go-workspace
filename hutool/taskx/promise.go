package taskx

import (
	"context"
	"sync/atomic"
	"time"
)

type Promise[T any] struct {
	res       T
	resChan   chan T
	ctx       context.Context
	cancel    context.CancelFunc
	completed atomic.Bool
}

func NewPromise[T any]() *Promise[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &Promise[T]{
		ctx:       ctx,
		cancel:    cancel,
		resChan:   make(chan T, 1),
		completed: atomic.Bool{},
	}
}

func (p *Promise[T]) OnTaskDone(r T) {
	p.resChan <- r
}

func (p *Promise[T]) Cancel() {
	p.cancel()
}

func (p *Promise[T]) Wait() T {
	if p.completed.CompareAndSwap(false, true) {
		res := p.wait()
		p.res = res
	}
	return p.res
}

func (p *Promise[T]) WaitWithTimeout(timeout time.Duration) T {
	if p.completed.CompareAndSwap(false, true) {
		res := p.waitWithTimeout(timeout)
		p.res = res
	}
	return p.res
}

func (p *Promise[T]) wait() T {
	select {
	case <-p.ctx.Done():
		var zero T
		return zero
	case res := <-p.resChan:
		return res
	}
}

func (p *Promise[T]) waitWithTimeout(timeout time.Duration) T {
	tk := time.NewTicker(timeout)
	var zero T
	select {
	case <-p.ctx.Done():
		return zero
	case <-tk.C:
		return zero
	case <-p.resChan:
		return p.res
	}
}
