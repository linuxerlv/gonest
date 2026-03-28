package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/extensions"
)

func TestExtension_UseCORS(t *testing.T) {
	builder := core.NewWebApplicationBuilder()

	builder.Services().AddCORS(&extensions.CORSMiddlewareOptions{
		AllowOrigins:     []string{"http://example.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	})

	app := builder.Build()

	mixin := core.NewMiddlewareMixin(app.(*core.WebApplication), app.Services().(*core.ServiceCollection))
	mixin.UseCORS()

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

func TestExtension_ChainedMiddleware(t *testing.T) {
	builder := core.NewWebApplicationBuilder()

	builder.Services().AddRecovery(nil)
	builder.Services().AddLogging(nil)
	builder.Services().AddCORS(&extensions.CORSMiddlewareOptions{
		AllowOrigins: []string{"*"},
	})
	builder.Services().AddSecurity(nil)
	builder.Services().AddRequestID(nil)

	app := builder.Build()

	mixin := core.NewMiddlewareMixin(app.(*core.WebApplication), app.Services().(*core.ServiceCollection))
	mixin.UseRecovery().
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

func TestExtension_AllMiddlewares(t *testing.T) {
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

	mixin := core.NewMiddlewareMixin(app.(*core.WebApplication), app.Services().(*core.ServiceCollection))
	mixin.UseCORS().
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

func TestMixin_Application(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	mixin := core.NewMiddlewareMixin(app.(*core.WebApplication), app.Services().(*core.ServiceCollection))

	returnedApp := mixin.Application()
	if returnedApp != app {
		t.Error("Application() should return the same app instance")
	}
}
