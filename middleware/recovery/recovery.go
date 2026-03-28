package recovery

import (
	"fmt"
	"log"

	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	PrintStack bool
}

func DefaultConfig() *Config {
	return &Config{
		PrintStack: true,
	}
}

func New(cfg *Config) abstract.Middleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v recovered at %s", r, ctx.Path())
				if cfg.PrintStack {
				}
				err = abstract.InternalError(fmt.Sprintf("internal server error: %v", r))
			}
		}()
		return next()
	})
}
