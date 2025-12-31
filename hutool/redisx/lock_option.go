package redisx

import "time"

type LockConfig struct {
	maxTryTime      time.Duration //重试时间
	retryInterval   time.Duration //最小重试间隔
	extendValidTime time.Duration //锁失效时间
	extendInterval  time.Duration //刷新锁间隔
}

type LockOption func(config *LockConfig)

func WithLockMaxTryTime(maxTryTime time.Duration) LockOption {
	return func(config *LockConfig) {
		config.maxTryTime = maxTryTime
	}
}

func WithLockRetryInterval(retryInterval time.Duration) LockOption {
	return func(config *LockConfig) {
		config.retryInterval = retryInterval
	}
}

func WithLockExtendInterval(extendInterval time.Duration) LockOption {
	return func(config *LockConfig) {
		config.extendInterval = extendInterval
	}
}

func WithLockExtendValidTime(extendValidTime time.Duration) LockOption {
	return func(config *LockConfig) {
		config.extendValidTime = extendValidTime
	}
}

func DefaultLockConfig() *LockConfig {
	return &LockConfig{
		maxTryTime:      time.Second * 10,
		retryInterval:   time.Millisecond * 100,
		extendInterval:  time.Millisecond * 100,
		extendValidTime: time.Second * 10,
	}
}
