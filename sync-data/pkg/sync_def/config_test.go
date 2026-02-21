package sync_def

import (
	"hutool/logx"
	"testing"
)

type TestA struct {
	ID    int64  `sync:"primary=1,cache=1"`
	Name  string `sync:"primary=2,cache=2"`
	Age   int32
	Score int64 `sync:"isScore=true"`
	Seq   int64 `sync:"isSeq=true"`
}

func TestConfig(t *testing.T) {
	dbCfg := BuildFieldConfig[TestA]()
	logx.Infof("%+v", dbCfg)
}
