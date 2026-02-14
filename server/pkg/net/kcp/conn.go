package kcp

import (
	"encoding/binary"
	"hutool/bytex"
	"hutool/iox"
	"hutool/logx"
	"server/pkg/net/conn_id"
	"server/pkg/net/inet"
	"sync/atomic"

	"github.com/xtaci/kcp-go/v5"
)

const MaxPacketLen = 4 * 1024

type Conn struct {
	rawConn *kcp.UDPSession
	connId  uint32
	svc     inet.IService
	closed  atomic.Bool
}

func (c *Conn) Read(p []byte) (n int, err error) {
	return c.rawConn.Read(p)
}

func NewConn(rawConn *kcp.UDPSession, svc inet.IService) *Conn {
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

func (c *Conn) Write(b []byte) error {
	packetLen := len(b)
	packet := bytex.Allocate(packetLen + 4)
	defer bytex.Return(packet)
	binary.BigEndian.PutUint32(packet, uint32(packetLen))
	copy(packet[4:], b)
	err := iox.WriteLimit(c.rawConn, packet, MaxPacketLen)
	if err != nil {
		return err
	}
	return nil
}

func (c *Conn) Close() {
	if c.closed.CompareAndSwap(false, true) {
		err := c.rawConn.Close()
		if err != nil {
			logx.Errorf("tcp conn close err %+v", err)
		}
		c.svc.OnConnStop(c)
	}
}
