package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("token has expired")
	ErrMissingToken         = errors.New("missing token")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
)

type Claims struct {
	UserID   string         `json:"user_id"`
	Username string         `json:"username"`
	Roles    []string       `json:"roles,omitempty"`
	Extra    map[string]any `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type TokenStore interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (any, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type JWTConfig struct {
	Secret          string
	SigningMethod   jwt.SigningMethod
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
	Audience        []string
}

func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		Secret:          "your-secret-key",
		SigningMethod:   jwt.SigningMethodHS256,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "gonest",
	}
}

type JWTProvider struct {
	config *JWTConfig
	store  TokenStore
}

func NewJWTProvider(config *JWTConfig, store TokenStore) *JWTProvider {
	if config == nil {
		config = DefaultJWTConfig()
	}
	return &JWTProvider{
		config: config,
		store:  store,
	}
}

func (p *JWTProvider) GenerateTokenPair(userID, username string, roles []string, extra map[string]any) (*TokenPair, error) {
	now := time.Now()
	accessExpiresAt := now.Add(p.config.AccessTokenTTL)
	refreshExpiresAt := now.Add(p.config.RefreshTokenTTL)

	accessClaims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		Extra:    extra,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userID,
			Issuer:    p.config.Issuer,
			Audience:  p.config.Audience,
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken, err := p.generateToken(&accessClaims)
	if err != nil {
		return nil, err
	}

	refreshID := uuid.New().String()
	refreshClaims := jwt.RegisteredClaims{
		ID:        refreshID,
		Subject:   userID,
		Issuer:    p.config.Issuer,
		ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	refreshToken, err := p.generateToken(&refreshClaims)
	if err != nil {
		return nil, err
	}

	if p.store != nil {
		ctx := context.Background()
		refreshKey := "refresh:" + refreshID
		if err := p.store.Set(ctx, refreshKey, map[string]any{
			"user_id":    userID,
			"username":   username,
			"roles":      roles,
			"created_at": now,
		}, p.config.RefreshTokenTTL); err != nil {
			return nil, err
		}
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(p.config.AccessTokenTTL.Seconds()),
	}, nil
}

func (p *JWTProvider) generateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(p.config.SigningMethod, claims)
	return token.SignedString([]byte(p.config.Secret))
}

func (p *JWTProvider) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != p.config.SigningMethod.Alg() {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(p.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (p *JWTProvider) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := p.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if p.store != nil {
		ctx := context.Background()
		refreshKey := "refresh:" + claims.ID
		exists, err := p.store.Exists(ctx, refreshKey)
		if err != nil || !exists {
			return nil, ErrInvalidToken
		}

		p.store.Delete(ctx, refreshKey)
	}

	return p.GenerateTokenPair(claims.Subject, "", nil, nil)
}

func (p *JWTProvider) RevokeToken(ctx context.Context, tokenID string) error {
	if p.store == nil {
		return nil
	}
	refreshKey := "refresh:" + tokenID
	return p.store.Delete(ctx, refreshKey)
}

func (p *JWTProvider) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if p.store == nil {
		return nil
	}
	userKey := "user_tokens:" + userID
	return p.store.Delete(ctx, userKey)
}
