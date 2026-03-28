package tests

import (
	"reflect"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
)

func TestWireProvider_NewWireProvider(t *testing.T) {
	provider := core.NewWireProvider(func() *struct{ Name string } {
		return &struct{ Name string }{Name: "test"}
	})

	result := provider()
	if result == nil {
		t.Fatal("Expected provider to return non-nil result")
	}

	typed := result.(*struct{ Name string })
	if typed.Name != "test" {
		t.Errorf("Expected Name to be 'test', got '%s'", typed.Name)
	}
}

func TestWireSet_NewWireSet(t *testing.T) {
	provider1 := core.NewWireProvider(func() string { return "hello" })
	provider2 := core.NewWireProvider(func() int { return 42 })

	wireSet := core.NewWireSet(provider1, provider2)

	if len(wireSet) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(wireSet))
	}
}

func TestSimpleWireModule_Providers(t *testing.T) {
	provider := core.NewWireProvider(func() string { return "test" })
	wireSet := core.NewWireSet(provider)

	module := core.NewSimpleWireModule(wireSet)

	providers := module.Providers()
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(providers))
	}
}

func TestSimpleWireModule_Imports(t *testing.T) {
	provider := core.NewWireProvider(func() string { return "test" })
	wireSet := core.NewWireSet(provider)

	subModule := core.NewSimpleWireModule(wireSet)
	mainModule := core.NewSimpleWireModule(wireSet, subModule)

	imports := mainModule.Imports()
	if len(imports) != 1 {
		t.Errorf("Expected 1 import, got %d", len(imports))
	}
}

func TestServiceCollectionWireAdapter_ProvideSingleton(t *testing.T) {
	collection := core.NewServiceCollection()
	adapter := core.NewServiceCollectionWireAdapter(collection)

	type TestService struct{ Name string }
	service := &TestService{Name: "adapter-test"}

	adapter.ProvideSingleton(service)

	retrieved := collection.GetService(reflect.TypeOf(service))
	if retrieved == nil {
		t.Fatal("Expected service to be retrieved")
	}

	if retrieved.(*TestService).Name != "adapter-test" {
		t.Errorf("Expected Name to be 'adapter-test', got '%s'", retrieved.(*TestService).Name)
	}
}

func TestWireServiceRegistrar_AddProvider(t *testing.T) {
	registrar := core.NewWireServiceRegistrar()

	provider := core.NewWireProvider(func() string { return "registered" })
	registrar.AddProvider(provider)

	collection := core.NewServiceCollection()
	registrar.RegisterTo(collection)

	serviceType := reflect.TypeOf("")
	retrieved := collection.GetService(serviceType)
	if retrieved == nil {
		t.Fatal("Expected service to be retrieved")
	}

	if retrieved.(string) != "registered" {
		t.Errorf("Expected 'registered', got '%s'", retrieved.(string))
	}
}

func TestWireServiceRegistrar_RegisterTo(t *testing.T) {
	registrar := core.NewWireServiceRegistrar()

	type TestService struct{ Name string }
	provider := core.NewWireProvider(func() *TestService {
		return &TestService{Name: "registered-service"}
	})
	registrar.AddProvider(provider)

	collection := core.NewServiceCollection()
	registrar.RegisterTo(collection)

	serviceType := reflect.TypeOf(&TestService{})
	retrieved := collection.GetService(serviceType)
	if retrieved == nil {
		t.Fatal("Expected service to be retrieved")
	}

	if retrieved.(*TestService).Name != "registered-service" {
		t.Errorf("Expected Name to be 'registered-service', got '%s'", retrieved.(*TestService).Name)
	}
}

func TestCollectWireProviders(t *testing.T) {
	provider1 := core.NewWireProvider(func() string { return "module1" })
	provider2 := core.NewWireProvider(func() string { return "module2" })

	module1 := core.NewSimpleWireModule(core.NewWireSet(provider1))
	module2 := core.NewSimpleWireModule(core.NewWireSet(provider2))

	allProviders := core.CollectWireProviders(module1, module2)

	if len(allProviders) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(allProviders))
	}
}

func TestCollectWireProviders_WithImports(t *testing.T) {
	provider1 := core.NewWireProvider(func() string { return "main" })
	provider2 := core.NewWireProvider(func() string { return "sub" })

	subModule := core.NewSimpleWireModule(core.NewWireSet(provider2))
	mainModule := core.NewSimpleWireModule(core.NewWireSet(provider1), subModule)

	allProviders := core.CollectWireProviders(mainModule)

	if len(allProviders) != 2 {
		t.Errorf("Expected 2 providers (including imports), got %d", len(allProviders))
	}
}

func TestRegisterWireProviders(t *testing.T) {
	type TestService struct{ Name string }

	provider := core.NewWireProvider(func() *TestService {
		return &TestService{Name: "wire-registered"}
	})
	wireSet := core.NewWireSet(provider)

	collection := core.NewServiceCollection()
	core.RegisterWireProviders(collection, wireSet)

	serviceType := reflect.TypeOf(&TestService{})
	retrieved := collection.GetService(serviceType)
	if retrieved == nil {
		t.Fatal("Expected service to be retrieved")
	}

	if retrieved.(*TestService).Name != "wire-registered" {
		t.Errorf("Expected Name to be 'wire-registered', got '%s'", retrieved.(*TestService).Name)
	}
}

func TestWireApplicationBuilder(t *testing.T) {
	builder := core.NewWireApplicationBuilder()

	if builder == nil {
		t.Fatal("Expected builder to be created")
	}

	if builder.ApplicationBuilder == nil {
		t.Error("Expected ApplicationBuilder to be embedded")
	}

	if builder.Services() == nil {
		t.Error("Expected Services to be available")
	}
}

func TestWireApplicationBuilder_AddProvider(t *testing.T) {
	builder := core.NewWireApplicationBuilder()

	provider := core.NewWireProvider(func() string { return "test" })
	builder.AddProvider(provider)

	app := builder.Build()
	if app == nil {
		t.Fatal("Expected application to be built after adding provider")
	}
}

func TestWireApplicationBuilder_AddModule(t *testing.T) {
	builder := core.NewWireApplicationBuilder()

	module := core.NewSimpleWireModule(core.NewWireSet())
	builder.AddModule(module)

	app := builder.Build()
	if app == nil {
		t.Fatal("Expected application to be built after adding module")
	}
}

func TestWireApplicationBuilder_Build(t *testing.T) {
	builder := core.NewWireApplicationBuilder()
	app := builder.Build()

	if app == nil {
		t.Fatal("Expected application to be built")
	}
}

func TestWireWebApplicationBuilder(t *testing.T) {
	builder := core.NewWireWebApplicationBuilder()

	if builder == nil {
		t.Fatal("Expected builder to be created")
	}

	if builder.WebApplicationBuilder == nil {
		t.Error("Expected WebApplicationBuilder to be embedded")
	}

	if builder.Services() == nil {
		t.Error("Expected Services to be available")
	}
}

func TestWireWebApplicationBuilder_Build(t *testing.T) {
	builder := core.NewWireWebApplicationBuilder()
	app := builder.Build()

	if app == nil {
		t.Fatal("Expected web application to be built")
	}
}

func TestProvideApplicationBuilder(t *testing.T) {
	builder := core.ProvideApplicationBuilder()

	if builder == nil {
		t.Fatal("Expected builder to be provided")
	}

	if _, ok := interface{}(builder).(*core.ApplicationBuilder); !ok {
		t.Error("Expected ApplicationBuilder type")
	}
}

func TestProvideWebApplicationBuilder(t *testing.T) {
	builder := core.ProvideWebApplicationBuilder()

	if builder == nil {
		t.Fatal("Expected builder to be provided")
	}

	if _, ok := interface{}(builder).(*core.WebApplicationBuilder); !ok {
		t.Error("Expected WebApplicationBuilder type")
	}
}

func TestProvideServiceCollection(t *testing.T) {
	collection := core.ProvideServiceCollection()

	if collection == nil {
		t.Fatal("Expected collection to be provided")
	}

	if _, ok := interface{}(collection).(*core.ServiceCollection); !ok {
		t.Error("Expected ServiceCollection type")
	}
}

func TestProvideApplication(t *testing.T) {
	builder := core.NewApplicationBuilder()
	app := core.ProvideApplication(builder)

	if app == nil {
		t.Fatal("Expected application to be provided")
	}

	if _, ok := app.(abstract.Application); !ok {
		t.Error("Expected Application interface")
	}
}

func TestProvideWebApplication(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := core.ProvideWebApplication(builder)

	if app == nil {
		t.Fatal("Expected web application to be provided")
	}

	if _, ok := app.(abstract.WebApplication); !ok {
		t.Error("Expected WebApplication interface")
	}
}

func TestWireApp(t *testing.T) {
	builder := core.NewApplicationBuilder()
	app := builder.Build()
	collection := core.NewServiceCollection()

	wireApp := core.ProvideWireApp(app, collection)

	if wireApp == nil {
		t.Fatal("Expected WireApp to be created")
	}

	if wireApp.Application == nil {
		t.Error("Expected Application to be set")
	}

	if wireApp.Services == nil {
		t.Error("Expected Services to be set")
	}
}

func TestWireInjectorFunc(t *testing.T) {
	called := false
	injector := core.WireInjectorFunc(func(collection abstract.ServiceCollection) {
		called = true
	})

	collection := core.NewServiceCollection()
	injector.Inject(collection)

	if !called {
		t.Error("Expected injector function to be called")
	}
}
