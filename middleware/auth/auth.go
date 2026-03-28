package auth

import (
	"crypto/subtle"
	"strings"

	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	TokenLookup    string
	TokenHeader    string
	AuthScheme     string
	ContextKey     string
	SkipPaths      []string
	SkipFunc       func(ctx abstract.Context) bool
	SuccessHandler func(ctx abstract.Context) error
	ErrorHandler   func(ctx abstract.Context, err error) error
}

func DefaultConfig() *Config {
	return &Config{
		TokenLookup: "header:Authorization",
		TokenHeader: "Authorization",
		AuthScheme:  "Bearer",
		ContextKey:  "user",
		SkipPaths:   []string{},
	}
}

type AuthMiddleware struct {
	provider *JWTProvider
	config   *Config
}

func New(provider *JWTProvider, config *Config) *AuthMiddleware {
	if config == nil {
		config = DefaultConfig()
	} else {
		if config.TokenLookup == "" {
			config.TokenLookup = "header:Authorization"
		}
		if config.TokenHeader == "" {
			config.TokenHeader = "Authorization"
		}
		if config.AuthScheme == "" {
			config.AuthScheme = "Bearer"
		}
		if config.ContextKey == "" {
			config.ContextKey = "user"
		}
		if config.SkipPaths == nil {
			config.SkipPaths = []string{}
		}
	}
	return &AuthMiddleware{
		provider: provider,
		config:   config,
	}
}

func (m *AuthMiddleware) Handle(ctx abstract.Context, next func() error) error {
	if m.shouldSkip(ctx) {
		return next()
	}

	token, err := m.extractToken(ctx)
	if err != nil {
		return m.handleError(ctx, err)
	}

	claims, err := m.provider.ValidateToken(token)
	if err != nil {
		return m.handleError(ctx, err)
	}

	ctx.Set(m.config.ContextKey, claims)
	ctx.Set("jwt_claims", claims)
	ctx.Set("user_id", claims.UserID)
	ctx.Set("username", claims.Username)
	ctx.Set("roles", claims.Roles)

	if m.config.SuccessHandler != nil {
		if err := m.config.SuccessHandler(ctx); err != nil {
			return err
		}
	}

	return next()
}

func (m *AuthMiddleware) shouldSkip(ctx abstract.Context) bool {
	if m.config.SkipFunc != nil && m.config.SkipFunc(ctx) {
		return true
	}

	path := ctx.Path()
	for _, skipPath := range m.config.SkipPaths {
		if skipPath == path || strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

func (m *AuthMiddleware) extractToken(ctx abstract.Context) (string, error) {
	parts := strings.Split(m.config.TokenLookup, ":")
	if len(parts) != 2 {
		return "", ErrMissingToken
	}

	source := parts[0]
	key := parts[1]

	switch source {
	case "header":
		authHeader := ctx.Header(key)
		if authHeader == "" {
			return "", ErrMissingToken
		}

		if m.config.AuthScheme != "" {
			if !strings.HasPrefix(authHeader, m.config.AuthScheme+" ") {
				return "", ErrInvalidToken
			}
			return strings.TrimPrefix(authHeader, m.config.AuthScheme+" "), nil
		}
		return authHeader, nil

	case "query":
		token := ctx.Query(key)
		if token == "" {
			return "", ErrMissingToken
		}
		return token, nil

	case "cookie":
		cookie, err := ctx.Request().Cookie(key)
		if err != nil || cookie.Value == "" {
			return "", ErrMissingToken
		}
		return cookie.Value, nil

	case "param":
		token := ctx.Param(key)
		if token == "" {
			return "", ErrMissingToken
		}
		return token, nil

	default:
		return "", ErrMissingToken
	}
}

func (m *AuthMiddleware) handleError(ctx abstract.Context, err error) error {
	if m.config.ErrorHandler != nil {
		return m.config.ErrorHandler(ctx, err)
	}

	switch err {
	case ErrMissingToken:
		return abstract.Unauthorized("missing authentication token")
	case ErrExpiredToken:
		return abstract.Unauthorized("token has expired")
	case ErrInvalidToken:
		return abstract.Unauthorized("invalid token")
	default:
		return abstract.Unauthorized("authentication failed")
	}
}

func (m *AuthMiddleware) AsMiddleware() abstract.Middleware {
	return abstract.MiddlewareFunc(m.Handle)
}

func GetClaims(ctx abstract.Context) *Claims {
	if claims, ok := ctx.Get("jwt_claims").(*Claims); ok {
		return claims
	}
	return nil
}

func GetUserID(ctx abstract.Context) string {
	if claims := GetClaims(ctx); claims != nil {
		return claims.UserID
	}
	if userID, ok := ctx.Get("user_id").(string); ok {
		return userID
	}
	return ""
}

func GetUsername(ctx abstract.Context) string {
	if claims := GetClaims(ctx); claims != nil {
		return claims.Username
	}
	if username, ok := ctx.Get("username").(string); ok {
		return username
	}
	return ""
}

func GetRoles(ctx abstract.Context) []string {
	if claims := GetClaims(ctx); claims != nil {
		return claims.Roles
	}
	if roles, ok := ctx.Get("roles").([]string); ok {
		return roles
	}
	return nil
}

func GetString(ctx abstract.Context, key string) string {
	if val, ok := ctx.Get(key).(string); ok {
		return val
	}
	return ""
}

func HasRole(ctx abstract.Context, role string) bool {
	roles := GetRoles(ctx)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func HasAnyRole(ctx abstract.Context, roles ...string) bool {
	userRoles := GetRoles(ctx)
	for _, userRole := range userRoles {
		for _, requiredRole := range roles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

func HasAllRoles(ctx abstract.Context, roles ...string) bool {
	userRoles := GetRoles(ctx)
	for _, requiredRole := range roles {
		found := false
		for _, userRole := range userRoles {
			if userRole == requiredRole {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (m *AuthMiddleware) WithRefresh(config *RefreshConfig) *AuthMiddleware {
	_ = NewRefreshMiddleware(m.provider, config, nil)
	return m
}

type BasicAuthConfig struct {
	Users        map[string]string
	Realm        string
	ContextKey   string
	SkipPaths    []string
	ValidateFunc func(username, password string) bool
}

func NewBasicAuth(config *BasicAuthConfig) abstract.Middleware {
	if config == nil {
		config = &BasicAuthConfig{
			Users:      make(map[string]string),
			Realm:      "Restricted",
			ContextKey: "user",
		}
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		username, password, ok := ctx.Request().BasicAuth()
		if !ok {
			ctx.ResponseWriter().Header().Set("WWW-Authenticate", `Basic realm="`+config.Realm+`"`)
			return abstract.Unauthorized("authentication required")
		}

		if config.ValidateFunc != nil {
			if !config.ValidateFunc(username, password) {
				return abstract.Unauthorized("invalid credentials")
			}
		} else {
			expectedPassword, exists := config.Users[username]
			if !exists || subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) != 1 {
				return abstract.Unauthorized("invalid credentials")
			}
		}

		ctx.Set(config.ContextKey, username)
		return next()
	})
}

type APIKeyConfig struct {
	Keys         []string
	HeaderName   string
	QueryParam   string
	ContextKey   string
	SkipPaths    []string
	ValidateFunc func(key string) bool
}

func NewAPIKey(config *APIKeyConfig) abstract.Middleware {
	if config == nil {
		config = &APIKeyConfig{
			HeaderName: "X-API-Key",
			ContextKey: "api_key",
		}
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		var key string

		if config.HeaderName != "" {
			key = ctx.Header(config.HeaderName)
		}

		if key == "" && config.QueryParam != "" {
			key = ctx.Query(config.QueryParam)
		}

		if key == "" {
			return abstract.Unauthorized("missing API key")
		}

		if config.ValidateFunc != nil {
			if !config.ValidateFunc(key) {
				return abstract.Unauthorized("invalid API key")
			}
		} else {
			valid := false
			for _, validKey := range config.Keys {
				if subtle.ConstantTimeCompare([]byte(key), []byte(validKey)) == 1 {
					valid = true
					break
				}
			}
			if !valid {
				return abstract.Unauthorized("invalid API key")
			}
		}

		ctx.Set(config.ContextKey, key)
		return next()
	})
}
