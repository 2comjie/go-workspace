package taskx

import (
	"hutool/logx"
	"testing"
	"time"
)

func TestTask(t *testing.T) {
	pool := NewTaskPool[int]()
	promise := NewPromise[int]()

	_ = pool.Add(func() int {
		time.Sleep(time.Second * 4)
		return 100
	}, promise)
	res := promise.Wait()
	logx.Infof("%d", res)
}
