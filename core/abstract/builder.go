package abstract

import (
	"context"
	"net/http"
)

// ApplicationBuilder 通用应用构建器接口（ASP.NET Core风格）
type ApplicationBuilder interface {
	Services() ServiceCollection
	Configuration() Config
	Environment() Env
	Logging() Logger
	Build() Application
}

// Application 通用应用接口
type Application interface {
	Services() ServiceCollection
	Configuration() Config
	Environment() Env
	Logging() Logger
	Run() error
	RunAsync() <-chan error
	Start() error
	StartAsync() <-chan error
	Stop() error
	Shutdown(ctx context.Context) error
	WaitForShutdown() error
}

// WebApplicationBuilder Web应用构建器接口
type WebApplicationBuilder interface {
	Services() ServiceCollection
	Configuration() Config
	Environment() Env
	Logging() Logger
	WebHost() WebHostBuilder
	Build() WebApplication
}

// WebApplication Web应用接口
type WebApplication interface {
	Application
	Router
	MiddlewareUser
	Use(middleware Middleware) WebApplication
	UseGlobalGuards(guards ...Guard) WebApplication
	UseGlobalInterceptors(interceptors ...Interceptor) WebApplication
	UseGlobalPipes(pipes ...Pipe) WebApplication
	UseGlobalFilters(filters ...ExceptionFilter) WebApplication
	MapGet(path string, handler any) RouteBuilder
	MapPost(path string, handler any) RouteBuilder
	MapPut(path string, handler any) RouteBuilder
	MapDelete(path string, handler any) RouteBuilder
	MapPatch(path string, handler any) RouteBuilder
	Map(method string, path string, handler any) RouteBuilder
	MapGroup(prefix string) RouteGroup
	UseRouting() WebApplication
	UseEndpoints(configure func(EndpointRouteBuilder)) WebApplication
	UseAuthentication() WebApplication
	UseAuthorization() WebApplication
	Urls() []string
	Addresses() []string
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Listen(addr string) error
}

// MiddlewareUser 中间件使用接口
type MiddlewareUser interface {
	UseCORS() WebApplication
	UseRecovery() WebApplication
	UseLogging() WebApplication
	UseRateLimit() WebApplication
	UseGzip() WebApplication
	UseSecurity() WebApplication
	UseRequestID() WebApplication
	UseTimeout() WebApplication
}

// EndpointRouteBuilder 端点路由构建器接口
type EndpointRouteBuilder interface {
	MapGet(path string, handler any) RouteBuilder
	MapPost(path string, handler any) RouteBuilder
	MapPut(path string, handler any) RouteBuilder
	MapDelete(path string, handler any) RouteBuilder
	MapPatch(path string, handler any) RouteBuilder
	MapControllers() EndpointRouteBuilder
	MapControllerRoute(name string, pattern string, defaults map[string]string) EndpointRouteBuilder
	MapAreaControllerRoute(name string, areaName string, pattern string, defaults map[string]string) EndpointRouteBuilder
	MapDefaultControllerRoute() EndpointRouteBuilder
}

// WebHostBuilder Web主机构建器接口
type WebHostBuilder interface {
	UseUrls(urls ...string) WebHostBuilder
	ConfigureKestrel(configure func(interface{})) WebHostBuilder
	Build() WebHost
}

// WebHost Web主机接口
type WebHost interface {
	Start() error
	Stop(ctx context.Context) error
	Addresses() []string
}

// HostBuilder 主机构建器接口
type HostBuilder interface {
	UseContentRoot(path string) HostBuilder
	UseEnvironment(env string) HostBuilder
	ContentRoot() string
	Environment() string
	Args() []string
	SetArgs(args []string)
}

// HostApplication 主机应用接口
type HostApplication interface {
	Application
}

// MapRoute 路由映射接口
type MapRoute interface {
	MapGet(path string, handler any) RouteBuilder
	MapPost(path string, handler any) RouteBuilder
	MapPut(path string, handler any) RouteBuilder
	MapDelete(path string, handler any) RouteBuilder
	MapPatch(path string, handler any) RouteBuilder
}
