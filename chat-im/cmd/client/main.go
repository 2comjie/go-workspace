package main

import (
	"hutool/syncx"
	"server/pkg/codec"
	"server/pkg/net/tcp"
	"time"
)

func main() {
	cl := tcp.NewClient(codec.JsonSerializer{}, 3*time.Second)
	_ = cl.Dial("127.0.0.1", 8080)

	// 1. 打印登陆页面

	syncx.WaitUntilSignaled()
}
