package ws

import (
	"errors"
	"hutool/logx"
	"net"
	"server/pkg/net/inet"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	listener *Listener
	svc      inet.IService
	wg       sync.WaitGroup
}

func NewServer() *Server {
	return &Server{
		wg: sync.WaitGroup{},
	}
}

func (s *Server) ListenAndServe(host string, port int, svc inet.IService, upgrader websocket.Upgrader) error {
	s.svc = svc
	s.listener = NewListener()
	err := s.listener.Listen(host, port, upgrader)
	if err != nil {
		return err
	}
	s.wg.Add(1)
	go s.serve()
	return nil
}

func (s *Server) serve() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept(s.svc)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				logx.Errorf("accept err %+v", err)
			}
			return
		}
		s.svc.OnConnStart(conn)
		logx.Debugf("new ws conn %s", conn.RemoteAddr())
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn *Conn) {
	defer conn.Close()

	for {
		messageType, message, err := conn.rawConn.ReadMessage()
		if err != nil {
			// 检查是否是正常的关闭错误
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				return
			}
			// 检查是否是意外的关闭错
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				logx.Errorf("read conn %v unexpected close err %+v", conn.RemoteAddr(), err)
			} else {
				// 其他错误
				logx.Errorf("read conn %v err %+v", conn.RemoteAddr(), err)
			}
			return
		}
		if messageType != websocket.BinaryMessage {
			logx.Warnf("unsupported message type: %d", messageType)
			continue
		}
		if len(message) > MaxPacketLen {
			logx.Errorf("read conn %v packet too large", conn.RemoteAddr())
			return
		}
		s.svc.OnConnRead(conn, message)
	}
}

func (s *Server) Stop() {
	s.listener.Close()
	s.wg.Wait()
	logx.Infof("ws server stop")
}
