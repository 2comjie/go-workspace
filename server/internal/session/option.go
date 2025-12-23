package session

import (
	"time"
)

type Config struct {
	sessionExpireTime    time.Duration
	sessionCheckInterval time.Duration
	onSessionEnd         func(session *Session)
	onSessionBind        func(session *Session)
}

type Option func(*Config)

func WithSessionExpireTime(expireTime time.Duration) Option {
	return func(c *Config) {
		c.sessionExpireTime = expireTime
	}
}

func WithSessionCheckInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.sessionCheckInterval = interval
	}
}

func WithOnSessionEnd(onSessionEnd func(session *Session)) Option {
	return func(c *Config) {
		c.onSessionEnd = onSessionEnd
	}
}

func WithOnSessionBind(onSessionBind func(session *Session)) Option {
	return func(c *Config) {
		c.onSessionBind = onSessionBind
	}
}

func DefaultConfig() *Config {
	return &Config{
		sessionExpireTime:    10 * time.Second,
		sessionCheckInterval: 5 * time.Second,
		onSessionEnd:         nil,
		onSessionBind:        nil,
	}
}
