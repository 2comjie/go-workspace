package service

import (
	"hutool/container"
	"hutool/logx"
	"hutool/reflectx"
	router2 "server/pkg/router"
	"server/pkg/service"
	session2 "server/pkg/session"
)

type Service struct {
	NetService *service.Service
	Registry   *router2.Registry

	uidToSessionId container.IMap[uint32, uint32]
}

func NewService() *Service {
	svc := &Service{
		NetService:     nil,
		Registry:       router2.NewRouter(),
		uidToSessionId: container.NewSyncMap[uint32, uint32](),
	}

	svc.InitRouter()
	return svc
}

func (s *Service) PostReadRequest(session *session2.Session, req any) {
	rType := reflectx.TypeOf(req)
	connId := session.GetConnId()
	logx.Debugf("conn %d req type %+v req body %+v", connId, rType, req)
}

func (s *Service) OnSessionEnd(session *session2.Session) {
	uid, ok := session.Get("uid")
	if !ok {
		return
	}
	s.uidToSessionId.Delete(uid.(uint32))
}
