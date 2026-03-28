package logger

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/linuxerlv/gonest/core/abstract"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ZapLogger struct {
	logger     *zap.Logger
	sugar      *zap.SugaredLogger
	level      zapcore.Level
	fields     []abstract.Field
	name       string
	mu         sync.RWMutex
	lumberjack *lumberjack.Logger
}

func NewZapLogger(cfg Config) (*ZapLogger, error) {
	core, level, lj, err := buildZapCore(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build zap core: %w", err)
	}

	opts := buildZapOptions(cfg)
	logger := zap.New(core, opts...)

	return &ZapLogger{
		logger:     logger,
		sugar:      logger.Sugar(),
		level:      level,
		lumberjack: lj,
	}, nil
}

func NewZapLoggerFromZap(zapLogger *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: zapLogger,
		sugar:  zapLogger.Sugar(),
		level:  zapLogger.Level(),
	}
}

func buildZapCore(cfg Config) (zapcore.Core, zapcore.Level, *lumberjack.Logger, error) {
	level := toZapLevel(cfg.Level)
	encoder := buildEncoder(cfg)

	var ws zapcore.WriteSyncer
	var lj *lumberjack.Logger

	switch cfg.Output {
	case "stdout":
		ws = zapcore.AddSync(os.Stdout)
	case "stderr":
		ws = zapcore.AddSync(os.Stderr)
	case "file":
		if cfg.FilePath == "" {
			return nil, level, nil, fmt.Errorf("file path is required when output is 'file'")
		}
		lj = &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		ws = zapcore.AddSync(lj)
	default:
		ws = zapcore.AddSync(os.Stdout)
	}

	return zapcore.NewCore(encoder, ws, level), level, lj, nil
}

func buildEncoder(cfg Config) zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        FieldKeyTime,
		LevelKey:       FieldKeyLevel,
		NameKey:        FieldKeyLoggerName,
		CallerKey:      FieldKeyCaller,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     FieldKeyMsg,
		StacktraceKey:  FieldKeyStack,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Development {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	switch cfg.Format {
	case "json":
		return zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return zapcore.NewConsoleEncoder(encoderConfig)
	case "text":
		return zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return zapcore.NewJSONEncoder(encoderConfig)
	}
}

func buildZapOptions(cfg Config) []zap.Option {
	opts := make([]zap.Option, 0)

	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
		opts = append(opts, zap.AddCallerSkip(cfg.CallerSkip))
	}

	if !cfg.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	if len(cfg.InitialFields) > 0 {
		fields := make([]zap.Field, 0, len(cfg.InitialFields))
		for k, v := range cfg.InitialFields {
			fields = append(fields, zap.Any(k, v))
		}
		opts = append(opts, zap.Fields(fields...))
	}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	return opts
}

func toZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case DPanicLevel:
		return zapcore.DPanicLevel
	case PanicLevel:
		return zapcore.PanicLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func toZapFields(fields []abstract.Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key(), f.Value()))
	}
	return zapFields
}

func (l *ZapLogger) Debug(msg string, fields ...abstract.Field) {
	l.logger.Debug(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Info(msg string, fields ...abstract.Field) {
	l.logger.Info(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Warn(msg string, fields ...abstract.Field) {
	l.logger.Warn(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Error(msg string, fields ...abstract.Field) {
	l.logger.Error(msg, toZapFields(fields)...)
}

func (l *ZapLogger) DPanic(msg string, fields ...abstract.Field) {
	l.logger.DPanic(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Panic(msg string, fields ...abstract.Field) {
	l.logger.Panic(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Fatal(msg string, fields ...abstract.Field) {
	l.logger.Fatal(msg, toZapFields(fields)...)
}

func (l *ZapLogger) DebugCtx(ctx context.Context, msg string, fields ...abstract.Field) {
	l.logger.Debug(msg, toZapFields(fields)...)
}

func (l *ZapLogger) InfoCtx(ctx context.Context, msg string, fields ...abstract.Field) {
	l.logger.Info(msg, toZapFields(fields)...)
}

func (l *ZapLogger) WarnCtx(ctx context.Context, msg string, fields ...abstract.Field) {
	l.logger.Warn(msg, toZapFields(fields)...)
}

func (l *ZapLogger) ErrorCtx(ctx context.Context, msg string, fields ...abstract.Field) {
	l.logger.Error(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Debugf(format string, args ...any) {
	l.sugar.Debugf(format, args...)
}

func (l *ZapLogger) Infof(format string, args ...any) {
	l.sugar.Infof(format, args...)
}

func (l *ZapLogger) Warnf(format string, args ...any) {
	l.sugar.Warnf(format, args...)
}

func (l *ZapLogger) Errorf(format string, args ...any) {
	l.sugar.Errorf(format, args...)
}

func (l *ZapLogger) With(fields ...abstract.Field) Logger {
	l.mu.RLock()
	existingFields := make([]abstract.Field, len(l.fields))
	copy(existingFields, l.fields)
	l.mu.RUnlock()

	allFields := append(existingFields, fields...)
	zapFields := toZapFields(fields)

	anyFields := make([]any, len(zapFields))
	for i, f := range zapFields {
		anyFields[i] = f
	}

	return &ZapLogger{
		logger: l.logger.With(zapFields...),
		sugar:  l.sugar.With(anyFields...),
		level:  l.level,
		fields: allFields,
		name:   l.name,
	}
}

func (l *ZapLogger) WithName(name string) Logger {
	return &ZapLogger{
		logger: l.logger.Named(name),
		sugar:  l.sugar.Named(name),
		level:  l.level,
		fields: l.fields,
		name:   name,
	}
}

func (l *ZapLogger) WithCtx(ctx context.Context) Logger {
	return l
}

func (l *ZapLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = toZapLevel(level)
}

func (l *ZapLogger) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	switch l.level {
	case zapcore.DebugLevel:
		return DebugLevel
	case zapcore.InfoLevel:
		return InfoLevel
	case zapcore.WarnLevel:
		return WarnLevel
	case zapcore.ErrorLevel:
		return ErrorLevel
	case zapcore.DPanicLevel:
		return DPanicLevel
	case zapcore.PanicLevel:
		return PanicLevel
	case zapcore.FatalLevel:
		return FatalLevel
	default:
		return InfoLevel
	}
}

func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

func (l *ZapLogger) Close() error {
	var err error
	if l.lumberjack != nil {
		err = l.lumberjack.Close()
	}
	if syncErr := l.logger.Sync(); syncErr != nil && err == nil {
		err = syncErr
	}
	return err
}

func (l *ZapLogger) Raw() *zap.Logger {
	return l.logger
}

var (
	globalLogger Logger = &NopLogger{}
	globalMu     sync.RWMutex
)

func SetGlobalLogger(l Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger = l
}

func GetGlobalLogger() Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalLogger
}

func Debug(msg string, fields ...abstract.Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...abstract.Field) {
	GetGlobalLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...abstract.Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...abstract.Field) {
	GetGlobalLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...abstract.Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}

func With(fields ...abstract.Field) Logger {
	return GetGlobalLogger().With(fields...)
}

func Sync() error {
	return GetGlobalLogger().Sync()
}

func Close() error {
	return GetGlobalLogger().Close()
}
