package ws

import (
	"context"
	"fmt"
	"hutool/logx"
	"hutool/reflectx"
	"reflect"
	"server/pkg/codec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	reqId             atomic.Uint32
	rawConn           *websocket.Conn
	rspMsgHandlerMap  sync.Map
	pushMsgHandlerMap map[uint32]map[uint32]S2CMsgHandler
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
		pushMsgHandlerMap: make(map[uint32]map[uint32]S2CMsgHandler),
		heartbeatInterval: heartbeatInterval,
		serializer:        serializer,
		ctx:               ctx,
		cancel:            cancel,
		wg:                sync.WaitGroup{},
	}
	return c
}

func (c *Client) Dial(host string, port int) error {
	url := fmt.Sprintf("ws://%s:%d/ws", host, port)

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	rawConn, _, err := dialer.Dial(url, nil)
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
	for {
		// WebSocket 直接读取消息，不需要长度前缀
		messageType, message, err := c.rawConn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logx.Errorf("read conn err %+v", err)
			}
			return
		}
		if messageType != websocket.BinaryMessage {
			logx.Warnf("unsupported message type: %d", messageType)
			continue
		}
		if len(message) > MaxPacketLen {
			logx.Errorf("packet len %d too large", len(message))
			continue
		}
		c.handleOnMsg(message)
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
		serviceId := msgPacket.ServiceId()
		routerId := msgPacket.RouterId()
		handler, ok := c.getPushHandler(serviceId, routerId)
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
	if !ok {
		return S2CMsgHandler{}, false
	}
	return handler.(S2CMsgHandler), true
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
	return c.rawConn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Client) Close() {
	c.cancel()
	if c.rawConn != nil {
		_ = c.rawConn.Close()
	}
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

func RegisterPushHandler[Push any](client *Client, serviceId uint32, routerId uint32, handler func(push Push)) {
	client.registerPushHandler(serviceId, routerId, S2CMsgHandler{
		msgType: reflectx.GenericTypeOf[Push](),
		handler: func(msg any) {
			handler(msg.(Push))
		},
	})
}

func (c *Client) registerPushHandler(serviceId uint32, routerId uint32, handler S2CMsgHandler) {
	serviceMap, ok := c.pushMsgHandlerMap[serviceId]
	if !ok {
		serviceMap = make(map[uint32]S2CMsgHandler)
		c.pushMsgHandlerMap[serviceId] = serviceMap
	}
	serviceMap[routerId] = handler
}

func (c *Client) getPushHandler(serviceId uint32, routerId uint32) (S2CMsgHandler, bool) {
	serviceMap, ok := c.pushMsgHandlerMap[serviceId]
	if !ok {
		return S2CMsgHandler{}, false
	}
	msgHandler, ok := serviceMap[routerId]
	if !ok {
		return S2CMsgHandler{}, false
	}
	return msgHandler, true
}
