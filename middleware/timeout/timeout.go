package timeout

import (
	"context"
	"net/http"
	"time"

	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	Timeout   time.Duration
	ErrorCode int
	ErrorMsg  string
}

func DefaultConfig() *Config {
	return &Config{
		Timeout:   30 * time.Second,
		ErrorCode: http.StatusRequestTimeout,
		ErrorMsg:  "request timeout",
	}
}

func New(cfg *Config) abstract.Middleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		done := make(chan error, 1)

		go func() {
			done <- next()
		}()

		select {
		case err := <-done:
			return err
		case <-time.After(cfg.Timeout):
			return abstract.NewHttpException(cfg.ErrorCode, cfg.ErrorMsg)
		}
	})
}

func NewWithContext(cfg *Config) abstract.Middleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		c, cancel := context.WithTimeout(ctx.Context(), cfg.Timeout)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- next()
		}()

		select {
		case err := <-done:
			return err
		case <-c.Done():
			return abstract.NewHttpException(cfg.ErrorCode, cfg.ErrorMsg)
		}
	})
}
