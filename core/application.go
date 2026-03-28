package core

import (
	"context"
	"net/http"
	"sync"

	"github.com/linuxerlv/gonest/config"
	"github.com/linuxerlv/gonest/core/abstract"
)

type Application struct {
	Config       config.Config
	Env          abstract.EnvAbstract
	Services     *ServiceCollection
	router       *HttpRouter
	controllers  []abstract.ControllerAbstract
	middlewares  []abstract.MiddlewareAbstract
	guards       []abstract.GuardAbstract
	interceptors []abstract.InterceptorAbstract
	pipes        []abstract.PipeAbstract
	filters      []abstract.ExceptionFilterAbstract
	server       *http.Server
	values       map[string]any
	mu           sync.RWMutex
}

func NewApplication() *Application {
	return &Application{
		Config:       nil,
		Env:          NewEnv(),
		Services:     NewServiceCollection(),
		router:       NewRouter(),
		controllers:  make([]abstract.ControllerAbstract, 0),
		middlewares:  make([]abstract.MiddlewareAbstract, 0),
		guards:       make([]abstract.GuardAbstract, 0),
		interceptors: make([]abstract.InterceptorAbstract, 0),
		pipes:        make([]abstract.PipeAbstract, 0),
		filters:      make([]abstract.ExceptionFilterAbstract, 0),
		values:       make(map[string]any),
	}
}

func (a *Application) Controller(controller abstract.ControllerAbstract) abstract.ApplicationAbstract {
	a.controllers = append(a.controllers, controller)
	controller.Routes(a.router)
	return a
}

func (a *Application) Use(middleware abstract.MiddlewareAbstract) abstract.ApplicationAbstract {
	a.middlewares = append(a.middlewares, middleware)
	return a
}

func (a *Application) UseGlobalGuards(guards ...abstract.GuardAbstract) abstract.ApplicationAbstract {
	a.guards = append(a.guards, guards...)
	return a
}

func (a *Application) UseGlobalInterceptors(interceptors ...abstract.InterceptorAbstract) abstract.ApplicationAbstract {
	a.interceptors = append(a.interceptors, interceptors...)
	return a
}

func (a *Application) UseGlobalPipes(pipes ...abstract.PipeAbstract) abstract.ApplicationAbstract {
	a.pipes = append(a.pipes, pipes...)
	return a
}

func (a *Application) UseGlobalFilters(filters ...abstract.ExceptionFilterAbstract) abstract.ApplicationAbstract {
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

func (a *Application) Shutdown(ctx context.Context) error {
	if a.server != nil {
		return a.server.Shutdown(ctx)
	}
	return nil
}

func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r, a)
}

func (a *Application) Router() *HttpRouter {
	return a.router
}

func (a *Application) GET(path string, handler abstract.RouteHandlerAbstract) abstract.RouteBuilderAbstract {
	return a.router.GET(path, handler)
}

func (a *Application) POST(path string, handler abstract.RouteHandlerAbstract) abstract.RouteBuilderAbstract {
	return a.router.POST(path, handler)
}

func (a *Application) PUT(path string, handler abstract.RouteHandlerAbstract) abstract.RouteBuilderAbstract {
	return a.router.PUT(path, handler)
}

func (a *Application) DELETE(path string, handler abstract.RouteHandlerAbstract) abstract.RouteBuilderAbstract {
	return a.router.DELETE(path, handler)
}

func (a *Application) PATCH(path string, handler abstract.RouteHandlerAbstract) abstract.RouteBuilderAbstract {
	return a.router.PATCH(path, handler)
}

func (a *Application) OPTIONS(path string, handler abstract.RouteHandlerAbstract) abstract.RouteBuilderAbstract {
	return a.router.OPTIONS(path, handler)
}

func (a *Application) Group(prefix string) abstract.RouteGroupAbstract {
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

func (a *Application) AddRoute(method, path string, handler abstract.RouteHandlerAbstract) abstract.RouteBuilderAbstract {
	return a.router.addRoute(method, path, handler)
}

func (a *Application) Match(req *http.Request) (abstract.RouteAbstract, map[string]string) {
	return a.router.Match(req)
}

var _ abstract.ApplicationAbstract = (*Application)(nil)
