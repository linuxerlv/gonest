package requestid

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	HeaderName string
	Generator  func() string
}

func DefaultConfig() *Config {
	return &Config{
		HeaderName: "X-Request-ID",
		Generator:  generateRequestID,
	}
}

func New(cfg *Config) abstract.Middleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-Request-ID"
	}
	if cfg.Generator == nil {
		cfg.Generator = generateRequestID
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		requestID := ctx.Header(cfg.HeaderName)
		if requestID == "" {
			requestID = cfg.Generator()
		}
		ctx.Set("request-id", requestID)

		hc := ctx.(*core.HttpContext)
		hc.ResponseWriter().Header().Set(cfg.HeaderName, requestID)

		return next()
	})
}

func generateRequestID() string {
	return uuid.New().String()
}

func generateRequestIDLegacy() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

func GetRequestID(ctx abstract.Context) string {
	if id, ok := ctx.Get("request-id").(string); ok {
		return id
	}
	return ""
}
