package core

import (
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/logger"
)

type LoggerAdapter struct {
	log logger.Logger
}

func NewLoggerAdapter(log logger.Logger) *LoggerAdapter {
	return &LoggerAdapter{log: log}
}

func (l *LoggerAdapter) Debug(msg string, fields ...abstract.Field) {
	if l.log == nil {
		return
	}
	l.log.Debug(msg, fields...)
}

func (l *LoggerAdapter) Info(msg string, fields ...abstract.Field) {
	if l.log == nil {
		return
	}
	l.log.Info(msg, fields...)
}

func (l *LoggerAdapter) Warn(msg string, fields ...abstract.Field) {
	if l.log == nil {
		return
	}
	l.log.Warn(msg, fields...)
}

func (l *LoggerAdapter) Error(msg string, fields ...abstract.Field) {
	if l.log == nil {
		return
	}
	l.log.Error(msg, fields...)
}

func (l *LoggerAdapter) Fatal(msg string, fields ...abstract.Field) {
	if l.log == nil {
		return
	}
	l.log.Fatal(msg, fields...)
}

func (l *LoggerAdapter) Unwrap() logger.Logger {
	return l.log
}

var _ abstract.Logger = (*LoggerAdapter)(nil)
