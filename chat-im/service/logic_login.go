package service

import (
	"chat-im/req_rsp"
	"server/pkg/codec"
)

func (s *Service) LoginReq(ctx codec.ReqCtx, req *req_rsp.LoginReq) *req_rsp.LoginRsp {
	uid := req.Uid
	s.uidToSessionId.Store(uid, ctx.GetSession().GetConnId())
	ctx.GetSession().Set("uid", uid)
	return &req_rsp.LoginRsp{}
}

func (s *Service) ChatReq(ctx codec.ReqCtx, req *req_rsp.ChatReq) *req_rsp.LoginRsp {
	return nil
}
