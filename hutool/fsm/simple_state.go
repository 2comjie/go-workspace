package fsm

import (
	"hutool/generic"
	"time"
)

type SimpleState[A any, T generic.Primitive] struct {
	MTotalDuration         time.Duration
	MPreEnter              func(oldState T, arg A, isTimeout bool)
	MPostEnter             func(oldState, newState T, arg A, isTimeout bool)
	MPreExit               func(newState T, arg A, isTimeout bool)
	MPostExit              func(oldState, newState T, arg A, isTimeout bool)
	MUpdate                func(arg A, dt time.Duration)
	MOnEventTrigger        func(arg A, event any) (T, bool, any)
	MGetNextStateOnTimeout func(arg A) T
}
