package core

import (
	"context"
	"net/http"
	"sync"

	"github.com/linuxerlv/gonest/config"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/logger"
)

type Application struct {
	config       config.Config
	env          abstract.Env
	services     *ServiceCollection
	logger       logger.Logger
	router       *HttpRouter
	controllers  []abstract.Controller
	middlewares  []abstract.Middleware
	guards       []abstract.Guard
	interceptors []abstract.Interceptor
	pipes        []abstract.Pipe
	filters      []abstract.ExceptionFilter
	server       *http.Server
	values       map[string]any
	mu           sync.RWMutex
}

func NewApplication() *Application {
	return &Application{
		config:       nil,
		env:          NewEnv(),
		services:     NewServiceCollection(),
		router:       NewRouter(),
		controllers:  make([]abstract.Controller, 0),
		middlewares:  make([]abstract.Middleware, 0),
		guards:       make([]abstract.Guard, 0),
		interceptors: make([]abstract.Interceptor, 0),
		pipes:        make([]abstract.Pipe, 0),
		filters:      make([]abstract.ExceptionFilter, 0),
		values:       make(map[string]any),
	}
}

func (a *Application) Services() abstract.ServiceCollection {
	return a.services
}

func (a *Application) Configuration() abstract.Config {
	if a.config != nil {
		return NewConfigAdapter(a.config)
	}
	return nil
}

func (a *Application) Environment() abstract.Env {
	return a.env
}

func (a *Application) Logging() abstract.Logger {
	if a.logger != nil {
		return NewLoggerAdapter(a.logger)
	}
	return NewLoggerAdapter(logger.GetGlobalLogger())
}

func (a *Application) Run() error {
	if err := a.Start(); err != nil {
		return err
	}
	return a.WaitForShutdown()
}

func (a *Application) RunAsync() <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- a.Run()
	}()
	return ch
}

func (a *Application) Start() error {
	return nil
}

func (a *Application) StartAsync() <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- a.Start()
	}()
	return ch
}

func (a *Application) Stop() error {
	return nil
}

func (a *Application) WaitForShutdown() error {
	return nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	if a.server != nil {
		return a.server.Shutdown(ctx)
	}
	return nil
}

func (a *Application) Controller(controller abstract.Controller) abstract.Application {
	a.controllers = append(a.controllers, controller)
	controller.Routes(a.router)
	return a
}

func (a *Application) Use(middleware abstract.Middleware) abstract.Application {
	a.middlewares = append(a.middlewares, middleware)
	return a
}

func (a *Application) UseGlobalGuards(guards ...abstract.Guard) abstract.Application {
	a.guards = append(a.guards, guards...)
	return a
}

func (a *Application) UseGlobalInterceptors(interceptors ...abstract.Interceptor) abstract.Application {
	a.interceptors = append(a.interceptors, interceptors...)
	return a
}

func (a *Application) UseGlobalPipes(pipes ...abstract.Pipe) abstract.Application {
	a.pipes = append(a.pipes, pipes...)
	return a
}

func (a *Application) UseGlobalFilters(filters ...abstract.ExceptionFilter) abstract.Application {
	a.filters = append(a.filters, filters...)
	return a
}

func (a *Application) Listen(addr string) error {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.router.ServeHTTP(w, r, a)
	})

	a.server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	return a.server.ListenAndServe()
}

func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r, a)
}

func (a *Application) Router() *HttpRouter {
	return a.router
}

func (a *Application) GET(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return a.router.GET(path, handler)
}

func (a *Application) POST(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return a.router.POST(path, handler)
}

func (a *Application) PUT(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return a.router.PUT(path, handler)
}

func (a *Application) DELETE(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return a.router.DELETE(path, handler)
}

func (a *Application) PATCH(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return a.router.PATCH(path, handler)
}

func (a *Application) OPTIONS(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return a.router.OPTIONS(path, handler)
}

func (a *Application) Group(prefix string) abstract.RouteGroup {
	return a.router.Group(prefix)
}

func (a *Application) Set(key string, value any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.values[key] = value
}

func (a *Application) Get(key string) any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.values[key]
}

func GetValue[T any](a *Application, key string) T {
	v := a.Get(key)
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

func (a *Application) AddRoute(method, path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return a.router.addRoute(method, path, handler)
}

func (a *Application) Match(req *http.Request) (abstract.Route, map[string]string) {
	return a.router.Match(req)
}

var _ abstract.Application = (*Application)(nil)
