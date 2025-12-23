package tcp

import (
	"context"
	"encoding/binary"
	"errors"
	"hutool/bytex"
	"hutool/iox"
	"hutool/logx"
	"io"
	"net"
	"server/internal/net/inet"
	"sync"
)

type Server struct {
	listener *Listener
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	svc      inet.IService
}

func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		ctx:    ctx,
		cancel: cancel,
		wg:     sync.WaitGroup{},
	}
}

func (s *Server) ListenAndServe(host string, port int, svc inet.IService) error {
	s.svc = svc
	s.listener = NewListener()
	err := s.listener.Listen(host, port)
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
				logx.Errorf("accept tcp err %+v", err)
			}
			return
		}
		s.svc.OnConnStart(conn)
		logx.Debugf("new tcp conn %s", conn.RemoteAddr())
		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn *Conn) {
	defer s.wg.Done()
	defer conn.Close()

	lenBytes := bytex.Allocate(4)
	defer bytex.Return(lenBytes)
	for {
		// 读取长度头
		lenBytes = lenBytes[:4]
		err := iox.ReadFixBytes(conn, lenBytes)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
				logx.Errorf("read conn %v err %+v", conn.RemoteAddr(), err)
			}
			return
		}
		packetLen := binary.BigEndian.Uint32(lenBytes)
		if packetLen > MaxPacketLen || packetLen <= 0 {
			logx.Errorf("packet len %d err", packetLen)
			return
		}
		// 读取包体
		ok := func() bool {
			packetBytes := bytex.Allocate(int(packetLen))
			defer bytex.Return(packetBytes)
			err = iox.ReadFixBytes(conn, packetBytes)
			if err != nil {
				if !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
					logx.Errorf("read conn %s err %+v", conn.RemoteAddr(), err)
				}
				return false
			}
			s.svc.OnConnRead(conn, packetBytes)
			return true
		}()
		if !ok {
			return
		}
	}
}

func (s *Server) Stop() {
	s.listener.Close()
	s.cancel()
	s.wg.Wait()
	logx.Infof("tcp server stop")
}
