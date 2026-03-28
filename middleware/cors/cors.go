package cors

import (
	"fmt"
	"net/http"

	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

func DefaultConfig() *Config {
	return &Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

func New(cfg *Config) abstract.Middleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		origin := ctx.Header("Origin")
		if origin == "" {
			return next()
		}

		allowed := false
		for _, o := range cfg.AllowOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if !allowed {
			return next()
		}

		w := ctx.ResponseWriter()

		w.Header().Set("Access-Control-Allow-Origin", origin)

		if cfg.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if len(cfg.ExposeHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", joinStrings(cfg.ExposeHeaders, ", "))
		}

		if ctx.Method() == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", joinStrings(cfg.AllowMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", joinStrings(cfg.AllowHeaders, ", "))
			if cfg.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.MaxAge))
			}
			w.WriteHeader(http.StatusNoContent)
			return nil
		}

		return next()
	})
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
