package actor

import (
	"context"
	"hutool/chanx"
	"hutool/taskx"
	"sync"
	"sync/atomic"
	"time"
)

type IActor interface {
	Start()
	Update(dt time.Duration)
	Stop()
}

type Runner[T IActor] struct {
	actor      T
	msgChan    chan func(T)
	dt         time.Duration
	closed     atomic.Bool
	canDropMsg bool
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewRunner[T IActor](actor T, opts ...Option) *Runner[T] {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())
	r := &Runner[T]{
		actor:      actor,
		msgChan:    make(chan func(T), cfg.msgChanLen),
		dt:         cfg.dt,
		closed:     atomic.Bool{},
		canDropMsg: cfg.canDropMsg,
		ctx:        ctx,
		cancel:     cancel,
		wg:         sync.WaitGroup{},
	}
	actor.Start()
	go r.work()
	return r
}

func (r *Runner[T]) Stop() {
	if r.closed.CompareAndSwap(false, true) {
		r.cancel()
		r.wg.Wait()
		r.actor.Stop()
	}
}

func (r *Runner[T]) tell(msg func(T)) error {
	if r.closed.Load() {
		return ErrMsgActorClosed
	}
	select {
	case r.msgChan <- msg:
		return nil
	default:
		return ErrMsgChanFull
	}
}

func (r *Runner[T]) ask(msg func(T) any, timeout ...time.Duration) (any, error) {
	if r.closed.Load() {
		return nil, ErrMsgActorClosed
	}
	promise := taskx.NewPromise[any]()
	wrapper := func(a T) {
		res := msg(a)
		promise.OnTaskDone(res)
	}
	select {
	case r.msgChan <- wrapper:
		if len(timeout) == 1 {
			res := promise.WaitWithTimeout(timeout[0])
			return res, nil
		} else {
			res := promise.Wait()
			return res, nil
		}
	default:
		return nil, ErrMsgChanFull
	}
}

func (r *Runner[T]) work() {
	defer r.wg.Done()
	tk := time.NewTicker(r.dt)
	for {
		select {
		case <-r.ctx.Done():
			if !r.canDropMsg {
				msgs := chanx.DrainNow[func(T)](r.msgChan)
				for _, msg := range msgs {
					msg(r.actor)
				}
			}
			return
		case <-tk.C:
			r.actor.Update(r.dt)
		case msg := <-r.msgChan:
			msg(r.actor)
		}
	}
}
