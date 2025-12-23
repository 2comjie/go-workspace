package session

import (
	"server/internal/net/inet"
	"sync"
	"time"
)

type Session struct {
	bindConn       inet.IConn
	ctx            sync.Map
	lastActiveTime time.Time
	sync.RWMutex
}

func (s *Session) Get(key string) (any, bool) {
	v, ok := s.ctx.Load(key)
	return v, ok
}

func (s *Session) Set(key string, value any) {
	s.ctx.Store(key, value)
}

func (s *Session) Remove(key string) {
	s.ctx.Delete(key)
}

func (s *Session) GetConnId() uint32 {
	return s.bindConn.GetConnId()
}

func (s *Session) Expired(expireDuration time.Duration) bool {
	s.RLock()
	defer s.RUnlock()
	return time.Now().Sub(s.lastActiveTime) > expireDuration
}

func (s *Session) keepAlive() {
	s.Lock()
	defer s.Unlock()
	s.lastActiveTime = time.Now()
}
