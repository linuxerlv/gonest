package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/linuxerlv/gonest"
	"github.com/linuxerlv/gonest/middleware/auth"
)

// MockStore for testing TokenStore interface
type MockStore struct {
	data map[string]any
}

func NewMockStore() *MockStore {
	return &MockStore{data: make(map[string]any)}
}

func (s *MockStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	s.data[key] = value
	return nil
}

func (s *MockStore) Get(ctx context.Context, key string) (any, error) {
	return s.data[key], nil
}

func (s *MockStore) Delete(ctx context.Context, key string) error {
	delete(s.data, key)
	return nil
}

func (s *MockStore) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := s.data[key]
	return ok, nil
}

// Helper to create a test context with custom path
func newTestContextWithPath(method, path string) gonest.Context {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	return gonest.NewContext(w, req)
}

// Helper to create a test context with Authorization header
func newTestContextWithAuth(method, path, token string) gonest.Context {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	return gonest.NewContext(w, req)
}

// Helper to create a JWT token for testing
func createTestJWT(t *testing.T, secret string) string {
	t.Helper()
	claims := &auth.Claims{
		UserID:   "user123",
		Username: "testuser",
		Roles:    []string{"admin", "user"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "user123",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to create test JWT: %v", err)
	}
	return tokenString
}

// Tests for AuthMiddleware_Handle
func TestAuthMiddleware_Handle(t *testing.T) {
	secret := "test-secret-key"
	store := NewMockStore()
	provider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          secret,
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}, store)

	config := &auth.Config{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		ContextKey:  "user",
	}
	mw := auth.New(provider, config)

	validToken := createTestJWT(t, secret)

	tests := []struct {
		name          string
		path          string
		authHeader    string
		expectSuccess bool
		expectStatus  int
	}{
		{
			name:          "valid token in header",
			path:          "/protected",
			authHeader:    "Bearer " + validToken,
			expectSuccess: true,
			expectStatus:  http.StatusOK,
		},
		{
			name:          "valid token with custom scheme",
			path:          "/api/data",
			authHeader:    "Bearer " + validToken,
			expectSuccess: true,
			expectStatus:  http.StatusOK,
		},
		{
			name:          "no auth header",
			path:          "/protected",
			authHeader:    "",
			expectSuccess: false,
			expectStatus:  http.StatusUnauthorized,
		},
		{
			name:          "empty token",
			path:          "/protected",
			authHeader:    "Bearer ",
			expectSuccess: false,
			expectStatus:  http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContextWithPath("GET", tt.path)
			if tt.authHeader != "" {
				ctx.Request().Header.Set("Authorization", tt.authHeader)
			}

			var nextCalled bool
			err := mw.Handle(ctx, func() error {
				nextCalled = true
				return nil
			})

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
				}
				if !nextCalled {
					t.Error("Expected next() to be called")
				}

				// Verify claims are set
				userID := auth.GetUserID(ctx)
				if userID != "user123" {
					t.Errorf("Expected user_id to be 'user123', got '%s'", userID)
				}

				username := auth.GetUsername(ctx)
				if username != "testuser" {
					t.Errorf("Expected username to be 'testuser', got '%s'", username)
				}

				roles := auth.GetRoles(ctx)
				if len(roles) != 2 || roles[0] != "admin" || roles[1] != "user" {
					t.Errorf("Expected roles ['admin', 'user'], got %v", roles)
				}
			} else {
				if err == nil {
					t.Error("Expected error, got nil")
				} else {
					// Verify it's an unauthorized error
					httpErr, ok := err.(*gonest.HttpError)
					if !ok {
						t.Errorf("Expected HttpError, got %T", err)
					} else if httpErr.Status() != tt.expectStatus {
						t.Errorf("Expected status %d, got %d", tt.expectStatus, httpErr.Status())
					}
				}
				if nextCalled {
					t.Error("Expected next() NOT to be called")
				}
			}
		})
	}
}

// Tests for AuthMiddleware_SkipPaths
func TestAuthMiddleware_SkipPaths(t *testing.T) {
	secret := "test-secret-key"
	store := NewMockStore()
	provider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          secret,
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}, store)

	config := &auth.Config{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		ContextKey:  "user",
		SkipPaths:   []string{"/public", "/health", "/open"},
	}
	mw := auth.New(provider, config)

	tests := []struct {
		name    string
		path    string
		skipped bool
	}{
		{name: "/public endpoint", path: "/public", skipped: true},
		{name: "/public/sub endpoint", path: "/public/sub", skipped: true},
		{name: "/health endpoint", path: "/health", skipped: true},
		{name: "/health/status endpoint", path: "/health/status", skipped: true},
		{name: "/open endpoint", path: "/open", skipped: true},
		{name: "/protected endpoint", path: "/protected", skipped: false},
		{name: "/api/data endpoint", path: "/api/data", skipped: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContextWithPath("GET", tt.path)
			// No auth token set

			var nextCalled bool
			err := mw.Handle(ctx, func() error {
				nextCalled = true
				return nil
			})

			if tt.skipped {
				if err != nil {
					t.Errorf("Expected no error for skipped path, got: %v", err)
				}
				if !nextCalled {
					t.Error("Expected next() to be called for skipped path")
				}
			} else {
				if err == nil {
					t.Error("Expected error for non-skipped path without token")
				}
				if nextCalled {
					t.Error("Expected next() NOT to be called for non-skipped path without token")
				}
			}
		})
	}
}

// Tests for AuthMiddleware_SkipFunc
func TestAuthMiddleware_SkipFunc(t *testing.T) {
	secret := "test-secret-key"
	store := NewMockStore()
	provider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          secret,
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}, store)

	config := &auth.Config{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		ContextKey:  "user",
		SkipFunc: func(ctx gonest.Context) bool {
			return ctx.Path() == "/skip-this"
		},
	}
	mw := auth.New(provider, config)

	tests := []struct {
		name    string
		path    string
		skipped bool
	}{
		{name: "skip func returns true", path: "/skip-this", skipped: true},
		{name: "skip func returns false", path: "/protected", skipped: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContextWithPath("GET", tt.path)

			var nextCalled bool
			err := mw.Handle(ctx, func() error {
				nextCalled = true
				return nil
			})

			if tt.skipped {
				if err != nil {
					t.Errorf("Expected no error when skipped, got: %v", err)
				}
				if !nextCalled {
					t.Error("Expected next() to be called when skipped")
				}
			} else {
				if err == nil {
					t.Error("Expected error when not skipped without token")
				}
				if nextCalled {
					t.Error("Expected next() NOT to be called when not skipped without token")
				}
			}
		})
	}
}

func TestAuthMiddleware_ExtractToken(t *testing.T) {
	secret := "test-secret-key"
	store := NewMockStore()
	provider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          secret,
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}, store)

	tests := []struct {
		name        string
		tokenLookup string
		authHeader  string
		queryParam  string
		expectError bool
	}{
		{
			name:        "extract from header with Bearer",
			tokenLookup: "header:Authorization",
			authHeader:  "Bearer " + createTestJWT(t, secret),
			expectError: false,
		},
		{
			name:        "missing header returns error",
			tokenLookup: "header:Authorization",
			authHeader:  "",
			expectError: true,
		},
		{
			name:        "wrong scheme returns error",
			tokenLookup: "header:Authorization",
			authHeader:  "MAC test-token",
			expectError: true,
		},
		{
			name:        "extract from query",
			tokenLookup: "query:token",
			queryParam:  createTestJWT(t, secret),
			expectError: false,
		},
		{
			name:        "missing query param returns error",
			tokenLookup: "query:token",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContextWithPath("GET", "/test")
			if tt.authHeader != "" {
				ctx.Request().Header.Set("Authorization", tt.authHeader)
			}
			if tt.queryParam != "" {
				ctx.Request().URL.RawQuery = "token=" + tt.queryParam
			}

			mwCustom := auth.New(provider, &auth.Config{
				TokenLookup: tt.tokenLookup,
				AuthScheme:  "Bearer",
			})

			err := mwCustom.Handle(ctx, func() error {
				return nil
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// Tests for AuthMiddleware_ErrorHandling
func TestAuthMiddleware_ErrorHandling(t *testing.T) {
	secret := "test-secret-key"
	store := NewMockStore()
	provider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          secret,
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}, store)

	config := &auth.Config{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		ContextKey:  "user",
	}
	mw := auth.New(provider, config)

	tests := []struct {
		name        string
		authHeader  string
		expectError bool
	}{
		{
			name:        "missing token",
			authHeader:  "",
			expectError: true,
		},
		{
			name:        "invalid token format",
			authHeader:  "Bearer invalid.token.format",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContextWithPath("GET", "/protected")
			if tt.authHeader != "" {
				ctx.Request().Header.Set("Authorization", tt.authHeader)
			}

			err := mw.Handle(ctx, func() error {
				t.Error("next() should not be called on error")
				return nil
			})

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else {
					httpErr, ok := err.(*gonest.HttpError)
					if !ok {
						t.Errorf("Expected HttpError, got %T", err)
					} else if httpErr.Status() != http.StatusUnauthorized {
						t.Errorf("Expected status 401, got %d", httpErr.Status())
					}
				}
			}
		})
	}
}

// Tests for AuthMiddleware_CustomErrorHandler
func TestAuthMiddleware_CustomErrorHandler(t *testing.T) {
	secret := "test-secret-key"
	store := NewMockStore()
	provider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          secret,
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}, store)

	customErrorCalled := false
	config := &auth.Config{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		ContextKey:  "user",
		ErrorHandler: func(ctx gonest.Context, err error) error {
			customErrorCalled = true
			return &gonest.HttpError{
				Code: http.StatusPreconditionFailed,
				Msg:  "custom: " + err.Error(),
			}
		},
	}
	mw := auth.New(provider, config)

	ctx := newTestContextWithPath("GET", "/protected")
	err := mw.Handle(ctx, func() error {
		return nil
	})

	if !customErrorCalled {
		t.Error("Expected custom error handler to be called")
	}

	if err == nil {
		t.Error("Expected error, got nil")
	} else {
		httpErr := err.(*gonest.HttpError)
		if httpErr.Status() != http.StatusPreconditionFailed {
			t.Errorf("Expected custom status 412, got %d", httpErr.Status())
		}
	}
}

// Tests for BasicAuth
func TestAuthMiddleware_BasicAuth(t *testing.T) {
	tests := []struct {
		name       string
		username   string
		password   string
		valid      bool
		validateFn bool
	}{
		{
			name:       "valid credentials with mapped users",
			username:   "admin",
			password:   "secret123",
			valid:      true,
			validateFn: false,
		},
		{
			name:       "valid credentials with custom validator",
			username:   "admin",
			password:   "custom-pass",
			valid:      true,
			validateFn: true,
		},
		{
			name:       "invalid credentials",
			username:   "admin",
			password:   "wrong-password",
			valid:      false,
			validateFn: false,
		},
		{
			name:       "no auth header",
			username:   "",
			password:   "",
			valid:      false,
			validateFn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config *auth.BasicAuthConfig
			if tt.validateFn {
				config = &auth.BasicAuthConfig{
					ValidateFunc: func(username, password string) bool {
						return username == "admin" && password == "custom-pass"
					},
					ContextKey: "user",
				}
			} else {
				config = &auth.BasicAuthConfig{
					Users: map[string]string{
						"admin": "secret123",
						"user":  "password",
					},
					ContextKey: "user",
				}
			}

			mw := auth.NewBasicAuth(config)

			ctx := newTestContextWithPath("GET", "/protected")
			if tt.username != "" && tt.password != "" {
				ctx.Request().SetBasicAuth(tt.username, tt.password)
			}

			err := mw.Handle(ctx, func() error {
				return nil
			})

			if tt.valid {
				if err != nil {
					t.Errorf("Expected no error for valid credentials, got: %v", err)
				}

				// Verify user is set in context
				user := ctx.Get("user")
				if user == nil {
					t.Error("Expected user to be set in context")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid credentials")
				}

				httpErr := err.(*gonest.HttpError)
				if httpErr.Status() != http.StatusUnauthorized {
					t.Errorf("Expected status 401, got %d", httpErr.Status())
				}
			}
		})
	}
}

// Tests for BasicAuth_Realm
func TestAuthMiddleware_BasicAuth_Realm(t *testing.T) {
	config := &auth.BasicAuthConfig{
		Users: map[string]string{
			"admin": "secret",
		},
		Realm: "MyApp",
	}
	mw := auth.NewBasicAuth(config)

	// Test with no auth header - should get WWW-Authenticate
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	ctx := gonest.NewContext(rec, req)

	_ = mw.Handle(ctx, func() error {
		return nil
	})

	wwwAuth := rec.Header().Get("WWW-Authenticate")
	if wwwAuth != `Basic realm="MyApp"` {
		t.Errorf("Expected WWW-Authenticate header 'Basic realm=\"MyApp\"', got '%s'", wwwAuth)
	}
}

// Tests for APIKey
func TestAuthMiddleware_APIKey(t *testing.T) {
	validKeys := []string{"key123", "key456", "key789"}

	tests := []struct {
		name      string
		headerKey string
		queryKey  string
		valid     bool
	}{
		{
			name:      "valid API key in header",
			headerKey: "key123",
			valid:     true,
		},
		{
			name:     "valid API key in query",
			queryKey: "key456",
			valid:    true,
		},
		{
			name:      "invalid API key",
			headerKey: "invalid-key",
			valid:     false,
		},
		{
			name:  "no API key provided",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &auth.APIKeyConfig{
				Keys:       validKeys,
				HeaderName: "X-API-Key",
				QueryParam: "api_key",
				ContextKey: "api_key",
			}
			mw := auth.NewAPIKey(config)

			ctx := newTestContextWithPath("GET", "/protected")

			if tt.headerKey != "" {
				ctx.Request().Header.Set("X-API-Key", tt.headerKey)
			}
			if tt.queryKey != "" {
				q := ctx.Request().URL.Query()
				q.Set("api_key", tt.queryKey)
				ctx.Request().URL.RawQuery = q.Encode()
			}

			err := mw.Handle(ctx, func() error {
				return nil
			})

			if tt.valid {
				if err != nil {
					t.Errorf("Expected no error for valid API key, got: %v", err)
				}

				// Verify API key is set in context
				key := ctx.Get("api_key")
				if key == nil {
					t.Error("Expected api_key to be set in context")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid API key")
				}

				httpErr := err.(*gonest.HttpError)
				if httpErr.Status() != http.StatusUnauthorized {
					t.Errorf("Expected status 401, got %d", httpErr.Status())
				}
			}
		})
	}
}

// Tests for APIKey_CustomValidator
func TestAuthMiddleware_APIKey_CustomValidator(t *testing.T) {
	config := &auth.APIKeyConfig{
		ValidateFunc: func(key string) bool {
			return key == "custom-valid-key"
		},
		HeaderName: "X-API-Key",
		ContextKey: "api_key",
	}
	mw := auth.NewAPIKey(config)

	ctx := newTestContextWithPath("GET", "/protected")
	ctx.Request().Header.Set("X-API-Key", "custom-valid-key")

	err := mw.Handle(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error for valid custom API key, got: %v", err)
	}

	ctx2 := newTestContextWithPath("GET", "/protected")
	ctx2.Request().Header.Set("X-API-Key", "invalid-key")

	err2 := mw.Handle(ctx2, func() error {
		return nil
	})

	if err2 == nil {
		t.Error("Expected error for invalid custom API key")
	}
}

// Tests for HasRole helpers
func TestAuthMiddleware_HasRoleHelpers(t *testing.T) {
	secret := "test-secret-key"
	store := NewMockStore()
	provider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          secret,
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}, store)

	mw := auth.New(provider, &auth.Config{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		ContextKey:  "user",
	})

	validToken := createTestJWT(t, secret)

	tests := []struct {
		name       string
		hasRole    string
		anyOfRoles []string
		allOfRoles []string
		expect     bool
	}{
		{
			name:    "has admin role",
			hasRole: "admin",
			expect:  true,
		},
		{
			name:    "has user role",
			hasRole: "user",
			expect:  true,
		},
		{
			name:       "has any role - matches",
			anyOfRoles: []string{"guest", "user"},
			expect:     true,
		},
		{
			name:       "has any role - no match",
			anyOfRoles: []string{"guest", "moderator"},
			expect:     false,
		},
		{
			name:       "has all roles - matches",
			allOfRoles: []string{"admin", "user"},
			expect:     true,
		},
		{
			name:       "has all roles - partial match fails",
			allOfRoles: []string{"admin", "guest"},
			expect:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContextWithPath("GET", "/protected")
			ctx.Request().Header.Set("Authorization", "Bearer "+validToken)

			_ = mw.Handle(ctx, func() error {
				return nil
			})

			if tt.hasRole != "" {
				result := auth.HasRole(ctx, tt.hasRole)
				if result != tt.expect {
					t.Errorf("HasRole('%s') = %v, expected %v", tt.hasRole, result, tt.expect)
				}
			}

			if len(tt.anyOfRoles) > 0 {
				result := auth.HasAnyRole(ctx, tt.anyOfRoles...)
				if result != tt.expect {
					t.Errorf("HasAnyRole(%v) = %v, expected %v", tt.anyOfRoles, result, tt.expect)
				}
			}

			if len(tt.allOfRoles) > 0 {
				result := auth.HasAllRoles(ctx, tt.allOfRoles...)
				if result != tt.expect {
					t.Errorf("HasAllRoles(%v) = %v, expected %v", tt.allOfRoles, result, tt.expect)
				}
			}
		})
	}
}

// Tests for HasRoleHelpers on empty context
func TestAuthMiddleware_HasRoleHelpers_EmptyContext(t *testing.T) {
	ctx := newTestContextWithPath("GET", "/")

	// Should return false/nil/empty for empty context
	if auth.HasRole(ctx, "admin") {
		t.Error("HasRole should return false when no claims in context")
	}

	if auth.HasAnyRole(ctx, "admin", "user") {
		t.Error("HasAnyRole should return false when no claims in context")
	}

	if auth.HasAllRoles(ctx, "admin", "user") {
		t.Error("HasAllRoles should return false when no claims in context")
	}

	if auth.GetUserID(ctx) != "" {
		t.Error("GetUserID should return empty string when no claims in context")
	}

	if auth.GetUsername(ctx) != "" {
		t.Error("GetUsername should return empty string when no claims in context")
	}

	if roles := auth.GetRoles(ctx); roles != nil {
		t.Errorf("GetRoles should return nil when no claims in context, got %v", roles)
	}
}

// Tests for AuthMiddleware_WithRefresh
func TestAuthMiddleware_WithRefresh(t *testing.T) {
	secret := "test-secret-key"
	store := NewMockStore()
	provider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          secret,
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}, store)

	config := &auth.Config{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
	}
	mw := auth.New(provider, config)

	refreshedMW := mw.WithRefresh(&auth.RefreshConfig{
		Enabled:           true,
		Threshold:         5 * time.Minute,
		RefreshHeaderName: "X-Refresh-Token",
	})
	if refreshedMW == nil {
		t.Error("WithRefresh should return non-nil middleware")
	}
}

// TestGetClaims helper function
func TestAuthMiddleware_GetClaims(t *testing.T) {
	secret := "test-secret-key"
	store := NewMockStore()
	provider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          secret,
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}, store)

	mw := auth.New(provider, &auth.Config{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
	})

	validToken := createTestJWT(t, secret)
	ctx := newTestContextWithPath("GET", "/protected")
	ctx.Request().Header.Set("Authorization", "Bearer "+validToken)

	var nextCalled bool
	_ = mw.Handle(ctx, func() error {
		nextCalled = true

		// Test GetClaims
		claims := auth.GetClaims(ctx)
		if claims == nil {
			t.Error("GetClaims should return non-nil claims")
		} else {
			if claims.UserID != "user123" {
				t.Errorf("Expected UserID 'user123', got '%s'", claims.UserID)
			}
			if claims.Username != "testuser" {
				t.Errorf("Expected Username 'testuser', got '%s'", claims.Username)
			}
		}
		return nil
	})

	if !nextCalled {
		t.Error("next() should be called")
	}

	// Test GetClaims on empty context
	emptyCtx := newTestContextWithPath("GET", "/")
	claims := auth.GetClaims(emptyCtx)
	if claims != nil {
		t.Error("GetClaims should return nil when no claims in context")
	}
}

// TestGetString helper function
func TestAuthMiddleware_GetString(t *testing.T) {
	ctx := newTestContextWithPath("GET", "/")
	ctx.Set("test_key", "test_value")
	ctx.Set("empty_key", "")

	val1 := auth.GetString(ctx, "test_key")
	if val1 != "test_value" {
		t.Errorf("GetString('test_key') = '%s', expected 'test_value'", val1)
	}

	val2 := auth.GetString(ctx, "empty_key")
	if val2 != "" {
		t.Errorf("GetString('empty_key') = '%s', expected ''", val2)
	}

	val3 := auth.GetString(ctx, "nonexistent")
	if val3 != "" {
		t.Errorf("GetString('nonexistent') = '%s', expected ''", val3)
	}
}

// Test for DefaultConfig
func TestDefaultConfig(t *testing.T) {
	config := auth.DefaultConfig()

	if config.TokenLookup != "header:Authorization" {
		t.Errorf("Expected TokenLookup 'header:Authorization', got '%s'", config.TokenLookup)
	}
	if config.TokenHeader != "Authorization" {
		t.Errorf("Expected TokenHeader 'Authorization', got '%s'", config.TokenHeader)
	}
	if config.AuthScheme != "Bearer" {
		t.Errorf("Expected AuthScheme 'Bearer', got '%s'", config.AuthScheme)
	}
	if config.ContextKey != "user" {
		t.Errorf("Expected ContextKey 'user', got '%s'", config.ContextKey)
	}
	if len(config.SkipPaths) != 0 {
		t.Errorf("Expected SkipPaths to be empty, got %v", config.SkipPaths)
	}
}

// Test JWTProvider integration with AuthMiddleware
func TestAuthMiddleware_JWTProviderIntegration(t *testing.T) {
	secret := "test-secret-key"
	store := NewMockStore()
	provider := auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          secret,
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}, store)

	mw := auth.New(provider, &auth.Config{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
	})

	// Generate a token
	tokenPair, err := provider.GenerateTokenPair("user123", "testuser", []string{"admin", "user"}, nil)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	validToken := tokenPair.AccessToken

	ctx := newTestContextWithPath("GET", "/protected")
	ctx.Request().Header.Set("Authorization", "Bearer "+validToken)

	var nextCalled bool
	err = mw.Handle(ctx, func() error {
		nextCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error with valid token, got: %v", err)
	}
	if !nextCalled {
		t.Error("next() should be called with valid token")
	}

	// Create expired token manually
	claims := &auth.Claims{
		UserID:   "user123",
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   "user123",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, _ := token.SignedString([]byte(secret))

	ctx2 := newTestContextWithPath("GET", "/protected")
	ctx2.Request().Header.Set("Authorization", "Bearer "+expiredToken)

	err2 := mw.Handle(ctx2, func() error {
		t.Error("next() should not be called with expired token")
		return nil
	})

	if err2 == nil {
		t.Error("Expected error with expired token")
	} else {
		httpErr := err2.(*gonest.HttpError)
		if httpErr.Status() != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", httpErr.Status())
		}
	}
}
