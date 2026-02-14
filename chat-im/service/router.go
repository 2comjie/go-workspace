package service

import (
	"chat-im/req_rsp"
	router2 "server/pkg/router"
)

func (s *Service) InitRouter() {
	router2.RegisterAskRouter[*req_rsp.LoginReq, *req_rsp.LoginRsp](s.Registry, req_rsp.Login, s.LoginReq)
}
