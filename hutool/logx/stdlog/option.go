package stdlog

import "hutool/logx/logdef"

type Config struct {
	skip  int
	level logdef.Level
	hook  Hook
}

type Option func(config *Config)

func WithSkip(skip int) Option {
	return func(config *Config) {
		config.skip = skip
	}
}

func WithLevel(level logdef.Level) Option {
	return func(config *Config) {
		config.level = level
	}
}

func WithHook(hook Hook) Option {
	return func(config *Config) {
		config.hook = hook
	}
}

func DefaultConfig() *Config {
	return &Config{
		skip:  4,
		level: logdef.LevelInfo,
		hook:  nil,
	}
}
