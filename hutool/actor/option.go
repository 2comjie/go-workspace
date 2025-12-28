package actor

import "time"

type Config struct {
	dt         time.Duration
	canDropMsg bool
	msgChanLen int
}

type Option func(*Config)

func WithDropMsg() Option {
	return func(c *Config) {
		c.canDropMsg = true
	}
}

func WithDt(dt time.Duration) Option {
	return func(c *Config) {
		c.dt = dt
	}
}

func WithMsgChanLen(l int) Option {
	return func(c *Config) {
		c.msgChanLen = l
	}
}

func DefaultConfig() *Config {
	return &Config{
		dt:         time.Second,
		canDropMsg: false,
		msgChanLen: 1000,
	}
}
