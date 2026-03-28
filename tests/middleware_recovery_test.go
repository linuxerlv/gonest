package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/recovery"
)

func TestRecovery_DefaultConfig(t *testing.T) {
	cfg := recovery.DefaultConfig()

	if cfg == nil {
		t.Fatal("Expected config to be created")
	}

	if !cfg.PrintStack {
		t.Error("Expected PrintStack to be true by default")
	}
}

func TestRecovery_New_NilConfig(t *testing.T) {
	mw := recovery.New(nil)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestRecovery_New_WithConfig(t *testing.T) {
	cfg := &recovery.Config{
		PrintStack: false,
	}

	mw := recovery.New(cfg)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestRecovery_Middleware_NoPanic(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	app.Use(recovery.New(nil))

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

func TestRecovery_Middleware_WithPanic(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	app.Use(recovery.New(&recovery.Config{
		PrintStack: true,
	}))

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
