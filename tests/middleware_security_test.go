package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/security"
)

func TestSecurity_DefaultConfig(t *testing.T) {
	cfg := security.DefaultConfig()

	if cfg == nil {
		t.Fatal("Expected config to be created")
	}

	if !cfg.XSSProtection {
		t.Error("Expected XSSProtection to be true")
	}

	if !cfg.ContentTypeNosniff {
		t.Error("Expected ContentTypeNosniff to be true")
	}

	if cfg.XFrameOptions != "DENY" {
		t.Errorf("Expected XFrameOptions 'DENY', got '%s'", cfg.XFrameOptions)
	}
}

func TestSecurity_DevelopmentConfig(t *testing.T) {
	cfg := security.DevelopmentConfig()

	if cfg == nil {
		t.Fatal("Expected config to be created")
	}

	if cfg.XFrameOptions != "SAMEORIGIN" {
		t.Errorf("Expected XFrameOptions 'SAMEORIGIN', got '%s'", cfg.XFrameOptions)
	}
}

func TestSecurity_New_NilConfig(t *testing.T) {
	mw := security.New(nil)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestSecurity_New_WithConfig(t *testing.T) {
	cfg := &security.Config{
		XSSProtection:         true,
		ContentTypeNosniff:    true,
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "no-referrer",
	}

	mw := security.New(cfg)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestSecurity_Middleware_SetsHeaders(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	app.Use(security.New(nil))

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("Expected XSS-Protection header, got '%s'", w.Header().Get("X-XSS-Protection"))
	}

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected Content-Type-Options header, got '%s'", w.Header().Get("X-Content-Type-Options"))
	}

	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options header, got '%s'", w.Header().Get("X-Frame-Options"))
	}

	if w.Header().Get("Strict-Transport-Security") == "" {
		t.Error("Expected Strict-Transport-Security header")
	}

	if w.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Errorf("Expected Referrer-Policy header, got '%s'", w.Header().Get("Referrer-Policy"))
	}
}

func TestSecurity_Middleware_CustomConfig(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.BuildWeb()

	app.Use(security.New(&security.Config{
		XSSProtection:         false,
		ContentTypeNosniff:    true,
		XFrameOptions:         "ALLOW-FROM http://example.com",
		HSTSMaxAge:            0,
		HSTSIncludeSubdomains: false,
	}))

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Header().Get("X-XSS-Protection") != "" {
		t.Error("Expected no XSS-Protection header")
	}

	if w.Header().Get("X-Frame-Options") != "ALLOW-FROM http://example.com" {
		t.Errorf("Expected custom X-Frame-Options, got '%s'", w.Header().Get("X-Frame-Options"))
	}
}
