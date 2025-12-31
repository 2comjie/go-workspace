package redisx

import (
	"context"
	"errors"
	"hutool/logx"
	"hutool/safe"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrLockTimeout = errors.New("lock timeout")

type Lock struct {
	rc              redis.UniversalClient
	key             string
	ctx             context.Context
	cancel          context.CancelFunc
	lockValue       string
	maxTryTime      time.Duration //重试时间
	retryInterval   time.Duration //最小重试间隔
	extendValidTime time.Duration //锁失效时间
	extendInterval  time.Duration //刷新锁间隔
	wg              sync.WaitGroup
}

func NewLock(key string, rc redis.UniversalClient, opts ...LockOption) *Lock {
	cfg := DefaultLockConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Lock{
		rc:              rc,
		key:             key,
		ctx:             ctx,
		cancel:          cancel,
		lockValue:       "",
		maxTryTime:      cfg.maxTryTime,
		retryInterval:   cfg.retryInterval,
		extendValidTime: cfg.extendValidTime,
		extendInterval:  cfg.extendInterval,
		wg:              sync.WaitGroup{},
	}
}

func (r *Lock) TryLock() (bool, error) {
	success, err := r.tryLock()
	if success {
		return true, nil
	} else {
		if err != nil {
			return false, err
		}
		return false, nil
	}
}

func (r *Lock) Lock() error {
	success, _ := r.tryLock()
	if success {
		return nil
	}
	ticker := time.NewTicker(r.retryInterval)
	defer ticker.Stop()
	lockCtx, lockCancel := context.WithTimeout(r.ctx, r.maxTryTime)
	defer lockCancel()
	for {
		select {
		case <-lockCtx.Done():
			return ErrLockTimeout
		case <-ticker.C:
			success, err := r.tryLock()
			if success {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}
}

func (r *Lock) Unlock() {
	if r.lockValue == "" {
		return
	}
	r.cancel()
	r.wg.Wait()
	_, err := r.rc.Eval(r.ctx, `
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("DEL", KEYS[1])
    else
        return 0
    end`, []string{r.key}, r.lockValue).Result()
	if err != nil {
		logx.Warnf("unlock %s failed: %v", r.key, err)
	}
}

func (r *Lock) tryLock() (bool, error) {
	if r.lockValue != "" {
		return false, nil
	}
	lockValue := uuid.New().String()
	ok, err := r.rc.SetNX(r.ctx, r.key, lockValue, r.extendValidTime).Result()
	if ok {
		r.lockValue = lockValue
		r.wg.Add(1)
		go safe.Run(func() {
			r.extendBackground()
		})
		return true, nil
	}
	return false, err
}

func (r *Lock) extendBackground() {
	defer r.wg.Done()
	defer r.cancel()
	extendTicker := time.NewTicker(r.extendInterval)
	defer extendTicker.Stop()
	for {
		select {
		case <-extendTicker.C:
			err := r.extendOnce()
			if err != nil {
				logx.Warnf("extend lock %s failed: %v", r.key, err)
			}
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Lock) extendOnce() error {
	_, err := r.rc.Eval(r.ctx, `
local currentValue=redis.Call("GET", KEYS[1])
if currentValue == ARGV[1] then
    return redis.Call("pexpire", KEYS[1], ARGV[2])
else
    if currentValue==false then
        return redis.error_reply("key not found")
    end
    local errMsg="current value not match,saved " .. currentValue .. " input " .. ARGV[1]
    return redis.error_reply(errMsg)
end`, []string{r.key}, r.lockValue, r.extendValidTime).Result()
	return err
}
