package ws

import (
	"hutool/logx"
	"server/pkg/net/conn_id"
	"server/pkg/net/inet"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

const MaxPacketLen = 4 * 1024

type Conn struct {
	rawConn *websocket.Conn
	connId  uint32
	svc     inet.IService
	closed  atomic.Bool
}

func NewConn(rawConn *websocket.Conn, svc inet.IService) *Conn {
	c := &Conn{
		rawConn: rawConn,
		connId:  conn_id.NextId(),
		closed:  atomic.Bool{},
		svc:     svc,
	}
	return c
}

func (c *Conn) GetConnId() uint32 {
	return c.connId
}

func (c *Conn) RemoteAddr() string {
	return c.rawConn.RemoteAddr().String()
}

func (c *Conn) Write(data []byte) error {
	return c.rawConn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Conn) Close() {
	if c.closed.CompareAndSwap(false, true) {
		err := c.rawConn.Close()
		if err != nil {
			logx.Errorf("ws conn close err %+v", err)
		}
		c.svc.OnConnStop(c)
	}
}
