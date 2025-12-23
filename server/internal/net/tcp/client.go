package tcp

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hutool/iox"
	"hutool/logx"
	"hutool/reflectx"
	"io"
	"net"
	"reflect"
	"server/internal/codec"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	reqId             atomic.Uint32
	rawConn           *net.TCPConn
	rspMsgHandlerMap  sync.Map
	heartbeatInterval time.Duration
	serializer        codec.ISerializer
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

func NewClient(serializer codec.ISerializer, heartbeatInterval time.Duration) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		reqId:             atomic.Uint32{},
		rawConn:           nil,
		rspMsgHandlerMap:  sync.Map{},
		heartbeatInterval: heartbeatInterval,
		serializer:        serializer,
		ctx:               ctx,
		cancel:            cancel,
		wg:                sync.WaitGroup{},
	}
	return c
}

func (c *Client) Dial(host string, port int) error {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	rawConn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}
	c.rawConn = rawConn
	c.wg.Add(1)
	go c.handleMsgFromServer()
	c.wg.Add(1)
	go c.keepAlive()

	return nil
}

type S2CMsgHandler struct {
	msgType reflect.Type
	handler func(msg any)
}

func (c *Client) ask(serviceId uint32, routerId uint32, reqBody any, handler S2CMsgHandler) error {
	reqId := c.reqId.Add(1)
	reqBodyBytes, err := c.serializer.Marshal(reqBody)
	if err != nil {
		return err
	}
	reqPacket := codec.NewC2SReqPacket(serviceId, routerId, reqId, false, reqBodyBytes)
	c.rspMsgHandlerMap.Store(reqId, handler)
	return c.writeToServer(reqPacket.Bytes())
}

func (c *Client) tell(serviceId uint32, routerId uint32, reqBody any) error {
	reqId := c.reqId.Add(1)
	reqBodyBytes, err := c.serializer.Marshal(reqBody)
	if err != nil {
		return err
	}
	reqPacket := codec.NewC2SReqPacket(serviceId, routerId, reqId, true, reqBodyBytes)
	return c.writeToServer(reqPacket.Bytes())
}

func (c *Client) handleMsgFromServer() {
	defer c.wg.Done()
	lenBytes := make([]byte, 4)
	for {
		lenBytes = lenBytes[:4]
		err := iox.ReadFixBytes(c.rawConn, lenBytes)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
				logx.Errorf("read conn err %+v", err)
			}
			return
		}
		packetLen := binary.BigEndian.Uint32(lenBytes)
		if packetLen > MaxPacketLen || packetLen <= 0 {
			logx.Errorf("packet len %d err", packetLen)
			return
		}
		packetBytes := make([]byte, packetLen)
		err = iox.ReadFixBytes(c.rawConn, packetBytes)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				logx.Errorf("read conn  err %+v", err)
			}
			return
		}
		c.handleOnMsg(packetBytes)
	}
}

func (c *Client) handleOnMsg(readData []byte) {
	msgPacket, err := codec.BytesToS2CPacket(readData)
	if err != nil {
		logx.Errorf("unmarshal err %+v", err)
		return
	}
	isPushPacket := msgPacket.IsPushPacket()
	msgBodyBytes := msgPacket.Body()
	if isPushPacket {
		// TODO
	} else {
		reqId := msgPacket.ReqId()
		handler, ok := c.getRspMsgHandler(reqId)
		if !ok {
			return
		}
		msgBody := reflectx.NewPointerIns(handler.msgType)
		err := c.serializer.Unmarshal(msgBodyBytes, msgBody)
		if err != nil {
			logx.Errorf("unmarshal err %+v", err)
			return
		}
		handler.handler(msgBody)
		c.rspMsgHandlerMap.Delete(reqId)
	}
}

func (c *Client) getRspMsgHandler(reqId uint32) (S2CMsgHandler, bool) {
	handler, ok := c.rspMsgHandlerMap.Load(reqId)
	return handler.(S2CMsgHandler), ok
}

func (c *Client) keepAlive() {
	defer c.wg.Done()
	tk := time.NewTicker(c.heartbeatInterval)
	defer tk.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-tk.C:
			heartBeatPacket := codec.NewC2SHeartBeatPacket()
			_ = c.writeToServer(heartBeatPacket.Bytes())
		}
	}
}

func (c *Client) writeToServer(data []byte) error {
	packetLen := len(data)
	p := make([]byte, packetLen+4)
	binary.BigEndian.PutUint32(p, uint32(packetLen))
	copy(p[4:], data)
	return iox.WriteLimit(c.rawConn, p, MaxPacketLen)
}

func (c *Client) Close() {
	c.cancel()
	_ = c.rawConn.Close()
	c.wg.Wait()
}

func Ask[Req any, Rsp any](client *Client, serviceId uint32, routerId uint32, req Req, handler func(rsp Rsp)) error {
	return client.ask(serviceId, routerId, req, S2CMsgHandler{
		msgType: reflectx.GenericTypeOf[Rsp](),
		handler: func(msg any) {
			handler(msg.(Rsp))
		},
	})
}

func Tell[Req any](client *Client, serviceId uint32, routerId uint32, req any) error {
	return client.tell(serviceId, routerId, req)
}
