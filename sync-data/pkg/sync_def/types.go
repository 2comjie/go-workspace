package sync_def

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type MemberCoder[T any] interface {
	Encode(data *T) (string, error)
	Decode(data string) (*T, error)
}

type DbHandler[T any] interface {
	LoadOne(key *T) (*T, error)
	SaveOne(key *T) error
	DelOne(key *T) error
	LoadBatch(keys []*T) ([]*T, error)
}

type RedisHandler[T any] interface {
	LoadOne(key *T) (*T, error)
	SaveOne(key *T) error
	DelOne(key *T) error
	Expire(key *T, duration time.Duration) error
}

type Marker[T any] interface {
	MarkUpdate(key *T) error
	DelUpdate(key *T) error
	FetchUpdateList(max int64) ([]*T, error)
}

type FlushConfig struct {
	FlushInterval time.Duration // 刷新的周期
	Batch         int64         // 批量刷新大小
	ExpireTime    time.Duration // 过期时间
}

type BaseConfig[T any] struct {
	Config *DataConfig    // 数据字段配置
	Coder  MemberCoder[T] // 编码器
}

type RedisConfig[T any] struct {
	*BaseConfig[T]
	DataRedisPrefix string // redis前缀
	ExpireDuration  time.Duration
	Rc              redis.UniversalClient
}

type DbConfig[T any] struct {
	*BaseConfig[T]
	TableName string // 数据库表名字
}

type BaseOption[T any] struct {
	Coder MemberCoder[T] // 编码器
}

type RedisOption[T any] struct {
	BaseOption[T]
	DataRedisPrefix string // redis前缀
	ExpireDuration  time.Duration
}

type DbOption[T any] struct {
	BaseOption[T]
	TableName string
}

type SyncOption[T any] struct {
	BaseOption[T]
	RedisLockKeyPrefix string
}
