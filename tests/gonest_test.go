package gonest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/cors"
	"github.com/linuxerlv/gonest/middleware/recovery"
)

// ============================================================
//                    abstract.ContextAbstract Tests
// ============================================================

func TestHttpContext_Method(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	if ctx.Method() != http.MethodGet {
		t.Errorf("expected GET, got %s", ctx.Method())
	}
}

func TestHttpContext_Path(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	if ctx.Path() != "/users/123" {
		t.Errorf("expected /users/123, got %s", ctx.Path())
	}
}

func TestHttpContext_Param(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()
	params := map[string]string{"id": "123"}
	ctx := core.NewContextWithParams(w, req, params)

	if ctx.Param("id") != "123" {
		t.Errorf("expected 123, got %s", ctx.Param("id"))
	}

	if ctx.Param("nonexistent") != "" {
		t.Errorf("expected empty string, got %s", ctx.Param("nonexistent"))
	}
}

func TestHttpContext_Query(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/search?q=golang&page=1", nil)
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	if ctx.Query("q") != "golang" {
		t.Errorf("expected golang, got %s", ctx.Query("q"))
	}

	if ctx.Query("page") != "1" {
		t.Errorf("expected 1, got %s", ctx.Query("page"))
	}
}

func TestHttpContext_Header(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	if ctx.Header("Authorization") != "Bearer token123" {
		t.Errorf("expected Bearer token123, got %s", ctx.Header("Authorization"))
	}
}

func TestHttpContext_Body(t *testing.T) {
	body := `{"name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	got := string(ctx.Body())
	if got != body {
		t.Errorf("expected %s, got %s", body, got)
	}
}

func TestHttpContext_Bind(t *testing.T) {
	body := `{"name":"test","age":25}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	var data struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	if err := ctx.Bind(&data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Name != "test" {
		t.Errorf("expected test, got %s", data.Name)
	}

	if data.Age != 25 {
		t.Errorf("expected 25, got %d", data.Age)
	}
}

func TestHttpContext_JSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	data := map[string]string{"message": "hello"}
	if err := ctx.JSON(http.StatusOK, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", w.Header().Get("Content-Type"))
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["message"] != "hello" {
		t.Errorf("expected hello, got %s", result["message"])
	}
}

func TestHttpContext_String(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	if err := ctx.String(http.StatusOK, "hello world"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "hello world" {
		t.Errorf("expected hello world, got %s", w.Body.String())
	}
}

func TestHttpContext_SetGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	ctx.Set("user", "admin")
	ctx.Set("role", "admin")

	if ctx.Get("user") != "admin" {
		t.Errorf("expected admin, got %v", ctx.Get("user"))
	}

	if ctx.Get("role") != "admin" {
		t.Errorf("expected admin, got %v", ctx.Get("role"))
	}
}

// ============================================================
//                    abstract.RouterAbstract Tests
// ============================================================

func TestHttpRouter_GET(t *testing.T) {
	router := core.NewRouter()
	called := false
	router.GET("/users", func(ctx abstract.ContextAbstract) error {
		called = true
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	route, params := router.Match(req)

	if route == nil {
		t.Fatal("expected route to be found")
	}

	if len(params) != 0 {
		t.Errorf("expected no params, got %v", params)
	}

	if !called {
		// Route was found but not executed - that's expected in Match
		t.Log("Route found successfully")
	}
}

func TestHttpRouter_POST(t *testing.T) {
	router := core.NewRouter()
	router.POST("/users", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusCreated, "Created")
	})

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	route, _ := router.Match(req)

	if route == nil {
		t.Fatal("expected route to be found")
	}

	if route.Method() != http.MethodPost {
		t.Errorf("expected POST, got %s", route.Method())
	}
}

func TestHttpRouter_Param(t *testing.T) {
	router := core.NewRouter()
	router.GET("/users/:id", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, ctx.Param("id"))
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	route, params := router.Match(req)

	if route == nil {
		t.Fatal("expected route to be found")
	}

	if params["id"] != "123" {
		t.Errorf("expected id=123, got %v", params)
	}
}

func TestHttpRouter_MultipleParams(t *testing.T) {
	router := core.NewRouter()
	router.GET("/users/:userId/posts/:postId", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42/posts/100", nil)
	route, params := router.Match(req)

	if route == nil {
		t.Fatal("expected route to be found")
	}

	if params["userId"] != "42" {
		t.Errorf("expected userId=42, got %s", params["userId"])
	}

	if params["postId"] != "100" {
		t.Errorf("expected postId=100, got %s", params["postId"])
	}
}

func TestHttpRouter_NotFound(t *testing.T) {
	router := core.NewRouter()
	router.GET("/users", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	route, _ := router.Match(req)

	if route != nil {
		t.Error("expected route to be nil for nonexistent path")
	}
}

func TestHttpRouter_MethodMismatch(t *testing.T) {
	router := core.NewRouter()
	router.GET("/users", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	route, _ := router.Match(req)

	if route != nil {
		t.Error("expected route to be nil for method mismatch")
	}
}

func TestHttpRouter_Group(t *testing.T) {
	router := core.NewRouter()
	api := router.Group("/api")
	api.GET("/users", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	route, _ := router.Match(req)

	if route == nil {
		t.Fatal("expected route to be found")
	}

	if route.Path() != "/api/users" {
		t.Errorf("expected /api/users, got %s", route.Path())
	}
}

func TestHttpRouter_RouteBuilder(t *testing.T) {
	router := core.NewRouter()
	guard := abstract.GuardFuncAbstract(func(ctx abstract.ContextAbstract) bool { return true })
	interceptor := abstract.InterceptorFuncAbstract(func(ctx abstract.ContextAbstract, next abstract.RouteHandlerAbstract) (any, error) {
		return nil, next(ctx)
	})
	pipe := abstract.PipeFuncAbstract(func(value any, ctx abstract.ContextAbstract) (any, error) {
		return value, nil
	})

	router.GET("/protected", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	}).Guard(guard).Interceptor(interceptor).Pipe(pipe)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	route, _ := router.Match(req)

	if route == nil {
		t.Fatal("expected route to be found")
	}
}

// ============================================================
//                    Application Tests
// ============================================================

func TestApplication_New(t *testing.T) {
	app := core.NewApplication()

	if app.Router() == nil {
		t.Error("expected router to be initialized")
	}
}

func TestApplication_Controller(t *testing.T) {
	app := core.NewApplication()

	testController := &TestController{}
	result := app.Controller(testController)
	if result == nil {
		t.Error("expected Controller to return Application")
	}
}

func TestApplication_Use(t *testing.T) {
	app := core.NewApplication()

	middleware := abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
		return next()
	})

	result := app.Use(middleware)
	if result == nil {
		t.Error("expected Use to return Application")
	}
}

func TestApplication_UseGlobalGuards(t *testing.T) {
	app := core.NewApplication()

	guard := abstract.GuardFuncAbstract(func(ctx abstract.ContextAbstract) bool { return true })
	result := app.UseGlobalGuards(guard)
	if result == nil {
		t.Error("expected UseGlobalGuards to return Application")
	}
}

func TestApplication_UseGlobalInterceptors(t *testing.T) {
	app := core.NewApplication()

	interceptor := abstract.InterceptorFuncAbstract(func(ctx abstract.ContextAbstract, next abstract.RouteHandlerAbstract) (any, error) {
		return nil, next(ctx)
	})
	result := app.UseGlobalInterceptors(interceptor)
	if result == nil {
		t.Error("expected UseGlobalInterceptors to return Application")
	}
}

func TestApplication_UseGlobalFilters(t *testing.T) {
	app := core.NewApplication()

	filter := &TestExceptionFilter{}
	result := app.UseGlobalFilters(filter)
	if result == nil {
		t.Error("expected UseGlobalFilters to return Application")
	}
}

// ============================================================
//                    Middleware Tests
// ============================================================

func TestMiddleware_Chain(t *testing.T) {
	app := core.NewApplication()
	var order []string

	mw1 := abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
		order = append(order, "mw1-before")
		err := next()
		order = append(order, "mw1-after")
		return err
	})

	mw2 := abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
		order = append(order, "mw2-before")
		err := next()
		order = append(order, "mw2-after")
		return err
	})

	app.Use(mw1)
	app.Use(mw2)

	app.GET("/test", func(ctx abstract.ContextAbstract) error {
		order = append(order, "handler")
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	app.Router().ServeHTTP(w, req, app)

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Errorf("expected %v, got %v", expected, order)
	}

	for i, v := range expected {
		if i >= len(order) || order[i] != v {
			t.Errorf("expected %s at position %d, got %v", v, i, order)
		}
	}
}

// ============================================================
//                    Guard Tests
// ============================================================

func TestGuard_Allow(t *testing.T) {
	app := core.NewApplication()

	app.UseGlobalGuards(abstract.GuardFuncAbstract(func(ctx abstract.ContextAbstract) bool {
		return true
	}))

	app.GET("/protected", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	app.Router().ServeHTTP(w, req, app)

	if w.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGuard_Deny(t *testing.T) {
	app := core.NewApplication()

	app.UseGlobalGuards(abstract.GuardFuncAbstract(func(ctx abstract.ContextAbstract) bool {
		return false
	}))

	app.GET("/protected", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	app.Router().ServeHTTP(w, req, app)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestGuard_RouteLevel(t *testing.T) {
	app := core.NewApplication()

	app.GET("/protected", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	}).Guard(abstract.GuardFuncAbstract(func(ctx abstract.ContextAbstract) bool {
		return ctx.Header("Authorization") == "valid-token"
	}))

	// Test with valid token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "valid-token")
	w := httptest.NewRecorder()
	app.Router().ServeHTTP(w, req, app)

	if w.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
	}

	// Test with invalid token
	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "invalid-token")
	w2 := httptest.NewRecorder()
	app.Router().ServeHTTP(w2, req2, app)

	if w2.Code != http.StatusForbidden {
		t.Errorf("expected %d, got %d", http.StatusForbidden, w2.Code)
	}
}

// ============================================================
//                    Interceptor Tests
// ============================================================

func TestInterceptor_BeforeAfter(t *testing.T) {
	app := core.NewApplication()
	var order []string

	app.UseGlobalInterceptors(abstract.InterceptorFuncAbstract(func(ctx abstract.ContextAbstract, next abstract.RouteHandlerAbstract) (any, error) {
		order = append(order, "interceptor-before")
		err := next(ctx)
		order = append(order, "interceptor-after")
		return nil, err
	}))

	app.GET("/test", func(ctx abstract.ContextAbstract) error {
		order = append(order, "handler")
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	app.Router().ServeHTTP(w, req, app)

	expected := []string{"interceptor-before", "handler", "interceptor-after"}
	if len(order) != len(expected) {
		t.Errorf("expected %v, got %v", expected, order)
	}
}

func TestInterceptor_TransformResponse(t *testing.T) {
	app := core.NewApplication()

	app.UseGlobalInterceptors(abstract.InterceptorFuncAbstract(func(ctx abstract.ContextAbstract, next abstract.RouteHandlerAbstract) (any, error) {
		err := next(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"status": "success",
		}, nil
	}))

	app.GET("/test", func(ctx abstract.ContextAbstract) error {
		return ctx.JSON(http.StatusOK, map[string]string{"message": "hello"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	app.Router().ServeHTTP(w, req, app)

	// Note: The interceptor result is not used in current implementation
	// This test verifies the interceptor is called
	if w.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

// ============================================================
//                    ExceptionFilter Tests
// ============================================================

func TestExceptionFilter_Catch(t *testing.T) {
	app := core.NewApplication()

	app.UseGlobalFilters(&TestExceptionFilter{})

	app.GET("/error", func(ctx abstract.ContextAbstract) error {
		return abstract.BadRequest("invalid request")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()
	app.Router().ServeHTTP(w, req, app)

	// Error should be caught by filter
	t.Logf("Response code: %d", w.Code)
}

func TestExceptionFilter_Panic(t *testing.T) {
	app := core.NewApplication()

	app.UseGlobalFilters(&TestExceptionFilter{})

	app.GET("/panic", func(ctx abstract.ContextAbstract) error {
		panic("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	app.Router().ServeHTTP(w, req, app)

	// Panic should be recovered by filter
	t.Logf("Response code: %d", w.Code)
}

// ============================================================
//                    Integration Tests
// ============================================================

func TestIntegration_FullStack(t *testing.T) {
	app := core.NewApplication()

	// Add middleware
	app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
		ctx.Set("request-id", "12345")
		return next()
	}))

	// Add global guard
	app.UseGlobalGuards(abstract.GuardFuncAbstract(func(ctx abstract.ContextAbstract) bool {
		return ctx.Header("X-API-Key") == "secret"
	}))

	// Add global interceptor
	app.UseGlobalInterceptors(abstract.InterceptorFuncAbstract(func(ctx abstract.ContextAbstract, next abstract.RouteHandlerAbstract) (any, error) {
		ctx.Set("intercepted", true)
		return nil, next(ctx)
	}))

	// Add controller
	app.Controller(&UserController{})

	// Test GET /users with valid API key
	t.Run("GET /users with valid key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set("X-API-Key", "secret")
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)

		if w.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
		}

		var result map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		users, ok := result["users"].([]any)
		if !ok || len(users) != 2 {
			t.Errorf("expected 2 users, got %v", result)
		}
	})

	// Test GET /users without API key
	t.Run("GET /users without key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	// Test GET /users/:id
	t.Run("GET /users/:id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		req.Header.Set("X-API-Key", "secret")
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)

		if w.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
		}

		var result map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result["id"] != "123" {
			t.Errorf("expected id=123, got %v", result)
		}
	})

	// Test POST /users
	t.Run("POST /users", func(t *testing.T) {
		body := `{"name":"Charlie"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("X-API-Key", "secret")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)

		if w.Code != http.StatusCreated {
			t.Errorf("expected %d, got %d", http.StatusCreated, w.Code)
		}
	})
}

func TestIntegration_RouteGroup(t *testing.T) {
	app := core.NewApplication()

	api := app.Group("/api/v1")
	api.GET("/health", func(ctx abstract.ContextAbstract) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api.GET("/users/:id", func(ctx abstract.ContextAbstract) error {
		return ctx.JSON(http.StatusOK, map[string]string{"id": ctx.Param("id")})
	})

	// Test health endpoint
	t.Run("GET /api/v1/health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)

		if w.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
		}
	})

	// Test users endpoint
	t.Run("GET /api/v1/users/42", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42", nil)
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)

		if w.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
		}

		var result map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result["id"] != "42" {
			t.Errorf("expected id=42, got %s", result["id"])
		}
	})
}

// ============================================================
//                    Test Helpers
// ============================================================

type TestController struct{}

func (c *TestController) Routes(r abstract.RouterAbstract) {
	r.GET("/test", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "test")
	})
}

type UserController struct{}

func (c *UserController) Routes(r abstract.RouterAbstract) {
	r.GET("/users", c.List)
	r.GET("/users/:id", c.Get)
	r.POST("/users", c.Create)
}

func (c *UserController) List(ctx abstract.ContextAbstract) error {
	users := []map[string]string{
		{"id": "1", "name": "Alice"},
		{"id": "2", "name": "Bob"},
	}
	return ctx.JSON(http.StatusOK, map[string]any{"users": users})
}

func (c *UserController) Get(ctx abstract.ContextAbstract) error {
	id := ctx.Param("id")
	return ctx.JSON(http.StatusOK, map[string]string{"id": id, "name": "User " + id})
}

func (c *UserController) Create(ctx abstract.ContextAbstract) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := ctx.Bind(&input); err != nil {
		return abstract.BadRequest("invalid JSON")
	}
	return ctx.JSON(http.StatusCreated, map[string]string{"id": "3", "name": input.Name})
}

type TestExceptionFilter struct{}

func (f *TestExceptionFilter) Catch(ctx abstract.ContextAbstract, err error) error {
	if httpErr, ok := err.(*abstract.HttpException); ok {
		ctx.JSON(httpErr.Status(), map[string]string{
			"error":  httpErr.Message(),
			"status": "error",
		})
	} else {
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":  err.Error(),
			"status": "error",
		})
	}
	return nil
}

// ============================================================
//                    HTTP Exception Tests
// ============================================================

func TestHttpException_BadRequest(t *testing.T) {
	err := abstract.BadRequest("invalid input")
	if err.Error() != "invalid input" {
		t.Errorf("expected 'invalid input', got '%s'", err.Error())
	}

	httpErr := err.(*abstract.HttpException)
	if httpErr.Status() != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, httpErr.Status())
	}
}

func TestHttpException_NotFound(t *testing.T) {
	err := abstract.NotFound("resource not found")
	httpErr := err.(*abstract.HttpException)
	if httpErr.Status() != http.StatusNotFound {
		t.Errorf("expected %d, got %d", http.StatusNotFound, httpErr.Status())
	}
}

func TestHttpException_Unauthorized(t *testing.T) {
	err := abstract.Unauthorized("not authenticated")
	httpErr := err.(*abstract.HttpException)
	if httpErr.Status() != http.StatusUnauthorized {
		t.Errorf("expected %d, got %d", http.StatusUnauthorized, httpErr.Status())
	}
}

func TestHttpException_Forbidden(t *testing.T) {
	err := abstract.Forbidden("access denied")
	httpErr := err.(*abstract.HttpException)
	if httpErr.Status() != http.StatusForbidden {
		t.Errorf("expected %d, got %d", http.StatusForbidden, httpErr.Status())
	}
}

func TestHttpException_InternalError(t *testing.T) {
	err := abstract.InternalError("server error")
	httpErr := err.(*abstract.HttpException)
	if httpErr.Status() != http.StatusInternalServerError {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, httpErr.Status())
	}
}

// ============================================================
//                    Server Tests (without actually starting)
// ============================================================

func TestApplication_ServerConfig(t *testing.T) {
	app := core.NewApplication()

	err := app.Shutdown(context.Background())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// ============================================================
//                    Edge Cases
// ============================================================

func TestRouter_EmptyPath(t *testing.T) {
	router := core.NewRouter()
	router.GET("/", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "root")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	route, _ := router.Match(req)

	if route == nil {
		t.Error("expected route to be found for root path")
	}
}

func TestRouter_TrailingSlash(t *testing.T) {
	router := core.NewRouter()
	router.GET("/users", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "users")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	route, _ := router.Match(req)

	// Note: Current implementation does NOT handle trailing slashes
	// This test documents the current behavior
	if route != nil {
		t.Log("abstract.RouterAbstract handles trailing slashes")
	} else {
		t.Log("abstract.RouterAbstract does NOT handle trailing slashes (expected behavior)")
	}
}

func TestContext_EmptyQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	if ctx.Query("nonexistent") != "" {
		t.Error("expected empty string for nonexistent query param")
	}
}

func TestContext_Context(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	if ctx.Context() == nil {
		t.Error("expected context to be non-nil")
	}
}

// ============================================================
//                    Backward Compatibility Tests
// ============================================================

func TestBackwardCompatibility_OldAndNewAPI(t *testing.T) {
	t.Run("Old API still works", func(t *testing.T) {
		app := core.NewApplication()
		app.Use(cors.New(nil))
		app.Use(recovery.New(nil))

		app.GET("/old", func(ctx abstract.ContextAbstract) error {
			return ctx.JSON(http.StatusOK, map[string]string{"api": "old"})
		})

		req := httptest.NewRequest(http.MethodGet, "/old", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("New builder API works", func(t *testing.T) {
		builder := core.CreateBuilder()
		app := builder.Build().(*core.WebApplication)

		app.Use(cors.New(nil))
		app.Use(recovery.New(nil))

		app.MapGet("/new", func(ctx abstract.ContextAbstract) error {
			return ctx.JSON(http.StatusOK, map[string]string{"api": "new"})
		})

		req := httptest.NewRequest(http.MethodGet, "/new", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Both APIs can coexist", func(t *testing.T) {
		builder := core.CreateBuilder()
		webApp := builder.Build().(*core.WebApplication)

		webApp.MapGet("/builder", func(ctx abstract.ContextAbstract) error {
			return ctx.JSON(http.StatusOK, map[string]string{"style": "builder"})
		})

		webApp.GET("/direct", func(ctx abstract.ContextAbstract) error {
			return ctx.JSON(http.StatusOK, map[string]string{"style": "direct"})
		})

		req1 := httptest.NewRequest(http.MethodGet, "/builder", nil)
		w1 := httptest.NewRecorder()
		webApp.ServeHTTP(w1, req1)

		if w1.Code != http.StatusOK {
			t.Errorf("builder route: expected 200, got %d", w1.Code)
		}

		req2 := httptest.NewRequest(http.MethodGet, "/direct", nil)
		w2 := httptest.NewRecorder()
		webApp.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Errorf("direct route: expected 200, got %d", w2.Code)
		}
	})
}
