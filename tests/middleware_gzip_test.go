package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/gzip"
)

func TestGzip_DefaultConfig(t *testing.T) {
	cfg := gzip.DefaultConfig()

	if cfg == nil {
		t.Fatal("Expected config to be created")
	}

	if cfg.Level != 6 {
		t.Errorf("Expected default level 6, got %d", cfg.Level)
	}
}

func TestGzip_New_NilConfig(t *testing.T) {
	mw := gzip.New(nil)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestGzip_New_WithConfig(t *testing.T) {
	cfg := &gzip.Config{
		Level: 9,
	}

	mw := gzip.New(cfg)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestGzip_Middleware_NoGzipHeader(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	app.Use(gzip.New(nil))

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

func TestGzip_Middleware_WithGzipHeader(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	app.Use(gzip.New(nil))

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
