package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	Limit     int
	Window    time.Duration
	KeyFunc   func(ctx abstract.Context) string
	ErrorCode int
	ErrorMsg  string
	SkipFunc  func(ctx abstract.Context) bool
}

func DefaultConfig() *Config {
	return &Config{
		Limit:     100,
		Window:    time.Minute,
		ErrorCode: http.StatusTooManyRequests,
		ErrorMsg:  "rate limit exceeded",
		KeyFunc: func(ctx abstract.Context) string {
			return ctx.Request().RemoteAddr
		},
	}
}

type Limiter struct {
	mu       sync.Mutex
	requests map[string]*clientInfo
	limit    int
	window   time.Duration
}

type clientInfo struct {
	count     int
	expiresAt time.Time
}

func NewLimiter(limit int, window time.Duration) *Limiter {
	rl := &Limiter{
		requests: make(map[string]*clientInfo),
		limit:    limit,
		window:   window,
	}

	go rl.cleanup()

	return rl
}

func (rl *Limiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.requests[key]

	if !exists || now.After(info.expiresAt) {
		rl.requests[key] = &clientInfo{
			count:     1,
			expiresAt: now.Add(rl.window),
		}
		return true
	}

	if info.count >= rl.limit {
		return false
	}

	info.count++
	return true
}

func (rl *Limiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.requests, key)
}

func (rl *Limiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, info := range rl.requests {
			if now.After(info.expiresAt) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

func New(cfg *Config) abstract.Middleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = DefaultConfig().KeyFunc
	}

	limiter := NewLimiter(cfg.Limit, cfg.Window)

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		if cfg.SkipFunc != nil && cfg.SkipFunc(ctx) {
			return next()
		}

		key := cfg.KeyFunc(ctx)

		if !limiter.Allow(key) {
			return abstract.NewHttpException(cfg.ErrorCode, cfg.ErrorMsg)
		}

		return next()
	})
}

func NewWithLimiter(limiter *Limiter, keyFunc func(ctx abstract.Context) string) abstract.Middleware {
	if keyFunc == nil {
		keyFunc = DefaultConfig().KeyFunc
	}

	return abstract.MiddlewareFunc(func(ctx abstract.Context, next func() error) error {
		key := keyFunc(ctx)

		if !limiter.Allow(key) {
			return abstract.NewHttpException(http.StatusTooManyRequests, "rate limit exceeded")
		}

		return next()
	})
}
