package core

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/linuxerlv/gonest/config"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/logger"
)

type ServiceDescriptor struct {
	serviceType reflect.Type
	instance    any
	factory     any
	lifetime    abstract.ServiceLifetimeAbstract
}

func (d *ServiceDescriptor) ServiceType() reflect.Type                  { return d.serviceType }
func (d *ServiceDescriptor) Instance() any                              { return d.instance }
func (d *ServiceDescriptor) Factory() any                               { return d.factory }
func (d *ServiceDescriptor) Lifetime() abstract.ServiceLifetimeAbstract { return d.lifetime }

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

func (s *ServiceCollection) AddSingleton(instance any) abstract.ServiceRegistrarAbstract {
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

func (s *ServiceCollection) AddSingletonFactory(serviceType reflect.Type, factory any) abstract.ServiceRegistrarAbstract {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[serviceType] = &ServiceDescriptor{
		serviceType: serviceType,
		factory:     factory,
		lifetime:    abstract.Singleton,
	}
	return s
}

func (s *ServiceCollection) AddScoped(serviceType reflect.Type, factory any) abstract.ServiceRegistrarAbstract {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.descriptors[serviceType] = &ServiceDescriptor{
		serviceType: serviceType,
		factory:     factory,
		lifetime:    abstract.Scoped,
	}
	return s
}

func (s *ServiceCollection) AddTransient(serviceType reflect.Type, factory any) abstract.ServiceRegistrarAbstract {
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
			if fn, ok := desc.Factory().(func(abstract.ServiceCollectionAbstract) any); ok {
				instance := fn(s)
				s.instances[serviceType] = instance
				return instance
			}
		}
	case abstract.Scoped, abstract.Transient:
		if desc.Factory() != nil {
			if fn, ok := desc.Factory().(func(abstract.ServiceCollectionAbstract) any); ok {
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

func AddSingletonFunc[T any](s *ServiceCollection, factory func(abstract.ServiceCollectionAbstract) T) *ServiceCollection {
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

func AddScoped[T any](s *ServiceCollection, factory func(abstract.ServiceCollectionAbstract) T) *ServiceCollection {
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

func AddTransient[T any](s *ServiceCollection, factory func(abstract.ServiceCollectionAbstract) T) *ServiceCollection {
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

func GetService[T any](s abstract.ServiceCollectionAbstract) T {
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

func GetRequiredService[T any](s abstract.ServiceCollectionAbstract) T {
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

var _ abstract.ServiceCollectionAbstract = (*ServiceCollection)(nil)

type WebApplicationBuilder struct {
	Services *ServiceCollection
	Config   config.Config
	Env      abstract.EnvAbstract
	Logger   logger.Logger
	Host     *HostBuilder
	mu       sync.RWMutex
}

func NewWebApplicationBuilder() *WebApplicationBuilder {
	return &WebApplicationBuilder{
		Services: NewServiceCollection(),
		Env:      NewEnv(),
		Host:     NewHostBuilder(),
	}
}

func (b *WebApplicationBuilder) UseConfig(cfg abstract.ConfigAbstract) abstract.WebApplicationBuilderAbstract {
	b.mu.Lock()
	if cfg != nil {
		if adapter, ok := cfg.(*ConfigAdapter); ok {
			b.Config = adapter.Unwrap()
		}
	}
	b.mu.Unlock()
	return b
}

func (b *WebApplicationBuilder) UseLogger(log abstract.LoggerAbstract) abstract.WebApplicationBuilderAbstract {
	b.mu.Lock()
	if log != nil {
		if adapter, ok := log.(*LoggerAdapter); ok {
			b.Logger = adapter.Unwrap()
		}
	}
	b.mu.Unlock()
	return b
}

func (b *WebApplicationBuilder) ConfigureServices(configure func(abstract.ServiceCollectionAbstract)) abstract.WebApplicationBuilderAbstract {
	configure(b.Services)
	return b
}

func (b *WebApplicationBuilder) Build() abstract.WebApplicationAbstract {
	app := &WebApplication{
		Application: NewApplication(),
		Config:      b.Config,
		Env:         b.Env,
		Services:    b.Services,
		Logger:      b.Logger,
		builder:     b,
	}
	if b.Logger != nil {
		logger.SetGlobalLogger(b.Logger)
	}
	return app
}

type WebApplication struct {
	*Application
	Config   config.Config
	Env      abstract.EnvAbstract
	Services *ServiceCollection
	Logger   logger.Logger
	builder  *WebApplicationBuilder
}

func (WebApplication) CreateBuilder(args ...string) *WebApplicationBuilder {
	builder := NewWebApplicationBuilder()
	if len(args) > 0 {
		builder.Host.SetArgs(args)
	}
	return builder
}

func (a *WebApplication) Configuration() abstract.ConfigAbstract {
	if a.Config != nil {
		return NewConfigAdapter(a.Config)
	}
	return nil
}

func (a *WebApplication) Log() abstract.LoggerAbstract {
	if a.Logger != nil {
		return NewLoggerAdapter(a.Logger)
	}
	return NewLoggerAdapter(logger.GetGlobalLogger())
}

func (a *WebApplication) ConfigurationConcrete() config.Config {
	return a.Config
}

func (a *WebApplication) LogConcrete() logger.Logger {
	if a.Logger != nil {
		return a.Logger
	}
	return logger.GetGlobalLogger()
}

func (a *WebApplication) Run() error {
	addr := ":8080"
	cfg := a.ConfigurationConcrete()
	if cfg != nil {
		if port := cfg.GetString("server.port"); port != "" {
			addr = ":" + port
		}
	}
	log := a.LogConcrete()
	if log != nil {
		log.Info(fmt.Sprintf("Server starting on %s", addr))
	}
	return a.Listen(addr)
}

func (a *WebApplication) RunAsync() <-chan error {
	errCh := make(chan error, 1)
	go func() { errCh <- a.Run() }()
	return errCh
}

func (a *WebApplication) WaitForShutdown() error {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		a.Shutdown(ctx)
		cancel()
	}()
	return a.Run()
}

func (a *WebApplication) MapGet(path string, handler any) abstract.RouteBuilderAbstract {
	return a.registerRoute("GET", path, handler)
}

func (a *WebApplication) MapPost(path string, handler any) abstract.RouteBuilderAbstract {
	return a.registerRoute("POST", path, handler)
}

func (a *WebApplication) MapPut(path string, handler any) abstract.RouteBuilderAbstract {
	return a.registerRoute("PUT", path, handler)
}

func (a *WebApplication) MapDelete(path string, handler any) abstract.RouteBuilderAbstract {
	return a.registerRoute("DELETE", path, handler)
}

func (a *WebApplication) MapPatch(path string, handler any) abstract.RouteBuilderAbstract {
	return a.registerRoute("PATCH", path, handler)
}

func (a *WebApplication) registerRoute(method, path string, handler any) abstract.RouteBuilderAbstract {
	routeHandler := a.wrapHandler(handler)
	return a.Application.router.addRoute(method, path, routeHandler)
}

func (a *WebApplication) wrapHandler(handler any) abstract.RouteHandlerAbstract {
	if h, ok := handler.(abstract.RouteHandlerAbstract); ok {
		return h
	}
	hv := reflect.ValueOf(handler)
	ht := reflect.TypeOf(handler)
	if ht.Kind() == reflect.Func && ht.NumIn() == 1 {
		switch ht.NumOut() {
		case 1:
			return func(ctx abstract.ContextAbstract) error {
				result := hv.Call([]reflect.Value{reflect.ValueOf(ctx)})
				if len(result) > 0 {
					if err, ok := result[0].Interface().(error); ok {
						return err
					}
				}
				return nil
			}
		case 2:
			return func(ctx abstract.ContextAbstract) error {
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
	return func(ctx abstract.ContextAbstract) error {
		return fmt.Errorf("invalid handler type: %T", handler)
	}
}

func CreateBuilder(args ...string) *WebApplicationBuilder {
	return WebApplication{}.CreateBuilder(args...)
}

func CreateApplication(args ...string) *WebApplication {
	return CreateBuilder(args...).Build().(*WebApplication)
}

var _ abstract.WebApplicationAbstract = (*WebApplication)(nil)
