package sync

import (
	"hutool/logx"
	"hutool/safe"
	lock2 "sync-data/pkg/lock"
	"sync-data/pkg/sync_def"
	"time"

	"github.com/redis/go-redis/v9"
)

type Synchronizer[T any] struct {
	*sync_def.BaseConfig[T]
	redisHandler       sync_def.RedisHandler[T]
	dbHandler          sync_def.DbHandler[T]
	marker             sync_def.Marker[T]
	flushConfigGetter  func() sync_def.FlushConfig
	Rc                 redis.UniversalClient
	RedisLockKeyPrefix string
}

func NewSynchronizer[T any](rc redis.UniversalClient, redisHandler sync_def.RedisHandler[T], dbHandler sync_def.DbHandler[T], marker sync_def.Marker[T], flushConfigGetter func() sync_def.FlushConfig, option sync_def.SyncOption[T]) *Synchronizer[T] {
	sc := &Synchronizer[T]{
		BaseConfig: &sync_def.BaseConfig[T]{
			Config: sync_def.BuildFieldConfig[T](),
			Coder:  option.Coder,
		},
		redisHandler:       redisHandler,
		dbHandler:          dbHandler,
		marker:             marker,
		flushConfigGetter:  flushConfigGetter,
		Rc:                 rc,
		RedisLockKeyPrefix: option.RedisLockKeyPrefix,
	}
	sc.run()
	return sc
}

func (sc *Synchronizer[T]) SaveOne(data *T, needLock bool) error {
	if needLock {
		lock := lock2.GenLock(sc.Rc, data, sc.RedisLockKeyPrefix, sc.Config)
		err := lock.Lock()
		if err != nil {
			return err
		}
		defer lock.Unlock()
	}

	err := sc.redisHandler.SaveOne(data)
	if err != nil {
		return err
	}
	return sc.marker.MarkUpdate(data)
}

func (sc *Synchronizer[T]) LoadOne(key *T, needLock bool) (*T, error) {
	if needLock {
		lock := lock2.GenLock(sc.Rc, key, sc.RedisLockKeyPrefix, sc.Config)
		err := lock.Lock()
		if err != nil {
			return nil, err
		}
		defer lock.Unlock()
	}

	redisData, err := sc.redisHandler.LoadOne(key)
	if err != nil {
		return nil, err
	}

	if redisData == nil {
		redisData, err = sc.dbHandler.LoadOne(key)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	err = sc.redisHandler.Expire(key, sc.flushConfigGetter().ExpireTime)
	if err != nil {
		return nil, err
	}

	return redisData, nil
}

func (sc *Synchronizer[T]) run() {
	go safe.Run(func() {
		flushConfig := sc.flushConfigGetter()
		timer := time.NewTimer(flushConfig.FlushInterval)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				sc.flushRedisToDb()
				timer.Reset(sc.flushConfigGetter().FlushInterval)
			}
		}
	})
}

func (sc *Synchronizer[T]) flushRedisToDb() {
	flushConfig := sc.flushConfigGetter()

	keyList, err := sc.marker.FetchUpdateList(flushConfig.Batch)
	if err != nil {
		logx.Errorf("fetch update list err %+v", err)
		return
	}

	for _, key := range keyList {
		err := sc.flushOneKey(key)
		if err != nil {
			logx.Errorf("flush redis to db err %+v key %v", err, key)
		}
	}
}

func (sc *Synchronizer[T]) flushOneKey(key *T) error {
	lock := lock2.GenLock(sc.Rc, key, sc.RedisLockKeyPrefix, sc.Config)
	err := lock.Lock()
	if err != nil {
		return err
	}
	defer lock.Unlock()

	data, err := sc.redisHandler.LoadOne(key)
	if err != nil {
		return err
	}

	err = sc.dbHandler.SaveOne(data)
	if err != nil {
		return err
	}
	return nil
}
