package main

import (
	"chat-im/req_rsp"
	"server/pkg/codec"
	"server/pkg/net/tcp"
	"time"
)

type Client struct {
	rawClient *tcp.Client
}

func (cl *Client) Init() {
	cl.rawClient = tcp.NewClient(codec.JsonSerializer{}, 3*time.Second)
	_ = cl.rawClient.Dial("127.0.0.1", 8080)
}

func (cl *Client) SendLoginReq(uid uint32) {
	tcp.Ask[*req_rsp.LoginReq, *req_rsp.LoginRsp](cl.rawClient, 0, req_rsp.Login, &req_rsp.LoginReq{
		Uid: uid,
	}, func(rsp *req_rsp.LoginRsp) {

	})
}
