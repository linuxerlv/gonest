package benchmark

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/linuxerlv/gonest"
	"github.com/linuxerlv/gonest/middleware/auth"
	"github.com/linuxerlv/gonest/middleware/casbin"
	"github.com/linuxerlv/gonest/middleware/cors"
	"github.com/linuxerlv/gonest/middleware/gzip"
	"github.com/linuxerlv/gonest/middleware/ratelimit"
	"github.com/linuxerlv/gonest/middleware/recovery"
	"github.com/linuxerlv/gonest/middleware/requestid"
	"github.com/linuxerlv/gonest/middleware/session"
	"github.com/linuxerlv/gonest/middleware/timeout"
)

func BenchmarkMiddleware(b *testing.B) {
	b.Run("Recovery", func(b *testing.B) {
		mw := recovery.New(nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := gonest.NewContext(w, req)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mw.Handle(ctx, func() error { return nil })
		}
	})

	b.Run("RequestID", func(b *testing.B) {
		mw := requestid.New(nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := gonest.NewContext(w, req)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mw.Handle(ctx, func() error { return nil })
		}
	})

	b.Run("CORS", func(b *testing.B) {
		mw := cors.New(nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := gonest.NewContext(w, req)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mw.Handle(ctx, func() error { return nil })
		}
	})

	b.Run("Timeout", func(b *testing.B) {
		mw := timeout.New(&timeout.Config{Timeout: 30 * time.Second})
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := gonest.NewContext(w, req)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mw.Handle(ctx, func() error { return nil })
		}
	})

	b.Run("RateLimit", func(b *testing.B) {
		mw := ratelimit.New(&ratelimit.Config{Limit: 10000, Window: time.Minute})
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := gonest.NewContext(w, req)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mw.Handle(ctx, func() error { return nil })
		}
	})

	b.Run("Gzip", func(b *testing.B) {
		mw := gzip.New(nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := gonest.NewContext(w, req)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mw.Handle(ctx, func() error { return nil })
		}
	})
}

func BenchmarkAuth(b *testing.B) {
	jwtProvider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          "benchmark-secret-key-for-testing",
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	}, nil)

	tokenPair, _ := jwtProvider.GenerateTokenPair("user1", "testuser", []string{"admin"}, nil)

	b.Run("TokenGeneration", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			jwtProvider.GenerateTokenPair("user1", "testuser", []string{"admin"}, nil)
		}
	})

	b.Run("TokenValidation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			jwtProvider.ValidateToken(tokenPair.AccessToken)
		}
	})

	b.Run("AuthMiddleware", func(b *testing.B) {
		mw := auth.New(jwtProvider, &auth.Config{
			TokenLookup: "header:Authorization",
			AuthScheme:  "Bearer",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		ctx := gonest.NewContext(w, req)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mw.Handle(ctx, func() error { return nil })
		}
	})
}

func BenchmarkSession(b *testing.B) {
	sm := session.WithMemoryStore()
	mw := session.New(&session.Config{SessionManager: sm})

	b.Run("SessionMiddleware", func(b *testing.B) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := gonest.NewContext(w, req)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mw.Handle(ctx, func() error { return nil })
		}
	})

	b.Run("SessionOperations", func(b *testing.B) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := gonest.NewContext(w, req)
		ctx.Set("session", sm)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			session.Put(ctx, "key", "value")
			session.Get(ctx, "key")
		}
	})
}

func BenchmarkCasbin(b *testing.B) {
	enforcer, _ := casbin.NewMemoryEnforcer()
	enforcer.AddPolicy("admin", "/admin/*", "*")
	enforcer.AddPolicy("user", "/api/*", "GET")
	enforcer.AddRoleForUser("user1", "admin")

	b.Run("Enforce", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			enforcer.Enforce("user1", "/admin/users", "GET")
		}
	})

	b.Run("CasbinMiddleware", func(b *testing.B) {
		mw, _ := casbin.New(&casbin.Config{Enforcer: enforcer})
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/admin/users", nil)
		ctx := gonest.NewContext(w, req)
		ctx.Set("user_id", "user1")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mw.Handle(ctx, func() error { return nil })
		}
	})
}

func BenchmarkFullChain(b *testing.B) {
	jwtProvider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          "benchmark-secret",
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	}, nil)

	tokenPair, _ := jwtProvider.GenerateTokenPair("user1", "test", []string{"admin"}, nil)

	enforcer, _ := casbin.NewMemoryEnforcer()
	enforcer.AddPolicy("admin", "/*", "*")
	enforcer.AddRoleForUser("user1", "admin")

	sm := session.WithMemoryStore()

	app := gonest.NewApplication()
	app.Use(recovery.New(nil))
	app.Use(requestid.New(nil))
	app.Use(cors.New(nil))
	app.Use(timeout.New(&timeout.Config{Timeout: 30 * time.Second}))
	app.Use(ratelimit.New(&ratelimit.Config{Limit: 10000, Window: time.Minute}))
	app.Use(gzip.New(nil))
	app.Use(session.New(&session.Config{SessionManager: sm}).AsMiddleware())
	app.Use(auth.New(jwtProvider, nil).AsMiddleware())
	casbinMW, _ := casbin.New(&casbin.Config{Enforcer: enforcer})
	app.Use(casbinMW.AsMiddleware())

	app.GET("/test", func(ctx gonest.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
			w := httptest.NewRecorder()
			app.Router().ServeHTTP(w, req, app)
		}
	})
}

func BenchmarkJSONResponse(b *testing.B) {
	app := gonest.NewApplication()
	app.GET("/json", func(ctx gonest.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{
			"message": "hello",
			"data":    []int{1, 2, 3, 4, 5},
			"nested": map[string]any{
				"key": "value",
			},
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)
	}
}

func BenchmarkJSONBinding(b *testing.B) {
	type Request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	app := gonest.NewApplication()
	app.POST("/bind", func(ctx gonest.Context) error {
		var req Request
		if err := ctx.Bind(&req); err != nil {
			return err
		}
		return ctx.JSON(http.StatusOK, req)
	})

	body, _ := json.Marshal(Request{Name: "test", Email: "test@example.com", Age: 25})
	req := httptest.NewRequest(http.MethodPost, "/bind", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		app.Router().ServeHTTP(w, req, app)
	}
}
