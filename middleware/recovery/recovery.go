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

func New(cfg *Config) abstract.MiddlewareAbstract {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v recovered at %s", r, ctx.Path())
				if cfg.PrintStack {
					// stack printing logic can be added here
				}
				err = abstract.InternalError(fmt.Sprintf("internal server error: %v", r))
			}
		}()
		return next()
	})
}
