package lock

import (
	"hutool/convert"
	"hutool/redisx"
	"hutool/reflectx"
	"strings"
	"sync-data/pkg/sync_def"

	"github.com/redis/go-redis/v9"
)

func GenLock[T any](rc redis.UniversalClient, key *T, redisPrefix string, dataConfig *sync_def.DataConfig) *redisx.Lock {
	lock := redisx.NewLock(BuildLockKey[T](key, redisPrefix, dataConfig), rc)
	return lock
}

func BuildLockKey[T any](key *T, prefix string, dataConfig *sync_def.DataConfig) string {
	sb := &strings.Builder{}
	sb.WriteString(prefix)
	sb.WriteString(":")
	rv := reflectx.IndirectValue(key)
	for index, fieldConfig := range dataConfig.CacheKeyFields {
		if index != 0 {
			sb.WriteString(":")
		}
		fieldV := rv.FieldByName(fieldConfig.GoFieldName)
		sb.WriteString(convert.String(fieldV))
	}
	return sb.String()
}
