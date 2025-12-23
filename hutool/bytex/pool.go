package bytex

import (
	"hutool/mathx"
	"sync"
)

var pools []*sync.Pool

const MaxIndex = 20

func init() {
	pools = make([]*sync.Pool, MaxIndex+1)
	for i := 0; i <= MaxIndex; i++ {
		pools[i] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, 1<<i)
			},
		}
	}
}

func index(c int) int {
	return mathx.FastLog2(c)
}

func Allocate(c int) []byte {
	idx := index(c)
	if idx > MaxIndex {
		return make([]byte, c)
	}
	return pools[idx].Get().([]byte)[:c]
}

func Return(bytes []byte) {
	idx := index(cap(bytes))
	if idx > MaxIndex {
		return
	}
	pools[idx].Put(bytes)
}
