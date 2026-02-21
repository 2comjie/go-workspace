package redis_impl

import (
	"context"
	"encoding/json"
	"errors"
	"hutool/convert"
	"hutool/reflectx"
	"strings"
	"sync-data/pkg/sync_def"
	"time"

	"github.com/redis/go-redis/v9"
)

type Field struct {
	Name string
	Data any
}

type BaseRedisSyncHandler[T any] struct {
	*sync_def.RedisConfig[T]
}

func NewBaseRedisSyncHandler[T any](rc redis.UniversalClient, option sync_def.RedisOption[T]) *BaseRedisSyncHandler[T] {
	return &BaseRedisSyncHandler[T]{
		&sync_def.RedisConfig[T]{
			BaseConfig: &sync_def.BaseConfig[T]{
				Config: sync_def.BuildFieldConfig[T](),
				Coder:  option.Coder,
			},
			DataRedisPrefix: option.DataRedisPrefix,
			Rc:              rc,
			ExpireDuration:  option.ExpireDuration,
		},
	}
}

func (b *BaseRedisSyncHandler[T]) LoadOne(key *T) (*T, error) {
	redisKey := b.GetRedisKey(key)
	str, err := b.Rc.Get(context.Background(), redisKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	if str == "" {
		return nil, nil
	}
	if b.Coder != nil {
		ret, err := b.Coder.Decode(str)
		if err != nil {
			return nil, err
		}
		return ret, nil
	}

	var ret = new(T)
	err = json.Unmarshal([]byte(str), ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (b *BaseRedisSyncHandler[T]) SaveOne(data *T) error {
	redisKey := b.GetRedisKey(data)
	if b.Coder != nil {
		dataStr, err := b.Coder.Encode(data)
		if err != nil {
			return err
		}
		return b.Rc.Set(context.Background(), redisKey, dataStr, b.ExpireDuration).Err()
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return b.Rc.Set(context.Background(), redisKey, jsonStr, b.ExpireDuration).Err()
}

func (b *BaseRedisSyncHandler[T]) DelOne(key *T) error {
	redisKey := b.GetRedisKey(key)
	return b.Rc.Del(context.Background(), redisKey).Err()
}

func (b *BaseRedisSyncHandler[T]) Expire(key *T, duration time.Duration) error {
	redisKey := b.GetRedisKey(key)
	if duration == time.Duration(-1) {
		return b.Rc.Persist(context.Background(), redisKey).Err()
	}
	return b.Rc.Expire(context.Background(), redisKey, duration).Err()
}

func (b *BaseRedisSyncHandler[T]) GetRedisKey(data *T) string {
	keyList := make([]string, 0, len(b.Config.CacheKeyFields))
	rv := reflectx.IndirectValue(data)
	for _, fieldCf := range b.Config.CacheKeyFields {
		fieldV := rv.FieldByName(fieldCf.GoFieldName).Interface()
		keyList = append(keyList, convert.String(fieldV))
	}

	bd := &strings.Builder{}
	bd.WriteString(b.DataRedisPrefix)
	for index, key := range keyList {
		if index != 0 {
			bd.WriteString(":")
		}
		bd.WriteString(key)
	}
	return bd.String()
}

func (b *BaseRedisSyncHandler[T]) EncodePrimaryKey(data *T) (string, error) {
	fields := make([]Field, 0, len(b.Config.PrimaryFields))
	rv := reflectx.IndirectValue(data)
	for _, fieldCf := range b.Config.PrimaryFields {
		fieldV := rv.FieldByName(fieldCf.GoFieldName).Interface()
		fields = append(fields, Field{
			Name: fieldCf.GoFieldName,
			Data: fieldV,
		})
	}

	jsonStr, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	return string(jsonStr), nil
}

func (b *BaseRedisSyncHandler[T]) DecodePrimaryKey(jsonStr string) (*T, error) {
	var fields []Field
	err := json.Unmarshal([]byte(jsonStr), &fields)
	if err != nil {
		return nil, err
	}

	var ret = new(T)
	rv := reflectx.IndirectValue(ret)

	for _, field := range fields {
		_, ok := b.Config.Fields[field.Name]
		if !ok {
			continue
		}
		reflectx.ConvertAndSet(field.Data, rv.FieldByName(field.Name))
	}

	return ret, nil
}
