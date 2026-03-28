package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/logger"
)

func TestLogger_DefaultConfig(t *testing.T) {
	cfg := logger.DefaultConfig()

	if cfg == nil {
		t.Fatal("Expected config to be created")
	}

	if cfg.TimeFormat == "" {
		t.Error("Expected TimeFormat to be set")
	}
}

func TestLogger_New_NilConfig(t *testing.T) {
	mw := logger.New(nil)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestLogger_New_WithConfig(t *testing.T) {
	cfg := &logger.Config{
		SkipPaths:  []string{"/health"},
		TimeFormat: "2006-01-02",
	}

	mw := logger.New(cfg)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestLogger_Middleware_LogsRequest(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	app.Use(logger.New(nil))

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

func TestLogger_Middleware_SkipPath(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	app.Use(logger.New(&logger.Config{
		SkipPaths: []string{"/health"},
	}))

	handlerCalled := false
	app.MapGet("/health", func(ctx abstract.Context) error {
		handlerCalled = true
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}
