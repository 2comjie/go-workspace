package router

import (
	"hutool/reflectx"
	"reflect"
	"server/internal/codec"
)

type askHandler func(ctx codec.ReqCtx, req any) any

type tellHandler func(ctx codec.ReqCtx, req any)

type AskHandler[Req any, Rsp any] func(ctx codec.ReqCtx, req Req) Rsp
type TellHandler[Req any] func(ctx codec.ReqCtx, req Req)

type Registry struct {
	askRouterMap  map[uint32]AskRouter
	tellRouterMap map[uint32]TellRouter
}

func NewRouter() *Registry {
	return &Registry{
		askRouterMap:  make(map[uint32]AskRouter),
		tellRouterMap: make(map[uint32]TellRouter),
	}
}

type AskRouter struct {
	ReqType reflect.Type
	RspType reflect.Type
	Handler askHandler
}

type TellRouter struct {
	ReqType reflect.Type
	Handler tellHandler
}

func (r *Registry) registerAskRouter(routerId uint32, reqType reflect.Type, rspType reflect.Type, handler askHandler) {
	if r.askRouterMap == nil {
		r.askRouterMap = make(map[uint32]AskRouter)
	}
	r.askRouterMap[routerId] = AskRouter{
		ReqType: reqType,
		RspType: rspType,
		Handler: handler,
	}
}

func (r *Registry) registerTellRouter(routerId uint32, reqType reflect.Type, handler tellHandler) {
	if r.tellRouterMap == nil {
		r.tellRouterMap = make(map[uint32]TellRouter)
	}
	r.tellRouterMap[routerId] = TellRouter{
		ReqType: reqType,
		Handler: handler,
	}
}

func RegisterAskRouter[Req any, Rsp any](registry *Registry, routerId uint32, handler AskHandler[Req, Rsp]) {
	reqType := reflectx.GenericTypeOf[Req]()
	rspType := reflectx.GenericTypeOf[Rsp]()
	registry.registerAskRouter(routerId, reqType, rspType, func(ctx codec.ReqCtx, req any) any {
		rsp := handler(ctx, req.(Req))
		return rsp
	})
}

func RegisterTellRouter[Req any](registry *Registry, routerId uint32, handler TellHandler[Req]) {
	reqType := reflectx.GenericTypeOf[Req]()
	registry.registerTellRouter(routerId, reqType, func(ctx codec.ReqCtx, req any) {
		handler(ctx, req.(Req))
	})
}
