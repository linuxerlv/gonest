package session

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	SessionName    string
	ContextKey     string
	SkipPaths      []string
	ErrorHandler   func(ctx abstract.Context, err error) error
	SessionManager *scs.SessionManager
}

func DefaultConfig() *Config {
	return &Config{
		SessionName: "session",
		ContextKey:  "session",
		SkipPaths:   []string{},
	}
}

type SessionMiddleware struct {
	config    *Config
	skipPaths map[string]bool
}

func New(cfg *Config) *SessionMiddleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return &SessionMiddleware{
		config:    cfg,
		skipPaths: skipPaths,
	}
}

func (m *SessionMiddleware) Handle(ctx abstract.Context, next func() error) error {
	if m.shouldSkip(ctx) {
		return next()
	}

	sm := m.config.SessionManager
	if sm == nil {
		return next()
	}

	ctx.Set(m.config.ContextKey, sm)

	var handlerErr error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.(*core.HttpContext).SetRequest(r)
		handlerErr = next()
	})

	hc := ctx.(*core.HttpContext)
	sm.LoadAndSave(handler).ServeHTTP(hc.ResponseWriter(), ctx.Request())

	return handlerErr
}

func (m *SessionMiddleware) shouldSkip(ctx abstract.Context) bool {
	path := ctx.Path()
	return m.skipPaths[path]
}

func (m *SessionMiddleware) AsMiddleware() abstract.Middleware {
	return abstract.MiddlewareFunc(m.Handle)
}

func GetSession(ctx abstract.Context) *scs.SessionManager {
	if sm, ok := ctx.Get("session").(*scs.SessionManager); ok {
		return sm
	}
	return nil
}

func Get(ctx abstract.Context, key string) any {
	sm := GetSession(ctx)
	if sm == nil {
		return nil
	}
	return sm.Get(ctx.Context(), key)
}

func GetString(ctx abstract.Context, key string) string {
	sm := GetSession(ctx)
	if sm == nil {
		return ""
	}
	return sm.GetString(ctx.Context(), key)
}

func GetInt(ctx abstract.Context, key string) int {
	sm := GetSession(ctx)
	if sm == nil {
		return 0
	}
	return sm.GetInt(ctx.Context(), key)
}

func GetInt64(ctx abstract.Context, key string) int64 {
	sm := GetSession(ctx)
	if sm == nil {
		return 0
	}
	return sm.GetInt64(ctx.Context(), key)
}

func GetBool(ctx abstract.Context, key string) bool {
	sm := GetSession(ctx)
	if sm == nil {
		return false
	}
	return sm.GetBool(ctx.Context(), key)
}

func GetBytes(ctx abstract.Context, key string) []byte {
	sm := GetSession(ctx)
	if sm == nil {
		return nil
	}
	return sm.GetBytes(ctx.Context(), key)
}

func GetTime(ctx abstract.Context, key string) time.Time {
	sm := GetSession(ctx)
	if sm == nil {
		return time.Time{}
	}
	return sm.GetTime(ctx.Context(), key)
}

func Put(ctx abstract.Context, key string, value any) {
	sm := GetSession(ctx)
	if sm != nil {
		sm.Put(ctx.Context(), key, value)
	}
}

func Remove(ctx abstract.Context, key string) {
	sm := GetSession(ctx)
	if sm != nil {
		sm.Remove(ctx.Context(), key)
	}
}

func Clear(ctx abstract.Context) {
	sm := GetSession(ctx)
	if sm != nil {
		sm.Clear(ctx.Context())
	}
}

func Destroy(ctx abstract.Context) {
	sm := GetSession(ctx)
	if sm != nil {
		sm.Destroy(ctx.Context())
	}
}

func RenewToken(ctx abstract.Context) error {
	sm := GetSession(ctx)
	if sm != nil {
		return sm.RenewToken(ctx.Context())
	}
	return nil
}

func Status(ctx abstract.Context) bool {
	sm := GetSession(ctx)
	if sm != nil {
		return sm.Status(ctx.Context()) != scs.Unmodified
	}
	return false
}

func GetUserID(ctx abstract.Context) string {
	return GetString(ctx, "user_id")
}

func SetUserID(ctx abstract.Context, userID string) {
	Put(ctx, "user_id", userID)
}

func GetUser(ctx abstract.Context, dest any) bool {
	data := GetBytes(ctx, "user")
	if data == nil {
		return false
	}
	return json.Unmarshal(data, dest) == nil
}

func SetUser(ctx abstract.Context, user any) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	Put(ctx, "user", data)
	return nil
}

func ClearUser(ctx abstract.Context) {
	Remove(ctx, "user_id")
	Remove(ctx, "user")
}

func IsAuthenticated(ctx abstract.Context) bool {
	return GetUserID(ctx) != ""
}

type inMemoryItem struct {
	data   []byte
	expiry time.Time
}

type InMemoryStore struct {
	items map[string]*inMemoryItem
	mu    sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		items: make(map[string]*inMemoryItem),
	}
}

func (s *InMemoryStore) Delete(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, token)
	return nil
}

func (s *InMemoryStore) Find(token string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, found := s.items[token]
	if !found || item.expiry.Before(time.Now()) {
		return nil, false, nil
	}
	return item.data, true, nil
}

func (s *InMemoryStore) Commit(token string, b []byte, expiry time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[token] = &inMemoryItem{
		data:   b,
		expiry: expiry,
	}
	return nil
}

func NewSessionManager(store scs.Store) *scs.SessionManager {
	sm := scs.New()
	sm.Store = store
	sm.Cookie.Name = "session_id"
	sm.Cookie.HttpOnly = true
	sm.Cookie.Secure = true
	sm.Cookie.SameSite = http.SameSiteStrictMode
	sm.Lifetime = 24 * time.Hour
	sm.IdleTimeout = 2 * time.Hour
	return sm
}

func WithMemoryStore() *scs.SessionManager {
	return NewSessionManager(NewInMemoryStore())
}

func WithRedisStore(addr string, prefix string, poolSize int) (*scs.SessionManager, error) {
	return NewSessionManager(nil), nil
}

func WithBadgerStore(path string) (*scs.SessionManager, error) {
	return NewSessionManager(nil), nil
}
