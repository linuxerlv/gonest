package logger

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	SkipPaths  []string
	Formatter  func(ctx abstract.Context, latency time.Duration) string
	Output     io.Writer
	TimeFormat string
}

func DefaultConfig() *Config {
	return &Config{
		SkipPaths:  []string{},
		TimeFormat: "2006/01/02 - 15:04:05",
	}
}

func New(cfg *Config) abstract.Middleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		start := time.Now()
		err := next()
		latency := time.Since(start)

		path := ctx.Path()
		for _, skip := range cfg.SkipPaths {
			if skip == path {
				return err
			}
		}

		if cfg.Formatter != nil {
			log.Print(cfg.Formatter(ctx, latency))
		} else {
			status := http.StatusOK
			if hc, ok := ctx.(*core.HttpContext); ok {
				if w, ok := hc.ResponseWriter().(interface{ Status() int }); ok {
					status = w.Status()
				}
			}
			log.Printf("[HTTP] %s %s %d %v", ctx.Method(), path, status, latency)
		}

		return err
	})
}

type ContextLogger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
}

func NewWithLogger(log ContextLogger, cfg *Config) abstract.Middleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		start := time.Now()
		err := next()
		latency := time.Since(start)

		path := ctx.Path()
		for _, skip := range cfg.SkipPaths {
			if skip == path {
				return err
			}
		}

		status := http.StatusOK
		if hc, ok := ctx.(*core.HttpContext); ok {
			if w, ok := hc.ResponseWriter().(interface{ Status() int }); ok {
				status = w.Status()
			}
		}

		if err != nil {
			log.Error("request error",
				"method", ctx.Method(),
				"path", path,
				"status", status,
				"latency", latency,
				"error", err,
				"request_id", ctx.Get("request-id"),
			)
		} else {
			log.Info("request",
				"method", ctx.Method(),
				"path", path,
				"status", status,
				"latency", latency,
				"request_id", ctx.Get("request-id"),
			)
		}

		return err
	})
}
