package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/ratelimit"
)

func TestRateLimit_DefaultConfig(t *testing.T) {
	cfg := ratelimit.DefaultConfig()

	if cfg == nil {
		t.Fatal("Expected config to be created")
	}

	if cfg.Limit != 100 {
		t.Errorf("Expected limit 100, got %d", cfg.Limit)
	}

	if cfg.Window != time.Minute {
		t.Errorf("Expected window 1 minute, got %v", cfg.Window)
	}

	if cfg.KeyFunc == nil {
		t.Error("Expected KeyFunc to be set")
	}
}

func TestRateLimit_New_NilConfig(t *testing.T) {
	mw := ratelimit.New(nil)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestRateLimit_New_WithConfig(t *testing.T) {
	cfg := &ratelimit.Config{
		Limit:     10,
		Window:    30 * time.Second,
		ErrorCode: http.StatusTooManyRequests,
		ErrorMsg:  "too many requests",
		KeyFunc: func(ctx abstract.Context) string {
			return "test-key"
		},
	}

	mw := ratelimit.New(cfg)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}

func TestRateLimit_Middleware_AllowsRequest(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	app.Use(ratelimit.New(&ratelimit.Config{
		Limit:  10,
		Window: time.Minute,
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

func TestRateLimit_Middleware_BlocksExcessRequests(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	limiter := ratelimit.NewLimiter(2, time.Minute)

	app.Use(ratelimit.NewWithLimiter(limiter, func(ctx abstract.Context) string {
		return "same-key"
	}))

	app.MapGet("/test", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestRateLimit_Middleware_SkipFunc(t *testing.T) {
	builder := core.NewWebApplicationBuilder()
	app := builder.Build()

	app.Use(ratelimit.New(&ratelimit.Config{
		Limit:  1,
		Window: time.Minute,
		KeyFunc: func(ctx abstract.Context) string {
			return "same-key"
		},
		SkipFunc: func(ctx abstract.Context) bool {
			return ctx.Path() == "/health"
		},
	}))

	app.MapGet("/health", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimit_NewLimiter(t *testing.T) {
	limiter := ratelimit.NewLimiter(10, time.Minute)

	if limiter == nil {
		t.Fatal("Expected limiter to be created")
	}
}

func TestRateLimit_Limiter_Allow(t *testing.T) {
	limiter := ratelimit.NewLimiter(3, time.Minute)

	if !limiter.Allow("key1") {
		t.Error("Expected first request to be allowed")
	}

	if !limiter.Allow("key1") {
		t.Error("Expected second request to be allowed")
	}

	if !limiter.Allow("key1") {
		t.Error("Expected third request to be allowed")
	}

	if limiter.Allow("key1") {
		t.Error("Expected fourth request to be blocked")
	}
}

func TestRateLimit_Limiter_Reset(t *testing.T) {
	limiter := ratelimit.NewLimiter(1, time.Minute)

	if !limiter.Allow("key1") {
		t.Error("Expected first request to be allowed")
	}

	if limiter.Allow("key1") {
		t.Error("Expected second request to be blocked")
	}

	limiter.Reset("key1")

	if !limiter.Allow("key1") {
		t.Error("Expected request after reset to be allowed")
	}
}

func TestRateLimit_NewWithLimiter(t *testing.T) {
	limiter := ratelimit.NewLimiter(10, time.Minute)
	mw := ratelimit.NewWithLimiter(limiter, nil)

	if mw == nil {
		t.Fatal("Expected middleware to be created")
	}
}
