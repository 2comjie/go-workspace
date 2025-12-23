package stdlog

import (
	"fmt"
	"hutool/logx/logdef"
	"log"
	"os"
	"reflect"
	"slices"
	"strings"
)

type Field struct {
	K string
	V any
}

type Hook func(level logdef.Level, fields []Field, msg string)

type Logger struct {
	skip   int
	fields []Field
	level  logdef.Level
	logger *log.Logger
	hook   Hook
}

func NewLogger(opts ...Option) *Logger {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	lg := log.New(os.Stdout, "", log.Lshortfile|log.LstdFlags|log.Lmicroseconds)
	l := &Logger{
		logger: lg,
		skip:   cfg.skip,
		hook:   cfg.hook,
		level:  cfg.level,
	}
	l.level = logdef.LevelInfo
	return l
}

func (l Logger) Debug(args ...any) {
	if logdef.LevelDebug.IntValue() < l.level.IntValue() {
		return
	}
	l.print(logdef.LevelDebug, fmt.Sprint(args...))
}

func (l Logger) Info(args ...any) {
	if logdef.LevelInfo.IntValue() < l.level.IntValue() {
		return
	}
	l.print(logdef.LevelInfo, fmt.Sprint(args...))
}

func (l Logger) Warn(args ...any) {
	if logdef.LevelWarn.IntValue() < l.level.IntValue() {
		return
	}
	l.print(logdef.LevelWarn, fmt.Sprint(args...))
}

func (l Logger) Error(args ...any) {
	if logdef.LevelError.IntValue() < l.level.IntValue() {
		return
	}
	l.print(logdef.LevelError, fmt.Sprint(args...))
}

func (l Logger) Debugf(format string, args ...any) {
	if logdef.LevelDebug.IntValue() < l.level.IntValue() {
		return
	}
	l.print(logdef.LevelDebug, fmt.Sprintf(format, args...))
}

func (l Logger) Infof(format string, args ...any) {
	if logdef.LevelInfo.IntValue() < l.level.IntValue() {
		return
	}
	l.print(logdef.LevelInfo, fmt.Sprintf(format, args...))
}

func (l Logger) Warnf(format string, args ...any) {
	if logdef.LevelWarn.IntValue() < l.level.IntValue() {
		return
	}
	l.print(logdef.LevelWarn, fmt.Sprint(format, args))
}

func (l Logger) Errorf(format string, args ...any) {
	if logdef.LevelError.IntValue() < l.level.IntValue() {
		return
	}
	l.print(logdef.LevelError, fmt.Sprintf(format, args...))
}

func (l Logger) WithField(key string, value any) logdef.ILogger {
	c := l.deepClone()
	c.fields = slices.DeleteFunc(c.fields, func(f Field) bool {
		return f.K == key
	})
	c.fields = append(c.fields, Field{K: key, V: value})
	return c
}

func (l Logger) WithFields(fields map[string]any) logdef.ILogger {
	c := l.deepClone()
	for key, value := range fields {
		c.fields = slices.DeleteFunc(c.fields, func(f Field) bool {
			return f.K == key
		})
		c.fields = append(c.fields, Field{K: key, V: value})
	}
	return c
}

func (l Logger) WithSkip(skip int) logdef.ILogger {
	c := l.deepClone()
	c.skip += skip
	return c
}

func (l Logger) print(level logdef.Level, msg string) {
	if level.IntValue() < l.level.IntValue() {
		return
	}
	l.updatePrefix(level)
	sb := &strings.Builder{}
	sb.Grow(len(l.fields)*20 + len(msg) + 10)

	for _, f := range l.fields {
		sb.WriteString(" ")
		sb.WriteString(f.K)
		sb.WriteString(" ")

		v := f.V
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr && !rv.IsNil() {
			v = rv.Elem().Interface()
		}

		fmt.Fprint(sb, v)
	}

	sb.WriteString(" ")
	sb.WriteString(msg)

	result := sb.String()

	if l.hook != nil {
		l.hook(level, l.fields, result)
	}
	_ = l.logger.Output(l.skip, result)
}

func (l Logger) deepClone() Logger {
	c := Logger{
		skip:   l.skip,
		level:  l.level,
		logger: log.New(l.logger.Writer(), l.logger.Prefix(), l.logger.Flags()),
		hook:   l.hook,
	}
	for _, fd := range l.fields {
		c.fields = append(c.fields, Field{K: fd.K, V: fd.V})
	}
	return c
}

func (l Logger) updatePrefix(level logdef.Level) {
	l.logger.SetPrefix(fmt.Sprintf("%s ", level))
}
