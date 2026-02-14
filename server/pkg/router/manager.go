package router

type Manager struct {
	registry *Registry
}

func NewManager(registry *Registry) *Manager {
	return &Manager{
		registry: registry,
	}
}

func (r *Manager) GetAskRouter(routerId uint32) (AskRouter, bool) {
	router, ok := r.registry.askRouterMap[routerId]
	return router, ok
}

func (r *Manager) GetTellRouter(routerId uint32) (TellRouter, bool) {
	router, ok := r.registry.tellRouterMap[routerId]
	return router, ok
}
