package abstract

import "reflect"

// ServiceLifetime 服务生命周期类型
type ServiceLifetime int

const (
	Singleton ServiceLifetime = iota
	Scoped
	Transient
)

// ServiceDescriptor 服务描述符接口
type ServiceDescriptor interface {
	ServiceType() reflect.Type
	Instance() any
	Factory() any
	Lifetime() ServiceLifetime
}

// ServiceResolver 服务解析接口
type ServiceResolver interface {
	GetService(serviceType reflect.Type) any
	GetRequiredService(serviceType reflect.Type) any
}

// ServiceRegistrar 服务注册接口
type ServiceRegistrar interface {
	AddSingleton(instance any) ServiceRegistrar
	AddSingletonFactory(serviceType reflect.Type, factory any) ServiceRegistrar
	AddScoped(serviceType reflect.Type, factory any) ServiceRegistrar
	AddTransient(serviceType reflect.Type, factory any) ServiceRegistrar
}

// MiddlewareRegistrar 中间件注册接口
type MiddlewareRegistrar interface {
	AddCORS(config any) MiddlewareRegistrar
	AddRecovery(config any) MiddlewareRegistrar
	AddLogging(config any) MiddlewareRegistrar
	AddRateLimit(config any) MiddlewareRegistrar
	AddGzip(config any) MiddlewareRegistrar
	AddSecurity(config any) MiddlewareRegistrar
	AddRequestID(config any) MiddlewareRegistrar
	AddTimeout(config any) MiddlewareRegistrar
}

// ServiceCollection 服务集合接口（组合）
type ServiceCollection interface {
	ServiceResolver
	ServiceRegistrar
	MiddlewareRegistrar
}

// Scope 作用域接口
type Scope interface {
	Dispose()
	IsDisposed() bool
}

// Disposable 可释放接口
type Disposable interface {
	Dispose() error
}
