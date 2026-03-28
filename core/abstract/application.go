package abstract

import (
	"time"
)

// MiddlewareApp 中间件应用接口
type MiddlewareApp interface {
	UseCORS(cfg *CORSConfig) Application
	UseRecovery() Application
	UseRequestID(headerName string) Application
	UseRateLimit(limit int, window time.Duration) Application
	UseGzip(level int) Application
	UseSecurity(cfg *SecurityConfig) Application
	UseTimeout(timeout time.Duration) Application
	UseScope() Application
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	XSSProtection         bool
	ContentTypeNosniff    bool
	XFrameOptions         string
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	ContentSecurityPolicy string
	ReferrerPolicy        string
	PermissionsPolicy     string
}

// LoggerMiddlewareConfig 日志中间件配置
type LoggerMiddlewareConfig struct {
	SkipPaths []string
	Formatter func(ctx Context, latency time.Duration) string
}
