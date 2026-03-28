package gonest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
)

func BenchmarkContext_New(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/users/123?q=test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = core.NewContext(w, req)
	}
}

func BenchmarkContext_JSON(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	data := map[string]string{"message": "hello world", "status": "ok"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w = httptest.NewRecorder()
		ctx := core.NewContext(w, req)
		ctx.JSON(http.StatusOK, data)
	}
}

func BenchmarkContext_Bind(b *testing.B) {
	body := `{"name":"test","email":"test@example.com","age":25}`
	w := httptest.NewRecorder()

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := core.NewContext(w, httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body)))
		ctx.Bind(&input)
	}
}

func BenchmarkContext_Query(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/search?q=golang&page=1&limit=10&sort=desc", nil)
	w := httptest.NewRecorder()
	ctx := core.NewContext(w, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Query("q")
		ctx.Query("page")
		ctx.Query("limit")
		ctx.Query("sort")
	}
}

func BenchmarkRouter_Match_Static(b *testing.B) {
	router := core.NewRouter()
	router.GET("/users", func(ctx abstract.ContextAbstract) error { return nil })
	router.GET("/posts", func(ctx abstract.ContextAbstract) error { return nil })
	router.GET("/comments", func(ctx abstract.ContextAbstract) error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/users", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.Match(req)
	}
}

func BenchmarkRouter_Match_Param(b *testing.B) {
	router := core.NewRouter()
	router.GET("/users/:id", func(ctx abstract.ContextAbstract) error { return nil })
	router.GET("/users/:id/posts/:postId", func(ctx abstract.ContextAbstract) error { return nil })
	router.GET("/users/:id/comments/:commentId", func(ctx abstract.ContextAbstract) error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.Match(req)
	}
}

func BenchmarkRouter_Match_DeepPath(b *testing.B) {
	router := core.NewRouter()
	router.GET("/api/v1/users/:id/posts/:postId/comments/:commentId", func(ctx abstract.ContextAbstract) error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123/posts/456/comments/789", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.Match(req)
	}
}

func BenchmarkRouter_AddRoute(b *testing.B) {
	router := core.NewRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.GET("/test-route", func(ctx abstract.ContextAbstract) error { return nil })
	}
}

func BenchmarkRouter_Group(b *testing.B) {
	router := core.NewRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api := router.Group("/api/v1")
		api.GET("/users", func(ctx abstract.ContextAbstract) error { return nil })
	}
}

func BenchmarkMiddleware_Chain(b *testing.B) {
	app := core.NewApplication()

	app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
		ctx.Set("mw1", true)
		return next()
	}))
	app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
		ctx.Set("mw2", true)
		return next()
	}))
	app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
		ctx.Set("mw3", true)
		return next()
	}))

	app.GET("/test", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)
	}
}

func BenchmarkGuard_Check(b *testing.B) {
	app := core.NewApplication()

	app.UseGlobalGuards(abstract.GuardFuncAbstract(func(ctx abstract.ContextAbstract) bool {
		return ctx.Header("Authorization") != ""
	}))

	app.GET("/protected", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer token")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)
	}
}

func BenchmarkInterceptor_Execute(b *testing.B) {
	app := core.NewApplication()

	app.UseGlobalInterceptors(abstract.InterceptorFuncAbstract(func(ctx abstract.ContextAbstract, next abstract.RouteHandlerAbstract) (any, error) {
		ctx.Set("before", true)
		err := next(ctx)
		ctx.Set("after", true)
		return nil, err
	}))

	app.GET("/test", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)
	}
}

func BenchmarkFullRequest_Simple(b *testing.B) {
	app := core.NewApplication()
	app.GET("/hello", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "Hello, World!")
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)
	}
}

func BenchmarkFullRequest_JSON(b *testing.B) {
	app := core.NewApplication()
	app.GET("/users", func(ctx abstract.ContextAbstract) error {
		users := []map[string]string{
			{"id": "1", "name": "Alice"},
			{"id": "2", "name": "Bob"},
			{"id": "3", "name": "Charlie"},
		}
		return ctx.JSON(http.StatusOK, map[string]any{"users": users})
	})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)
	}
}

func BenchmarkFullRequest_WithAllFeatures(b *testing.B) {
	app := core.NewApplication()

	app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
		ctx.Set("request-id", "12345")
		return next()
	}))

	app.UseGlobalGuards(abstract.GuardFuncAbstract(func(ctx abstract.ContextAbstract) bool {
		return ctx.Header("X-API-Key") == "secret"
	}))

	app.UseGlobalInterceptors(abstract.InterceptorFuncAbstract(func(ctx abstract.ContextAbstract, next abstract.RouteHandlerAbstract) (any, error) {
		return nil, next(ctx)
	}))

	app.UseGlobalFilters(&benchErrorFilter{})

	app.GET("/users/:id", func(ctx abstract.ContextAbstract) error {
		user := map[string]string{
			"id":   ctx.Param("id"),
			"name": "User " + ctx.Param("id"),
		}
		return ctx.JSON(http.StatusOK, user)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	req.Header.Set("X-API-Key", "secret")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)
	}
}

func BenchmarkJSONEncoding_Small(b *testing.B) {
	data := map[string]string{"id": "1", "name": "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(data)
	}
}

func BenchmarkJSONEncoding_Large(b *testing.B) {
	users := make([]map[string]string, 100)
	for i := 0; i < 100; i++ {
		users[i] = map[string]string{
			"id":    string(rune(i)),
			"name":  "User Name",
			"email": "user@example.com",
		}
	}
	data := map[string]any{"users": users, "total": 100}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(data)
	}
}

func BenchmarkRouter_ManyRoutes(b *testing.B) {
	router := core.NewRouter()

	for i := 0; i < 100; i++ {
		router.GET("/route"+string(rune(i)), func(ctx abstract.ContextAbstract) error { return nil })
	}

	req := httptest.NewRequest(http.MethodGet, "/route50", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.Match(req)
	}
}

func BenchmarkRouter_Concurrent(b *testing.B) {
	router := core.NewRouter()
	router.GET("/users/:id", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		for pb.Next() {
			router.Match(req)
		}
	})
}

func BenchmarkFullRequest_Concurrent(b *testing.B) {
	app := core.NewApplication()
	app.GET("/test", func(ctx abstract.ContextAbstract) error {
		return ctx.String(http.StatusOK, "OK")
	})

	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		for pb.Next() {
			w := httptest.NewRecorder()
			app.Router().ServeHTTP(w, req, app)
		}
	})
}

type benchErrorFilter struct{}

func (f *benchErrorFilter) Catch(ctx abstract.ContextAbstract, err error) error {
	return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
