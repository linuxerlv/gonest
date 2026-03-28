package logger

import (
	"context"

	"github.com/linuxerlv/gonest/core/abstract"
)

var _ abstract.Logger = (*ZapLogger)(nil)
var _ abstract.Logger = (*NopLogger)(nil)
var _ abstract.Field = (*Field)(nil)

type Level int

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

type Field struct {
	key   string
	value any
}

func (f *Field) Key() string { return f.key }
func (f *Field) Value() any  { return f.value }

type Logger interface {
	Debug(msg string, fields ...abstract.Field)
	Info(msg string, fields ...abstract.Field)
	Warn(msg string, fields ...abstract.Field)
	Error(msg string, fields ...abstract.Field)
	DPanic(msg string, fields ...abstract.Field)
	Panic(msg string, fields ...abstract.Field)
	Fatal(msg string, fields ...abstract.Field)

	DebugCtx(ctx context.Context, msg string, fields ...abstract.Field)
	InfoCtx(ctx context.Context, msg string, fields ...abstract.Field)
	WarnCtx(ctx context.Context, msg string, fields ...abstract.Field)
	ErrorCtx(ctx context.Context, msg string, fields ...abstract.Field)

	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)

	With(fields ...abstract.Field) Logger
	WithName(name string) Logger
	WithCtx(ctx context.Context) Logger

	SetLevel(level Level)
	GetLevel() Level

	Sync() error

	Close() error
}

type FieldConstructor func(key string, value any) *Field

var (
	String = func(key string, value string) *Field {
		return &Field{key: key, value: value}
	}

	Int = func(key string, value int) *Field {
		return &Field{key: key, value: value}
	}

	Int64 = func(key string, value int64) *Field {
		return &Field{key: key, value: value}
	}

	Float64 = func(key string, value float64) *Field {
		return &Field{key: key, value: value}
	}

	Bool = func(key string, value bool) *Field {
		return &Field{key: key, value: value}
	}

	Err = func(err error) *Field {
		return &Field{key: "error", value: err}
	}

	Any = func(key string, value any) *Field {
		return &Field{key: key, value: value}
	}

	Duration = func(key string, value interface{ String() string }) *Field {
		return &Field{key: key, value: value}
	}

	Time = func(key string, value interface{ String() string }) *Field {
		return &Field{key: key, value: value}
	}

	Stringer = func(key string, value interface{ String() string }) *Field {
		return &Field{key: key, value: value.String()}
	}

	Namespace = func(key string) *Field {
		return &Field{key: key, value: nil}
	}
)

const (
	FieldKeyMsg           = "msg"
	FieldKeyLevel         = "level"
	FieldKeyTime          = "time"
	FieldKeyLoggerName    = "logger"
	FieldKeyCaller        = "caller"
	FieldKeyStack         = "stack"
	FieldKeyError         = "error"
	FieldKeyRequestID     = "request_id"
	FieldKeyTraceID       = "trace_id"
	FieldKeySpanID        = "span_id"
	FieldKeyUserID        = "user_id"
	FieldKeyMethod        = "method"
	FieldKeyPath          = "path"
	FieldKeyStatusCode    = "status"
	FieldKeyLatency       = "latency"
	FieldKeyIP            = "ip"
	FieldKeyUserAgent     = "user_agent"
	FieldKeyContentLength = "content_length"
)

type Config struct {
	Level             Level
	Format            string
	Output            string
	FilePath          string
	MaxSize           int
	MaxBackups        int
	MaxAge            int
	Compress          bool
	CallerSkip        int
	Development       bool
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	InitialFields     map[string]any
}

func DefaultConfig() Config {
	return Config{
		Level:         InfoLevel,
		Format:        "json",
		Output:        "stdout",
		MaxSize:       100,
		MaxBackups:    5,
		MaxAge:        30,
		Compress:      true,
		CallerSkip:    1,
		Development:   false,
		DisableCaller: false,
	}
}

func DevelopmentConfig() Config {
	return Config{
		Level:         DebugLevel,
		Format:        "console",
		Output:        "stdout",
		MaxSize:       100,
		MaxBackups:    3,
		MaxAge:        7,
		Compress:      false,
		CallerSkip:    1,
		Development:   true,
		DisableCaller: false,
	}
}

func ProductionConfig() Config {
	return Config{
		Level:         InfoLevel,
		Format:        "json",
		Output:        "file",
		MaxSize:       100,
		MaxBackups:    10,
		MaxAge:        30,
		Compress:      true,
		CallerSkip:    1,
		Development:   false,
		DisableCaller: false,
	}
}

type NopLogger struct{}

func NewNopLogger() *NopLogger { return &NopLogger{} }

func (l *NopLogger) Debug(msg string, fields ...abstract.Field)                         {}
func (l *NopLogger) Info(msg string, fields ...abstract.Field)                          {}
func (l *NopLogger) Warn(msg string, fields ...abstract.Field)                          {}
func (l *NopLogger) Error(msg string, fields ...abstract.Field)                         {}
func (l *NopLogger) DPanic(msg string, fields ...abstract.Field)                        {}
func (l *NopLogger) Panic(msg string, fields ...abstract.Field)                         {}
func (l *NopLogger) Fatal(msg string, fields ...abstract.Field)                         {}
func (l *NopLogger) DebugCtx(ctx context.Context, msg string, fields ...abstract.Field) {}
func (l *NopLogger) InfoCtx(ctx context.Context, msg string, fields ...abstract.Field)  {}
func (l *NopLogger) WarnCtx(ctx context.Context, msg string, fields ...abstract.Field)  {}
func (l *NopLogger) ErrorCtx(ctx context.Context, msg string, fields ...abstract.Field) {}
func (l *NopLogger) Debugf(format string, args ...any)                                  {}
func (l *NopLogger) Infof(format string, args ...any)                                   {}
func (l *NopLogger) Warnf(format string, args ...any)                                   {}
func (l *NopLogger) Errorf(format string, args ...any)                                  {}
func (l *NopLogger) With(fields ...abstract.Field) Logger                               { return l }
func (l *NopLogger) WithName(name string) Logger                                        { return l }
func (l *NopLogger) WithCtx(ctx context.Context) Logger                                 { return l }
func (l *NopLogger) SetLevel(level Level)                                               {}
func (l *NopLogger) GetLevel() Level                                                    { return InfoLevel }
func (l *NopLogger) Sync() error                                                        { return nil }
func (l *NopLogger) Close() error                                                       { return nil }
