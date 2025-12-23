package session

import (
	"context"
	"server/internal/net/inet"
	"sync"
	"time"
)

type Manager struct {
	expireTime      time.Duration
	checkInterval   time.Duration
	connIdToSession sync.Map
	onSessionEnd    func(session *Session)
	onSessionBind   func(session *Session)
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

func NewManager(opts ...Option) *Manager {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		expireTime:      cfg.sessionExpireTime,
		checkInterval:   cfg.sessionCheckInterval,
		onSessionEnd:    cfg.onSessionEnd,
		onSessionBind:   cfg.onSessionBind,
		connIdToSession: sync.Map{},
		ctx:             ctx,
		cancel:          cancel,
		wg:              sync.WaitGroup{},
	}
	m.wg.Add(1)
	go m.checkAlive()
	return m
}

func (m *Manager) BindSession(conn inet.IConn) *Session {
	v, loaded := m.connIdToSession.LoadOrStore(conn.GetConnId(), &Session{
		bindConn:       conn,
		ctx:            sync.Map{},
		lastActiveTime: time.Now(),
		RWMutex:        sync.RWMutex{},
	})
	if !loaded {
		if m.onSessionBind != nil {
			m.onSessionBind(v.(*Session))
		}
	}
	return v.(*Session)
}

func (m *Manager) GetSession(connId uint32) (*Session, bool) {
	session, ok := m.connIdToSession.Load(connId)
	return session.(*Session), ok
}

func (m *Manager) RemoveSession(connId uint32) {
	v, ok := m.connIdToSession.LoadAndDelete(connId)
	if ok {
		session := v.(*Session)
		if m.onSessionEnd != nil {
			m.onSessionEnd(session)
		}
		session.bindConn.Close()
	}
}

func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
}

func (m *Manager) KeepAlive(connId uint32) {
	session, ok := m.GetSession(connId)
	if ok {
		session.keepAlive()
	}
}

func (m *Manager) checkAlive() {
	defer m.wg.Done()
	tk := time.NewTicker(m.checkInterval)
	defer tk.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-tk.C:
			m.connIdToSession.Range(func(k, v any) bool {
				session, _ := v.(*Session)
				if session.Expired(m.expireTime) {
					m.RemoveSession(session.bindConn.GetConnId())
				}
				return true
			})
		}
	}
}

func (m *Manager) Push(connId uint32, data []byte) error {
	session, ok := m.GetSession(connId)
	if ok {
		return session.bindConn.Write(data)
	}
	return NotFoundErr
}
