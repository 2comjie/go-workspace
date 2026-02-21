package mark

import (
	"context"
	"hutool/convert"
	"hutool/reflectx"
	"reflect"
	"strings"
	"sync-data/pkg/sync_def"

	"github.com/redis/go-redis/v9"
)

type SimpleMarker[T any] struct {
	*sync_def.BaseConfig[T]
	UpdateSetRedisKey string
	Rc                redis.UniversalClient
}

func (m *SimpleMarker[T]) MarkUpdate(key *T) error {
	markValue := m.buildMarkValue(key)
	err := m.Rc.SAdd(context.Background(), m.UpdateSetRedisKey, markValue).Err()
	if err != nil {
		return err
	}
	return nil
}

func (m *SimpleMarker[T]) buildMarkValue(key *T) string {
	sb := &strings.Builder{}
	rv := reflectx.IndirectValue(key)
	sb.WriteString("{")
	for index, value := range m.Config.CacheKeyFields {
		if index != 0 {
			sb.WriteString(":")
		}
		fieldV := rv.FieldByName(value.GoFieldName).Interface()
		sb.WriteString(convert.String(fieldV))
	}
	sb.WriteString("}")
	return sb.String()
}

func (m *SimpleMarker[T]) DelUpdate(key *T) error {
	markValue := m.buildMarkValue(key)
	err := m.Rc.SRem(context.Background(), m.UpdateSetRedisKey, markValue).Err()
	if err != nil {
		return err
	}
	return nil
}

func (m *SimpleMarker[T]) FetchUpdateList(max int64) ([]*T, error) {
	keys, err := m.Rc.SRandMemberN(context.Background(), m.UpdateSetRedisKey, max).Result()
	if err != nil {
		return nil, err
	}
	ret := make([]*T, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimPrefix(key, "{")
		key = strings.TrimSuffix(key, "}")
		ss := strings.Split(key, ":")
		if len(ss) != len(m.Config.CacheKeyFields) {
			continue
		}

		oneData := new(T)
		rv := reflect.Indirect(reflect.ValueOf(oneData))
		for index, fieldCf := range m.Config.CacheKeyFields {
			stringV := ss[index]
			fieldValue := rv.FieldByName(fieldCf.GoFieldName)
			success := reflectx.ConvertAndSet(stringV, fieldValue)
			if !success {
				continue
			}
		}
		ret = append(ret, oneData)
	}
	return ret, nil
}
