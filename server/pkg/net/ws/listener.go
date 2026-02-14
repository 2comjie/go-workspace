package ws

import (
	"context"
	"errors"
	"fmt"
	"hutool/logx"
	"net"
	"net/http"
	"server/pkg/net/inet"

	"github.com/gorilla/websocket"
)

type Listener struct {
	httpServer *http.Server
	upgrader   websocket.Upgrader
	connChan   chan *websocket.Conn
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewListener() *Listener {
	ctx, cancel := context.WithCancel(context.Background())
	return &Listener{
		ctx:      ctx,
		cancel:   cancel,
		connChan: make(chan *websocket.Conn, 10),
	}
}

func (l *Listener) Accept(svc inet.IService) (*Conn, error) {
	select {
	case <-l.ctx.Done():
		return nil, net.ErrClosed
	case rawConn := <-l.connChan:
		if rawConn == nil {
			return nil, net.ErrClosed
		}
		conn := NewConn(rawConn, svc)
		return conn, nil
	}
}

func (l *Listener) Listen(host string, port int, upgrader websocket.Upgrader) error {
	address := fmt.Sprintf("%s:%d", host, port)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", l.handleWebSocket)
	l.httpServer = &http.Server{
		Addr:    address,
		Handler: mux,
	}
	l.upgrader = upgrader

	go func() {
		err := l.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logx.Errorf("ws listener err %+v", err)
		}
	}()

	logx.Infof("ws listener start %s", address)
	return nil
}

func (l *Listener) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	rawConn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Errorf("upgrade err %+v", err)
		return
	}
	select {
	case l.connChan <- rawConn:
	case <-l.ctx.Done():
		rawConn.Close()
	default:
		logx.Errorf("ws listener too many conn")
		rawConn.Close()
	}
}

func (l *Listener) Close() {
	close(l.connChan)
	err := l.httpServer.Close()
	if err != nil {
		logx.Errorf("ws listener close err %+v", err)
	}
	l.cancel()
}
