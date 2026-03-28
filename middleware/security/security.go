package security

import (
	"fmt"

	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	XSSProtection         bool
	ContentTypeNosniff    bool
	XFrameOptions         string
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	ContentSecurityPolicy string
	ReferrerPolicy        string
	PermissionsPolicy     string
}

func DefaultConfig() *Config {
	return &Config{
		XSSProtection:         true,
		ContentTypeNosniff:    true,
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

func DevelopmentConfig() *Config {
	return &Config{
		XSSProtection:      true,
		ContentTypeNosniff: true,
		XFrameOptions:      "SAMEORIGIN",
	}
}

func New(cfg *Config) abstract.Middleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		w := ctx.ResponseWriter()

		if cfg.XSSProtection {
			w.Header().Set("X-XSS-Protection", "1; mode=block")
		}

		if cfg.ContentTypeNosniff {
			w.Header().Set("X-Content-Type-Options", "nosniff")
		}

		if cfg.XFrameOptions != "" {
			w.Header().Set("X-Frame-Options", cfg.XFrameOptions)
		}

		if cfg.HSTSMaxAge > 0 {
			hsts := fmt.Sprintf("max-age=%d", cfg.HSTSMaxAge)
			if cfg.HSTSIncludeSubdomains {
				hsts += "; includeSubDomains"
			}
			w.Header().Set("Strict-Transport-Security", hsts)
		}

		if cfg.ContentSecurityPolicy != "" {
			w.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
		}

		if cfg.ReferrerPolicy != "" {
			w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
		}

		if cfg.PermissionsPolicy != "" {
			w.Header().Set("Permissions-Policy", cfg.PermissionsPolicy)
		}

		return next()
	})
}
