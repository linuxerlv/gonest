package auth

import (
	"context"
	"time"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
)

type RefreshConfig struct {
	Enabled           bool
	Threshold         time.Duration
	RefreshHeaderName string
	MaxRefreshCount   int
}

func DefaultRefreshConfig() *RefreshConfig {
	return &RefreshConfig{
		Enabled:           true,
		Threshold:         5 * time.Minute,
		RefreshHeaderName: "X-Refresh-Token",
		MaxRefreshCount:   100,
	}
}

type RefreshMiddleware struct {
	provider  *JWTProvider
	config    *RefreshConfig
	blacklist TokenBlacklist
}

type TokenBlacklist interface {
	Add(ctx context.Context, token string, ttl time.Duration) error
	Exists(ctx context.Context, token string) (bool, error)
}

func NewRefreshMiddleware(provider *JWTProvider, config *RefreshConfig, blacklist TokenBlacklist) *RefreshMiddleware {
	if config == nil {
		config = DefaultRefreshConfig()
	}
	return &RefreshMiddleware{
		provider:  provider,
		config:    config,
		blacklist: blacklist,
	}
}

func (m *RefreshMiddleware) shouldRefresh(expiresAt time.Time) bool {
	if !m.config.Enabled {
		return false
	}
	remaining := time.Until(expiresAt)
	return remaining > 0 && remaining <= m.config.Threshold
}

func (m *RefreshMiddleware) generateNewToken(claims *Claims) (*TokenPair, error) {
	return m.provider.GenerateTokenPair(
		claims.UserID,
		claims.Username,
		claims.Roles,
		claims.Extra,
	)
}

func (m *RefreshMiddleware) Handle(ctx abstract.Context, next func() error) error {
	err := next()

	if !m.config.Enabled {
		return err
	}

	claims, ok := ctx.Get("jwt_claims").(*Claims)
	if !ok {
		return err
	}

	if m.shouldRefresh(claims.ExpiresAt.Time) {
		newTokenPair, err := m.generateNewToken(claims)
		if err != nil {
			return err
		}

		hc := ctx.(*core.HttpContext)
		hc.ResponseWriter().Header().Set(m.config.RefreshHeaderName, newTokenPair.AccessToken)
	}

	return nil
}

func (m *RefreshMiddleware) WithRefreshEndpoint(path string) abstract.RouteHandler {
	return func(ctx abstract.Context) error {
		refreshToken := ctx.Header("X-Refresh-Token")
		if refreshToken == "" {
			return abstract.Unauthorized("missing refresh token")
		}

		if m.blacklist != nil {
			blacklisted, err := m.blacklist.Exists(ctx.Context(), refreshToken)
			if err != nil {
				return abstract.InternalError("failed to check token status")
			}
			if blacklisted {
				return abstract.Unauthorized("token has been revoked")
			}
		}

		newTokenPair, err := m.provider.RefreshToken(refreshToken)
		if err != nil {
			return abstract.Unauthorized("invalid refresh token")
		}

		return ctx.JSON(200, map[string]any{
			"success": true,
			"data":    newTokenPair,
		})
	}
}
