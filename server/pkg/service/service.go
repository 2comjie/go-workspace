package service

import (
	"hutool/logx"
	"hutool/reflectx"
	"hutool/taskx"
	"server/pkg/codec"
	"server/pkg/net/inet"
	"server/pkg/net/kcp"
	"server/pkg/net/tcp"
	"server/pkg/net/ws"
	router2 "server/pkg/router"
	session2 "server/pkg/session"
	zip2 "server/pkg/zip"

	"github.com/gorilla/websocket"
)

type Service struct {
	svcId uint32
	// zips
	zip zip2.IZip

	// net
	tcpServer *tcp.Server
	wsServer  *ws.Server
	kcpServer *kcp.Server

	// codec
	serializer codec.ISerializer

	// router
	routerManager *router2.Manager

	// session
	sessionManger *session2.Manager

	// plugin
	pluginContainer *PluginContainer

	writerPool *taskx.TaskPool[struct{}]
}

func NewService(svcId uint32, opts ...Option) *Service {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	s := &Service{
		svcId:           svcId,
		tcpServer:       nil,
		serializer:      cfg.serializer,
		routerManager:   router2.NewManager(cfg.router),
		sessionManger:   session2.NewManager(cfg.sessionOpts...),
		pluginContainer: NewPluginContainer(cfg.plugins),
		writerPool:      taskx.NewTaskPool[struct{}](cfg.writerPoolOptions...),
		zip:             cfg.zip,
	}

	return s
}

func (s *Service) StartTCPServer(host string, port int) error {
	s.tcpServer = tcp.NewServer()
	err := s.tcpServer.ListenAndServe(host, port, s)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) StartKcpServer(host string, port int) error {
	s.kcpServer = kcp.NewServer()
	err := s.kcpServer.ListenAndServe(host, port, s)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) StartWsServer(host string, port int, upgrader websocket.Upgrader) error {
	s.wsServer = ws.NewServer()
	err := s.wsServer.ListenAndServe(host, port, s, upgrader)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Stop() {
	s.pluginContainer.doPreSvcStop(s)
	if s.tcpServer != nil {
		s.tcpServer.Stop()
	}
	if s.wsServer != nil {
		s.wsServer.Stop()
	}
	if s.kcpServer != nil {
		s.kcpServer.Stop()
	}
	s.sessionManger.Stop()
	s.writerPool.Stop()
	s.pluginContainer.doPostSvcStop(s)
}

func (s *Service) OnConnStart(conn inet.IConn) {
	s.sessionManger.BindSession(conn)
	logx.Debugf("bind session: %d", conn.GetConnId())
}

func (s *Service) OnConnRead(conn inet.IConn, readData []byte) {
	reqPacket, err := codec.BytesToC2SPacket(readData)
	if err != nil {
		logx.Errorf("unmashal err %+v", err)
		return
	}

	session, ok := s.sessionManger.GetSession(conn.GetConnId())
	if !ok {
		logx.Warnf("session not found: %d", conn.GetConnId())
		return
	}

	if reqPacket.IsHeartbeatPacket() {
		s.sessionManger.KeepAlive(conn.GetConnId())
		s.pluginContainer.doHeartBeat(session)
	} else {
		s.handleOnPacket(session, reqPacket)
	}
}

func (s *Service) handleOnPacket(session *session2.Session, reqPacket codec.C2SPacket) {
	if !s.pluginContainer.doPreReadRequest(session, reqPacket) {
		return
	}

	svcId := reqPacket.ServiceId()
	if svcId != s.svcId {
		return
	}
	isOneWay := reqPacket.IsOneWay()
	reqBodyBytes := reqPacket.Body()
	routerId := reqPacket.RouterId()
	reqCtx := codec.NewReqCtx(reqPacket, session)

	if isOneWay {
		router, ok := s.routerManager.GetTellRouter(routerId)
		if !ok {
			return
		}

		reqBody := reflectx.NewPointerIns(router.ReqType)
		err := s.serializer.Unmarshal(reqBodyBytes, reqBody)
		if err != nil {
			logx.Errorf("unmashal err %+v", err)
			return
		}

		s.pluginContainer.doPostReadRequest(session, reqBody)
		router.Handler(reqCtx, reqBody)
	} else {
		reqId := reqPacket.ReqId()

		router, ok := s.routerManager.GetAskRouter(routerId)
		if !ok {
			return
		}

		reqBody := reflectx.NewPointerIns(router.ReqType)
		err := s.serializer.Unmarshal(reqBodyBytes, reqBody)
		if err != nil {
			logx.Errorf("unmashal err %+v", err)
			return
		}

		s.pluginContainer.doPostReadRequest(session, reqBody)
		rspBody := router.Handler(reqCtx, reqBody)
		rspBodyBytes, err := s.serializer.Marshal(rspBody)
		if err != nil {
			logx.Errorf("marshal err %+v", err)
			return
		}

		rspPacket := codec.NewS2CRspPacket(reqId, rspBodyBytes)
		err = s.writeAsync(session.GetConnId(), rspPacket.Bytes())
		if err != nil {
			logx.Errorf("push err %d %+v", session.GetConnId(), err)
		}
	}
}

func (s *Service) OnConnStop(conn inet.IConn) {
	s.sessionManger.RemoveSession(conn.GetConnId())
}

func (s *Service) Push(connId uint32, routerId uint32, data any) error {
	pushBodyBytes, err := s.serializer.Marshal(data)
	if err != nil {
		logx.Errorf("marshal err %+v", err)
		return err
	}
	pushPacket := codec.NewS2CPushPacket(s.svcId, routerId, pushBodyBytes)

	err = s.writeAsync(connId, pushPacket.Bytes())
	if err != nil {
		logx.Errorf("push err %d %+v", connId, err)
		return err
	}
	return nil
}

func (s *Service) writeAsync(connId uint32, data []byte) error {
	err := s.writerPool.Add(func() struct{} {
		zipData, err := s.zip.Zip(data)
		if err != nil {
			logx.Errorf("zip err %d %+v", connId, err)
			return struct{}{}
		}
		err = s.sessionManger.Push(connId, zipData)
		if err != nil {
			logx.Errorf("push err conn %d %+v", connId, err)
			return struct{}{}
		}
		return struct{}{}
	}, nil, connId)
	return err
}

func (s *Service) RemoveSession(connId uint32) {
	s.sessionManger.RemoveSession(connId)
}
