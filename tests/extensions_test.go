package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/extensions"
)

func TestUseCORS(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseCORS(app, &extensions.CORSMiddlewareOptions{
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

func TestUseCORS_NilOptions(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseCORS(app, nil)

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseRecovery(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseRecovery(app, &extensions.RecoveryMiddlewareOptions{
		PrintStack: true,
	})

	app.MapGet("/panic", func(ctx abstract.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestUseRecovery_NilOptions(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseRecovery(app, nil)

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseLogging(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseLogging(app, &extensions.LoggingMiddlewareOptions{
		SkipPaths: []string{"/health"},
	})

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseLogging_NilOptions(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseLogging(app, nil)

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseRateLimit(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseRateLimit(app, &extensions.RateLimitMiddlewareOptions{
		Limit:  10,
		Window: 60,
	})

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseRateLimit_NilOptions(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseRateLimit(app, nil)

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseGzip(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseGzip(app, &extensions.GzipMiddlewareOptions{
		Level: 6,
	})

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseGzip_NilOptions(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseGzip(app, nil)

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseSecurity(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseSecurity(app, &extensions.SecurityMiddlewareOptions{
		XSSProtection:      true,
		ContentTypeNosniff: true,
		XFrameOptions:      "DENY",
	})

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options header, got %s", w.Header().Get("X-Frame-Options"))
	}
}

func TestUseSecurity_NilOptions(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseSecurity(app, nil)

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseRequestID(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseRequestID(app, &extensions.RequestIDMiddlewareOptions{
		HeaderName: "X-Request-ID",
	})

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header to be set")
	}
}

func TestUseRequestID_NilOptions(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseRequestID(app, nil)

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseTimeout(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseTimeout(app, &extensions.TimeoutMiddlewareOptions{
		Timeout: 30,
	})

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUseTimeout_NilOptions(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	extensions.UseTimeout(app, nil)

	app.MapGet("/test", func(ctx abstract.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
