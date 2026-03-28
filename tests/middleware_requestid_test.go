package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/requestid"
)

func TestRequestID_DefaultConfig(t *testing.T) {
	cfg := requestid.DefaultConfig()

	if cfg == nil {
		t.Fatal("Expected config to be created")
	}

	if cfg.HeaderName != "X-Request-ID" {
		t.Errorf("Expected HeaderName 'X-Request-ID', got '%s'", cfg.HeaderName)
	}

	if cfg.Generator == nil {
		t.Error("Expected Generator to be set")
	}
}

func TestRequestID_New_NilConfig(t *testing.T) {
	mw := requestid.New(nil)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestRequestID_New_WithConfig(t *testing.T) {
	cfg := &requestid.Config{
		HeaderName: "X-Custom-ID",
		Generator:  func() string { return "custom-id" },
	}

	mw := requestid.New(cfg)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestRequestID_Middleware_GeneratesID(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	app.Use(requestid.New(nil))

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header to be set")
	}
}

func TestRequestID_Middleware_UsesExistingID(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	app.Use(requestid.New(nil))

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "existing-id")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != "existing-id" {
		t.Errorf("Expected X-Request-ID to be 'existing-id', got '%s'", w.Header().Get("X-Request-ID"))
	}
}

func TestRequestID_GetRequestID(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	app.Use(requestid.New(nil))

	var capturedID string
	app.MapGet("/test", func(ctx abstract.Context) error {
		capturedID = requestid.GetRequestID(ctx)
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if capturedID == "" {
		t.Error("Expected request ID to be captured")
	}
}
