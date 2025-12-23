package service

import (
	"hutool/taskx"
	"server/internal/codec"
	router2 "server/internal/router"
	"server/internal/session"
	zip2 "server/internal/zip"
)

type Config struct {
	router            *router2.Registry
	serializer        codec.ISerializer
	sessionOpts       []session.Option
	plugins           []any
	writerPoolOptions []taskx.TaskPoolOption
	zip               zip2.IZip
}

type Option func(*Config)

func WithRouter(router *router2.Registry) Option {
	return func(c *Config) {
		c.router = router
	}
}

func WithSerializer(serializer codec.ISerializer) Option {
	return func(c *Config) {
		c.serializer = serializer
	}
}

func WithSessionOpts(opts ...session.Option) Option {
	return func(c *Config) {
		c.sessionOpts = opts
	}
}

func WithPlugin(plugin any) Option {
	return func(c *Config) {
		c.plugins = append(c.plugins, plugin)
	}
}

func WithZip(zip zip2.IZip) Option {
	return func(c *Config) {
		c.zip = zip
	}
}

func WithWriterPoolOptions(options ...taskx.TaskPoolOption) Option {
	return func(c *Config) {
		c.writerPoolOptions = options
	}
}

func DefaultConfig() *Config {
	return &Config{
		router:            router2.NewRouter(),
		serializer:        codec.ProtoSerializer{},
		sessionOpts:       []session.Option{},
		plugins:           []any{},
		writerPoolOptions: make([]taskx.TaskPoolOption, 0),
		zip:               zip2.None{},
	}
}
