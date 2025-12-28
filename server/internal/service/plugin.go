package service

import (
	"server/internal/codec"
	session2 "server/internal/session"
)

type PluginContainer struct {
	plugins []any
}

func NewPluginContainer(plugins []any) *PluginContainer {
	p := &PluginContainer{}
	for _, plugin := range plugins {
		p.Register(plugin)
	}
	return p
}

type (
	PreReadRequestPlugin interface {
		PreReadReadRequest(session *session2.Session, reqPacket codec.C2SPacket) bool
	}

	PostReadRequestPlugin interface {
		PostReadRequest(session *session2.Session, req any)
	}

	HeartBeatPlugin interface {
		HeartBeat(session *session2.Session)
	}

	PreSvcStopPlugin interface {
		PreSvcStop(svc *Service)
	}

	PostSvcStopPlugin interface {
		PostSvcStop(svc *Service)
	}
)

func (p *PluginContainer) Register(plugin any) {
	p.plugins = append(p.plugins, plugin)
}

func (p *PluginContainer) Remove(plugin any) {
	if p.plugins == nil {
		return
	}
	plugins := make([]any, 0, len(p.plugins))
	for _, p := range p.plugins {
		if p != plugin {
			plugins = append(plugins, p)
		}
	}
	p.plugins = plugins
}

func (p *PluginContainer) doPreReadRequest(session *session2.Session, reqPacket codec.C2SPacket) bool {
	for _, plugin := range p.plugins {
		if plugin, ok := plugin.(PreReadRequestPlugin); ok {
			if !plugin.PreReadReadRequest(session, reqPacket) {
				return false
			}
		}
	}
	return true
}

func (p *PluginContainer) doPostReadRequest(session *session2.Session, req any) {
	for _, plugin := range p.plugins {
		if plugin, ok := plugin.(PostReadRequestPlugin); ok {
			plugin.PostReadRequest(session, req)
		}
	}
}

func (p *PluginContainer) doHeartBeat(session *session2.Session) {
	for _, plugin := range p.plugins {
		if plugin, ok := plugin.(HeartBeatPlugin); ok {
			plugin.HeartBeat(session)
		}
	}
}

func (p *PluginContainer) doPreSvcStop(svc *Service) {
	for _, plugin := range p.plugins {
		if plugin, ok := plugin.(PreSvcStopPlugin); ok {
			plugin.PreSvcStop(svc)
		}
	}
}

func (p *PluginContainer) doPostSvcStop(svc *Service) {
	for _, plugin := range p.plugins {
		if plugin, ok := plugin.(PostSvcStopPlugin); ok {
			plugin.PostSvcStop(svc)
		}
	}
}
