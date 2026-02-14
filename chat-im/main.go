package main

import (
	"chat-im/service"
	"hutool/logx"
	"hutool/logx/logdef"
	"hutool/logx/stdlog"
	"hutool/syncx"
	"server/pkg/codec"
	net "server/pkg/service"
	session2 "server/pkg/session"
	"time"
)

func main() {
	logger := stdlog.NewLogger(stdlog.WithLevel(logdef.LevelDebug))
	logx.SetLogger(logger)
	svc := service.NewService()
	netService := net.NewService(0,
		net.WithPlugin(svc),
		net.WithSerializer(codec.JsonSerializer{}),
		net.WithRouter(svc.Registry),
		net.WithSessionOpts(session2.WithSessionExpireTime(10*time.Second)),
	)
	svc.NetService = netService
	_ = netService.StartTCPServer("127.0.0.1", 8080)
	syncx.WaitUntilSignaled()
	netService.Stop()
}
