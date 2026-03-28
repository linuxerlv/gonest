package abstract

import "io"

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

// Field 日志字段接口
type Field interface {
	Key() string
	Value() any
}

// LoggerWriter 日志写入接口
type LoggerWriter interface {
	Write(p []byte) (n int, err error)
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string
	Output     io.Writer
	TimeFormat string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	FileName   string
}

// GlobalLogger 全局日志接口
type GlobalLogger interface {
	SetGlobalLogger(log Logger)
	GetGlobalLogger() Logger
}

// FieldImpl 日志字段实现
type FieldImpl struct {
	key   string
	value any
}

func NewField(key string, value any) Field {
	return &FieldImpl{key: key, value: value}
}

func (f *FieldImpl) Key() string { return f.key }
func (f *FieldImpl) Value() any  { return f.value }

// 常用字段工厂函数
func Err(err error) Field             { return NewField("error", err) }
func String(key, value string) Field  { return NewField(key, value) }
func Int(key string, value int) Field { return NewField(key, value) }
func Any(key string, value any) Field { return NewField(key, value) }
