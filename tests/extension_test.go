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
	app := extensions.Extend(builder.Build())

	app.UseCORS(&extensions.CORSMiddlewareOptions{
		AllowOrigins:     []string{"http://example.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	})

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
	app := extensions.Extend(builder.Build())

	app.UseRecovery(nil).
		UseLogging(nil).
		UseCORS(&extensions.CORSMiddlewareOptions{
			AllowOrigins: []string{"*"},
		}).
		UseSecurity(nil).
		UseRequestID(nil)

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header to be set")
	}
}

func TestExtension_AllMiddlewares(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := extensions.Extend(builder.Build())

	app.UseCORS(nil).
		UseRecovery(nil).
		UseLogging(nil).
		UseRateLimit(nil).
		UseGzip(nil).
		UseSecurity(nil).
		UseRequestID(nil).
		UseTimeout(nil)

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
