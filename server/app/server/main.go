package main

import (
	"hutool/logx"
	"hutool/syncx"
	"net/http"
	"server/app/test"
	"server/pkg/codec"
	router2 "server/pkg/router"
	"server/pkg/service"
	session2 "server/pkg/session"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

func main() {
	uid := atomic.Uint32{}

	router := router2.NewRouter()
	router2.RegisterAskRouter[*test.HelloAsk, *test.HelloRsp](router, uint32(test.RouterId_Hello), func(ctx codec.ReqCtx, req *test.HelloAsk) *test.HelloRsp {
		v, _ := ctx.GetSession().Get("uid")
		id := v.(uint32)
		logx.Infof("msg from client %v %v", id, req.Msg)
		return &test.HelloRsp{Msg: "hello world"}
	})

	router2.RegisterTellRouter[*test.HiTell](router, uint32(test.RouterId_Hi), func(ctx codec.ReqCtx, req *test.HiTell) {
		v, _ := ctx.GetSession().Get("uid")
		id := v.(uint32)
		logx.Infof("msg from client %v %v", id, req.Msg)
	})

	svc := service.NewService(
		0,
		service.WithSerializer(codec.ProtoSerializer{}),
		service.WithRouter(router),
		service.WithSessionOpts(
			session2.WithOnSessionBind(func(session *session2.Session) {
				v := uid.Add(1)
				session.Set("uid", v)
			}),
			session2.WithOnSessionEnd(func(session *session2.Session) {
				v, _ := session.Get("uid")
				logx.Infof("session end: %d", v.(uint32))
			})),
		service.WithPlugin(Plugin{}),
	)

	err := svc.StartWsServer("127.0.0.1", 8080, websocket.Upgrader{
		ReadBufferSize:  1024 * 4,
		WriteBufferSize: 1024 * 4,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})
	if err != nil {
		logx.Errorf("start tcp server error: %v", err)
	}

	syncx.WaitUntilSignaled()

	svc.Stop()
}

type Plugin struct {
}

func (h Plugin) PreReadReadRequest(session *session2.Session, reqPacket codec.C2SPacket) bool {
	uid, _ := session.Get("uid")
	logx.Infof("new req_rsp %v", uid)
	if reqPacket.ServiceId() != 0 {
		logx.Infof("intercept")
		return false
	}
	return true
}

func (h Plugin) HeartBeat(session *session2.Session) {
	uid, _ := session.Get("uid")
	logx.Infof("heartbeat %v", uid)
}
