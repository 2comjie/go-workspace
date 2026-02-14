package codec

import "server/pkg/session"

type ReqCtx struct {
	reqId   uint32
	session *session.Session
}

func NewReqCtx(packet C2SPacket, session *session.Session) ReqCtx {
	return ReqCtx{
		session: session,
		reqId:   packet.ReqId(),
	}
}

func (c ReqCtx) GetReqId() uint32 {
	return c.reqId
}

func (c ReqCtx) GetSession() *session.Session {
	return c.session
}
