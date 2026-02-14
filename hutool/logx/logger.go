package logx

import (
	"hutool/logx/logdef"
	"hutool/logx/stdlog"
)

var globalLog logdef.ILogger = nil

func init() {
	globalLog = stdlog.NewLogger()
}

func SetLogger(logger logdef.ILogger) {
	globalLog = logger.WithSkip(1)
}

func Debug(args ...interface{}) {
	globalLog.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	globalLog.Debugf(format, args...)
}

func Info(args ...interface{}) {
	globalLog.Info(args...)
}

func Infof(format string, args ...interface{}) {
	globalLog.Infof(format, args...)
}

func Warn(args ...interface{}) {
	globalLog.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	globalLog.Warnf(format, args...)
}

func Error(args ...interface{}) {
	globalLog.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	globalLog.Errorf(format, args...)
}
