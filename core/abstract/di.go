package abstract

import "reflect"

// ServiceLifetimeAbstract 服务生命周期类型
type ServiceLifetimeAbstract int

const (
	Singleton ServiceLifetimeAbstract = iota
	Scoped
	Transient
)

// ServiceDescriptorAbstract 服务描述符接口
type ServiceDescriptorAbstract interface {
	ServiceType() reflect.Type
	Instance() any
	Factory() any
	Lifetime() ServiceLifetimeAbstract
}

// ServiceResolverAbstract 服务解析接口
type ServiceResolverAbstract interface {
	GetService(serviceType reflect.Type) any
	GetRequiredService(serviceType reflect.Type) any
}

// ServiceRegistrarAbstract 服务注册接口
type ServiceRegistrarAbstract interface {
	AddSingleton(instance any) ServiceRegistrarAbstract
	AddSingletonFactory(serviceType reflect.Type, factory any) ServiceRegistrarAbstract
	AddScoped(serviceType reflect.Type, factory any) ServiceRegistrarAbstract
	AddTransient(serviceType reflect.Type, factory any) ServiceRegistrarAbstract
}

// ServiceCollectionAbstract 服务集合接口（组合）
type ServiceCollectionAbstract interface {
	ServiceResolverAbstract
	ServiceRegistrarAbstract
}

// ScopeAbstract 作用域接口
type ScopeAbstract interface {
	Dispose()
	IsDisposed() bool
}

// DisposableAbstract 可释放接口
type DisposableAbstract interface {
	Dispose() error
}
