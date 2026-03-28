package tests

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/linuxerlv/gonest"
	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
)

func TestServiceCollection_AddSingleton(t *testing.T) {
	collection := core.NewServiceCollection()

	type TestService struct{ Name string }
	service := &TestService{Name: "test"}

	collection.AddSingleton(service)

	retrieved := collection.GetService(reflect.TypeOf(service))
	if retrieved == nil {
		t.Error("Expected service to be retrieved")
	}

	if retrieved.(*TestService).Name != "test" {
		t.Errorf("Expected Name to be 'test', got '%s'", retrieved.(*TestService).Name)
	}
}

func TestServiceCollection_AddSingletonFactory(t *testing.T) {
	collection := core.NewServiceCollection()

	type TestService struct{ Name string }
	serviceType := reflect.TypeOf(&TestService{})

	collection.AddSingletonFactory(serviceType, func(sc abstract.ServiceCollection) any {
		return &TestService{Name: "factory"}
	})

	retrieved := collection.GetService(serviceType)
	if retrieved == nil {
		t.Error("Expected service to be retrieved")
	}

	if retrieved.(*TestService).Name != "factory" {
		t.Errorf("Expected Name to be 'factory', got '%s'", retrieved.(*TestService).Name)
	}
}

func TestServiceCollection_AddScoped(t *testing.T) {
	collection := core.NewServiceCollection()

	type TestService struct{ ID int }
	serviceType := reflect.TypeOf(&TestService{})

	callCount := 0
	collection.AddScoped(serviceType, func(sc abstract.ServiceCollection) any {
		callCount++
		return &TestService{ID: callCount}
	})

	retrieved1 := collection.GetService(serviceType)
	retrieved2 := collection.GetService(serviceType)

	if retrieved1.(*TestService).ID != 1 {
		t.Errorf("Expected first ID to be 1, got %d", retrieved1.(*TestService).ID)
	}

	if retrieved2.(*TestService).ID != 2 {
		t.Errorf("Expected second ID to be 2, got %d", retrieved2.(*TestService).ID)
	}
}

func TestServiceCollection_AddTransient(t *testing.T) {
	collection := core.NewServiceCollection()

	type TestService struct{ ID int }
	serviceType := reflect.TypeOf(&TestService{})

	callCount := 0
	collection.AddTransient(serviceType, func(sc abstract.ServiceCollection) any {
		callCount++
		return &TestService{ID: callCount}
	})

	retrieved1 := collection.GetService(serviceType)
	retrieved2 := collection.GetService(serviceType)

	if retrieved1.(*TestService).ID != 1 {
		t.Errorf("Expected first ID to be 1, got %d", retrieved1.(*TestService).ID)
	}

	if retrieved2.(*TestService).ID != 2 {
		t.Errorf("Expected second ID to be 2, got %d", retrieved2.(*TestService).ID)
	}
}

func TestServiceCollection_GetRequiredService(t *testing.T) {
	collection := core.NewServiceCollection()

	type TestService struct{ Name string }
	service := &TestService{Name: "required"}
	collection.AddSingleton(service)

	retrieved := collection.GetRequiredService(reflect.TypeOf(service))
	if retrieved.(*TestService).Name != "required" {
		t.Errorf("Expected Name to be 'required', got '%s'", retrieved.(*TestService).Name)
	}
}

func TestServiceCollection_GetRequiredService_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for non-existent service")
		}
	}()

	collection := core.NewServiceCollection()
	type TestService struct{ Name string }
	var service *TestService
	collection.GetRequiredService(reflect.TypeOf(service))
}

func TestApplicationBuilder_New(t *testing.T) {
	builder := core.NewApplicationBuilder()

	if builder == nil {
		t.Fatal("Expected builder to be created")
	}

	if builder.Services() == nil {
		t.Error("Expected Services to be initialized")
	}

	if builder.Environment() == nil {
		t.Error("Expected Environment to be initialized")
	}
}

func TestApplicationBuilder_Build(t *testing.T) {
	builder := core.NewApplicationBuilder()
	app := builder.Build()

	if app == nil {
		t.Fatal("Expected application to be built")
	}

	if app.Services() == nil {
		t.Error("Expected Services to be available")
	}
}

func TestWebApplicationBuilder_New(t *testing.T) {
	builder := core.NewWebApplicationBuilder()

	if builder == nil {
		t.Fatal("Expected builder to be created")
	}

	if builder.Services() == nil {
		t.Error("Expected Services to be initialized")
	}

	if builder.Host == nil {
		t.Error("Expected Host to be initialized")
	}
}

func TestWebApplicationBuilder_BuildWeb(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	if app == nil {
		t.Fatal("Expected web application to be built")
	}

	if app.Services() == nil {
		t.Error("Expected Services to be available")
	}
}

func TestWebApplication_Use(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	middlewareCalled := false
	app.Use(abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		middlewareCalled = true
		return next()
	}))

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Expected middleware to be called")
	}
}

func TestWebApplication_MapGet(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	handlerCalled := false
	app.MapGet("/test", func(ctx abstract.Context) error {
		handlerCalled = true
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestWebApplication_MapPost(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	handlerCalled := false
	app.MapPost("/create", func(ctx abstract.Context) error {
		handlerCalled = true
		return ctx.JSON(http.StatusCreated, map[string]string{"created": "true"})
	})

	req := httptest.NewRequest(http.MethodPost, "/create", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestCreateBuilder(t *testing.T) {
	builder := core.CreateBuilder()

	if builder == nil {
		t.Fatal("Expected builder to be created")
	}

	if _, ok := interface{}(builder).(*core.WebApplicationBuilder); !ok {
		t.Error("Expected WebApplicationBuilder type")
	}
}

func TestCreateApplicationBuilder(t *testing.T) {
	builder := core.CreateApplicationBuilder()

	if builder == nil {
		t.Fatal("Expected builder to be created")
	}

	if _, ok := interface{}(builder).(*core.ApplicationBuilder); !ok {
		t.Error("Expected ApplicationBuilder type")
	}
}

func TestEnv_Get(t *testing.T) {
	env := core.NewEnv()
	env.Set("TEST_KEY", "test_value")

	value := env.Get("TEST_KEY")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}
}

func TestEnv_GetOrDefault(t *testing.T) {
	env := core.NewEnv()

	value := env.GetOrDefault("NON_EXISTENT", "default")
	if value != "default" {
		t.Errorf("Expected 'default', got '%s'", value)
	}
}

func TestEnv_Has(t *testing.T) {
	env := core.NewEnv()
	env.Set("TEST_KEY", "test_value")

	if !env.Has("TEST_KEY") {
		t.Error("Expected Has to return true")
	}

	if env.Has("NON_EXISTENT") {
		t.Error("Expected Has to return false for non-existent key")
	}
}

func TestEnv_All(t *testing.T) {
	env := core.NewEnv()
	env.Set("KEY1", "value1")
	env.Set("KEY2", "value2")

	all := env.All()
	if all["KEY1"] != "value1" {
		t.Error("Expected KEY1 to be 'value1'")
	}
	if all["KEY2"] != "value2" {
		t.Error("Expected KEY2 to be 'value2'")
	}
}

func TestEnv_Unset(t *testing.T) {
	env := core.NewEnv()
	env.Set("TEST_KEY", "test_value")

	if !env.Has("TEST_KEY") {
		t.Error("Expected Has to return true")
	}

	env.Unset("TEST_KEY")

	if env.Has("TEST_KEY") {
		t.Error("Expected Has to return false after Unset")
	}
}

func TestBadRequest(t *testing.T) {
	err := gonest.BadRequest("test error")
	if err == nil {
		t.Fatal("Expected error to be created")
	}

	httpErr := err.(*gonest.HttpError)
	if httpErr.Status() != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", httpErr.Status())
	}

	if httpErr.Message() != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", httpErr.Message())
	}
}

func TestUnauthorized(t *testing.T) {
	err := gonest.Unauthorized("unauthorized")
	httpErr := err.(*gonest.HttpError)
	if httpErr.Status() != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", httpErr.Status())
	}
}

func TestForbidden(t *testing.T) {
	err := gonest.Forbidden("forbidden")
	httpErr := err.(*gonest.HttpError)
	if httpErr.Status() != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", httpErr.Status())
	}
}

func TestNotFound(t *testing.T) {
	err := gonest.NotFound("not found")
	httpErr := err.(*gonest.HttpError)
	if httpErr.Status() != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", httpErr.Status())
	}
}

func TestInternalError(t *testing.T) {
	err := gonest.InternalError("internal error")
	httpErr := err.(*gonest.HttpError)
	if httpErr.Status() != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", httpErr.Status())
	}
}

func TestNewHttpException(t *testing.T) {
	err := gonest.NewHttpException(418, "I'm a teapot")
	httpErr := err.(*gonest.HttpError)
	if httpErr.Status() != 418 {
		t.Errorf("Expected status 418, got %d", httpErr.Status())
	}
}

func TestNewApplication(t *testing.T) {
	app := gonest.NewApplication()
	if app == nil {
		t.Fatal("Expected application to be created")
	}
}

func TestNewRouter(t *testing.T) {
	router := gonest.NewRouter()
	if router == nil {
		t.Fatal("Expected router to be created")
	}
}

func TestNewContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := gonest.NewContext(w, req)
	if ctx == nil {
		t.Fatal("Expected context to be created")
	}

	if ctx.Method() != http.MethodGet {
		t.Errorf("Expected method GET, got %s", ctx.Method())
	}

	if ctx.Path() != "/test" {
		t.Errorf("Expected path /test, got %s", ctx.Path())
	}
}

func TestNewContextWithParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()

	params := map[string]string{"id": "123"}
	ctx := gonest.NewContextWithParams(w, req, params)
	if ctx == nil {
		t.Fatal("Expected context to be created")
	}

	if ctx.Param("id") != "123" {
		t.Errorf("Expected param id=123, got %s", ctx.Param("id"))
	}
}

func TestNewServiceCollection(t *testing.T) {
	collection := gonest.NewServiceCollection()
	if collection == nil {
		t.Fatal("Expected service collection to be created")
	}
}

func TestMiddlewareFunc(t *testing.T) {
	called := false
	middleware := gonest.MiddlewareFunc(func(ctx gonest.Context, next func() error) error {
		called = true
		return next()
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := gonest.NewContext(w, req)

	err := middleware.Handle(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !called {
		t.Error("Expected middleware to be called")
	}
}

func TestGuardFunc(t *testing.T) {
	guard := gonest.GuardFunc(func(ctx gonest.Context) bool {
		return ctx.Header("Authorization") != ""
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()
	ctx := gonest.NewContext(w, req)

	if !guard.CanActivate(ctx) {
		t.Error("Expected guard to return true")
	}
}

func TestPipeFunc(t *testing.T) {
	pipe := gonest.PipeFunc(func(value any, ctx gonest.Context) (any, error) {
		if str, ok := value.(string); ok {
			return str + "_transformed", nil
		}
		return value, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := gonest.NewContext(w, req)

	result, err := pipe.Transform("test", ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != "test_transformed" {
		t.Errorf("Expected 'test_transformed', got '%s'", result)
	}
}

func TestInterceptorFunc(t *testing.T) {
	beforeCalled := false
	afterCalled := false

	interceptor := gonest.InterceptorFunc(func(ctx gonest.Context, next gonest.RouteHandler) (any, error) {
		beforeCalled = true
		err := next(ctx)
		afterCalled = true
		return nil, err
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := gonest.NewContext(w, req)

	result, err := interceptor.Intercept(ctx, func(ctx gonest.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	if !beforeCalled {
		t.Error("Expected before hook to be called")
	}

	if !afterCalled {
		t.Error("Expected after hook to be called")
	}
}

func TestRouter_Routes(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	handlerCalled := false
	app.MapGet("/test", func(ctx abstract.Context) error {
		handlerCalled = true
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}

func TestRouter_Group(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	handlerCalled := false
	api := app.MapGroup("/api")
	api.GET("/users", func(ctx abstract.Context) error {
		handlerCalled = true
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called for grouped route")
	}
}

func TestContext_JSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := gonest.NewContext(w, req)

	err := ctx.JSON(http.StatusOK, map[string]string{"message": "hello"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestContext_GetSet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := gonest.NewContext(w, req)

	ctx.Set("key", "value")
	if ctx.Get("key") != "value" {
		t.Errorf("Expected 'value', got '%s'", ctx.Get("key"))
	}
}

func TestContext_Query(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?name=john&age=30", nil)
	w := httptest.NewRecorder()
	ctx := gonest.NewContext(w, req)

	if ctx.Query("name") != "john" {
		t.Errorf("Expected 'john', got '%s'", ctx.Query("name"))
	}

	if ctx.Query("age") != "30" {
		t.Errorf("Expected '30', got '%s'", ctx.Query("age"))
	}
}

func TestContext_Header(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom-Header", "custom-value")
	w := httptest.NewRecorder()
	ctx := gonest.NewContext(w, req)

	if ctx.Header("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected 'custom-value', got '%s'", ctx.Header("X-Custom-Header"))
	}
}
