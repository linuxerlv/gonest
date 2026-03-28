package core

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/linuxerlv/gonest/config"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/logger"
)

const (
	defaultShutdownTimeout = 30 * time.Second
)

type ServiceDescriptor struct {
	serviceType reflect.Type
	instance    any
	factory     any
	lifetime    abstract.ServiceLifetime
}

func (d *ServiceDescriptor) ServiceType() reflect.Type          { return d.serviceType }
func (d *ServiceDescriptor) Instance() any                      { return d.instance }
func (d *ServiceDescriptor) Factory() any                       { return d.factory }
func (d *ServiceDescriptor) Lifetime() abstract.ServiceLifetime { return d.lifetime }

type ServiceCollection struct {
	descriptors map[reflect.Type]*ServiceDescriptor
	instances   map[reflect.Type]any
	mu          sync.RWMutex
}

func NewServiceCollection() *ServiceCollection {
	return &ServiceCollection{
		descriptors: make(map[reflect.Type]*ServiceDescriptor),
		instances:   make(map[reflect.Type]any),
	}
}

func (s *ServiceCollection) AddSingleton(instance any) abstract.ServiceRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	serviceType := reflect.TypeOf(instance)
	s.descriptors[serviceType] = &ServiceDescriptor{
		serviceType: serviceType,
		instance:    instance,
		lifetime:    abstract.Singleton,
	}
	return s
}

func (s *ServiceCollection) AddSingletonFactory(serviceType reflect.Type, factory any) abstract.ServiceRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[serviceType] = &ServiceDescriptor{
		serviceType: serviceType,
		factory:     factory,
		lifetime:    abstract.Singleton,
	}
	return s
}

func (s *ServiceCollection) AddScoped(serviceType reflect.Type, factory any) abstract.ServiceRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[serviceType] = &ServiceDescriptor{
		serviceType: serviceType,
		factory:     factory,
		lifetime:    abstract.Scoped,
	}
	return s
}

func (s *ServiceCollection) AddTransient(serviceType reflect.Type, factory any) abstract.ServiceRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[serviceType] = &ServiceDescriptor{
		serviceType: serviceType,
		factory:     factory,
		lifetime:    abstract.Transient,
	}
	return s
}

func (s *ServiceCollection) GetService(serviceType reflect.Type) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	desc, ok := s.descriptors[serviceType]
	if !ok {
		return nil
	}
	switch desc.Lifetime() {
	case abstract.Singleton:
		if desc.Instance() != nil {
			return desc.Instance()
		}
		if desc.Factory() != nil {
			if fn, ok := desc.Factory().(func(abstract.ServiceCollection) any); ok {
				instance := fn(s)
				s.instances[serviceType] = instance
				return instance
			}
		}
	case abstract.Scoped, abstract.Transient:
		if desc.Factory() != nil {
			if fn, ok := desc.Factory().(func(abstract.ServiceCollection) any); ok {
				return fn(s)
			}
		}
		if desc.Instance() != nil {
			return desc.Instance()
		}
	}
	return nil
}

func (s *ServiceCollection) GetRequiredService(serviceType reflect.Type) any {
	service := s.GetService(serviceType)
	if service == nil {
		panic(fmt.Sprintf("service of type %v not registered", serviceType))
	}
	return service
}

func (s *ServiceCollection) AddCORS(config any) abstract.MiddlewareRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[reflect.TypeOf("cors")] = &ServiceDescriptor{
		serviceType: reflect.TypeOf("cors"),
		instance:    config,
		lifetime:    abstract.Singleton,
	}
	return s
}

func (s *ServiceCollection) AddRecovery(config any) abstract.MiddlewareRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[reflect.TypeOf("recovery")] = &ServiceDescriptor{
		serviceType: reflect.TypeOf("recovery"),
		instance:    config,
		lifetime:    abstract.Singleton,
	}
	return s
}

func (s *ServiceCollection) AddLogging(config any) abstract.MiddlewareRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[reflect.TypeOf("logging")] = &ServiceDescriptor{
		serviceType: reflect.TypeOf("logging"),
		instance:    config,
		lifetime:    abstract.Singleton,
	}
	return s
}

func (s *ServiceCollection) AddRateLimit(config any) abstract.MiddlewareRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[reflect.TypeOf("ratelimit")] = &ServiceDescriptor{
		serviceType: reflect.TypeOf("ratelimit"),
		instance:    config,
		lifetime:    abstract.Singleton,
	}
	return s
}

func (s *ServiceCollection) AddGzip(config any) abstract.MiddlewareRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[reflect.TypeOf("gzip")] = &ServiceDescriptor{
		serviceType: reflect.TypeOf("gzip"),
		instance:    config,
		lifetime:    abstract.Singleton,
	}
	return s
}

func (s *ServiceCollection) AddSecurity(config any) abstract.MiddlewareRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[reflect.TypeOf("security")] = &ServiceDescriptor{
		serviceType: reflect.TypeOf("security"),
		instance:    config,
		lifetime:    abstract.Singleton,
	}
	return s
}

func (s *ServiceCollection) AddRequestID(config any) abstract.MiddlewareRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[reflect.TypeOf("requestid")] = &ServiceDescriptor{
		serviceType: reflect.TypeOf("requestid"),
		instance:    config,
		lifetime:    abstract.Singleton,
	}
	return s
}

func (s *ServiceCollection) AddTimeout(config any) abstract.MiddlewareRegistrar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[reflect.TypeOf("timeout")] = &ServiceDescriptor{
		serviceType: reflect.TypeOf("timeout"),
		instance:    config,
		lifetime:    abstract.Singleton,
	}
	return s
}

func AddSingleton[T any](s *ServiceCollection, instance T) *ServiceCollection {
	s.mu.Lock()
	defer s.mu.Unlock()
	serviceType := reflect.TypeOf(instance)
	s.descriptors[serviceType] = &ServiceDescriptor{
		serviceType: serviceType,
		instance:    instance,
		lifetime:    abstract.Singleton,
	}
	return s
}

func AddSingletonFunc[T any](s *ServiceCollection, factory func(abstract.ServiceCollection) T) *ServiceCollection {
	s.mu.Lock()
	defer s.mu.Unlock()
	var zero T
	serviceType := reflect.TypeOf(zero)
	if serviceType == nil {
		serviceType = reflect.TypeOf((*T)(nil)).Elem()
	}
	s.descriptors[serviceType] = &ServiceDescriptor{
		serviceType: serviceType,
		factory:     factory,
		lifetime:    abstract.Singleton,
	}
	return s
}

func AddScoped[T any](s *ServiceCollection, factory func(abstract.ServiceCollection) T) *ServiceCollection {
	s.mu.Lock()
	defer s.mu.Unlock()
	var zero T
	serviceType := reflect.TypeOf(zero)
	if serviceType == nil {
		serviceType = reflect.TypeOf((*T)(nil)).Elem()
	}
	s.descriptors[serviceType] = &ServiceDescriptor{
		serviceType: serviceType,
		factory:     factory,
		lifetime:    abstract.Scoped,
	}
	return s
}

func AddTransient[T any](s *ServiceCollection, factory func(abstract.ServiceCollection) T) *ServiceCollection {
	s.mu.Lock()
	defer s.mu.Unlock()
	var zero T
	serviceType := reflect.TypeOf(zero)
	if serviceType == nil {
		serviceType = reflect.TypeOf((*T)(nil)).Elem()
	}
	s.descriptors[serviceType] = &ServiceDescriptor{
		serviceType: serviceType,
		factory:     factory,
		lifetime:    abstract.Transient,
	}
	return s
}

func GetService[T any](s abstract.ServiceCollection) T {
	var zero T
	serviceType := reflect.TypeOf(zero)
	if serviceType == nil {
		serviceType = reflect.TypeOf((*T)(nil)).Elem()
	}
	v := s.GetService(serviceType)
	if v == nil {
		return zero
	}
	return v.(T)
}

func GetRequiredService[T any](s abstract.ServiceCollection) T {
	var zero T
	serviceType := reflect.TypeOf(zero)
	if serviceType == nil {
		serviceType = reflect.TypeOf((*T)(nil)).Elem()
	}
	v := s.GetRequiredService(serviceType)
	return v.(T)
}

func IsNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	}
	return false
}

var _ abstract.ServiceCollection = (*ServiceCollection)(nil)

// ApplicationBuilder 通用应用构建器
type ApplicationBuilder struct {
	services      *ServiceCollection
	configuration config.Config
	environment   abstract.Env
	logging       logger.Logger
	mu            sync.RWMutex
}

func NewApplicationBuilder() *ApplicationBuilder {
	return &ApplicationBuilder{
		services:    NewServiceCollection(),
		environment: NewEnv(),
	}
}

func (b *ApplicationBuilder) Services() abstract.ServiceCollection {
	return b.services
}

func (b *ApplicationBuilder) Configuration() abstract.Config {
	if b.configuration != nil {
		return NewConfigAdapter(b.configuration)
	}
	return nil
}

func (b *ApplicationBuilder) Environment() abstract.Env {
	return b.environment
}

func (b *ApplicationBuilder) Logging() abstract.Logger {
	if b.logging != nil {
		return NewLoggerAdapter(b.logging)
	}
	return NewLoggerAdapter(logger.GetGlobalLogger())
}

func (b *ApplicationBuilder) UseConfig(cfg abstract.Config) *ApplicationBuilder {
	b.mu.Lock()
	if cfg != nil {
		if adapter, ok := cfg.(*ConfigAdapter); ok {
			b.configuration = adapter.Unwrap()
		}
	}
	b.mu.Unlock()
	return b
}

func (b *ApplicationBuilder) UseLogger(log abstract.Logger) *ApplicationBuilder {
	b.mu.Lock()
	if log != nil {
		if adapter, ok := log.(*LoggerAdapter); ok {
			b.logging = adapter.Unwrap()
		}
	}
	b.mu.Unlock()
	return b
}

func (b *ApplicationBuilder) Build() abstract.Application {
	app := &HostApplication{
		config:   b.configuration,
		env:      b.environment,
		services: b.services,
		logger:   b.logging,
		values:   make(map[string]any),
		stopCh:   make(chan struct{}),
		runCh:    make(chan error, 1),
	}
	if b.logging != nil {
		logger.SetGlobalLogger(b.logging)
	}
	return app
}

var _ abstract.ApplicationBuilder = (*ApplicationBuilder)(nil)

// WebApplicationBuilder Web应用构建器
type WebApplicationBuilder struct {
	*ApplicationBuilder
	Host *HostBuilder
}

func NewWebApplicationBuilder() *WebApplicationBuilder {
	return &WebApplicationBuilder{
		ApplicationBuilder: NewApplicationBuilder(),
		Host:               NewHostBuilder(),
	}
}

func (b *WebApplicationBuilder) WebHost() abstract.WebHostBuilder {
	return b.Host
}

func (b *WebApplicationBuilder) Build() abstract.WebApplication {
	app := &WebApplication{
		Application: &Application{
			config:       b.configuration,
			env:          b.environment,
			services:     b.services,
			logger:       b.logging,
			router:       NewRouter(),
			controllers:  make([]abstract.Controller, 0),
			middlewares:  make([]abstract.Middleware, 0),
			guards:       make([]abstract.Guard, 0),
			interceptors: make([]abstract.Interceptor, 0),
			pipes:        make([]abstract.Pipe, 0),
			filters:      make([]abstract.ExceptionFilter, 0),
			values:       make(map[string]any),
		},
		builder: b,
	}
	if b.logging != nil {
		logger.SetGlobalLogger(b.logging)
	}
	return app
}

func (b *WebApplicationBuilder) UseConfig(cfg abstract.Config) *WebApplicationBuilder {
	b.ApplicationBuilder.UseConfig(cfg)
	return b
}

func (b *WebApplicationBuilder) UseLogger(log abstract.Logger) *WebApplicationBuilder {
	b.ApplicationBuilder.UseLogger(log)
	return b
}

var _ abstract.WebApplicationBuilder = (*WebApplicationBuilder)(nil)

// MiddlewareMixin 中间件 Mixin
type MiddlewareMixin struct {
	app      *WebApplication
	services *ServiceCollection
}

func NewMiddlewareMixin(app *WebApplication, services *ServiceCollection) *MiddlewareMixin {
	return &MiddlewareMixin{app: app, services: services}
}

func (m *MiddlewareMixin) UseCORS() abstract.MiddlewareUser {
	cfg := m.services.GetService(reflect.TypeOf("cors"))
	if mw := createCORSMiddleware(cfg); mw != nil {
		m.app.Application.middlewares = append(m.app.Application.middlewares, mw)
	}
	return m
}

func (m *MiddlewareMixin) UseRecovery() abstract.MiddlewareUser {
	cfg := m.services.GetService(reflect.TypeOf("recovery"))
	if mw := createRecoveryMiddleware(cfg); mw != nil {
		m.app.Application.middlewares = append(m.app.Application.middlewares, mw)
	}
	return m
}

func (m *MiddlewareMixin) UseLogging() abstract.MiddlewareUser {
	cfg := m.services.GetService(reflect.TypeOf("logging"))
	if mw := createLoggingMiddleware(cfg); mw != nil {
		m.app.Application.middlewares = append(m.app.Application.middlewares, mw)
	}
	return m
}

func (m *MiddlewareMixin) UseRateLimit() abstract.MiddlewareUser {
	cfg := m.services.GetService(reflect.TypeOf("ratelimit"))
	if mw := createRateLimitMiddleware(cfg); mw != nil {
		m.app.Application.middlewares = append(m.app.Application.middlewares, mw)
	}
	return m
}

func (m *MiddlewareMixin) UseGzip() abstract.MiddlewareUser {
	cfg := m.services.GetService(reflect.TypeOf("gzip"))
	if mw := createGzipMiddleware(cfg); mw != nil {
		m.app.Application.middlewares = append(m.app.Application.middlewares, mw)
	}
	return m
}

func (m *MiddlewareMixin) UseSecurity() abstract.MiddlewareUser {
	cfg := m.services.GetService(reflect.TypeOf("security"))
	if mw := createSecurityMiddleware(cfg); mw != nil {
		m.app.Application.middlewares = append(m.app.Application.middlewares, mw)
	}
	return m
}

func (m *MiddlewareMixin) UseRequestID() abstract.MiddlewareUser {
	cfg := m.services.GetService(reflect.TypeOf("requestid"))
	if mw := createRequestIDMiddleware(cfg); mw != nil {
		m.app.Application.middlewares = append(m.app.Application.middlewares, mw)
	}
	return m
}

func (m *MiddlewareMixin) UseTimeout() abstract.MiddlewareUser {
	cfg := m.services.GetService(reflect.TypeOf("timeout"))
	if mw := createTimeoutMiddleware(cfg); mw != nil {
		m.app.Application.middlewares = append(m.app.Application.middlewares, mw)
	}
	return m
}

func (m *MiddlewareMixin) Application() abstract.WebApplication {
	return m.app
}

var _ abstract.MiddlewareUser = (*MiddlewareMixin)(nil)

// WebApplication Web应用实现
type WebApplication struct {
	*Application
	builder *WebApplicationBuilder
}

func (a *WebApplication) Services() abstract.ServiceCollection {
	return a.Application.services
}

func (a *WebApplication) Use(middleware abstract.Middleware) abstract.WebApplication {
	a.Application.middlewares = append(a.Application.middlewares, middleware)
	return a
}

func createCORSMiddleware(cfg any) abstract.Middleware {
	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		return next()
	})
}

func createRecoveryMiddleware(cfg any) abstract.Middleware {
	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		return next()
	})
}

func createLoggingMiddleware(cfg any) abstract.Middleware {
	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		return next()
	})
}

func createRateLimitMiddleware(cfg any) abstract.Middleware {
	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		return next()
	})
}

func createGzipMiddleware(cfg any) abstract.Middleware {
	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		return next()
	})
}

func createSecurityMiddleware(cfg any) abstract.Middleware {
	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		return next()
	})
}

func createRequestIDMiddleware(cfg any) abstract.Middleware {
	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		return next()
	})
}

func createTimeoutMiddleware(cfg any) abstract.Middleware {
	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		return next()
	})
}

func (a *WebApplication) UseGlobalGuards(guards ...abstract.Guard) abstract.WebApplication {
	a.Application.guards = append(a.Application.guards, guards...)
	return a
}

func (a *WebApplication) UseGlobalInterceptors(interceptors ...abstract.Interceptor) abstract.WebApplication {
	a.Application.interceptors = append(a.Application.interceptors, interceptors...)
	return a
}

func (a *WebApplication) UseGlobalPipes(pipes ...abstract.Pipe) abstract.WebApplication {
	a.Application.pipes = append(a.Application.pipes, pipes...)
	return a
}

func (a *WebApplication) UseGlobalFilters(filters ...abstract.ExceptionFilter) abstract.WebApplication {
	a.Application.filters = append(a.Application.filters, filters...)
	return a
}

func (a *WebApplication) MapGet(path string, handler any) abstract.RouteBuilder {
	return a.registerRoute("GET", path, handler)
}

func (a *WebApplication) MapPost(path string, handler any) abstract.RouteBuilder {
	return a.registerRoute("POST", path, handler)
}

func (a *WebApplication) MapPut(path string, handler any) abstract.RouteBuilder {
	return a.registerRoute("PUT", path, handler)
}

func (a *WebApplication) MapDelete(path string, handler any) abstract.RouteBuilder {
	return a.registerRoute("DELETE", path, handler)
}

func (a *WebApplication) MapPatch(path string, handler any) abstract.RouteBuilder {
	return a.registerRoute("PATCH", path, handler)
}

func (a *WebApplication) Map(method string, path string, handler any) abstract.RouteBuilder {
	return a.registerRoute(method, path, handler)
}

func (a *WebApplication) MapGroup(prefix string) abstract.RouteGroup {
	return a.Application.router.Group(prefix)
}

func (a *WebApplication) UseRouting() abstract.WebApplication {
	return a
}

func (a *WebApplication) UseEndpoints(configure func(abstract.EndpointRouteBuilder)) abstract.WebApplication {
	endpointBuilder := &EndpointRouteBuilder{app: a}
	configure(endpointBuilder)
	return a
}

func (a *WebApplication) UseAuthentication() abstract.WebApplication {
	return a
}

func (a *WebApplication) UseAuthorization() abstract.WebApplication {
	return a
}

func (a *WebApplication) Urls() []string {
	if a.builder != nil && a.builder.Host != nil {
		return a.builder.Host.urls
	}
	return []string{"http://localhost:8080"}
}

func (a *WebApplication) Addresses() []string {
	return a.Urls()
}

func (a *WebApplication) registerRoute(method, path string, handler any) abstract.RouteBuilder {
	routeHandler := a.wrapHandler(handler)
	return a.Application.router.addRoute(method, path, routeHandler)
}

func (a *WebApplication) wrapHandler(handler any) abstract.RouteHandler {
	if h, ok := handler.(abstract.RouteHandler); ok {
		return h
	}
	hv := reflect.ValueOf(handler)
	ht := reflect.TypeOf(handler)
	if ht.Kind() == reflect.Func && ht.NumIn() == 1 {
		switch ht.NumOut() {
		case 1:
			return func(ctx abstract.Context) error {
				result := hv.Call([]reflect.Value{reflect.ValueOf(ctx)})
				if len(result) > 0 {
					if err, ok := result[0].Interface().(error); ok {
						return err
					}
				}
				return nil
			}
		case 2:
			return func(ctx abstract.Context) error {
				results := hv.Call([]reflect.Value{reflect.ValueOf(ctx)})
				if len(results) >= 2 {
					if err, ok := results[1].Interface().(error); ok && err != nil {
						return err
					}
					if results[0].Interface() != nil {
						return ctx.JSON(http.StatusOK, results[0].Interface())
					}
				}
				return nil
			}
		}
	}
	return func(ctx abstract.Context) error {
		return fmt.Errorf("invalid handler type: %T", handler)
	}
}

func (a *WebApplication) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.Application.router.ServeHTTP(w, r, a.Application)
}

func (a *WebApplication) Run() error {
	if err := a.Start(); err != nil {
		return err
	}
	return a.WaitForShutdown()
}

func (a *WebApplication) RunAsync() <-chan error {
	errCh := make(chan error, 1)
	go func() { errCh <- a.Run() }()
	return errCh
}

func (a *WebApplication) Start() error {
	addr := ":8080"
	if a.Application.config != nil {
		if port := a.Application.config.GetString("server.port"); port != "" {
			addr = ":" + port
		}
	}
	if a.Application.logger != nil {
		a.Application.logger.Info(fmt.Sprintf("Server starting on %s", addr))
	}
	return a.Listen(addr)
}

func (a *WebApplication) StartAsync() <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- a.Start()
	}()
	return ch
}

func (a *WebApplication) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()
	return a.Shutdown(ctx)
}

func (a *WebApplication) WaitForShutdown() error {
	select {}
}

var _ abstract.WebApplication = (*WebApplication)(nil)

// EndpointRouteBuilder 端点路由构建器实现
type EndpointRouteBuilder struct {
	app *WebApplication
}

func (e *EndpointRouteBuilder) MapGet(path string, handler any) abstract.RouteBuilder {
	return e.app.MapGet(path, handler)
}

func (e *EndpointRouteBuilder) MapPost(path string, handler any) abstract.RouteBuilder {
	return e.app.MapPost(path, handler)
}

func (e *EndpointRouteBuilder) MapPut(path string, handler any) abstract.RouteBuilder {
	return e.app.MapPut(path, handler)
}

func (e *EndpointRouteBuilder) MapDelete(path string, handler any) abstract.RouteBuilder {
	return e.app.MapDelete(path, handler)
}

func (e *EndpointRouteBuilder) MapPatch(path string, handler any) abstract.RouteBuilder {
	return e.app.MapPatch(path, handler)
}

func (e *EndpointRouteBuilder) MapControllers() abstract.EndpointRouteBuilder {
	return e
}

func (e *EndpointRouteBuilder) MapControllerRoute(name string, pattern string, defaults map[string]string) abstract.EndpointRouteBuilder {
	return e
}

func (e *EndpointRouteBuilder) MapAreaControllerRoute(name string, areaName string, pattern string, defaults map[string]string) abstract.EndpointRouteBuilder {
	return e
}

func (e *EndpointRouteBuilder) MapDefaultControllerRoute() abstract.EndpointRouteBuilder {
	e.MapControllerRoute("default", "{controller=Home}/{action=Index}/{id?}", nil)
	return e
}

var _ abstract.EndpointRouteBuilder = (*EndpointRouteBuilder)(nil)

// CreateBuilder 创建 WebApplicationBuilder
func CreateBuilder(args ...string) *WebApplicationBuilder {
	builder := NewWebApplicationBuilder()
	if len(args) > 0 {
		builder.Host.SetArgs(args)
	}
	return builder
}

// CreateApplication 创建 WebApplication
func CreateApplication(args ...string) *WebApplication {
	return CreateBuilder(args...).Build().(*WebApplication)
}

// CreateApplicationBuilder 创建通用 Application Builder（用于非 Web 场景）
func CreateApplicationBuilder(args ...string) *ApplicationBuilder {
	builder := NewApplicationBuilder()
	if len(args) > 0 {
		// 通用应用也可以处理命令行参数
	}
	return builder
}

// ApplicationCreateBuilder 创建通用 ApplicationBuilder（向后兼容）
func ApplicationCreateBuilder() *ApplicationBuilder {
	return NewApplicationBuilder()
}

// WebApplicationCreateBuilder 创建 WebApplicationBuilder（向后兼容）
func WebApplicationCreateBuilder() *WebApplicationBuilder {
	return NewWebApplicationBuilder()
}
