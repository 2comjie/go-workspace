package taskx

import "runtime"

type TaskPoolConfig struct {
	workerNum   int
	queueSize   int
	canDropTask bool
}

type TaskPoolOption func(*TaskPoolConfig)

func WithWorkerNum(workerNum int) TaskPoolOption {
	return func(config *TaskPoolConfig) {
		config.workerNum = workerNum
	}
}

func WithQueueSize(queueSize int) TaskPoolOption {
	return func(config *TaskPoolConfig) {
		config.queueSize = queueSize
	}
}

func WithCanDropTask(canDropTask bool) TaskPoolOption {
	return func(config *TaskPoolConfig) {
		config.canDropTask = canDropTask
	}
}

func DefaultTaskPoolConfig() *TaskPoolConfig {
	return &TaskPoolConfig{
		workerNum:   runtime.NumCPU(),
		queueSize:   100,
		canDropTask: false,
	}
}
