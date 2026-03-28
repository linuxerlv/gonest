package extensions

import (
	"time"

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

type CORSMiddlewareOptions struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

type RecoveryMiddlewareOptions struct {
	PrintStack bool
}

type LoggingMiddlewareOptions struct {
	SkipPaths []string
}

type RateLimitMiddlewareOptions struct {
	Limit  int
	Window int
}

type GzipMiddlewareOptions struct {
	Level int
}

type SecurityMiddlewareOptions struct {
	XSSProtection         bool
	ContentTypeNosniff    bool
	XFrameOptions         string
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
}

type RequestIDMiddlewareOptions struct {
	HeaderName string
}

type TimeoutMiddlewareOptions struct {
	Timeout int
}

func UseCORS(app abstract.WebApplication, options *CORSMiddlewareOptions) abstract.WebApplication {
	cfg := convertCORSOptions(options)
	app.Use(cors.New(cfg))
	return app
}

func UseRecovery(app abstract.WebApplication, options *RecoveryMiddlewareOptions) abstract.WebApplication {
	cfg := convertRecoveryOptions(options)
	app.Use(recovery.New(cfg))
	return app
}

func UseLogging(app abstract.WebApplication, options *LoggingMiddlewareOptions) abstract.WebApplication {
	cfg := convertLoggingOptions(options)
	app.Use(logger.New(cfg))
	return app
}

func UseRateLimit(app abstract.WebApplication, options *RateLimitMiddlewareOptions) abstract.WebApplication {
	cfg := convertRateLimitOptions(options)
	app.Use(ratelimit.New(cfg))
	return app
}

func UseGzip(app abstract.WebApplication, options *GzipMiddlewareOptions) abstract.WebApplication {
	cfg := convertGzipOptions(options)
	app.Use(gzip.New(cfg))
	return app
}

func UseSecurity(app abstract.WebApplication, options *SecurityMiddlewareOptions) abstract.WebApplication {
	cfg := convertSecurityOptions(options)
	app.Use(security.New(cfg))
	return app
}

func UseRequestID(app abstract.WebApplication, options *RequestIDMiddlewareOptions) abstract.WebApplication {
	cfg := convertRequestIDOptions(options)
	app.Use(requestid.New(cfg))
	return app
}

func UseTimeout(app abstract.WebApplication, options *TimeoutMiddlewareOptions) abstract.WebApplication {
	cfg := convertTimeoutOptions(options)
	app.Use(timeout.New(cfg))
	return app
}

func convertCORSOptions(options *CORSMiddlewareOptions) *cors.Config {
	if options == nil {
		return cors.DefaultConfig()
	}
	return &cors.Config{
		AllowOrigins:     options.AllowOrigins,
		AllowMethods:     options.AllowMethods,
		AllowHeaders:     options.AllowHeaders,
		ExposeHeaders:    options.ExposeHeaders,
		AllowCredentials: options.AllowCredentials,
		MaxAge:           options.MaxAge,
	}
}

func convertRecoveryOptions(options *RecoveryMiddlewareOptions) *recovery.Config {
	if options == nil {
		return recovery.DefaultConfig()
	}
	return &recovery.Config{
		PrintStack: options.PrintStack,
	}
}

func convertLoggingOptions(options *LoggingMiddlewareOptions) *logger.Config {
	if options == nil {
		return logger.DefaultConfig()
	}
	return &logger.Config{
		SkipPaths: options.SkipPaths,
	}
}

func convertRateLimitOptions(options *RateLimitMiddlewareOptions) *ratelimit.Config {
	if options == nil {
		return ratelimit.DefaultConfig()
	}
	return &ratelimit.Config{
		Limit:  options.Limit,
		Window: time.Duration(options.Window) * time.Second,
	}
}

func convertGzipOptions(options *GzipMiddlewareOptions) *gzip.Config {
	if options == nil {
		return gzip.DefaultConfig()
	}
	return &gzip.Config{
		Level: options.Level,
	}
}

func convertSecurityOptions(options *SecurityMiddlewareOptions) *security.Config {
	if options == nil {
		return security.DefaultConfig()
	}
	return &security.Config{
		XSSProtection:         options.XSSProtection,
		ContentTypeNosniff:    options.ContentTypeNosniff,
		XFrameOptions:         options.XFrameOptions,
		HSTSMaxAge:            options.HSTSMaxAge,
		HSTSIncludeSubdomains: options.HSTSIncludeSubdomains,
	}
}

func convertRequestIDOptions(options *RequestIDMiddlewareOptions) *requestid.Config {
	if options == nil {
		return requestid.DefaultConfig()
	}
	return &requestid.Config{
		HeaderName: options.HeaderName,
	}
}

func convertTimeoutOptions(options *TimeoutMiddlewareOptions) *timeout.Config {
	if options == nil {
		return timeout.DefaultConfig()
	}
	return &timeout.Config{
		Timeout: time.Duration(options.Timeout) * time.Second,
	}
}

func UseMiddleware(app abstract.WebApplication, middleware abstract.Middleware) abstract.WebApplication {
	app.Use(middleware)
	return app
}
