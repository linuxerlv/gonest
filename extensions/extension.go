package extensions

import (
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/cors"
	"github.com/linuxerlv/gonest/middleware/gzip"
	"github.com/linuxerlv/gonest/middleware/logger"
	"github.com/linuxerlv/gonest/middleware/ratelimit"
	"github.com/linuxerlv/gonest/middleware/recovery"
	"github.com/linuxerlv/gonest/middleware/requestid"
	"github.com/linuxerlv/gonest/middleware/security"
	"github.com/linuxerlv/gonest/middleware/timeout"
)

type WebAppExtensible interface {
	abstract.WebApplication
	UseCORS(options *CORSMiddlewareOptions) WebAppExtensible
	UseRecovery(options *RecoveryMiddlewareOptions) WebAppExtensible
	UseLogging(options *LoggingMiddlewareOptions) WebAppExtensible
	UseRateLimit(options *RateLimitMiddlewareOptions) WebAppExtensible
	UseGzip(options *GzipMiddlewareOptions) WebAppExtensible
	UseSecurity(options *SecurityMiddlewareOptions) WebAppExtensible
	UseRequestID(options *RequestIDMiddlewareOptions) WebAppExtensible
	UseTimeout(options *TimeoutMiddlewareOptions) WebAppExtensible
}

type WebAppExtension struct {
	abstract.WebApplication
}

func Extend(app abstract.WebApplication) WebAppExtensible {
	return &WebAppExtension{WebApplication: app}
}

func (e *WebAppExtension) UseCORS(options *CORSMiddlewareOptions) WebAppExtensible {
	cfg := convertCORSOptions(options)
	e.WebApplication.Use(cors.New(cfg))
	return e
}

func (e *WebAppExtension) UseRecovery(options *RecoveryMiddlewareOptions) WebAppExtensible {
	cfg := convertRecoveryOptions(options)
	e.WebApplication.Use(recovery.New(cfg))
	return e
}

func (e *WebAppExtension) UseLogging(options *LoggingMiddlewareOptions) WebAppExtensible {
	cfg := convertLoggingOptions(options)
	e.WebApplication.Use(logger.New(cfg))
	return e
}

func (e *WebAppExtension) UseRateLimit(options *RateLimitMiddlewareOptions) WebAppExtensible {
	cfg := convertRateLimitOptions(options)
	e.WebApplication.Use(ratelimit.New(cfg))
	return e
}

func (e *WebAppExtension) UseGzip(options *GzipMiddlewareOptions) WebAppExtensible {
	cfg := convertGzipOptions(options)
	e.WebApplication.Use(gzip.New(cfg))
	return e
}

func (e *WebAppExtension) UseSecurity(options *SecurityMiddlewareOptions) WebAppExtensible {
	cfg := convertSecurityOptions(options)
	e.WebApplication.Use(security.New(cfg))
	return e
}

func (e *WebAppExtension) UseRequestID(options *RequestIDMiddlewareOptions) WebAppExtensible {
	cfg := convertRequestIDOptions(options)
	e.WebApplication.Use(requestid.New(cfg))
	return e
}

func (e *WebAppExtension) UseTimeout(options *TimeoutMiddlewareOptions) WebAppExtensible {
	cfg := convertTimeoutOptions(options)
	e.WebApplication.Use(timeout.New(cfg))
	return e
}
