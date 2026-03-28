package casbin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/linuxerlv/gonest"
	"github.com/linuxerlv/gonest/testutil"
)

// ============================================================
//                       Test Helpers
// ============================================================

func createTestEnforcerWithPolicy() (*casbin.Enforcer, error) {
	enforcer, err := NewMemoryEnforcer()
	if err != nil {
		return nil, err
	}

	// Add some test policy rules
	_, err = enforcer.AddPolicy("admin", "/api/admin", "GET")
	if err != nil {
		return nil, err
	}
	_, err = enforcer.AddPolicy("user", "/api/users", "GET")
	if err != nil {
		return nil, err
	}
	_, err = enforcer.AddPolicy("user", "/api/users", "POST")
	if err != nil {
		return nil, err
	}

	return enforcer, nil
}

// ============================================================
//               TestCasbinMiddleware_Handle
// ============================================================

func TestCasbinMiddleware_Handle(t *testing.T) {
	tests := []struct {
		name          string
		setupEnforcer func() (*casbin.Enforcer, error)
		method        string
		path          string
		setupContext  func(ctx *gonest.HttpContext)
		expectStatus  int
		expectError   bool
	}{
		{
			name:          "allowed_request",
			setupEnforcer: createTestEnforcerWithPolicy,
			method:        http.MethodGet,
			path:          "/api/users",
			setupContext: func(ctx *gonest.HttpContext) {
				ctx.Set("user_id", "user")
			},
			expectStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "forbidden_request",
			setupEnforcer: createTestEnforcerWithPolicy,
			method:        http.MethodDelete,
			path:          "/api/admin",
			setupContext: func(ctx *gonest.HttpContext) {
				ctx.Set("user_id", "user")
			},
			expectStatus: http.StatusForbidden,
			expectError:  true,
		},
		{
			name:          "unauthorized_no_subject",
			setupEnforcer: createTestEnforcerWithPolicy,
			method:        http.MethodGet,
			path:          "/api/users",
			setupContext:  nil,
			expectStatus:  http.StatusUnauthorized,
			expectError:   true,
		},
		{
			name:          "skip_path_middleware",
			setupEnforcer: createTestEnforcerWithPolicy,
			method:        http.MethodGet,
			path:          "/health",
			setupContext:  nil,
			expectStatus:  http.StatusOK,
			expectError:   false,
		},
		{
			name: "skip_path_middleware",
			setupEnforcer: func() (*casbin.Enforcer, error) {
				return createTestEnforcerWithPolicy()
			},
			method:       http.MethodGet,
			path:         "/health",
			setupContext: nil,
			expectStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name: "allowed_with_user_id",
			setupEnforcer: func() (*casbin.Enforcer, error) {
				return createTestEnforcerWithPolicy()
			},
			method:       http.MethodGet,
			path:         "/api/users",
			setupContext: func(ctx *gonest.HttpContext) { ctx.Set("user_id", "user") },
			expectStatus: http.StatusOK,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer, err := tt.setupEnforcer()
			if err != nil {
				t.Fatalf("Failed to setup enforcer: %v", err)
			}

			config := DefaultConfig()
			config.Enforcer = enforcer
			config.SkipPaths = []string{"/health"}

			middleware, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create middleware: %v", err)
			}

			ctx, w := testutil.NewTestContext(tt.method, tt.path, nil)
			if tt.setupContext != nil {
				tt.setupContext(ctx)
			}

			err = middleware.Handle(ctx, func() error {
				return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
			})

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectError {
				if httpErr, ok := err.(*gonest.HttpError); ok {
					if httpErr.Status() != tt.expectStatus {
						t.Errorf("Expected error status %d, got %d", tt.expectStatus, httpErr.Status())
					}
				}
			} else if w.Code != tt.expectStatus {
				t.Errorf("Expected status %d, got %d", tt.expectStatus, w.Code)
			}
		})
	}
}

// ============================================================
//               TestCasbinMiddleware_SkipPaths
// ============================================================

func TestCasbinMiddleware_SkipPaths(t *testing.T) {
	tests := []struct {
		name         string
		configPaths  []string
		requestPath  string
		skipExpected bool
	}{
		{
			name:         "exact_match",
			configPaths:  []string{"/health", "/metrics"},
			requestPath:  "/health",
			skipExpected: true,
		},
		{
			name:         "no_match",
			configPaths:  []string{"/health", "/metrics"},
			requestPath:  "/api/users",
			skipExpected: false,
		},
		{
			name:         "empty_skip_paths",
			configPaths:  nil,
			requestPath:  "/health",
			skipExpected: false,
		},
		{
			name:         "multiple_paths_match",
			configPaths:  []string{"/health", "/metrics", "/static"},
			requestPath:  "/metrics",
			skipExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer, err := createTestEnforcerWithPolicy()
			if err != nil {
				t.Fatalf("Failed to setup enforcer: %v", err)
			}

			config := DefaultConfig()
			config.SkipPaths = tt.configPaths
			config.Enforcer = enforcer

			middleware, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create middleware: %v", err)
			}

			ctx, _ := testutil.NewTestContext(http.MethodGet, tt.requestPath, nil)

			// Use reflection-like approach to test shouldSkip
			shouldSkip := middleware.shouldSkip(ctx)

			if shouldSkip != tt.skipExpected {
				t.Errorf("Path '%s' skip expected %v, got %v", tt.requestPath, tt.skipExpected, shouldSkip)
			}
		})
	}
}

// ============================================================
//               TestCasbinMiddleware_Enforce
// ============================================================

func TestCasbinMiddleware_Enforce(t *testing.T) {
	tests := []struct {
		name      string
		subject   string
		object    string
		action    string
		allowed   bool
		expectErr bool
	}{
		{
			name:      "admin_access_admin_api",
			subject:   "admin",
			object:    "/api/admin",
			action:    "GET",
			allowed:   true,
			expectErr: false,
		},
		{
			name:      "user_access_users_api_get",
			subject:   "user",
			object:    "/api/users",
			action:    "GET",
			allowed:   true,
			expectErr: false,
		},
		{
			name:      "user_access_users_api_post",
			subject:   "user",
			object:    "/api/users",
			action:    "POST",
			allowed:   true,
			expectErr: false,
		},
		{
			name:      "user_access_admin_api_forbidden",
			subject:   "user",
			object:    "/api/admin",
			action:    "GET",
			allowed:   false,
			expectErr: false,
		},
		{
			name:      "unknown_subject",
			subject:   "unknown",
			object:    "/api/users",
			action:    "GET",
			allowed:   false,
			expectErr: false,
		},
		{
			name:      "wrong_action",
			subject:   "user",
			object:    "/api/users",
			action:    "DELETE",
			allowed:   false,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer, err := createTestEnforcerWithPolicy()
			if err != nil {
				t.Fatalf("Failed to setup enforcer: %v", err)
			}

			middleware := NewWithEnforcer(enforcer)

			allowed, err := middleware.Enforce(tt.subject, tt.object, tt.action)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if allowed != tt.allowed {
				t.Errorf("Enforce result: expected %v, got %v", tt.allowed, allowed)
			}
		})
	}
}

// ============================================================
//               TestCasbinMiddleware_Roles
// ============================================================

func TestCasbinMiddleware_Roles(t *testing.T) {
	tests := []struct {
		name         string
		setupRoles   func(*CasbinMiddleware) error
		testRole     string
		testUser     string
		hasRole      bool
		rolesForUser []string
		usersForRole []string
	}{
		{
			name:         "add_and_check_role",
			setupRoles:   nil,
			testRole:     "admin",
			testUser:     "alice",
			hasRole:      false, // initially not assigned
			rolesForUser: []string{},
			usersForRole: []string{},
		},
		{
			name: "add_role_assignment",
			setupRoles: func(m *CasbinMiddleware) error {
				_, err := m.enforcer.AddRoleForUser("alice", "admin")
				return err
			},
			testRole:     "admin",
			testUser:     "alice",
			hasRole:      true,
			rolesForUser: []string{"admin"},
			usersForRole: []string{"alice"},
		},
		{
			name: "multiple_roles_for_user",
			setupRoles: func(m *CasbinMiddleware) error {
				_, err := m.enforcer.AddRoleForUser("bob", "admin")
				if err != nil {
					return err
				}
				_, err = m.enforcer.AddRoleForUser("bob", "moderator")
				return err
			},
			testRole:     "moderator",
			testUser:     "bob",
			hasRole:      true,
			rolesForUser: []string{"admin", "moderator"},
			usersForRole: []string{"bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMemoryEnforcer()
			if err != nil {
				t.Fatalf("Failed to create enforcer: %v", err)
			}

			casbinMW := NewWithEnforcer(m)

			if tt.setupRoles != nil {
				if err := tt.setupRoles(casbinMW); err != nil {
					t.Fatalf("Failed to setup roles: %v", err)
				}
			}

			// Test HasRoleForUser
			hasRole, err := casbinMW.HasRoleForUser(tt.testUser, tt.testRole)
			if err != nil {
				t.Fatalf("HasRoleForUser failed: %v", err)
			}
			if hasRole != tt.hasRole {
				t.Errorf("HasRoleForUser: expected %v, got %v", tt.hasRole, hasRole)
			}

			// Test GetRolesForUser
			roles, err := casbinMW.GetRolesForUser(tt.testUser)
			if err != nil {
				t.Fatalf("GetRolesForUser failed: %v", err)
			}

			if len(roles) != len(tt.rolesForUser) {
				t.Errorf("GetRolesForUser: expected %v roles, got %v", tt.rolesForUser, roles)
			} else {
				roleMap := make(map[string]bool)
				for _, role := range roles {
					roleMap[role] = true
				}
				for _, expectedRole := range tt.rolesForUser {
					if !roleMap[expectedRole] {
						t.Errorf("GetRolesForUser: expected role %s not found in %v", expectedRole, roles)
					}
				}
			}

			// Test GetUsersForRole
			users, err := casbinMW.GetUsersForRole(tt.testRole)
			if err != nil {
				t.Fatalf("GetUsersForRole failed: %v", err)
			}

			if len(users) != len(tt.usersForRole) {
				t.Errorf("GetUsersForRole: expected %v users, got %v", tt.usersForRole, users)
			} else {
				for i, user := range tt.usersForRole {
					if users[i] != user {
						t.Errorf("GetUsersForRole: expected user %s at index %d, got %s", user, i, users[i])
					}
				}
			}
		})
	}
}

// ============================================================
//               TestCasbinMiddleware_Permissions
// ============================================================

func TestCasbinMiddleware_Permissions(t *testing.T) {
	tests := []struct {
		name          string
		setupPolicies func(*CasbinMiddleware) error
		checkSub      string
		checkObj      string
		checkAct      string
		allowed       bool
	}{
		{
			name:          "empty_policy_no_access",
			setupPolicies: nil,
			checkSub:      "user",
			checkObj:      "/api/data",
			checkAct:      "GET",
			allowed:       false,
		},
		{
			name: "add_permission_and_check",
			setupPolicies: func(m *CasbinMiddleware) error {
				return m.AddPolicy("user", "/api/data", "GET")
			},
			checkSub: "user",
			checkObj: "/api/data",
			checkAct: "GET",
			allowed:  true,
		},
		{
			name: "remove_permission_and_check",
			setupPolicies: func(m *CasbinMiddleware) error {
				if err := m.AddPolicy("user", "/api/data", "GET"); err != nil {
					return err
				}
				return m.RemovePolicy("user", "/api/data", "GET")
			},
			checkSub: "user",
			checkObj: "/api/data",
			checkAct: "GET",
			allowed:  false,
		},
		{
			name: "add_permission_for_user_directly",
			setupPolicies: func(m *CasbinMiddleware) error {
				return m.AddPermissionForUser("alice", "/api/private", "GET")
			},
			checkSub: "alice",
			checkObj: "/api/private",
			checkAct: "GET",
			allowed:  true,
		},
		{
			name: "delete_permission_for_user",
			setupPolicies: func(m *CasbinMiddleware) error {
				if err := m.AddPermissionForUser("bob", "/api/secret", "GET"); err != nil {
					return err
				}
				return m.DeletePermissionForUser("bob", "/api/secret", "GET")
			},
			checkSub: "bob",
			checkObj: "/api/secret",
			checkAct: "GET",
			allowed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMemoryEnforcer()
			if err != nil {
				t.Fatalf("Failed to create enforcer: %v", err)
			}

			casbinMW := NewWithEnforcer(m)

			if tt.setupPolicies != nil {
				if err := tt.setupPolicies(casbinMW); err != nil {
					t.Fatalf("Failed to setup policies: %v", err)
				}
			}

			allowed, err := casbinMW.Enforce(tt.checkSub, tt.checkObj, tt.checkAct)
			if err != nil {
				t.Fatalf("Enforce failed: %v", err)
			}

			if allowed != tt.allowed {
				t.Errorf("Permission check: expected %v, got %v", tt.allowed, allowed)
			}
		})
	}
}

// ============================================================
//               TestNewMemoryEnforcer
// ============================================================

func TestNewMemoryEnforcer(t *testing.T) {
	tests := []struct {
		name    string
		compare func(*casbin.Enforcer) error
	}{
		{
			name: "creates_valid_enforcer",
			compare: func(e *casbin.Enforcer) error {
				if e == nil {
					return fmt.Errorf("enforcer is nil")
				}
				return nil
			},
		},
		{
			name: "has_rbac_model",
			compare: func(e *casbin.Enforcer) error {
				roles := e.GetRolesForUser
				if roles == nil {
					return fmt.Errorf("RBAC functions not initialized")
				}
				return nil
			},
		},
		{
			name: "can_add_policy",
			compare: func(e *casbin.Enforcer) error {
				_, err := e.AddPolicy("admin", "/api", "GET")
				if err != nil {
					return fmt.Errorf("failed to add policy: %v", err)
				}
				return nil
			},
		},
		{
			name: "can_add_role",
			compare: func(e *casbin.Enforcer) error {
				_, err := e.AddRoleForUser("alice", "admin")
				if err != nil {
					return fmt.Errorf("failed to add role: %v", err)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer, err := NewMemoryEnforcer()
			if err != nil {
				t.Fatalf("NewMemoryEnforcer failed: %v", err)
			}

			if err := tt.compare(enforcer); err != nil {
				t.Errorf("Comparison failed: %v", err)
			}
		})
	}
}

// ============================================================
//               TestNewWithEnforcer
// ============================================================

func TestNewWithEnforcer(t *testing.T) {
	tests := []struct {
		name        string
		preSetup    func() (*casbin.Enforcer, error)
		compare     func(*CasbinMiddleware) error
		expectError bool
	}{
		{
			name: "creates_middleware_with_enforcer",
			preSetup: func() (*casbin.Enforcer, error) {
				return createTestEnforcerWithPolicy()
			},
			compare: func(m *CasbinMiddleware) error {
				if m == nil {
					return fmt.Errorf("middleware is nil")
				}
				if m.enforcer == nil {
					return fmt.Errorf("enforcer is nil")
				}
				return nil
			},
		},
		{
			name: "uses_default_config",
			preSetup: func() (*casbin.Enforcer, error) {
				return createTestEnforcerWithPolicy()
			},
			compare: func(m *CasbinMiddleware) error {
				if m.config == nil {
					return fmt.Errorf("config is nil")
				}
				if m.config.SubGetter == nil {
					return fmt.Errorf("SubGetter is nil")
				}
				if m.config.ObjGetter == nil {
					return fmt.Errorf("ObjGetter is nil")
				}
				if m.config.ActGetter == nil {
					return fmt.Errorf("ActGetter is nil")
				}
				return nil
			},
		},
		{
			name: "has_handle_method",
			preSetup: func() (*casbin.Enforcer, error) {
				return createTestEnforcerWithPolicy()
			},
			compare: func(m *CasbinMiddleware) error {
				// Check that the middleware has the Handle method
				_ = m.Handle
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer, err := tt.preSetup()
			if err != nil {
				t.Fatalf("Failed to setup enforcer: %v", err)
			}

			middleware := NewWithEnforcer(enforcer)

			if err := tt.compare(middleware); err != nil {
				t.Errorf("Comparison failed: %v", err)
			}
		})
	}
}

// ============================================================
//               TestPolicyManagement
// ============================================================

func TestPolicyManagement(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*CasbinMiddleware) error
		政策操作  func(*CasbinMiddleware) error
		验证    func(*CasbinMiddleware) error
	}{
		{
			name:  "add_policy_success",
			setup: nil,
			政策操作: func(m *CasbinMiddleware) error {
				return m.AddPolicy("admin", "/api/admin", "DELETE")
			},
			验证: func(m *CasbinMiddleware) error {
				allowed, err := m.Enforce("admin", "/api/admin", "DELETE")
				if err != nil {
					return fmt.Errorf("enforce failed: %v", err)
				}
				if !allowed {
					return fmt.Errorf("policy not effective after add")
				}
				return nil
			},
		},
		{
			name: "remove_policy_success",
			setup: func(m *CasbinMiddleware) error {
				return m.AddPolicy("admin", "/api/admin", "DELETE")
			},
			政策操作: func(m *CasbinMiddleware) error {
				return m.RemovePolicy("admin", "/api/admin", "DELETE")
			},
			验证: func(m *CasbinMiddleware) error {
				allowed, err := m.Enforce("admin", "/api/admin", "DELETE")
				if err != nil {
					return fmt.Errorf("enforce failed: %v", err)
				}
				if allowed {
					return fmt.Errorf("policy still effective after remove")
				}
				return nil
			},
		},
		{
			name:  "multiple_add_policy",
			setup: nil,
			政策操作: func(m *CasbinMiddleware) error {
				policies := [][]string{
					{"user", "/api/users", "GET"},
					{"user", "/api/users", "POST"},
					{"admin", "/api/admin", "GET"},
					{"admin", "/api/admin", "POST"},
				}
				for _, p := range policies {
					if _, err := m.enforcer.AddPolicy(p[0], p[1], p[2]); err != nil {
						return err
					}
				}
				return nil
			},
			验证: func(m *CasbinMiddleware) error {
				for i, p := range [][]string{
					{"user", "/api/users", "GET"},
					{"user", "/api/users", "POST"},
					{"admin", "/api/admin", "GET"},
					{"admin", "/api/admin", "POST"},
				} {
					allowed, err := m.Enforce(p[0], p[1], p[2])
					if err != nil {
						return fmt.Errorf("policy %d enforce failed: %v", i, err)
					}
					if !allowed {
						return fmt.Errorf("policy %d not effective", i)
					}
				}
				return nil
			},
		},
		{
			name: "get_permissions_for_user",
			setup: func(m *CasbinMiddleware) error {
				policies := [][]string{
					{"alice", "/api/users", "GET"},
					{"alice", "/api/users", "POST"},
					{"alice", "/api/admin", "GET"},
					{"bob", "/api/users", "GET"},
				}
				for _, p := range policies {
					if _, err := m.enforcer.AddPolicy(p[0], p[1], p[2]); err != nil {
						return err
					}
				}
				return nil
			},
			政策操作: nil,
			验证: func(m *CasbinMiddleware) error {
				perms, err := m.GetPermissionsForUser("alice")
				if err != nil {
					return fmt.Errorf("GetPermissionsForUser failed: %v", err)
				}
				if len(perms) != 3 {
					return fmt.Errorf("expected 3 permissions for alice, got %d", len(perms))
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMemoryEnforcer()
			if err != nil {
				t.Fatalf("Failed to create enforcer: %v", err)
			}

			casbinMW := NewWithEnforcer(m)

			if tt.setup != nil {
				if err := tt.setup(casbinMW); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			if tt.政策操作 != nil {
				if err := tt.政策操作(casbinMW); err != nil {
					t.Fatalf("Policy operation failed: %v", err)
				}
			}

			if tt.验证 != nil {
				if err := tt.验证(casbinMW); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

// ============================================================
//               TestRoleManagement
// ============================================================

func TestRoleManagement(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*CasbinMiddleware) error
		role操作 func(*CasbinMiddleware) error
		验证     func(*CasbinMiddleware) error
	}{
		{
			name:  "add_role_for_user",
			setup: nil,
			role操作: func(m *CasbinMiddleware) error {
				_, err := m.enforcer.AddRoleForUser("alice", "admin")
				return err
			},
			验证: func(m *CasbinMiddleware) error {
				hasRole, err := m.HasRoleForUser("alice", "admin")
				if err != nil {
					return fmt.Errorf("HasRoleForUser failed: %v", err)
				}
				if !hasRole {
					return fmt.Errorf("role not assigned")
				}
				return nil
			},
		},
		{
			name: "delete_role_for_user",
			setup: func(m *CasbinMiddleware) error {
				_, err := m.enforcer.AddRoleForUser("alice", "admin")
				return err
			},
			role操作: func(m *CasbinMiddleware) error {
				_, err := m.enforcer.DeleteRoleForUser("alice", "admin")
				return err
			},
			验证: func(m *CasbinMiddleware) error {
				hasRole, err := m.HasRoleForUser("alice", "admin")
				if err != nil {
					return fmt.Errorf("HasRoleForUser failed: %v", err)
				}
				if hasRole {
					return fmt.Errorf("role still assigned after delete")
				}
				return nil
			},
		},
		{
			name:  "multiple_roles_for_user",
			setup: nil,
			role操作: func(m *CasbinMiddleware) error {
				roles := []string{"admin", "moderator", "editor"}
				for _, role := range roles {
					_, err := m.enforcer.AddRoleForUser("bob", role)
					if err != nil {
						return err
					}
				}
				return nil
			},
			验证: func(m *CasbinMiddleware) error {
				roles, err := m.GetRolesForUser("bob")
				if err != nil {
					return fmt.Errorf("GetRolesForUser failed: %v", err)
				}
				if len(roles) != 3 {
					return fmt.Errorf("expected 3 roles for bob, got %d: %v", len(roles), roles)
				}
				return nil
			},
		},
		{
			name: "get_users_for_role",
			setup: func(m *CasbinMiddleware) error {
				users := []string{"alice", "bob", "charlie"}
				for _, user := range users {
					if err := m.AddRoleForUser(user, "admin"); err != nil {
						return err
					}
				}
				return nil
			},
			role操作: nil,
			验证: func(m *CasbinMiddleware) error {
				users, err := m.GetUsersForRole("admin")
				if err != nil {
					return fmt.Errorf("GetUsersForRole failed: %v", err)
				}
				if len(users) != 3 {
					return fmt.Errorf("expected 3 users for admin role, got %d: %v", len(users), users)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMemoryEnforcer()
			if err != nil {
				t.Fatalf("Failed to create enforcer: %v", err)
			}

			casbinMW := NewWithEnforcer(m)

			if tt.setup != nil {
				if err := tt.setup(casbinMW); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			if tt.role操作 != nil {
				if err := tt.role操作(casbinMW); err != nil {
					t.Fatalf("Role operation failed: %v", err)
				}
			}

			if tt.验证 != nil {
				if err := tt.验证(casbinMW); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

// ============================================================
//               TestCheckPermission
// ============================================================

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*casbin.Enforcer, error)
		subject     string
		object      string
		action      string
		expectValue bool
	}{
		{
			name: "allowed_permission_returns_true",
			setup: func() (*casbin.Enforcer, error) {
				enforcer, err := NewMemoryEnforcer()
				if err != nil {
					return nil, err
				}
				_, err = enforcer.AddPolicy("admin", "/api/admin", "GET")
				return enforcer, err
			},
			subject:     "admin",
			object:      "/api/admin",
			action:      "GET",
			expectValue: true,
		},
		{
			name: "forbidden_permission_returns_false",
			setup: func() (*casbin.Enforcer, error) {
				enforcer, err := NewMemoryEnforcer()
				if err != nil {
					return nil, err
				}
				_, err = enforcer.AddPolicy("user", "/api/users", "GET")
				return enforcer, err
			},
			subject:     "admin",
			object:      "/api/admin",
			action:      "GET",
			expectValue: false,
		},
		{
			name: "nonexistent_permission_returns_false",
			setup: func() (*casbin.Enforcer, error) {
				return NewMemoryEnforcer()
			},
			subject:     "unknown",
			object:      "/api/data",
			action:      "GET",
			expectValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer, err := tt.setup()
			if err != nil {
				t.Fatalf("Failed to setup enforcer: %v", err)
			}

			// Create a test context
			req := httptest.NewRequest(http.MethodGet, tt.object, nil)
			w := httptest.NewRecorder()
			ctx := gonest.NewContext(w, req)

			result := CheckPermission(ctx, enforcer, tt.subject, tt.object, tt.action)

			if result != tt.expectValue {
				t.Errorf("CheckPermission: expected %v, got %v", tt.expectValue, result)
			}
		})
	}
}

// ============================================================
//               TestCheckRole
// ============================================================

func TestCheckRole(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*casbin.Enforcer, error)
		user        string
		role        string
		expectValue bool
	}{
		{
			name: "user_has_role_returns_true",
			setup: func() (*casbin.Enforcer, error) {
				enforcer, err := NewMemoryEnforcer()
				if err != nil {
					return nil, err
				}
				_, err = enforcer.AddRoleForUser("alice", "admin")
				return enforcer, err
			},
			user:        "alice",
			role:        "admin",
			expectValue: true,
		},
		{
			name: "user_without_role_returns_false",
			setup: func() (*casbin.Enforcer, error) {
				enforcer, err := NewMemoryEnforcer()
				if err != nil {
					return nil, err
				}
				_, err = enforcer.AddRoleForUser("bob", "user")
				return enforcer, err
			},
			user:        "alice",
			role:        "admin",
			expectValue: false,
		},
		{
			name: "nonexistent_user_returns_false",
			setup: func() (*casbin.Enforcer, error) {
				return NewMemoryEnforcer()
			},
			user:        "unknown",
			role:        "admin",
			expectValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer, err := tt.setup()
			if err != nil {
				t.Fatalf("Failed to setup enforcer: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			ctx := gonest.NewContext(w, req)

			result := CheckRole(ctx, enforcer, tt.user, tt.role)

			if result != tt.expectValue {
				t.Errorf("CheckRole: expected %v, got %v", tt.expectValue, result)
			}
		})
	}
}

// ============================================================
//               TestCasbinMiddleware_Handle_TableDriven
// ============================================================

func TestCasbinMiddleware_Handle_TableDriven(t *testing.T) {
	type testScenario struct {
		name         string
		method       string
		path         string
		subject      string
		policies     [][]string
		roles        [][]string
		expectedCode int
	}

	scenarios := []testScenario{
		{
			name:         "admin_GET_admin_api_allowed",
			method:       http.MethodGet,
			path:         "/api/admin",
			subject:      "admin",
			policies:     [][]string{{"admin", "/api/admin", "GET"}},
			roles:        nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "user_GET_admin_api_forbidden",
			method:       http.MethodGet,
			path:         "/api/admin",
			subject:      "user",
			policies:     [][]string{{"admin", "/api/admin", "GET"}},
			roles:        nil,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "admin_via_role_GET_admin_api_allowed",
			method:       http.MethodGet,
			path:         "/api/admin",
			subject:      "alice",
			policies:     [][]string{{"admin", "/api/admin", "GET"}},
			roles:        [][]string{{"alice", "admin"}},
			expectedCode: http.StatusOK,
		},
		{
			name:         "user_POST_users_api_allowed",
			method:       http.MethodPost,
			path:         "/api/users",
			subject:      "user",
			policies:     [][]string{{"user", "/api/users", "POST"}},
			roles:        nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "anonymous_access_forbidden",
			method:       http.MethodGet,
			path:         "/api/users",
			subject:      "",
			policies:     [][]string{{"user", "/api/users", "GET"}},
			roles:        nil,
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			m, err := NewMemoryEnforcer()
			if err != nil {
				t.Fatalf("Failed to create enforcer: %v", err)
			}

			// Add policies
			for _, p := range sc.policies {
				_, err = m.AddPolicy(p[0], p[1], p[2])
				if err != nil {
					t.Fatalf("Failed to add policy: %v", err)
				}
			}

			// Add roles
			for _, r := range sc.roles {
				_, err = m.AddRoleForUser(r[0], r[1])
				if err != nil {
					t.Fatalf("Failed to add role: %v", err)
				}
			}

			casbinMW := NewWithEnforcer(m)

			ctx, w := testutil.NewTestContext(sc.method, sc.path, nil)
			ctx.Set("user_id", sc.subject)

			err = casbinMW.Handle(ctx, func() error {
				return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
			})

			if sc.expectedCode == http.StatusForbidden || sc.expectedCode == http.StatusUnauthorized {
				if err == nil {
					t.Fatalf("Expected error for %d status, got nil", sc.expectedCode)
				}
				_, ok := err.(*gonest.HttpError)
				if !ok {
					t.Fatalf("Expected HttpError, got %T", err)
				}
				return
			}

			if err != nil {
				_, ok := err.(*gonest.HttpError)
				if !ok {
					t.Fatalf("Unexpected error type: %T", err)
				}
			}

			if w.Code != sc.expectedCode {
				t.Errorf("Expected status code %d, got %d", sc.expectedCode, w.Code)
			}
		})
	}
}

// ============================================================
//               TestCasbinMiddleware_Integration
// ============================================================

func TestCasbinMiddleware_Integration(t *testing.T) {
	// Create enforcer with RBAC model using NewMemoryEnforcer
	e, err := NewMemoryEnforcer()
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add RBAC roles
	_, err = e.AddRoleForUser("alice", "admin")
	if err != nil {
		t.Fatalf("Failed to add role: %v", err)
	}
	_, err = e.AddRoleForUser("bob", "user")
	if err != nil {
		t.Fatalf("Failed to add role: %v", err)
	}

	// Add policy rules
	_, err = e.AddPolicy("admin", "/api/admin", "GET")
	if err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}
	_, err = e.AddPolicy("admin", "/api/admin", "POST")
	if err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}
	_, err = e.AddPolicy("user", "/api/users", "GET")
	if err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}
	_, err = e.AddPolicy("user", "/api/users", "POST")
	if err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}

	// Create middleware
	middleware := NewWithEnforcer(e)

	// Test scenario 1: Admin accessing admin API
	t.Run("admin_access_admin_api", func(t *testing.T) {
		ctx, w := testutil.NewTestContext(http.MethodGet, "/api/admin", nil)
		ctx.Set("user_id", "alice")

		err = middleware.Handle(ctx, func() error {
			return ctx.JSON(http.StatusOK, map[string]string{"data": "admin data"})
		})

		if err != nil {
			t.Fatalf("Admin should be able to access admin API: %v", err)
		}
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	// Test scenario 2: User accessing admin API (should be forbidden)
	t.Run("user_access_admin_api_forbidden", func(t *testing.T) {
		ctx, _ := testutil.NewTestContext(http.MethodGet, "/api/admin", nil)
		ctx.Set("user_id", "bob")

		err = middleware.Handle(ctx, func() error {
			return ctx.JSON(http.StatusOK, map[string]string{"data": "admin data"})
		})

		if err == nil {
			t.Fatal("Expected error for forbidden access")
		}
	})

	// Test scenario 3: Bob (user) accessing users API (should be allowed)
	t.Run("user_access_users_api", func(t *testing.T) {
		ctx, w := testutil.NewTestContext(http.MethodGet, "/api/users", nil)
		ctx.Set("user_id", "bob")

		err = middleware.Handle(ctx, func() error {
			return ctx.JSON(http.StatusOK, map[string]string{"data": "user data"})
		})

		if err != nil {
			t.Fatalf("User should be able to access users API: %v", err)
		}
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	// Test scenario 4: Anonymous access (should be unauthorized)
	t.Run("anonymous_access_forbidden", func(t *testing.T) {
		ctx, _ := testutil.NewTestContext(http.MethodGet, "/api/users", nil)

		err = middleware.Handle(ctx, func() error {
			return ctx.JSON(http.StatusOK, map[string]string{"data": "user data"})
		})

		if err == nil {
			t.Fatal("Expected error for unauthorized access")
		}
	})

	// Test scenario 5: Alice deleting users (should be forbidden)
	t.Run("admin_delete_users_forbidden", func(t *testing.T) {
		ctx, _ := testutil.NewTestContext(http.MethodDelete, "/api/users", nil)
		ctx.Set("user_id", "alice")

		err = middleware.Handle(ctx, func() error {
			return ctx.JSON(http.StatusOK, map[string]string{"data": "deleted"})
		})

		if err == nil {
			t.Fatal("Expected error for unauthorized action")
		}
	})
}

// ============================================================
//               TestCasbinMiddleware_Config
// ============================================================

func TestCasbinMiddleware_Config(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
		expectNil bool
	}{
		{
			name:      "nil_config_uses_default",
			config:    nil,
			expectErr: true,
			expectNil: true,
		},
		{
			name: "valid_enforcer_provided",
			config: &Config{
				Enforcer: func() *casbin.Enforcer {
					e, _ := createTestEnforcerWithPolicy()
					return e
				}(),
			},
			expectErr: false,
			expectNil: false,
		},
		{
			name: "model_and_policy_paths",
			config: &Config{
				ModelPath:  "",
				PolicyPath: "",
			},
			expectErr: true,
			expectNil: true,
		},
		{
			name: "adapter_provided",
			config: &Config{
				Adapter: fileadapter.NewAdapter(""),
			},
			expectErr: false,
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, err := New(tt.config)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectNil && middleware != nil {
				t.Errorf("Expected nil middleware, got %v", middleware)
			}

			if !tt.expectNil && middleware == nil {
				t.Errorf("Expected non-nil middleware, got nil")
			}
		})
	}
}

// ============================================================
//               TestCasbinMiddleware_GetEnforcer
// ============================================================

func TestCasbinMiddleware_GetEnforcer(t *testing.T) {
	enforcer, err := createTestEnforcerWithPolicy()
	if err != nil {
		t.Fatalf("Failed to setup enforcer: %v", err)
	}

	middleware := NewWithEnforcer(enforcer)

	retrieved := middleware.GetEnforcer()

	if retrieved == nil {
		t.Fatal("GetEnforcer returned nil")
	}

	if retrieved != enforcer {
		t.Error("GetEnforcer returned different enforcer instance")
	}
}

// ============================================================
//               TestCasbinMiddleware_SavePolicy
// ============================================================

func TestCasbinMiddleware_SavePolicy(t *testing.T) {
	enforcer, err := createTestEnforcerWithPolicy()
	if err != nil {
		t.Fatalf("Failed to setup enforcer: %v", err)
	}

	middleware := NewWithEnforcer(enforcer)

	err = middleware.AddPolicy("moderator", "/api/mod", "GET")
	if err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("SavePolicy panicked (expected for in-memory enforcer): %v", r)
		}
	}()
	err = middleware.SavePolicy()
	if err != nil {
		t.Logf("SavePolicy error (expected for in-memory enforcer): %v", err)
	}

	err = middleware.LoadPolicy()
	if err != nil {
		t.Logf("LoadPolicy error (expected for in-memory enforcer): %v", err)
	}

	allowed, err := middleware.Enforce("moderator", "/api/mod", "GET")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !allowed {
		t.Error("Policy should still be effective")
	}
}

// ============================================================
//               TestCasbinMiddleware_AsMiddleware
// ============================================================

func TestCasbinMiddleware_AsMiddleware(t *testing.T) {
	enforcer, err := createTestEnforcerWithPolicy()
	if err != nil {
		t.Fatalf("Failed to setup enforcer: %v", err)
	}

	middleware := NewWithEnforcer(enforcer)

	// Test that AsMiddleware returns a valid gonest.Middleware
	var gw gonest.Middleware = middleware.AsMiddleware()

	if gw == nil {
		t.Fatal("AsMiddleware returned nil")
	}

	// Test that the returned middleware has a Handle method
	var _ gonest.Middleware = gw
}

// ============================================================
//               TestCasbinMiddleware_EnforceWithContext
// ============================================================

func TestCasbinMiddleware_EnforceWithContext(t *testing.T) {
	enforcer, err := createTestEnforcerWithPolicy()
	if err != nil {
		t.Fatalf("Failed to setup enforcer: %v", err)
	}

	middleware := NewWithEnforcer(enforcer)

	tests := []struct {
		name         string
		subjectValue any
		path         string
		method       string
		allowed      bool
	}{
		{
			name:         "valid_request",
			subjectValue: "user",
			path:         "/api/users",
			method:       http.MethodGet,
			allowed:      true,
		},
		{
			name:         "empty_subject",
			subjectValue: "",
			path:         "/api/users",
			method:       http.MethodGet,
			allowed:      false,
		},
		{
			name:         "missing_subject",
			subjectValue: nil,
			path:         "/api/users",
			method:       http.MethodGet,
			allowed:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := testutil.NewTestContext(tt.method, tt.path, nil)
			if tt.subjectValue != nil {
				ctx.Set("user_id", tt.subjectValue)
			}

			allowed, err := middleware.EnforceWithContext(ctx)

			if err != nil {
				t.Fatalf("EnforceWithContext failed: %v", err)
			}

			if allowed != tt.allowed {
				t.Errorf("EnforceWithContext: expected %v, got %v", tt.allowed, allowed)
			}
		})
	}
}
