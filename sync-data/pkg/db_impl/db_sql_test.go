package db_impl

import (
	"sync-data/pkg/sync_def"
	"testing"

	"github.com/jmoiron/sqlx"
)

type TestA struct {
	ID    int64  `sync:"primary=1,cache=1"`
	Name  string `sync:"primary=2,cache=2"`
	Age   int32
	Score int64 `sync:"isScore=true"`
	Seq   int64 `sync:"isSeq=true"`
}

func TestSql(t *testing.T) {
	_ = NewMysqlHandler[TestA](sqlx.DB{}, sync_def.DbOption[TestA]{})
}
