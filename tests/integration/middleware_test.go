package integration

import (
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
	"github.com/linuxerlv/gonest/middleware/security"
	"github.com/linuxerlv/gonest/middleware/session"
	"github.com/linuxerlv/gonest/middleware/timeout"
)

func TestMiddlewareChain(t *testing.T) {
	jwtProvider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          "test-secret",
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	}, nil)

	enforcer, _ := casbin.NewMemoryEnforcer()
	enforcer.AddPolicy("admin", "/admin/*", "*")
	enforcer.AddRoleForUser("test-user", "admin")

	sm := session.WithMemoryStore()

	tests := []struct {
		name           string
		path           string
		method         string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "Health check passes through all middleware",
			path:           "/health",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Protected route without auth returns 401",
			path:           "/api/users",
			method:         "GET",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Protected route with valid auth returns 200",
			path:   "/api/users",
			method: "GET",
			headers: func() map[string]string {
				tokenPair, _ := jwtProvider.GenerateTokenPair("test-user", "testuser", []string{"admin"}, nil)
				return map[string]string{"Authorization": "Bearer " + tokenPair.AccessToken}
			}(),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := gonest.NewApplication()

			app.Use(recovery.New(nil))
			app.Use(requestid.New(nil))
			app.Use(security.New(nil))
			app.Use(cors.New(nil))
			app.Use(timeout.New(&timeout.Config{Timeout: 30 * time.Second}))
			app.Use(ratelimit.New(&ratelimit.Config{Limit: 1000, Window: time.Minute}))
			app.Use(gzip.New(nil))
			app.Use(session.New(&session.Config{SessionManager: sm}).AsMiddleware())
			app.Use(auth.New(jwtProvider, &auth.Config{
				SkipPaths: []string{"/health"},
			}).AsMiddleware())

			app.GET("/health", func(ctx gonest.Context) error {
				return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
			})
			app.GET("/api/users", func(ctx gonest.Context) error {
				return ctx.JSON(http.StatusOK, map[string]string{"users": "list"})
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()

			app.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthSessionIntegration(t *testing.T) {
	jwtProvider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          "test-secret",
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	}, nil)

	sm := session.WithMemoryStore()

	app := gonest.NewApplication()
	app.Use(session.New(&session.Config{SessionManager: sm}).AsMiddleware())
	app.Use(auth.New(jwtProvider, &auth.Config{
		SkipPaths: []string{"/login"},
	}).AsMiddleware())

	app.POST("/login", func(ctx gonest.Context) error {
		tokenPair, err := jwtProvider.GenerateTokenPair("user1", "testuser", []string{"user"}, nil)
		if err != nil {
			return err
		}
		session.SetUserID(ctx, "user1")
		return ctx.JSON(http.StatusOK, map[string]string{"token": tokenPair.AccessToken})
	})

	app.GET("/me", func(ctx gonest.Context) error {
		userID := auth.GetUserID(ctx)
		if userID == "" {
			return gonest.Unauthorized("not authenticated")
		}
		return ctx.JSON(http.StatusOK, map[string]string{"user_id": userID})
	})

	t.Run("Login and access protected route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Login failed: %d", w.Code)
		}

		var loginResp map[string]string
		json.Unmarshal(w.Body.Bytes(), &loginResp)
		token := loginResp["token"]

		cookie := w.Result().Cookies()
		if len(cookie) == 0 {
			t.Log("No session cookie set")
		}

		req2 := httptest.NewRequest(http.MethodGet, "/me", nil)
		req2.Header.Set("Authorization", "Bearer "+token)
		w2 := httptest.NewRecorder()
		app.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d: %s", w2.Code, w2.Body.String())
		}
	})
}

func TestCasbinWithAuthIntegration(t *testing.T) {
	jwtProvider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          "test-secret",
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	}, nil)

	enforcer, _ := casbin.NewMemoryEnforcer()
	enforcer.AddPolicy("admin", "/admin/*", "*")
	enforcer.AddPolicy("user", "/api/*", "GET")

	app := gonest.NewApplication()
	app.Use(auth.New(jwtProvider, &auth.Config{
		SkipPaths: []string{"/login"},
	}).AsMiddleware())

	casbinMW, _ := casbin.New(&casbin.Config{
		Enforcer:  enforcer,
		SkipPaths: []string{"/login"},
	})
	app.Use(casbinMW.AsMiddleware())

	app.POST("/login", func(ctx gonest.Context) error {
		role := ctx.Query("role")
		userID := role + "-user"
		tokenPair, _ := jwtProvider.GenerateTokenPair(userID, userID, []string{role}, nil)
		enforcer.AddRoleForUser(userID, role)
		return ctx.JSON(http.StatusOK, map[string]string{"token": tokenPair.AccessToken, "user_id": userID})
	})

	app.GET("/admin/dashboard", func(ctx gonest.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"message": "admin dashboard"})
	})

	app.GET("/api/data", func(ctx gonest.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"data": "public"})
	})

	tests := []struct {
		name           string
		role           string
		path           string
		method         string
		expectedStatus int
	}{
		{"Admin can access admin route", "admin", "/admin/dashboard", "GET", http.StatusOK},
		{"Regular user cannot access admin route", "user", "/admin/dashboard", "GET", http.StatusForbidden},
		{"Regular user can access API GET", "user", "/api/data", "GET", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login?role="+tt.role, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			var resp map[string]string
			json.Unmarshal(w.Body.Bytes(), &resp)
			token := resp["token"]

			req2 := httptest.NewRequest(tt.method, tt.path, nil)
			req2.Header.Set("Authorization", "Bearer "+token)
			w2 := httptest.NewRecorder()
			app.ServeHTTP(w2, req2)

			if w2.Code != tt.expectedStatus {
				t.Errorf("Expected %d, got %d: %s", tt.expectedStatus, w2.Code, w2.Body.String())
			}
		})
	}
}
