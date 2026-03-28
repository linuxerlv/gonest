package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type ProviderConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type Config struct {
	Providers      []ProviderConfig
	SuccessURL     string
	FailureURL     string
	SessionSecret  string
	SessionMaxAge  int
	UserContextKey string
	OnUserLogin    func(ctx abstract.Context, user goth.User) error
	SkipPaths      []string
}

func DefaultConfig() *Config {
	return &Config{
		SuccessURL:     "/",
		FailureURL:     "/login",
		SessionMaxAge:  86400,
		UserContextKey: "oauth_user",
	}
}

type OAuthMiddleware struct {
	config    *Config
	skipPaths map[string]bool
}

func New(cfg *Config) *OAuthMiddleware {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	for _, p := range cfg.Providers {
		setupProvider(p)
	}

	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return &OAuthMiddleware{
		config:    cfg,
		skipPaths: skipPaths,
	}
}

func setupProvider(cfg ProviderConfig) {
	switch cfg.Name {
	case "google":
		goth.UseProviders(
			google.New(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL, cfg.Scopes...),
		)
	}
}

func (m *OAuthMiddleware) BeginAuthHandler(provider string) abstract.RouteHandler {
	return func(ctx abstract.Context) error {
		w := ctx.ResponseWriter()
		r := ctx.Request()

		q := r.URL.Query()
		q.Add("provider", provider)
		r.URL.RawQuery = q.Encode()

		gothic.BeginAuthHandler(w, r)
		return nil
	}
}

func (m *OAuthMiddleware) CallbackHandler(provider string) abstract.RouteHandler {
	return func(ctx abstract.Context) error {
		w := ctx.ResponseWriter()
		r := ctx.Request()

		q := r.URL.Query()
		q.Add("provider", provider)
		r.URL.RawQuery = q.Encode()

		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			return abstract.BadRequest(fmt.Sprintf("oauth callback failed: %v", err))
		}

		ctx.Set(m.config.UserContextKey, user)

		if m.config.OnUserLogin != nil {
			if err := m.config.OnUserLogin(ctx, user); err != nil {
				return err
			}
		}

		return ctx.JSON(http.StatusOK, map[string]any{
			"success": true,
			"user":    user,
		})
	}
}

func (m *OAuthMiddleware) LogoutHandler() abstract.RouteHandler {
	return func(ctx abstract.Context) error {
		w := ctx.ResponseWriter()
		r := ctx.Request()

		err := gothic.Logout(w, r)
		if err != nil {
			return abstract.InternalError(fmt.Sprintf("logout failed: %v", err))
		}

		return ctx.JSON(http.StatusOK, map[string]any{
			"success": true,
			"message": "logged out",
		})
	}
}

func (m *OAuthMiddleware) GetUser(ctx abstract.Context) *goth.User {
	if user, ok := ctx.Get(m.config.UserContextKey).(goth.User); ok {
		return &user
	}
	return nil
}

func (m *OAuthMiddleware) IsAuthenticated(ctx abstract.Context) bool {
	return m.GetUser(ctx) != nil
}

func (m *OAuthMiddleware) Handle(ctx abstract.Context, next func() error) error {
	if m.shouldSkip(ctx) {
		return next()
	}

	w := ctx.ResponseWriter()
	r := ctx.Request()

	user, err := gothic.CompleteUserAuth(w, r)
	if err == nil {
		ctx.Set(m.config.UserContextKey, user)
	}

	return next()
}

func (m *OAuthMiddleware) shouldSkip(ctx abstract.Context) bool {
	path := ctx.Path()
	return m.skipPaths[path]
}

func (m *OAuthMiddleware) AsMiddleware() abstract.Middleware {
	return abstract.MiddlewareFunc(m.Handle)
}

func RegisterProviders(providers ...ProviderConfig) {
	for _, p := range providers {
		switch p.Name {
		case "google":
			goth.UseProviders(
				google.New(p.ClientID, p.ClientSecret, p.RedirectURL, p.Scopes...),
			)
		}
	}
}

func GetUser(ctx abstract.Context) *goth.User {
	if user, ok := ctx.Get("oauth_user").(goth.User); ok {
		return &user
	}
	return nil
}

func GetUserID(ctx abstract.Context) string {
	user := GetUser(ctx)
	if user != nil {
		return user.UserID
	}
	return ""
}

func GetUserName(ctx abstract.Context) string {
	user := GetUser(ctx)
	if user != nil {
		return user.Name
	}
	return ""
}

func GetEmail(ctx abstract.Context) string {
	user := GetUser(ctx)
	if user != nil {
		return user.Email
	}
	return ""
}

func GetAvatarURL(ctx abstract.Context) string {
	user := GetUser(ctx)
	if user != nil {
		return user.AvatarURL
	}
	return ""
}

func GetProvider(ctx abstract.Context) string {
	user := GetUser(ctx)
	if user != nil {
		return user.Provider
	}
	return ""
}

type OAuthUser struct {
	Provider    string `json:"provider"`
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	NickName    string `json:"nick_name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
	Location    string `json:"location"`
	AccessToken string `json:"-"`
	ExpiresAt   string `json:"expires_at"`
}

func ToOAuthUser(user goth.User) *OAuthUser {
	return &OAuthUser{
		Provider:    user.Provider,
		UserID:      user.UserID,
		Email:       user.Email,
		Name:        user.Name,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		NickName:    user.NickName,
		Description: user.Description,
		AvatarURL:   user.AvatarURL,
		Location:    user.Location,
		AccessToken: user.AccessToken,
		ExpiresAt:   user.ExpiresAt.String(),
	}
}

func (u *OAuthUser) ToJSON() string {
	data, _ := json.Marshal(u)
	return string(data)
}

type Router interface {
	GET(path string, handler abstract.RouteHandler)
}

func RegisterOAuthRoutes(router Router, mw *OAuthMiddleware, providers []string) {
	for _, provider := range providers {
		router.GET("/auth/"+provider, mw.BeginAuthHandler(provider))
		router.GET("/auth/"+provider+"/callback", mw.CallbackHandler(provider))
	}
	router.GET("/auth/logout", mw.LogoutHandler())
}
