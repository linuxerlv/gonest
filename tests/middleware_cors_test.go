package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/cors"
)

func TestCORS_DefaultConfig(t *testing.T) {
	cfg := cors.DefaultConfig()

	if cfg == nil {
		t.Fatal("Expected config to be created")
	}

	if len(cfg.AllowOrigins) == 0 {
		t.Error("Expected AllowOrigins to have values")
	}

	if len(cfg.AllowMethods) == 0 {
		t.Error("Expected AllowMethods to have values")
	}
}

func TestCORS_New_NilConfig(t *testing.T) {
	mw := cors.New(nil)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestCORS_New_WithConfig(t *testing.T) {
	cfg := &cors.Config{
		AllowOrigins:     []string{"http://example.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	mw := cors.New(cfg)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestCORS_Middleware_NoOrigin(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	app.Use(cors.New(&cors.Config{
		AllowOrigins: []string{"http://example.com"},
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
}

func TestCORS_Middleware_AllowedOrigin(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	app.Use(cors.New(&cors.Config{
		AllowOrigins: []string{"http://example.com"},
	}))

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("Expected CORS header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_Middleware_WildcardOrigin(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	app.Use(cors.New(&cors.Config{
		AllowOrigins: []string{"*"},
	}))

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://any-origin.com" {
		t.Errorf("Expected CORS header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_Middleware_OptionsRequest(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	app.Use(cors.New(&cors.Config{
		AllowOrigins:     []string{"http://example.com"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 204 or 404, got %d", w.Code)
	}
}
