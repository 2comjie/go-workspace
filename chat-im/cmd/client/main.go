package main

import (
	"chat-im/req_rsp"
	"hutool/logx"
	"hutool/syncx"
	"server/pkg/codec"
	"server/pkg/net/tcp"
	"time"
)

func main() {
	cl := tcp.NewClient(codec.JsonSerializer{}, 3*time.Second)
	_ = cl.Dial("127.0.0.1", 8080)
	_ = tcp.Ask[*req_rsp.LoginReq, *req_rsp.LoginRsp](cl, 0, req_rsp.Login, &req_rsp.LoginReq{
		Uid: 1,
	}, func(rsp *req_rsp.LoginRsp) {
		logx.Infof("登陆成功")
	})
	syncx.WaitUntilSignaled()
}
