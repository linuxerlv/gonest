package abstract

import "io"

// LoggerAbstract 日志接口
type LoggerAbstract interface {
	Debug(msg string, fields ...FieldAbstract)
	Info(msg string, fields ...FieldAbstract)
	Warn(msg string, fields ...FieldAbstract)
	Error(msg string, fields ...FieldAbstract)
	Fatal(msg string, fields ...FieldAbstract)
}

// FieldAbstract 日志字段接口
type FieldAbstract interface {
	Key() string
	Value() any
}

// LoggerWriterAbstract 日志写入接口
type LoggerWriterAbstract interface {
	Write(p []byte) (n int, err error)
}

// LoggerConfigAbstract 日志配置
type LoggerConfigAbstract struct {
	Level      string
	Output     io.Writer
	TimeFormat string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	FileName   string
}

// GlobalLoggerAbstract 全局日志接口
type GlobalLoggerAbstract interface {
	SetGlobalLogger(log LoggerAbstract)
	GetGlobalLogger() LoggerAbstract
}

// Field 日志字段实现
type Field struct {
	key   string
	value any
}

func NewField(key string, value any) FieldAbstract {
	return &Field{key: key, value: value}
}

func (f *Field) Key() string { return f.key }
func (f *Field) Value() any  { return f.value }

// 常用字段工厂函数
func Err(err error) FieldAbstract             { return NewField("error", err) }
func String(key, value string) FieldAbstract  { return NewField(key, value) }
func Int(key string, value int) FieldAbstract { return NewField(key, value) }
func Any(key string, value any) FieldAbstract { return NewField(key, value) }
