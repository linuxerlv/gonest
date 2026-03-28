package casbin

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	ModelPath  string
	PolicyPath string
	Adapter    persist.Adapter
	Enforcer   *casbin.Enforcer
	SkipPaths  []string
	SubGetter  func(ctx abstract.Context) string
	ObjGetter  func(ctx abstract.Context) string
	ActGetter  func(ctx abstract.Context) string
}

func DefaultConfig() *Config {
	return &Config{
		SubGetter: func(ctx abstract.Context) string {
			if userID := ctx.Get("user_id"); userID != nil {
				return fmt.Sprintf("%v", userID)
			}
			return ""
		},
		ObjGetter: func(ctx abstract.Context) string {
			return ctx.Path()
		},
		ActGetter: func(ctx abstract.Context) string {
			return ctx.Method()
		},
	}
}

type CasbinMiddleware struct {
	enforcer  *casbin.Enforcer
	config    *Config
	skipPaths map[string]bool
}

func New(cfg *Config) (*CasbinMiddleware, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	} else {
		if cfg.SubGetter == nil {
			cfg.SubGetter = DefaultConfig().SubGetter
		}
		if cfg.ObjGetter == nil {
			cfg.ObjGetter = DefaultConfig().ObjGetter
		}
		if cfg.ActGetter == nil {
			cfg.ActGetter = DefaultConfig().ActGetter
		}
		if cfg.SkipPaths == nil {
			cfg.SkipPaths = []string{}
		}
	}

	var enforcer *casbin.Enforcer
	var err error

	if cfg.Enforcer != nil {
		enforcer = cfg.Enforcer
	} else if cfg.ModelPath != "" && cfg.PolicyPath != "" {
		adapter := fileadapter.NewAdapter(cfg.PolicyPath)
		enforcer, err = casbin.NewEnforcer(cfg.ModelPath, adapter)
		if err != nil {
			return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
		}
	} else if cfg.Adapter != nil {
		m, err := model.NewModelFromString(defaultRBACModel())
		if err != nil {
			return nil, fmt.Errorf("failed to create model: %w", err)
		}
		enforcer, err = casbin.NewEnforcer(m, cfg.Adapter)
		if err != nil {
			return nil, fmt.Errorf("failed to create enforcer: %w", err)
		}
	} else {
		return nil, fmt.Errorf("either Enforcer or ModelPath+PolicyPath must be provided")
	}

	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return &CasbinMiddleware{
		enforcer:  enforcer,
		config:    cfg,
		skipPaths: skipPaths,
	}, nil
}

func NewWithEnforcer(enforcer *casbin.Enforcer) *CasbinMiddleware {
	return &CasbinMiddleware{
		enforcer: enforcer,
		config:   DefaultConfig(),
	}
}

func (m *CasbinMiddleware) Handle(ctx abstract.Context, next func() error) error {
	if m.shouldSkip(ctx) {
		return next()
	}

	sub := m.config.SubGetter(ctx)
	if sub == "" {
		return abstract.Unauthorized("unauthorized: no subject")
	}

	obj := m.config.ObjGetter(ctx)
	act := m.config.ActGetter(ctx)

	ok, err := m.enforcer.Enforce(sub, obj, act)
	if err != nil {
		return abstract.InternalError(fmt.Sprintf("casbin enforce error: %v", err))
	}

	if !ok {
		return abstract.Forbidden("forbidden: access denied")
	}

	return next()
}

func (m *CasbinMiddleware) shouldSkip(ctx abstract.Context) bool {
	path := ctx.Path()
	return m.skipPaths[path]
}

func (m *CasbinMiddleware) AsMiddleware() abstract.Middleware {
	return abstract.MiddlewareFunc(m.Handle)
}

func (m *CasbinMiddleware) Enforce(sub, obj, act string) (bool, error) {
	return m.enforcer.Enforce(sub, obj, act)
}

func (m *CasbinMiddleware) EnforceWithContext(ctx abstract.Context) (bool, error) {
	sub := m.config.SubGetter(ctx)
	obj := m.config.ObjGetter(ctx)
	act := m.config.ActGetter(ctx)
	return m.Enforce(sub, obj, act)
}

func (m *CasbinMiddleware) AddPolicy(sub, obj, act string) error {
	_, err := m.enforcer.AddPolicy(sub, obj, act)
	return err
}

func (m *CasbinMiddleware) RemovePolicy(sub, obj, act string) error {
	_, err := m.enforcer.RemovePolicy(sub, obj, act)
	return err
}

func (m *CasbinMiddleware) AddRoleForUser(user, role string) error {
	_, err := m.enforcer.AddRoleForUser(user, role)
	return err
}

func (m *CasbinMiddleware) DeleteRoleForUser(user, role string) error {
	_, err := m.enforcer.DeleteRoleForUser(user, role)
	return err
}

func (m *CasbinMiddleware) GetRolesForUser(user string) ([]string, error) {
	return m.enforcer.GetRolesForUser(user)
}

func (m *CasbinMiddleware) GetUsersForRole(role string) ([]string, error) {
	return m.enforcer.GetUsersForRole(role)
}

func (m *CasbinMiddleware) HasRoleForUser(user, role string) (bool, error) {
	return m.enforcer.HasRoleForUser(user, role)
}

func (m *CasbinMiddleware) GetPermissionsForUser(user string) ([][]string, error) {
	return m.enforcer.GetPermissionsForUser(user)
}

func (m *CasbinMiddleware) AddPermissionForUser(user, obj, act string) error {
	_, err := m.enforcer.AddPermissionForUser(user, obj, act)
	return err
}

func (m *CasbinMiddleware) DeletePermissionForUser(user, obj, act string) error {
	_, err := m.enforcer.DeletePermissionForUser(user, obj, act)
	return err
}

func (m *CasbinMiddleware) SavePolicy() error {
	return m.enforcer.SavePolicy()
}

func (m *CasbinMiddleware) LoadPolicy() error {
	return m.enforcer.LoadPolicy()
}

func (m *CasbinMiddleware) GetEnforcer() *casbin.Enforcer {
	return m.enforcer
}

func defaultRBACModel() string {
	return `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && keyMatch(r.act, p.act)
`
}

func NewMemoryEnforcer() (*casbin.Enforcer, error) {
	m, err := model.NewModelFromString(defaultRBACModel())
	if err != nil {
		return nil, err
	}
	return casbin.NewEnforcer(m)
}

func CheckPermission(ctx abstract.Context, enforcer *casbin.Enforcer, sub, obj, act string) bool {
	ok, err := enforcer.Enforce(sub, obj, act)
	return err == nil && ok
}

func CheckRole(ctx abstract.Context, enforcer *casbin.Enforcer, user, role string) bool {
	ok, err := enforcer.HasRoleForUser(user, role)
	return err == nil && ok
}
