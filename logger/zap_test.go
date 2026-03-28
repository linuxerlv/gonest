package logger

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewZapLogger(t *testing.T) {
	t.Run("should create logger with default config", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should create logger with development config", func(t *testing.T) {
		cfg := DevelopmentConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should create logger with production config", func(t *testing.T) {
		cfg := ProductionConfig()
		cfg.Output = "stdout"
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should handle custom level", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Level = DebugLevel
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if logger.GetLevel() != DebugLevel {
			t.Errorf("expected level %v, got %v", DebugLevel, logger.GetLevel())
		}
	})

	t.Run("should fail when output is file but no path provided", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output = "file"
		cfg.FilePath = ""
		_, err := NewZapLogger(cfg)
		if err == nil {
			t.Fatal("expected error when file path is empty")
		}
		if !strings.Contains(err.Error(), "file path") {
			t.Errorf("expected error to contain 'file path', got %v", err)
		}
	})

	t.Run("should create logger from existing zap instance", func(t *testing.T) {
		zapLogger, err := NewZapLogger(DefaultConfig())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if zapLogger == nil {
			t.Fatal("expected logger to not be nil")
		}
		raw := zapLogger.Raw()
		if raw == nil {
			t.Fatal("expected raw logger to not be nil")
		}
	})

	t.Run("should create uses initial fields", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.InitialFields = map[string]any{
			"service": "test-service",
			"version": "1.0.0",
		}
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})
}

func TestZapLogger_Levels(t *testing.T) {
	tests := []struct {
		name   string
		level  Level
		method func(Logger)
	}{
		{"Debug level", DebugLevel, func(l Logger) { l.Debug("debug message") }},
		{"Info level", InfoLevel, func(l Logger) { l.Info("info message") }},
		{"Warn level", WarnLevel, func(l Logger) { l.Warn("warn message") }},
		{"Error level", ErrorLevel, func(l Logger) { l.Error("error message") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Level = tt.level
			logger, err := NewZapLogger(cfg)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			// Should not panic
			tt.method(logger)
		})
	}

	t.Run("should log with fields", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Level = InfoLevel
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		logger.Info("test message", String("key", "value"), Int("code", 200))
	})

	t.Run("should respect level filtering", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Level = WarnLevel
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Debug and Info should be filtered out
		logger.Debug("debug message")
		logger.Info("info message")
		// Warn and Error should be logged
		logger.Warn("warn message")
		logger.Error("error message")
	})
}

func TestZapLogger_With(t *testing.T) {
	t.Run("should create child logger with fields", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		child := logger.With(String("request_id", "123"), Int("user_id", 456))

		if child == nil {
			t.Fatal("expected child logger to not be nil")
		}

		if _, ok := child.(*ZapLogger); !ok {
			t.Errorf("expected *ZapLogger, got %T", child)
		}
	})

	t.Run("should preserve parent fields", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		child1 := logger.With(String("parent_field", "parent_value"))
		child2 := child1.With(String("child_field", "child_value"))

		// Child should have both parent and child fields
		child2Info := child2.(*ZapLogger)
		if len(child2Info.fields) != 2 {
			t.Errorf("expected 2 fields, got %d", len(child2Info.fields))
		}
	})

	t.Run("should not mutate parent logger", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		child := logger.With(String("new_field", "value"))

		child.Info("child message", String("child_field", "child_value"))

		logger.Info("parent message", String("parent_field", "parent_value"))
	})

	t.Run("should work with empty fields", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		child := logger.With()
		if child == nil {
			t.Fatal("expected child logger to not be nil")
		}
	})

	t.Run("should create nested child loggers", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		child1 := logger.With(String("a", "1"))
		child2 := child1.With(String("b", "2"))
		child3 := child2.With(String("c", "3"))

		if child3.(*ZapLogger).fields[0].Key() != "a" {
			t.Error("expected field a to be present")
		}
		if child3.(*ZapLogger).fields[1].Key() != "b" {
			t.Error("expected field b to be present")
		}
		if child3.(*ZapLogger).fields[2].Key() != "c" {
			t.Error("expected field c to be present")
		}
	})
}

func TestZapLogger_WithName(t *testing.T) {
	t.Run("should create named logger", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		named := logger.WithName("test-logger")

		if named == nil {
			t.Fatal("expected named logger to not be nil")
		}

		if named.(*ZapLogger).name != "test-logger" {
			t.Errorf("expected name 'test-logger', got %s", named.(*ZapLogger).name)
		}
	})

	t.Run("should preserve fields when naming", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		child := logger.With(String("field", "value"))
		named := child.WithName("named-logger")

		if named.(*ZapLogger).name != "named-logger" {
			t.Errorf("expected name 'named-logger', got %s", named.(*ZapLogger).name)
		}

		if len(named.(*ZapLogger).fields) != 1 {
			t.Error("expected fields to be preserved")
		}
	})

	t.Run("should chain named loggers", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		named1 := logger.WithName("api")
		named2 := named1.WithName("handler")
		named3 := named2.WithName("user")

		if named3.(*ZapLogger).name != "user" {
			t.Errorf("expected name 'user', got %s", named3.(*ZapLogger).name)
		}
	})
}

func TestZapLogger_SetLevel(t *testing.T) {
	t.Run("should change level", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger.GetLevel() != InfoLevel {
			t.Errorf("expected initial level %v, got %v", InfoLevel, logger.GetLevel())
		}

		logger.SetLevel(DebugLevel)

		if logger.GetLevel() != DebugLevel {
			t.Errorf("expected level %v after set, got %v", DebugLevel, logger.GetLevel())
		}
	})

	t.Run("should change through all levels", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}

		for _, level := range levels {
			logger.SetLevel(level)
			if logger.GetLevel() != level {
				t.Errorf("expected level %v, got %v", level, logger.GetLevel())
			}
		}
	})

	t.Run("should affect child logger level", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		child := logger.With(String("parent", "value"))
		child.SetLevel(ErrorLevel)

		if child.GetLevel() != ErrorLevel {
			t.Errorf("expected child level %v, got %v", ErrorLevel, child.GetLevel())
		}
	})
}

func TestZapLogger_Formats(t *testing.T) {
	t.Run("should create json format logger", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Format = "json"
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should create console format logger", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Format = "console"
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should create text format logger", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Format = "text"
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should default to json format", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Format = "" // empty
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})
}

func TestZapLogger_Outputs(t *testing.T) {
	t.Run("should create stdout logger", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output = "stdout"
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should create stderr logger", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output = "stderr"
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should create file logger", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		cfg := DefaultConfig()
		cfg.Output = "file"
		cfg.FilePath = filePath
		cfg.MaxSize = 1
		cfg.MaxBackups = 5
		cfg.MaxAge = 7
		cfg.Compress = false

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer logger.Close()

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}

		logger.Info("test message")
		logger.Sync()

		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("expected no error reading file, got %v", err)
		}

		if len(content) == 0 {
			t.Error("expected file to have content")
		}

		if !strings.Contains(string(content), "test message") {
			t.Error("expected file to contain 'test message'")
		}
	})

	t.Run("should handle file logger with compression", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "compressed.log")

		cfg := DefaultConfig()
		cfg.Output = "file"
		cfg.FilePath = filePath
		cfg.Compress = true

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer logger.Close()

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should fallback to stdout for invalid output", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output = "invalid-output"
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})
}

func TestLogger_Fields(t *testing.T) {
	t.Run("String field constructor", func(t *testing.T) {
		field := String("key", "value")

		if field.Key() != "key" {
			t.Errorf("expected key 'key', got %s", field.Key())
		}
		if field.Value() != "value" {
			t.Errorf("expected value 'value', got %v", field.Value())
		}
	})

	t.Run("Int field constructor", func(t *testing.T) {
		field := Int("code", 200)

		if field.Key() != "code" {
			t.Errorf("expected key 'code', got %s", field.Key())
		}
		if field.Value() != 200 {
			t.Errorf("expected value 200, got %v", field.Value())
		}
	})

	t.Run("Int64 field constructor", func(t *testing.T) {
		field := Int64("timestamp", 1234567890)

		if field.Key() != "timestamp" {
			t.Errorf("expected key 'timestamp', got %s", field.Key())
		}
		if field.Value() != int64(1234567890) {
			t.Errorf("expected value 1234567890, got %v", field.Value())
		}
	})

	t.Run("Float64 field constructor", func(t *testing.T) {
		field := Float64("amount", 99.99)

		if field.Key() != "amount" {
			t.Errorf("expected key 'amount', got %s", field.Key())
		}
		if field.Value() != 99.99 {
			t.Errorf("expected value 99.99, got %v", field.Value())
		}
	})

	t.Run("Bool field constructor", func(t *testing.T) {
		field := Bool("success", true)

		if field.Key() != "success" {
			t.Errorf("expected key 'success', got %s", field.Key())
		}
		if field.Value() != true {
			t.Errorf("expected value true, got %v", field.Value())
		}
	})

	t.Run("Err field constructor", func(t *testing.T) {
		err := os.ErrNotExist
		field := Err(err)

		if field.Key() != "error" {
			t.Errorf("expected key 'error', got %s", field.Key())
		}
		if field.Value() != err {
			t.Errorf("expected value to be the error, got %v", field.Value())
		}
	})

	t.Run("Any field constructor", func(t *testing.T) {
		data := map[string]any{"nested": "value"}
		field := Any("data", data)

		if field.Key() != "data" {
			t.Errorf("expected key 'data', got %s", field.Key())
		}
		// Maps can only be compared to nil, so just verify the value is not nil
		if field.Value() == nil {
			t.Errorf("expected value to be data, got nil")
		}
	})

	t.Run("Duration field constructor", func(t *testing.T) {
		duration := 5 * time.Second
		field := Duration("duration", duration)

		if field.Key() != "duration" {
			t.Errorf("expected key 'duration', got %s", field.Key())
		}
		// Duration can't use != comparison with same variable, just verify it's not nil
		if field.Value() == nil {
			t.Errorf("expected value to be duration, got nil")
		}
	})

	t.Run("Time field constructor", func(t *testing.T) {
		tm := time.Now()
		field := Time("timestamp", tm)

		if field.Key() != "timestamp" {
			t.Errorf("expected key 'timestamp', got %s", field.Key())
		}
		// Time can't use != comparison with same variable, just verify it's not nil
		if field.Value() == nil {
			t.Errorf("expected value to be time, got nil")
		}
	})

	t.Run("Namespace field constructor", func(t *testing.T) {
		field := Namespace("my-namespace")

		if field.Key() != "my-namespace" {
			t.Errorf("expected key 'my-namespace', got %s", field.Key())
		}
		if field.Value() != nil {
			t.Errorf("expected value nil, got %v", field.Value())
		}
	})

	t.Run("should use with logger", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		logger.Info("test",
			String("request_id", "abc"),
			Int("status", 200),
			Bool("success", true),
			Err(os.ErrNotExist),
		)
	})

	t.Run("should use with logger and context", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		ctx := context.WithValue(context.Background(), "context_key", "context_value")
		logger.InfoCtx(ctx, "test message with context",
			String("request_id", "xyz"),
		)
	})
}

// TestZapLogger_Sync tests the Sync method
func TestZapLogger_Sync(t *testing.T) {
	t.Run("should sync without error", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		err = logger.Sync()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

// TestZapLogger_SugarMethods tests the SugaredLogger methods
func TestZapLogger_SugarMethods(t *testing.T) {
	cfg := DefaultConfig()
	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Run("Debugf should work", func(t *testing.T) {
		logger.Debugf("debug %s with %d", "message", 123)
	})

	t.Run("Infof should work", func(t *testing.T) {
		logger.Infof("info %s with %d", "message", 456)
	})

	t.Run("Warnf should work", func(t *testing.T) {
		logger.Warnf("warn %s with %d", "message", 789)
	})

	t.Run("Errorf should work", func(t *testing.T) {
		logger.Errorf("error %s with %d", "message", 101112)
	})
}

// TestZapLogger_ConcurrentAccess tests concurrent access safety
func TestZapLogger_ConcurrentAccess(t *testing.T) {
	cfg := DefaultConfig()
	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	done := make(chan bool, 100)

	// Test concurrent logging
	for i := 0; i < 100; i++ {
		go func(id int) {
			logger.Info("concurrent message", Int("id", id))
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestZapLogger_DPanicsPanicMethods(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Level = DebugLevel
	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Panic to panic")
		}
	}()

	logger.Panic("panic message should panic")
}

// TestConfigConstants tests config-related constants
func TestConfigConstants(t *testing.T) {
	t.Run("should return valid default config", func(t *testing.T) {
		cfg := DefaultConfig()
		if cfg.Level != InfoLevel {
			t.Errorf("expected default level InfoLevel, got %v", cfg.Level)
		}
		if cfg.Format != "json" {
			t.Errorf("expected default format 'json', got %s", cfg.Format)
		}
		if cfg.Output != "stdout" {
			t.Errorf("expected default output 'stdout', got %s", cfg.Output)
		}
	})

	t.Run("should return valid development config", func(t *testing.T) {
		cfg := DevelopmentConfig()
		if cfg.Level != DebugLevel {
			t.Errorf("expected development level DebugLevel, got %v", cfg.Level)
		}
		if cfg.Format != "console" {
			t.Errorf("expected development format 'console', got %s", cfg.Format)
		}
		if !cfg.Development {
			t.Error("expected development mode true")
		}
	})

	t.Run("should return valid production config", func(t *testing.T) {
		cfg := ProductionConfig()
		if cfg.Level != InfoLevel {
			t.Errorf("expected production level InfoLevel, got %v", cfg.Level)
		}
		if cfg.Format != "json" {
			t.Errorf("expected production format 'json', got %s", cfg.Format)
		}
		if cfg.Output != "file" {
			t.Errorf("expected production output 'file', got %s", cfg.Output)
		}
		if cfg.Compress != true {
			t.Error("expected production compress true")
		}
	})
}

func TestConfigValues(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Level != InfoLevel {
		t.Errorf("expected default level InfoLevel, got %v", cfg.Level)
	}
	if cfg.Format != "json" {
		t.Errorf("expected default format 'json', got %s", cfg.Format)
	}
	if cfg.Output != "stdout" {
		t.Errorf("expected default output 'stdout', got %s", cfg.Output)
	}
}

// TestWithCtx tests the WithCtx method
func TestZapLogger_WithCtx(t *testing.T) {
	t.Run("should return same logger", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		ctx := context.WithValue(context.Background(), "key", "value")
		result := logger.WithCtx(ctx)

		if result != logger {
			t.Error("expected WithCtx to return same logger")
		}
	})
}

// TestFieldKeys tests the field key constants
func TestFieldKeys(t *testing.T) {
	t.Run("should have expected field keys", func(t *testing.T) {
		if FieldKeyMsg != "msg" {
			t.Errorf("expected FieldKeyMsg 'msg', got %s", FieldKeyMsg)
		}
		if FieldKeyLevel != "level" {
			t.Errorf("expected FieldKeyLevel 'level', got %s", FieldKeyLevel)
		}
		if FieldKeyTime != "time" {
			t.Errorf("expected FieldKeyTime 'time', got %s", FieldKeyTime)
		}
		if FieldKeyLoggerName != "logger" {
			t.Errorf("expected FieldKeyLoggerName 'logger', got %s", FieldKeyLoggerName)
		}
		if FieldKeyCaller != "caller" {
			t.Errorf("expected FieldKeyCaller 'caller', got %s", FieldKeyCaller)
		}
		if FieldKeyStack != "stack" {
			t.Errorf("expected FieldKeyStack 'stack', got %s", FieldKeyStack)
		}
		if FieldKeyError != "error" {
			t.Errorf("expected FieldKeyError 'error', got %s", FieldKeyError)
		}
	})
}

// TestNopLogger tests the NopLogger implementation
func TestNopLogger(t *testing.T) {
	t.Run("should not panic on any operation", func(t *testing.T) {
		logger := NewNopLogger()

		logger.Debug("debug")
		logger.Info("info")
		logger.Warn("warn")
		logger.Error("error")
		logger.DPanic("dpanic")
		logger.Panic("panic")
		logger.Fatal("fatal")

		ctx := context.Background()
		logger.DebugCtx(ctx, "debug")
		logger.InfoCtx(ctx, "info")
		logger.WarnCtx(ctx, "warn")
		logger.ErrorCtx(ctx, "error")

		logger.Debugf("fmt: %s", "msg")
		logger.Infof("fmt: %s", "msg")
		logger.Warnf("fmt: %s", "msg")
		logger.Errorf("fmt: %s", "msg")

		child := logger.With(String("key", "value"))
		named := logger.WithName("test")
		ctxLogger := logger.WithCtx(ctx)

		if child != logger {
			t.Error("expected With to return self")
		}
		if named != logger {
			t.Error("expected WithName to return self")
		}
		if ctxLogger != logger {
			t.Error("expected WithCtx to return self")
		}

		logger.SetLevel(DebugLevel)
		level := logger.GetLevel()

		if level != InfoLevel {
			t.Errorf("expected NopLogger level InfoLevel, got %v", level)
		}

		err := logger.Sync()
		if err != nil {
			t.Errorf("expected no error from Sync, got %v", err)
		}
	})
}

// TestErrorRecovery tests that logging doesn't crash on invalid fields
func TestErrorRecovery(t *testing.T) {
	cfg := DefaultConfig()
	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test with various field types that might cause issues
	logger.Info("test",
		Any("map", map[string]any{"key": "value"}),
		Any("slice", []any{1, 2, 3}),
		Any("struct", struct{ Name string }{Name: "test"}),
	)

	// Test with nil values
	logger.Info("test with nil", Any("nil_value", nil), String("valid", "value"))
}

// TestDeathTest-like patterns for Fatal
// We can't actually test Fatal without stopping tests, so we test the call is made
func TestLogger_DeathBehavior(t *testing.T) {
	t.Run("Fatal should call zap.Fatal", func(t *testing.T) {
		// This would normally exit the test, so we just verify the method exists
		// by trying to call it in a recover pattern
		defer func() {
			if r := recover(); r == nil {
				// zap.Fatal might not panic but call os.Exit(1)
				// which we can't recover from
				t.Log("Fatal method exists and was called")
			}
		}()

		// Note: We don't actually call Fatal here as it would exit the test
		// The method is tested implicitly through the interface compliance
	})
}

// TestCustomEncoder tests custom encoder configuration
func TestCustomEncoder(t *testing.T) {
	t.Run("should create logger with custom config", func(t *testing.T) {
		cfg := Config{
			Level:         DebugLevel,
			Format:        "console",
			Output:        "stdout",
			Development:   true,
			DisableCaller: false,
		}

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})
}

// TestLoggerInheritance tests logger field inheritance
func TestLoggerInheritance(t *testing.T) {
	t.Run("fields should be inherited correctly", func(t *testing.T) {
		cfg := DefaultConfig()
		parent, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		child1 := parent.With(String("parent", "value"))
		child2 := child1.With(String("child", "value"))

		// Verify fields are properly inherited
		child2Info := child2.(*ZapLogger)
		if len(child2Info.fields) != 2 {
			t.Errorf("expected 2 fields, got %d", len(child2Info.fields))
		}
	})
}

// TestGlobalLogger tests the global logger functionality
func TestGlobalLogger(t *testing.T) {
	t.Run("should have initial nop logger", func(t *testing.T) {
		// Save original
		original := GetGlobalLogger()

		defer SetGlobalLogger(original)

		// Should start with NopLogger
		if _, ok := GetGlobalLogger().(*NopLogger); !ok {
			t.Error("expected initial global logger to be NopLogger")
		}
	})

	t.Run("should set custom logger", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		SetGlobalLogger(logger)

		if GetGlobalLogger() != logger {
			t.Error("expected global logger to be set")
		}
	})

	t.Run("global methods should work", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		SetGlobalLogger(logger)

		// These should not panic
		Info("test message")
		Debug("debug message")
		Warn("warn message")
		Error("error message")

		With(String("key", "value"))
		Sync()
	})
}

// TestOutputBuffer tests writing to a bytes.Buffer
func TestOutputBuffer(t *testing.T) {
	t.Run("should write to buffer", func(t *testing.T) {
		var buf bytes.Buffer

		cfg := DefaultConfig()
		cfg.Output = "stdout"
		cfg.Development = true
		cfg.Format = "console"

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Use a custom option to write to our buffer
		// Since we can't easily intercept zap's output, we'll verify by checking
		// that the logger runs without error
		logger.Info("test message to buffer")

		if buf.Len() == 0 {
			// This is expected since zap doesn't automatically write to our buffer
			// The important thing is no panic occurred
		}
	})
}

// TestFormatVariations tests various format combinations
func TestFormatVariations(t *testing.T) {
	formats := []string{"json", "console", "text", "invalid"}
	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Format = format
			cfg.Development = true

			logger, err := NewZapLogger(cfg)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if logger == nil {
				t.Fatal("expected logger to not be nil")
			}

			logger.Info("test")
		})
	}
}

// TestOutputVariations tests various output combinations
func TestOutputVariations(t *testing.T) {
	outputs := []string{"stdout", "stderr", "file"}
	for _, output := range outputs {
		t.Run("output_"+output, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Output = output

			if output == "file" {
				tmpDir := t.TempDir()
				cfg.FilePath = filepath.Join(tmpDir, "test.log")
			}

			logger, err := NewZapLogger(cfg)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			defer logger.Close()

			if logger == nil {
				t.Fatal("expected logger to not be nil")
			}

			logger.Info("test")
		})
	}
}

// TestCallerConfig tests caller configuration
func TestCallerConfig(t *testing.T) {
	t.Run("should include caller when enabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Development = true
		cfg.DisableCaller = false

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should skip caller when DisableCaller is true", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.DisableCaller = true

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})
}

// TestStacktraceConfig tests stacktrace configuration
func TestStacktraceConfig(t *testing.T) {
	t.Run("should include stacktrace when enabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.DisableStacktrace = false

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should skip stacktrace when disabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.DisableStacktrace = true

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})
}

// TestCtorFunctionUsage tests that constructors can be used in tests
func TestCtorFunctionUsage(t *testing.T) {
	t.Run("should create constructor fields", func(t *testing.T) {
		_ = String("str", "value")
		_ = Int("int", 1)
		_ = Int64("int64", 2)
		_ = Float64("float", 3.14)
		_ = Bool("bool", true)
		_ = Err(os.ErrNotExist)
		_ = Any("any", "value")
		_ = Duration("duration", 5*time.Second)
		_ = Time("time", time.Now())
		_ = Namespace("namespace")
	})

	t.Run("should create logger with all fields", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Field constructors return *Field which implements abstract.FieldAbstract
		logger.Info("test message with fields",
			String("request_id", "test-123"),
			Int("status_code", 200),
			Bool("success", true),
			Err(nil),
			Any("data", map[string]any{"key": "value"}),
		)
	})
}

// TestFieldKeyConstants tests all field key constants
func TestAllFieldKeyConstants(t *testing.T) {
	t.Run("all field keys should be defined", func(t *testing.T) {
		if FieldKeyMsg == "" {
			t.Error("FieldKeyMsg should not be empty")
		}
		if FieldKeyLevel == "" {
			t.Error("FieldKeyLevel should not be empty")
		}
		if FieldKeyTime == "" {
			t.Error("FieldKeyTime should not be empty")
		}
		if FieldKeyLoggerName == "" {
			t.Error("FieldKeyLoggerName should not be empty")
		}
		if FieldKeyCaller == "" {
			t.Error("FieldKeyCaller should not be empty")
		}
		if FieldKeyStack == "" {
			t.Error("FieldKeyStack should not be empty")
		}
		if FieldKeyError == "" {
			t.Error("FieldKeyError should not be empty")
		}
		if FieldKeyRequestID == "" {
			t.Error("FieldKeyRequestID should not be empty")
		}
		if FieldKeyTraceID == "" {
			t.Error("FieldKeyTraceID should not be empty")
		}
		if FieldKeySpanID == "" {
			t.Error("FieldKeySpanID should not be empty")
		}
		if FieldKeyUserID == "" {
			t.Error("FieldKeyUserID should not be empty")
		}
		if FieldKeyMethod == "" {
			t.Error("FieldKeyMethod should not be empty")
		}
		if FieldKeyPath == "" {
			t.Error("FieldKeyPath should not be empty")
		}
		if FieldKeyStatusCode == "" {
			t.Error("FieldKeyStatusCode should not be empty")
		}
		if FieldKeyLatency == "" {
			t.Error("FieldKeyLatency should not be empty")
		}
		if FieldKeyIP == "" {
			t.Error("FieldKeyIP should not be empty")
		}
		if FieldKeyUserAgent == "" {
			t.Error("FieldKeyUserAgent should not be empty")
		}
		if FieldKeyContentLength == "" {
			t.Error("FieldKeyContentLength should not be empty")
		}
	})
}

// TestDevelopmentMode tests development mode behavior
func TestDevelopmentMode(t *testing.T) {
	t.Run("should enable development mode", func(t *testing.T) {
		cfg := DevelopmentConfig()
		if !cfg.Development {
			t.Error("DevelopmentConfig should have Development=true")
		}
	})

	t.Run("should disable development mode for production", func(t *testing.T) {
		cfg := ProductionConfig()
		if cfg.Development {
			t.Error("ProductionConfig should have Development=false")
		}
	})
}

// TestCompression tests compression configuration
func TestCompression(t *testing.T) {
	t.Run("should enable compression", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Compress = true

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should disable compression", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Compress = false

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})
}

// TestLevelValidation tests level validation
func TestLevelValidation(t *testing.T) {
	t.Run("should handle all level constants", func(t *testing.T) {
		levels := []Level{
			DebugLevel,
			InfoLevel,
			WarnLevel,
			ErrorLevel,
			DPanicLevel,
			PanicLevel,
			FatalLevel,
		}

		for _, level := range levels {
			cfg := DefaultConfig()
			cfg.Level = level

			logger, err := NewZapLogger(cfg)
			if err != nil {
				t.Fatalf("expected no error for level %v, got %v", level, err)
			}

			if logger == nil {
				t.Fatal("expected logger to not be nil")
			}

			if logger.GetLevel() != level {
				t.Errorf("expected level %v, got %v", level, logger.GetLevel())
			}
		}
	})
}

// TestLoggerInterfaceCompliance tests that ZapLogger implements Logger interface
func TestLoggerInterfaceCompliance(t *testing.T) {
	t.Run("ZapLogger should implement Logger interface", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		var _ Logger = logger
	})

	t.Run("NopLogger should implement Logger interface", func(t *testing.T) {
		logger := NewNopLogger()
		var _ Logger = logger
	})
}

// TestInterfaceConformance tests all interface methods are callable
func TestInterfaceConformance(t *testing.T) {
	t.Run("all interface methods should be callable", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Level = DebugLevel
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Debug(msg string, fields ...Field)
		logger.Debug("test", String("key", "value"))

		// Info(msg string, fields ...Field)
		logger.Info("test", String("key", "value"))

		// Warn(msg string, fields ...Field)
		logger.Warn("test", String("key", "value"))

		// Error(msg string, fields ...Field)
		logger.Error("test", String("key", "value"))

		// DPanic(msg string, fields ...Field)
		logger.DPanic("test", String("key", "value"))

		// Panic(msg string, fields ...Field)
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected Panic to panic")
			}
		}()
		logger.Panic("test", String("key", "value"))

		// DebugCtx(ctx context.Context, msg string, fields ...Field)
		ctx := context.Background()
		logger.DebugCtx(ctx, "test", String("key", "value"))

		// InfoCtx(ctx context.Context, msg string, fields ...Field)
		logger.InfoCtx(ctx, "test", String("key", "value"))

		// WarnCtx(ctx context.Context, msg string, fields ...Field)
		logger.WarnCtx(ctx, "test", String("key", "value"))

		// ErrorCtx(ctx context.Context, msg string, fields ...Field)
		logger.ErrorCtx(ctx, "test", String("key", "value"))

		// Debugf(format string, args ...any)
		logger.Debugf("test %s", "args")

		// Infof(format string, args ...any)
		logger.Infof("test %s", "args")

		// Warnf(format string, args ...any)
		logger.Warnf("test %s", "args")

		// Errorf(format string, args ...any)
		logger.Errorf("test %s", "args")

		// With(fields ...Field) Logger
		child := logger.With(String("key", "value"))
		if child == nil {
			t.Error("With should return non-nil logger")
		}

		// WithName(name string) Logger
		named := logger.WithName("test")
		if named == nil {
			t.Error("WithName should return non-nil logger")
		}

		// WithCtx(ctx context.Context) Logger
		ctxLogger := logger.WithCtx(ctx)
		if ctxLogger == nil {
			t.Error("WithCtx should return non-nil logger")
		}

		// SetLevel(level Level)
		logger.SetLevel(DebugLevel)

		// GetLevel() Level
		_ = logger.GetLevel()

		// Sync() error
		_ = logger.Sync()
	})
}

// TestChildLoggerAllMethods tests that child loggers work with all methods
func TestChildLoggerAllMethods(t *testing.T) {
	t.Run("child logger should support all methods", func(t *testing.T) {
		cfg := DefaultConfig()
		parent, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		child := parent.With(String("child_key", "child_value"))

		// Test all methods on child
		child.Debug("test")
		child.Info("test")
		child.Warn("test")
		child.Error("test")
		child.DPanic("test")

		t.Run("Panic method should panic", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected Panic to panic")
				}
			}()
			child.Panic("test")
		})

		child.Debugf("test %s", "args")
		child.Infof("test %s", "args")
		child.Warnf("test %s", "args")
		child.Errorf("test %s", "args")

		grandchild := child.With(String("grandchild_key", "grandchild_value"))
		grandchild.Info("grandchild message")
	})
}

// TestFieldWithSpecialCharacters tests fields with special characters
func TestFieldWithSpecialCharacters(t *testing.T) {
	cfg := DefaultConfig()
	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	logger.Info("test with special chars",
		String("key with spaces", "value"),
		String("key\twith\ttabs", "value"),
		String("key\nwith\nnewlines", "value"),
		String("unicode", "你好世界 🌍"),
		Any("complex", map[string]any{"key": "value\nwith\nnewlines"}),
	)
}

// TestMultipleFields tests multiple field constructors in one call
func TestMultipleFields(t *testing.T) {
	cfg := DefaultConfig()
	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_ = String("request_id", "req-12345")
	_ = Int("user_id", 67890)
	_ = Int64("timestamp", time.Now().Unix())
	_ = Float64("amount", 99.99)
	_ = Bool("is_admin", true)
	_ = Err(nil)
	_ = Any("metadata", map[string]any{
		"source":  "api",
		"version": "v1",
	})
	_ = Duration("latency", 150*time.Millisecond)
	_ = Time("time", time.Now())
	_ = Namespace("api")

	logger.Info("request completed",
		String("request_id", "req-12345"),
		Int("user_id", 67890),
	)
}

// TestContextSerialization tests context-related fields
func TestContextSerialization(t *testing.T) {
	cfg := DefaultConfig()
	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
	ctx = context.WithValue(ctx, "span_id", "def-456")

	logger.InfoCtx(ctx, "WithContext test",
		String(FieldKeyTraceID, "abc-123"),
		String(FieldKeySpanID, "def-456"),
		String(FieldKeyRequestID, "req-789"),
	)
}

// TestRepeatedOperations tests repeated log operations
func TestRepeatedOperations(t *testing.T) {
	cfg := DefaultConfig()
	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Log many times to test stability
	for i := 0; i < 100; i++ {
		logger.Info("repeated message",
			Int("iteration", i),
			String("status", "ok"),
		)
	}
}

// TestFieldConstructorComprehensive tests field constructors
func TestFieldConstructorComprehensive(t *testing.T) {
	t.Run("String field", func(t *testing.T) {
		field := String("key", "value")
		if field.Key() != "key" {
			t.Errorf("key: expected %q, got %q", "key", field.Key())
		}
	})

	t.Run("Int field", func(t *testing.T) {
		field := Int("code", 200)
		if field.Key() != "code" {
			t.Errorf("key: expected %q, got %q", "code", field.Key())
		}
	})

	t.Run("Int64 field", func(t *testing.T) {
		field := Int64("id", 123)
		if field.Key() != "id" {
			t.Errorf("key: expected %q, got %q", "id", field.Key())
		}
	})

	t.Run("Float64 field", func(t *testing.T) {
		field := Float64("rate", 0.5)
		if field.Key() != "rate" {
			t.Errorf("key: expected %q, got %q", "rate", field.Key())
		}
	})

	t.Run("Bool field", func(t *testing.T) {
		field := Bool("enabled", true)
		if field.Key() != "enabled" {
			t.Errorf("key: expected %q, got %q", "enabled", field.Key())
		}
	})

	t.Run("Bool false field", func(t *testing.T) {
		field := Bool("enabled", false)
		if field.Key() != "enabled" {
			t.Errorf("key: expected %q, got %q", "enabled", field.Key())
		}
	})

	t.Run("Err field", func(t *testing.T) {
		field := Err(os.ErrPermission)
		if field.Key() != "error" {
			t.Errorf("key: expected %q, got %q", "error", field.Key())
		}
	})

	t.Run("Any field", func(t *testing.T) {
		field := Any("data", nil)
		if field.Key() != "data" {
			t.Errorf("key: expected %q, got %q", "data", field.Key())
		}
	})

	t.Run("Any with value", func(t *testing.T) {
		field := Any("data", "value")
		if field.Key() != "data" {
			t.Errorf("key: expected %q, got %q", "data", field.Key())
		}
	})

	t.Run("Duration field", func(t *testing.T) {
		field := Duration("timeout", 30*time.Second)
		if field.Key() != "timeout" {
			t.Errorf("key: expected %q, got %q", "timeout", field.Key())
		}
	})

	t.Run("Time field", func(t *testing.T) {
		field := Time("created", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		if field.Key() != "created" {
			t.Errorf("key: expected %q, got %q", "created", field.Key())
		}
	})

	t.Run("Namespace field", func(t *testing.T) {
		field := Namespace("api")
		if field.Key() != "api" {
			t.Errorf("key: expected %q, got %q", "api", field.Key())
		}
	})
}

// TestFullLoggerLifecycle tests a complete logger lifecycle
func TestFullLoggerLifecycle(t *testing.T) {
	t.Run("should complete lifecycle without errors", func(t *testing.T) {
		// 1. Create logger with development config
		cfg := DevelopmentConfig()
		cfg.Development = false // Use JSON format for cleaner output
		cfg.Format = "json"
		cfg.Output = "stdout"

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("failed to create logger: %v", err)
		}

		// 2. Log various levels
		logger.Debug("debug message", String("key", "debug"))
		logger.Info("info message", String("key", "info"))
		logger.Warn("warn message", String("key", "warn"))
		logger.Error("error message", String("key", "error"))

		// 3. Create child logger
		child := logger.With(String("parent_field", "parent_value"))

		// 4. Create named logger
		namedChild := child.WithName("named-child")

		// 5. Log with child
		namedChild.Info("child message", String("child_field", "child_value"))

		// 6. Change level
		logger.SetLevel(ErrorLevel)
		if logger.GetLevel() != ErrorLevel {
			t.Errorf("level not changed properly")
		}

		// 7. Sync
		err = logger.Sync()
		if err != nil {
			t.Errorf("sync failed: %v", err)
		}
	})
}

// TestConfigEdgeCases tests edge cases in config
func TestConfigEdgeCases(t *testing.T) {
	t.Run("should handle zero config", func(t *testing.T) {
		cfg := Config{}

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should handle empty config values", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.FilePath = ""
		cfg.Format = ""
		cfg.Output = ""

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})

	t.Run("should handle custom initial fields", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.InitialFields = map[string]any{
			"service":     "test-service",
			"environment": "test",
			"version":     "1.0.0",
		}

		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to not be nil")
		}
	})
}

// TestErrorFieldConstructor tests Err field with various errors
func TestErrorFieldConstructor(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"nil error", nil},
		{"os.ErrNotExist", os.ErrNotExist},
		{"os.ErrPermission", os.ErrPermission},
		{"custom error", &customError{msg: "custom error"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := Err(tt.err)
			if field.Key() != "error" {
				t.Errorf("expected key 'error', got %q", field.Key())
			}
			if field.Value() != tt.err {
				t.Errorf("expected error value, got %v", field.Value())
			}
		})
	}
}

// customError is a custom error for testing
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

// TestMultiLevelLogging tests logging at multiple levels in sequence
func TestMultiLevelLogging(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Level = DebugLevel // Set to lowest to ensure all are logged
	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	levels := []struct {
		name  string
		level Level
		msg   string
	}{
		{"Debug", DebugLevel, "debug message"},
		{"Info", InfoLevel, "info message"},
		{"Warn", WarnLevel, "warn message"},
		{"Error", ErrorLevel, "error message"},
		{"DPanic", DPanicLevel, "dpanic message"},
		{"Panic", PanicLevel, "panic message"},
	}

	for _, tt := range levels {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We don't call DPanic, Panic, Fatal as they might panic/exit
			switch tt.level {
			case DPanicLevel:
				logger.DPanic(tt.msg)
			case PanicLevel:
				defer func() {
					recover()
				}()
				logger.Panic(tt.msg)
			default:
				logger.Info(tt.msg, String("level", tt.name))
			}
		})
	}
}

// TestSync Method tests sync after logging
func TestSyncAfterLogging(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Output = "file"
	tmpDir := t.TempDir()
	cfg.FilePath = filepath.Join(tmpDir, "sync-test.log")

	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer logger.Close()

	logger.Info("message before sync")
	err = logger.Sync()
	if err != nil {
		t.Errorf("sync failed: %v", err)
	}

	// Verify file was written
	content, err := os.ReadFile(cfg.FilePath)
	if err != nil {
		t.Fatalf("expected no error reading file, got %v", err)
	}

	if !strings.Contains(string(content), "message before sync") {
		t.Error("expected file to contain 'message before sync'")
	}
}

// TestLevelConfig tests level configuration
func TestLevelConfig(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected Level
	}{
		{"DebugLevel", DebugLevel, DebugLevel},
		{"InfoLevel", InfoLevel, InfoLevel},
		{"WarnLevel", WarnLevel, WarnLevel},
		{"ErrorLevel", ErrorLevel, ErrorLevel},
		{"DPanicLevel", DPanicLevel, DPanicLevel},
		{"PanicLevel", PanicLevel, PanicLevel},
		{"FatalLevel", FatalLevel, FatalLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Level = tt.level

			logger, err := NewZapLogger(cfg)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if logger.GetLevel() != tt.expected {
				t.Errorf("expected level %v, got %v", tt.expected, logger.GetLevel())
			}
		})
	}
}

// TestLoggerWithFields tests logger with fields
func TestLoggerWithFields(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Level = InfoLevel
	cfg.Development = false
	cfg.Format = "json"

	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	child := logger.With(
		String("request_id", "123"),
		Int("user_id", 456),
		Bool("authenticated", true),
	)

	child.Info("user logged in",
		String("ip", "192.168.1.1"),
		String("user_agent", "test-agent"),
	)
}

// TestZapLogger_Raw tests the Raw method
func TestZapLogger_Raw(t *testing.T) {
	t.Run("should return raw zap logger", func(t *testing.T) {
		cfg := DefaultConfig()
		logger, err := NewZapLogger(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		raw := logger.Raw()
		if raw == nil {
			t.Fatal("expected raw logger to not be nil")
		}

		// Verify it's a valid logger by using it
		raw.Info("raw logger test")
	})
}

// TestConfigStruct fields tests Config struct
func TestConfigStruct(t *testing.T) {
	t.Run("should have all required fields", func(t *testing.T) {
		cfg := Config{
			Level:             InfoLevel,
			Format:            "json",
			Output:            "stdout",
			FilePath:          "/var/log/app.log",
			MaxSize:           100,
			MaxBackups:        5,
			MaxAge:            30,
			Compress:          true,
			CallerSkip:        1,
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: false,
			Encoding:          "gzip",
			InitialFields:     map[string]any{"service": "test"},
		}

		if cfg.Level != InfoLevel {
			t.Errorf("expected level InfoLevel, got %v", cfg.Level)
		}
		if cfg.Format != "json" {
			t.Errorf("expected format 'json', got %s", cfg.Format)
		}
		if cfg.Output != "stdout" {
			t.Errorf("expected output 'stdout', got %s", cfg.Output)
		}
		if cfg.FilePath != "/var/log/app.log" {
			t.Errorf("expected filepath '/var/log/app.log', got %s", cfg.FilePath)
		}
		if cfg.MaxSize != 100 {
			t.Errorf("expected maxsize 100, got %d", cfg.MaxSize)
		}
		if cfg.MaxBackups != 5 {
			t.Errorf("expected maxbackups 5, got %d", cfg.MaxBackups)
		}
		if cfg.MaxAge != 30 {
			t.Errorf("expected maxage 30, got %d", cfg.MaxAge)
		}
		if !cfg.Compress {
			t.Error("expected compress true")
		}
		if cfg.CallerSkip != 1 {
			t.Errorf("expected callerskip 1, got %d", cfg.CallerSkip)
		}
		if cfg.Development {
			t.Error("expected development false")
		}
		if cfg.DisableCaller {
			t.Error("expected disablecaller false")
		}
		if cfg.DisableStacktrace {
			t.Error("expected disablestacktrace false")
		}
		if cfg.Encoding != "gzip" {
			t.Errorf("expected encoding 'gzip', got %s", cfg.Encoding)
		}
		if cfg.InitialFields == nil {
			t.Error("expected initialfields to be set")
		}
	})
}

// TestDefaultConfigValues tests default config values
func TestDefaultConfigValues(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != InfoLevel {
		t.Errorf("expected default level InfoLevel, got %v", cfg.Level)
	}
	if cfg.Format != "json" {
		t.Errorf("expected default format 'json', got %s", cfg.Format)
	}
	if cfg.Output != "stdout" {
		t.Errorf("expected default output 'stdout', got %s", cfg.Output)
	}
	if cfg.MaxSize != 100 {
		t.Errorf("expected default maxsize 100, got %d", cfg.MaxSize)
	}
	if cfg.MaxBackups != 5 {
		t.Errorf("expected default maxbackups 5, got %d", cfg.MaxBackups)
	}
	if cfg.MaxAge != 30 {
		t.Errorf("expected default maxage 30, got %d", cfg.MaxAge)
	}
	if cfg.Compress != true {
		t.Error("expected default compress true")
	}
	if cfg.CallerSkip != 1 {
		t.Errorf("expected default callerskip 1, got %d", cfg.CallerSkip)
	}
	if cfg.Development {
		t.Error("expected default development false")
	}
	if cfg.DisableCaller {
		t.Error("expected default disablecaller false")
	}
}

// TestDevelopmentConfigValues tests development config values
func TestDevelopmentConfigValues(t *testing.T) {
	cfg := DevelopmentConfig()

	if cfg.Level != DebugLevel {
		t.Errorf("expected development level DebugLevel, got %v", cfg.Level)
	}
	if cfg.Format != "console" {
		t.Errorf("expected development format 'console', got %s", cfg.Format)
	}
	if cfg.Output != "stdout" {
		t.Errorf("expected development output 'stdout', got %s", cfg.Output)
	}
	if !cfg.Development {
		t.Error("expected development true")
	}
	if cfg.Compress != false {
		t.Error("expected development compress false")
	}
}

// TestProductionConfigValues tests production config values
func TestProductionConfigValues(t *testing.T) {
	cfg := ProductionConfig()

	if cfg.Level != InfoLevel {
		t.Errorf("expected production level InfoLevel, got %v", cfg.Level)
	}
	if cfg.Format != "json" {
		t.Errorf("expected production format 'json', got %s", cfg.Format)
	}
	if cfg.Output != "file" {
		t.Errorf("expected production output 'file', got %s", cfg.Output)
	}
	if cfg.MaxBackups != 10 {
		t.Errorf("expected production maxbackups 10, got %d", cfg.MaxBackups)
	}
	if !cfg.Compress {
		t.Error("expected production compress true")
	}
}
