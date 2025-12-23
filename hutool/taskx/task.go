package taskx

import (
	"context"
	"hutool/chanx"
	"sync"
	"sync/atomic"
)

type TaskCallback[T any] interface {
	OnTaskDone(r T)
	Cancel()
}

type Task[T any] struct {
	callback TaskCallback[T]
	fn       func() T
}

type TaskPool[T any] struct {
	nextId      atomic.Uint32
	chans       []chan Task[T]
	canDropTask bool
	workerNum   int
	queueSize   int
	ctx         context.Context
	cancelF     context.CancelFunc
	closed      atomic.Bool
	wg          sync.WaitGroup
}

func NewTaskPool[T any](opts ...TaskPoolOption) *TaskPool[T] {
	cfg := DefaultTaskPoolConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &TaskPool[T]{
		nextId:      atomic.Uint32{},
		ctx:         ctx,
		cancelF:     cancel,
		canDropTask: cfg.canDropTask,
		queueSize:   cfg.queueSize,
		workerNum:   cfg.workerNum,
		closed:      atomic.Bool{},
		wg:          sync.WaitGroup{},
	}

	for i := 0; i < cfg.workerNum; i++ {
		p.chans = append(p.chans, make(chan Task[T], cfg.queueSize))
	}

	for i := 0; i < p.workerNum; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	return p
}

func (t *TaskPool[T]) Add(fn func() T, callback TaskCallback[T], key ...uint32) error {
	if t.closed.Load() {
		return TaskPoolClosedErr
	}

	task := Task[T]{
		fn:       fn,
		callback: callback,
	}
	realKey := t.nextId.Add(1)
	if len(key) == 1 {
		realKey = key[0]
	}
	targetChan := t.chans[realKey%uint32(len(t.chans))]

	if !t.canDropTask {
		targetChan <- task
		return nil
	} else {
		select {
		case targetChan <- task:
			return nil
		default:
			return TaskChanFullErr
		}
	}
}

func (t *TaskPool[T]) worker(id int) {
	targetChan := t.chans[id]
	for {
		select {
		case <-t.ctx.Done():
			if !t.canDropTask {
				tasks := chanx.DrainNow[Task[T]](targetChan)
				for _, task := range tasks {
					t.processOneTask(task)
				}
			}
			return
		case task := <-targetChan:
			t.processOneTask(task)
		}
	}
}

func (t *TaskPool[T]) processOneTask(task Task[T]) {
	r := task.fn()
	if task.callback != nil {
		task.callback.OnTaskDone(r)
	}
}

func (t *TaskPool[T]) Stop() {
	if t.closed.CompareAndSwap(false, true) {
		t.cancelF()
		t.wg.Wait()
	}
}
