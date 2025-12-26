package main

import (
	"hutool/logx"
	"hutool/syncx"
	"server/app/test"
	"server/internal/codec"
	"server/internal/net/ws"
	"time"
)

func main() {
	cl := ws.NewClient(codec.ProtoSerializer{}, 5*time.Second)
	err := cl.Dial("127.0.0.1", 8080)
	if err != nil {
		logx.Errorf("dial err %+v", err)
		return
	}
	for i := 0; i < 5; i++ {
		err = ws.Ask[*test.HelloAsk, *test.HelloRsp](cl, uint32(0), uint32(test.RouterId_Hello), &test.HelloAsk{
			Msg: "client world",
		}, func(rsp *test.HelloRsp) {
			logx.Infof("client hello rsp %+v", rsp.Msg)
		})
		if err != nil {
			logx.Errorf("ask err %+v", err)
		}
		time.Sleep(time.Second)
	}

	for i := 0; i < 5; i++ {
		err = ws.Tell[*test.HiTell](cl, uint32(0), uint32(test.RouterId_Hi), &test.HiTell{Msg: "client hi"})
		if err != nil {
			logx.Errorf("tell err %+v", err)
		}
		time.Sleep(time.Second)
	}

	syncx.WaitUntilSignaled()
}
