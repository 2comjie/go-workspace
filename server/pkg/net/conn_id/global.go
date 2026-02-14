package conn_id

import "sync/atomic"

var globalId atomic.Uint32

func NextId() uint32 {
	return globalId.Add(1)
}
