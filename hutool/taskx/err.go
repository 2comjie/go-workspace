package taskx

import "errors"

var TaskChanFullErr = errors.New("task chan is full")

var TaskPoolClosedErr = errors.New("task pool is closed")
