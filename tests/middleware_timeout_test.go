package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/timeout"
)

func TestTimeout_DefaultConfig(t *testing.T) {
	cfg := timeout.DefaultConfig()

	if cfg == nil {
		t.Fatal("Expected config to be created")
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", cfg.Timeout)
	}

	if cfg.ErrorCode != http.StatusRequestTimeout {
		t.Errorf("Expected error code 408, got %d", cfg.ErrorCode)
	}
}

func TestTimeout_New_NilConfig(t *testing.T) {
	mw := timeout.New(nil)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestTimeout_New_WithConfig(t *testing.T) {
	cfg := &timeout.Config{
		Timeout:   10 * time.Second,
		ErrorCode: http.StatusGatewayTimeout,
		ErrorMsg:  "gateway timeout",
	}

	mw := timeout.New(cfg)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestTimeout_Middleware_NoTimeout(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	app.Use(timeout.New(&timeout.Config{
		Timeout: 5 * time.Second,
	}))

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

func TestTimeout_Middleware_WithTimeout(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	app.Use(timeout.New(&timeout.Config{
		Timeout:   100 * time.Millisecond,
		ErrorCode: http.StatusRequestTimeout,
		ErrorMsg:  "request timeout",
	}))

	app.MapGet("/slow", func(ctx abstract.Context) error {
		time.Sleep(200 * time.Millisecond)
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusRequestTimeout {
		t.Errorf("Expected status 408, got %d", w.Code)
	}
}

func TestTimeout_NewWithContext(t *testing.T) {
	mw := timeout.NewWithContext(nil)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}
