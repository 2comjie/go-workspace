package fsm

import (
	"hutool/generic"
	"time"
)

type IState[A any, T generic.Primitive] interface {
	TotalDuration(arg A) time.Duration
	PreEnter(oldState T, arg A, isTimeout bool)
	PostEnter(oldState T, arg A, isTimeout bool)
	PreExit(newState T, arg A, isTimeout bool)
	PostExit(newState T, arg A, isTimeout bool)
	Update(arg A, dt time.Duration)
	OnEventTrigger(arg A, event any) (T, bool, any)
	GetNextStateOnTimeout(arg A) T
}

type registry[A any, T generic.Primitive] struct {
	defaultState T
	states       map[T]IState[A, T]
	preSwitch    SwitchFunc[A, T]
	postSwitch   SwitchFunc[A, T]
}

type SwitchFunc[A any, T generic.Primitive] func(oldState T, newState T, arg A, isTimeout bool)

type FSM[A any, T generic.Primitive] struct {
	registry[A, T]
	curState      T
	leftDuration  time.Duration
	totalDuration time.Duration
	lastState     T
}

func (f *FSM[A, T]) Register(state T, info IState[A, T]) {
	f.states[state] = info
}

func (f *FSM[A, T]) Update(arg A, dt time.Duration) {
	curStateInfo := f.states[f.curState]
	if f.totalDuration >= 0 {
		f.leftDuration -= dt
		if f.leftDuration <= 0 {
			f.switchState(curStateInfo.GetNextStateOnTimeout(arg), arg, true)
		}
	}
	curStateInfo.Update(arg, dt)
}

func (f *FSM[A, T]) Trigger(arg A, event any) any {
	curStateInfo := f.states[f.curState]
	nextState, shouldSwitch, out := curStateInfo.OnEventTrigger(arg, event)
	if shouldSwitch {
		f.switchState(nextState, arg, false)
	}
	return out
}

func (f *FSM[A, T]) switchState(newState T, arg A, isTimeout bool) {
	oldState := f.curState
	newStateInfo := f.states[newState]
	oldStateInfo := f.states[oldState]
	if f.preSwitch != nil {
		f.preSwitch(oldState, newState, arg, isTimeout)
	}
	oldStateInfo.PreExit(newState, arg, isTimeout)
	newStateInfo.PreEnter(oldState, arg, isTimeout)

	f.lastState = oldState
	f.curState = newState
	f.totalDuration = newStateInfo.TotalDuration(arg)
	f.leftDuration = f.totalDuration

	oldStateInfo.PostExit(newState, arg, isTimeout)
	newStateInfo.PostEnter(oldState, arg, isTimeout)

	if f.postSwitch != nil {
		f.postSwitch(oldState, newState, arg, isTimeout)
	}
}

func (f *FSM[A, T]) CurState() T {
	return f.curState
}

func (f *FSM[A, T]) LeftDuration() time.Duration {
	return f.leftDuration
}

func (f *FSM[A, T]) LastState() T {
	return f.lastState
}
