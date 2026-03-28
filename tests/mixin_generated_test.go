package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/extensions"
)

// TestMixinGenerated 测试生成的 Mixin 方法
// 用户可以直接使用 app.UseCORS() 方法，无需额外的 Mixin 对象
func TestMixinGenerated_UseCORS(t *testing.T) {
	builder := core.NewWebApplicationBuilder()

	builder.Services().AddCORS(&extensions.CORSMiddlewareOptions{
		AllowOrigins:     []string{"http://example.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	})

	app := builder.Build()

	// 直接使用 WebApplication 的 UseCORS 方法
	// 这是通过 Mixin 代码生成器生成的原生方法
	app.(*core.WebApplication).UseCORS()

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestMixinGenerated_Chained 测试链式调用
func TestMixinGenerated_Chained(t *testing.T) {
	builder := core.NewWebApplicationBuilder()

	builder.Services().AddRecovery(nil)
	builder.Services().AddLogging(nil)
	builder.Services().AddCORS(&extensions.CORSMiddlewareOptions{
		AllowOrigins: []string{"*"},
	})
	builder.Services().AddSecurity(nil)
	builder.Services().AddRequestID(nil)

	app := builder.Build()

	// 链式调用，看起来完全像原生方法
	app.(*core.WebApplication).UseRecovery().
		UseLogging().
		UseCORS().
		UseSecurity().
		UseRequestID()

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestMixinGenerated_AllMiddlewares 测试所有中间件
func TestMixinGenerated_AllMiddlewares(t *testing.T) {
	builder := core.NewWebApplicationBuilder()

	builder.Services().AddCORS(nil)
	builder.Services().AddRecovery(nil)
	builder.Services().AddLogging(nil)
	builder.Services().AddRateLimit(nil)
	builder.Services().AddGzip(nil)
	builder.Services().AddSecurity(nil)
	builder.Services().AddRequestID(nil)
	builder.Services().AddTimeout(nil)

	app := builder.Build()

	// 所有中间件都可以直接调用
	app.(*core.WebApplication).UseCORS().
		UseRecovery().
		UseLogging().
		UseRateLimit().
		UseGzip().
		UseSecurity().
		UseRequestID().
		UseTimeout()

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
