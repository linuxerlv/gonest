package abstract

import (
	"context"
	"net/http"
	"time"
)

// ApplicationAbstract 应用接口
type ApplicationAbstract interface {
	RouterAbstract
	Use(middleware MiddlewareAbstract) ApplicationAbstract
	UseGlobalGuards(guards ...GuardAbstract) ApplicationAbstract
	UseGlobalInterceptors(interceptors ...InterceptorAbstract) ApplicationAbstract
	UseGlobalPipes(pipes ...PipeAbstract) ApplicationAbstract
	UseGlobalFilters(filters ...ExceptionFilterAbstract) ApplicationAbstract
	Controller(controller ControllerAbstract) ApplicationAbstract
	Listen(addr string) error
	Shutdown(ctx context.Context) error
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// MiddlewareAppAbstract 中间件应用接口
type MiddlewareAppAbstract interface {
	UseCORS(cfg *CORSConfigAbstract) ApplicationAbstract
	UseRecovery() ApplicationAbstract
	UseRequestID(headerName string) ApplicationAbstract
	UseRateLimit(limit int, window time.Duration) ApplicationAbstract
	UseGzip(level int) ApplicationAbstract
	UseSecurity(cfg *SecurityConfigAbstract) ApplicationAbstract
	UseTimeout(timeout time.Duration) ApplicationAbstract
	UseScope() ApplicationAbstract
}

// CORSConfigAbstract CORS配置
type CORSConfigAbstract struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// SecurityConfigAbstract 安全配置
type SecurityConfigAbstract struct {
	XSSProtection         bool
	ContentTypeNosniff    bool
	XFrameOptions         string
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	ContentSecurityPolicy string
	ReferrerPolicy        string
	PermissionsPolicy     string
}

// LoggerMiddlewareConfigAbstract 日志中间件配置
type LoggerMiddlewareConfigAbstract struct {
	SkipPaths []string
	Formatter func(ctx ContextAbstract, latency time.Duration) string
}
