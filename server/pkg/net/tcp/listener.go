package tcp

import (
	"fmt"
	"hutool/logx"
	"net"
	"server/pkg/net/inet"
)

type Listener struct {
	ln *net.TCPListener
}

func NewListener() *Listener {
	return &Listener{}
}

func (l *Listener) Accept(svc inet.IService) (*Conn, error) {
	rawConn, err := l.ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	conn := NewConn(rawConn, svc)
	return conn, nil
}

func (l *Listener) Listen(host string, port int) error {
	address := fmt.Sprintf("%s:%d", host, port)
	tcpAddress, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}
	ln, err := net.ListenTCP("tcp", tcpAddress)
	if err != nil {
		return err
	}
	l.ln = ln
	logx.Infof("tcp listener start at %s", address)
	return nil
}

func (l *Listener) Close() {
	err := l.ln.Close()
	if err != nil {
		logx.Errorf("tcp listener close err %+v", err)
	}
}
