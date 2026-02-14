package service

import (
	router2 "server/pkg/router"
	"server/pkg/service"
	"sync"
)

type Service struct {
	NetService  *service.Service
	Registry    *router2.Registry
	Uid2Session sync.Map
}
