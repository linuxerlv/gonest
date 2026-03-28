package core

import (
	"context"
	"reflect"

	"github.com/linuxerlv/gonest/core/abstract"
)

type WireProvider func() any

type WireProviderFunc[T any] func() T

func NewWireProvider[T any](factory func() T) WireProvider {
	return func() any {
		return factory()
	}
}

type WireModule interface {
	Providers() WireSet
	Imports() []WireModule
}

type WireSet []WireProvider

func NewWireSet(providers ...WireProvider) WireSet {
	return WireSet(providers)
}

type SimpleWireModule struct {
	providers WireSet
	imports   []WireModule
}

func NewSimpleWireModule(providers WireSet, imports ...WireModule) *SimpleWireModule {
	return &SimpleWireModule{
		providers: providers,
		imports:   imports,
	}
}

func (m *SimpleWireModule) Providers() WireSet {
	return m.providers
}

func (m *SimpleWireModule) Imports() []WireModule {
	return m.imports
}

type ServiceCollectionWireAdapter struct {
	collection *ServiceCollection
}

func NewServiceCollectionWireAdapter(collection *ServiceCollection) *ServiceCollectionWireAdapter {
	return &ServiceCollectionWireAdapter{collection: collection}
}

func (a *ServiceCollectionWireAdapter) ProvideSingleton(instance any) *ServiceCollectionWireAdapter {
	a.collection.AddSingleton(instance)
	return a
}

func (a *ServiceCollectionWireAdapter) ProvideSingletonFactory(serviceType reflect.Type, factory any) *ServiceCollectionWireAdapter {
	a.collection.AddSingletonFactory(serviceType, factory)
	return a
}

func (a *ServiceCollectionWireAdapter) ProvideScoped(serviceType reflect.Type, factory any) *ServiceCollectionWireAdapter {
	a.collection.AddScoped(serviceType, factory)
	return a
}

func (a *ServiceCollectionWireAdapter) ProvideTransient(serviceType reflect.Type, factory any) *ServiceCollectionWireAdapter {
	a.collection.AddTransient(serviceType, factory)
	return a
}

func (a *ServiceCollectionWireAdapter) Collection() *ServiceCollection {
	return a.collection
}

type WireInjector interface {
	Inject(collection abstract.ServiceCollection)
}

type WireInjectorFunc func(collection abstract.ServiceCollection)

func (f WireInjectorFunc) Inject(collection abstract.ServiceCollection) {
	f(collection)
}

func ProvideServiceCollection() *ServiceCollection {
	return NewServiceCollection()
}

func ProvideServiceCollectionAdapter(collection *ServiceCollection) *ServiceCollectionWireAdapter {
	return NewServiceCollectionWireAdapter(collection)
}

type WireServiceRegistrar struct {
	providers []WireProvider
}

func NewWireServiceRegistrar() *WireServiceRegistrar {
	return &WireServiceRegistrar{
		providers: make([]WireProvider, 0),
	}
}

func (r *WireServiceRegistrar) AddProvider(provider WireProvider) *WireServiceRegistrar {
	r.providers = append(r.providers, provider)
	return r
}

func (r *WireServiceRegistrar) AddProviders(providers ...WireProvider) *WireServiceRegistrar {
	r.providers = append(r.providers, providers...)
	return r
}

func (r *WireServiceRegistrar) RegisterTo(collection *ServiceCollection) {
	for _, provider := range r.providers {
		instance := provider()
		collection.AddSingleton(instance)
	}
}

func CollectWireProviders(modules ...WireModule) WireSet {
	var allProviders WireSet
	for _, module := range modules {
		allProviders = append(allProviders, module.Providers()...)
		for _, imp := range module.Imports() {
			allProviders = append(allProviders, CollectWireProviders(imp)...)
		}
	}
	return allProviders
}

func RegisterWireProviders(collection *ServiceCollection, providers WireSet) {
	for _, provider := range providers {
		instance := provider()
		collection.AddSingleton(instance)
	}
}

type WireApplicationBuilder struct {
	*ApplicationBuilder
	providers []any
	modules   []WireModule
}

func NewWireApplicationBuilder() *WireApplicationBuilder {
	return &WireApplicationBuilder{
		ApplicationBuilder: NewApplicationBuilder(),
		providers:          make([]any, 0),
		modules:            make([]WireModule, 0),
	}
}

func (b *WireApplicationBuilder) AddProvider(provider any) *WireApplicationBuilder {
	b.providers = append(b.providers, provider)
	return b
}

func (b *WireApplicationBuilder) AddProviders(providers ...any) *WireApplicationBuilder {
	b.providers = append(b.providers, providers...)
	return b
}

func (b *WireApplicationBuilder) AddModule(module WireModule) *WireApplicationBuilder {
	b.modules = append(b.modules, module)
	return b
}

func (b *WireApplicationBuilder) Services() abstract.ServiceCollection {
	return b.ApplicationBuilder.Services()
}

func (b *WireApplicationBuilder) Build() abstract.Application {
	return b.ApplicationBuilder.Build()
}

type WireWebApplicationBuilder struct {
	*WebApplicationBuilder
	providers []any
	modules   []WireModule
}

func NewWireWebApplicationBuilder() *WireWebApplicationBuilder {
	return &WireWebApplicationBuilder{
		WebApplicationBuilder: NewWebApplicationBuilder(),
		providers:            make([]any, 0),
		modules:              make([]WireModule, 0),
	}
}

func (b *WireWebApplicationBuilder) AddProvider(provider any) *WireWebApplicationBuilder {
	b.providers = append(b.providers, provider)
	return b
}

func (b *WireWebApplicationBuilder) AddProviders(providers ...any) *WireWebApplicationBuilder {
	b.providers = append(b.providers, providers...)
	return b
}

func (b *WireWebApplicationBuilder) AddModule(module WireModule) *WireWebApplicationBuilder {
	b.modules = append(b.modules, module)
	return b
}

func (b *WireWebApplicationBuilder) Services() abstract.ServiceCollection {
	return b.WebApplicationBuilder.Services()
}

func (b *WireWebApplicationBuilder) Build() abstract.WebApplication {
	return b.WebApplicationBuilder.Build()
}

func ProvideApplicationBuilder() *ApplicationBuilder {
	return NewApplicationBuilder()
}

func ProvideWebApplicationBuilder() *WebApplicationBuilder {
	return NewWebApplicationBuilder()
}

func ProvideApplication(builder *ApplicationBuilder) abstract.Application {
	return builder.Build()
}

func ProvideWebApplication(builder *WebApplicationBuilder) abstract.WebApplication {
	return builder.Build()
}

type WireApp struct {
	Application abstract.Application
	Services    *ServiceCollection
}

func ProvideWireApp(app abstract.Application, collection *ServiceCollection) *WireApp {
	return &WireApp{
		Application: app,
		Services:    collection,
	}
}

func (a *WireApp) Run() error {
	return a.Application.Run()
}

func (a *WireApp) RunAsync() <-chan error {
	return a.Application.RunAsync()
}

func (a *WireApp) Start() error {
	return a.Application.Start()
}

func (a *WireApp) Stop() error {
	return a.Application.Stop()
}

func (a *WireApp) Shutdown(ctx context.Context) error {
	return a.Application.Shutdown(ctx)
}

func (a *WireApp) WaitForShutdown() error {
	return a.Application.WaitForShutdown()
}
