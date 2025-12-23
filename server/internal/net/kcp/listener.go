package kcp

import (
	"fmt"
	"hutool/logx"
	"server/internal/net/inet"

	"github.com/xtaci/kcp-go/v5"
)

type Listener struct {
	ln *kcp.Listener
}

func NewListener() *Listener {
	return &Listener{}
}

func (l *Listener) Listen(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := kcp.Listen(addr)
	if err != nil {
		return err
	}
	l.ln = ln.(*kcp.Listener)
	logx.Infof("kcp listener start at %s", addr)
	return nil
}

func (l *Listener) Accept(svc inet.IService) (*Conn, error) {
	rawConn, err := l.ln.AcceptKCP()
	if err != nil {
		return nil, err
	}
	conn := NewConn(rawConn, svc)
	return conn, nil
}

func (l *Listener) Close() {
	err := l.ln.Close()
	if err != nil {
		logx.Errorf("tcp listener close err %+v", err)
	}
}
