package logger

import (
	"context"

	"github.com/linuxerlv/gonest/core/abstract"
)

var _ abstract.LoggerAbstract = (*ZapLogger)(nil)
var _ abstract.LoggerAbstract = (*NopLogger)(nil)
var _ abstract.FieldAbstract = (*Field)(nil)

// Level 日志级别
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

// Field 日志字段
type Field struct {
	key   string
	value any
}

func (f *Field) Key() string { return f.key }
func (f *Field) Value() any  { return f.value }

// Logger 日志接口
type Logger interface {
	// 基础日志方法
	Debug(msg string, fields ...abstract.FieldAbstract)
	Info(msg string, fields ...abstract.FieldAbstract)
	Warn(msg string, fields ...abstract.FieldAbstract)
	Error(msg string, fields ...abstract.FieldAbstract)
	DPanic(msg string, fields ...abstract.FieldAbstract)
	Panic(msg string, fields ...abstract.FieldAbstract)
	Fatal(msg string, fields ...abstract.FieldAbstract)

	// 带上下文的日志方法
	DebugCtx(ctx context.Context, msg string, fields ...abstract.FieldAbstract)
	InfoCtx(ctx context.Context, msg string, fields ...abstract.FieldAbstract)
	WarnCtx(ctx context.Context, msg string, fields ...abstract.FieldAbstract)
	ErrorCtx(ctx context.Context, msg string, fields ...abstract.FieldAbstract)

	// 格式化日志方法
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)

	// 创建子Logger
	With(fields ...abstract.FieldAbstract) Logger
	WithName(name string) Logger
	WithCtx(ctx context.Context) Logger

	// 级别控制
	SetLevel(level Level)
	GetLevel() Level

	// 同步刷新
	Sync() error

	// Close 关闭日志，释放资源（如文件句柄）
	Close() error
}

// FieldConstructor 字段构造函数
type FieldConstructor func(key string, value any) *Field

var (
	// String 字符串字段
	String = func(key string, value string) *Field {
		return &Field{key: key, value: value}
	}

	// Int 整数字段
	Int = func(key string, value int) *Field {
		return &Field{key: key, value: value}
	}

	// Int64 64位整数字段
	Int64 = func(key string, value int64) *Field {
		return &Field{key: key, value: value}
	}

	// Float64 浮点数字段
	Float64 = func(key string, value float64) *Field {
		return &Field{key: key, value: value}
	}

	// Bool 布尔字段
	Bool = func(key string, value bool) *Field {
		return &Field{key: key, value: value}
	}

	// Err 错误字段
	Err = func(err error) *Field {
		return &Field{key: "error", value: err}
	}

	// Any 任意类型字段
	Any = func(key string, value any) *Field {
		return &Field{key: key, value: value}
	}

	// Duration 时长字段
	Duration = func(key string, value interface{ String() string }) *Field {
		return &Field{key: key, value: value}
	}

	// Time 时间字段
	Time = func(key string, value interface{ String() string }) *Field {
		return &Field{key: key, value: value}
	}

	// Stringer 实现Stringer接口的字段
	Stringer = func(key string, value interface{ String() string }) *Field {
		return &Field{key: key, value: value.String()}
	}

	// Namespace 命名空间
	Namespace = func(key string) *Field {
		return &Field{key: key, value: nil}
	}
)

// Common field keys
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

// Config 日志配置
type Config struct {
	// Level 日志级别
	Level Level

	// Format 输出格式: json, text, console
	Format string

	// Output 输出目标: stdout, stderr, file
	Output string

	// FilePath 日志文件路径
	FilePath string

	// MaxSize 单个日志文件最大大小(MB)
	MaxSize int

	// MaxBackups 保留旧日志文件的最大数量
	MaxBackups int

	// MaxAge 保留旧日志文件的最大天数
	MaxAge int

	// Compress 是否压缩旧日志文件
	Compress bool

	// CallerSkip 调用者跳过层数
	CallerSkip int

	// Development 是否开发模式
	Development bool

	// DisableCaller 是否禁用调用者信息
	DisableCaller bool

	// DisableStacktrace 是否禁用堆栈跟踪
	DisableStacktrace bool

	// Encoding 编码方式
	Encoding string

	// InitialFields 初始字段
	InitialFields map[string]any
}

// DefaultConfig 默认配置
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

// DevelopmentConfig 开发环境配置
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

// ProductionConfig 生产环境配置
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

// NopLogger 空日志实现
type NopLogger struct{}

func NewNopLogger() *NopLogger { return &NopLogger{} }

func (l *NopLogger) Debug(msg string, fields ...abstract.FieldAbstract)                         {}
func (l *NopLogger) Info(msg string, fields ...abstract.FieldAbstract)                          {}
func (l *NopLogger) Warn(msg string, fields ...abstract.FieldAbstract)                          {}
func (l *NopLogger) Error(msg string, fields ...abstract.FieldAbstract)                         {}
func (l *NopLogger) DPanic(msg string, fields ...abstract.FieldAbstract)                        {}
func (l *NopLogger) Panic(msg string, fields ...abstract.FieldAbstract)                         {}
func (l *NopLogger) Fatal(msg string, fields ...abstract.FieldAbstract)                         {}
func (l *NopLogger) DebugCtx(ctx context.Context, msg string, fields ...abstract.FieldAbstract) {}
func (l *NopLogger) InfoCtx(ctx context.Context, msg string, fields ...abstract.FieldAbstract)  {}
func (l *NopLogger) WarnCtx(ctx context.Context, msg string, fields ...abstract.FieldAbstract)  {}
func (l *NopLogger) ErrorCtx(ctx context.Context, msg string, fields ...abstract.FieldAbstract) {}
func (l *NopLogger) Debugf(format string, args ...any)                                          {}
func (l *NopLogger) Infof(format string, args ...any)                                           {}
func (l *NopLogger) Warnf(format string, args ...any)                                           {}
func (l *NopLogger) Errorf(format string, args ...any)                                          {}
func (l *NopLogger) With(fields ...abstract.FieldAbstract) Logger                               { return l }
func (l *NopLogger) WithName(name string) Logger                                                { return l }
func (l *NopLogger) WithCtx(ctx context.Context) Logger                                         { return l }
func (l *NopLogger) SetLevel(level Level)                                                       {}
func (l *NopLogger) GetLevel() Level                                                            { return InfoLevel }
func (l *NopLogger) Sync() error                                                                { return nil }
func (l *NopLogger) Close() error                                                               { return nil }
