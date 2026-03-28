# Project Management System Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a complete enterprise-grade project management system using gonest framework + GORM + Vue3 + Element Plus, validating the framework's production capabilities.

**Architecture:** DDD modular backend with 7 business modules, shared infrastructure layer, WebSocket real-time notifications, and Vue3 Monorepo frontend with packages for core/utils/ui/features.

**Tech Stack:** 
- Backend: Go 1.22+, gonest framework, GORM, SQLite, JWT, Casbin RBAC, WebSocket
- Frontend: Vue 3, TypeScript, Pinia, Vue Router, Element Plus, Axios, pnpm workspace

**Design Doc:** `docs/superpowers/specs/2026-03-28-project-management-system-design.md`

---

## Project Structure

```
project-management/
├── backend/
│   ├── cmd/server/main.go
│   ├── modules/
│   │   ├── user/          # 用户认证
│   │   ├── team/          # 团队管理
│   │   ├── project/       # 项目管理
│   │   ├── task/          # 任务管理
│   │   ├── document/      # 文档管理
│   │   ├── activity/      # 活动日志
│   │   └── notification/  # 通知系统
│   ├── shared/
│   │   ├── database/      # GORM SQLite
│   │   ├── auth/          # JWT + Casbin
│   │   ├── websocket/     # WS Hub
│   │   ├── config/        # 配置管理
│   │   ├── logger/        # Zap日志
│   │   ├── middleware/    # 中间件
│   │   └── response/      # 响应格式
│   ├── config.yaml
│   ├── model.conf
│   ├── policy.csv
│   └── go.mod
│
└── frontend/
    ├── packages/
    │   ├── core/          # API客户端/类型/工具
    │   ├── ui/            # UI组件库
    │   ├── features/      # 功能模块
    │   │   ├── user/
    │   │   ├── team/
    │   │   ├── project/
    │   │   ├── task/
    │   │   ├── document/
    │   │   └── notification/
    │   └── app/           # 主应用
    ├── pnpm-workspace.yaml
    └── package.json
```

---

## Chunk 1: Backend Project Skeleton & Infrastructure

### Task 1.1: Initialize Backend Project

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/config.yaml`

- [ ] **Step 1: Create backend directory and go.mod**

```bash
mkdir -p backend/cmd/server
cd backend
go mod init github.com/linuxerlv/project-management
```

- [ ] **Step 2: Create basic main.go**

Create `backend/cmd/server/main.go`:

```go
package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Project Management System - Backend")
	log.Println("Server starting...")
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go run cmd/server/main.go`
Expected: "Project Management System - Backend"

- [ ] **Step 4: Create config.yaml**

Create `backend/config.yaml`:

```yaml
server:
  port: "8080"
  name: "pm-server"
  mode: "debug"

database:
  type: "sqlite"
  path: "data.db"

jwt:
  secret: "dev-secret-key-change-in-production"
  access_token_ttl: 3600
  refresh_token_ttl: 86400
  issuer: "pm-server"

casbin:
  model_path: "model.conf"
  policy_path: "policy.csv"

websocket:
  enabled: true
  path: "/ws"

log:
  level: "debug"
  format: "console"

cors:
  allow_origins:
    - "http://localhost:5173"
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allow_headers:
    - "Authorization"
    - "Content-Type"
  allow_credentials: true
  max_age: 86400
```

- [ ] **Step 5: Commit**

```bash
git add backend/
git commit -m "feat: initialize backend project skeleton"
```

---

### Task 1.2: Shared Database Layer

**Files:**
- Create: `backend/shared/database/db.go`
- Create: `backend/shared/database/migrate.go`

- [ ] **Step 1: Install GORM dependencies**

```bash
cd backend
go get -u gorm.io/gorm
go get -u gorm.io/driver/sqlite
```

- [ ] **Step 2: Create database package**

Create `backend/shared/database/db.go`:

```go
package database

import (
	"fmt"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	instance *gorm.DB
	once     sync.Once
)

// Config 数据库配置
type Config struct {
	Type string
	Path string
}

// NewDB 创建数据库连接（单例）
func NewDB(cfg *Config) (*gorm.DB, error) {
	var err error
	once.Do(func() {
		switch cfg.Type {
		case "sqlite":
			instance, err = gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Info),
			})
		default:
			err = fmt.Errorf("unsupported database type: %s", cfg.Type)
		}
	})
	return instance, err
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return instance
}

// Close 关闭数据库连接
func Close() error {
	if instance != nil {
		sqlDB, err := instance.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
```

- [ ] **Step 3: Create auto-migrate function**

Create `backend/shared/database/migrate.go`:

```go
package database

import (
	"gorm.io/gorm"
)

// AutoMigrate 自动迁移所有模型
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/shared/database/
git commit -m "feat: add database layer with GORM SQLite"
```

---

### Task 1.3: Shared Config Layer

**Files:**
- Create: `backend/shared/config/config.go`
- Create: `backend/shared/config/types.go`

- [ ] **Step 1: Install config dependencies**

```bash
cd backend
go get -u github.com/linuxerlv/gonest/config
```

- [ ] **Step 2: Create config types**

Create `backend/shared/config/types.go`:

```go
package config

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `koanf:"port"`
	Name string `koanf:"name"`
	Mode string `koanf:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type string `koanf:"type"`
	Path string `koanf:"path"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret          string `koanf:"secret"`
	AccessTokenTTL  int    `koanf:"access_token_ttl"`
	RefreshTokenTTL int    `koanf:"refresh_token_ttl"`
	Issuer          string `koanf:"issuer"`
}

// CasbinConfig Casbin配置
type CasbinConfig struct {
	ModelPath  string `koanf:"model_path"`
	PolicyPath string `koanf:"policy_path"`
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	Enabled bool   `koanf:"enabled"`
	Path    string `koanf:"path"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `koanf:"allow_origins"`
	AllowMethods     []string `koanf:"allow_methods"`
	AllowHeaders     []string `koanf:"allow_headers"`
	AllowCredentials bool     `koanf:"allow_credentials"`
	MaxAge           int      `koanf:"max_age"`
}

// AppConfig 应用配置
type AppConfig struct {
	Server    ServerConfig    `koanf:"server"`
	Database  DatabaseConfig  `koanf:"database"`
	JWT       JWTConfig       `koanf:"jwt"`
	Casbin    CasbinConfig    `koanf:"casbin"`
	WebSocket WebSocketConfig `koanf:"websocket"`
	Log       LogConfig       `koanf:"log"`
	CORS      CORSConfig      `koanf:"cors"`
}
```

- [ ] **Step 3: Create config loader**

Create `backend/shared/config/config.go`:

```go
package config

import (
	"github.com/linuxerlv/gonest/config"
)

// Load 加载配置文件
func Load(configPath string) (*AppConfig, error) {
	cfg := config.NewKoanfConfig(".")
	
	err := cfg.Load(
		config.NewFileProvider(configPath, config.NewYAMLParser()),
		config.NewYAMLParser(),
	)
	if err != nil {
		return nil, err
	}
	
	var appCfg AppConfig
	if err := cfg.Unmarshal("", &appCfg); err != nil {
		return nil, err
	}
	
	return &appCfg, nil
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/shared/config/
git commit -m "feat: add config layer with koanf"
```

---

### Task 1.4: Shared Logger Layer

**Files:**
- Create: `backend/shared/logger/logger.go`

- [ ] **Step 1: Install logger dependencies**

```bash
cd backend
go get -u github.com/linuxerlv/gonest/logger
```

- [ ] **Step 2: Create logger package**

Create `backend/shared/logger/logger.go`:

```go
package logger

import (
	"github.com/linuxerlv/gonest/logger"
)

// NewLogger 创建日志实例
func NewLogger(level, format string) (logger.Logger, error) {
	var cfg logger.Config
	
	switch format {
	case "json":
		cfg = logger.ProductionConfig()
	default:
		cfg = logger.DevelopmentConfig()
	}
	
	// 设置日志级别
	cfg.Level = level
	
	return logger.NewZapLogger(cfg)
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/shared/logger/
git commit -m "feat: add logger layer with zap"
```

---

### Task 1.5: Shared Response Utilities

**Files:**
- Create: `backend/shared/response/response.go`
- Create: `backend/shared/response/errors.go`

- [ ] **Step 1: Create response package**

Create `backend/shared/response/response.go`:

```go
package response

import (
	"time"

	"github.com/linuxerlv/gonest/core/abstract"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Errors    []FieldError `json:"errors,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// FieldError 字段错误
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Success 成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:      200,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// Error 错误响应
func Error(code int, message string, errors ...FieldError) *Response {
	return &Response{
		Code:      code,
		Message:   message,
		Errors:    errors,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// BadRequest 400错误
func BadRequest(message string, errors ...FieldError) error {
	return abstract.NewHttpException(400, message)
}

// Unauthorized 401错误
func Unauthorized(message string) error {
	return abstract.Unauthorized(message)
}

// Forbidden 403错误
func Forbidden(message string) error {
	return abstract.Forbidden(message)
}

// NotFound 404错误
func NotFound(message string) error {
	return abstract.NotFound(message)
}

// InternalError 500错误
func InternalError(message string) error {
	return abstract.InternalError(message)
}
```

- [ ] **Step 2: Create errors package**

Create `backend/shared/response/errors.go`:

```go
package response

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrProjectNotFound    = errors.New("project not found")
	ErrTaskNotFound       = errors.New("task not found")
	ErrTeamNotFound       = errors.New("team not found")
	ErrValidation         = errors.New("validation failed")
)
```

- [ ] **Step 3: Commit**

```bash
git add backend/shared/response/
git commit -m "feat: add response utilities"
```

---

### Task 1.6: Shared Middleware

**Files:**
- Create: `backend/shared/middleware/cors.go`
- Create: `backend/shared/middleware/recovery.go`

- [ ] **Step 1: Install middleware dependencies**

```bash
cd backend
go get -u github.com/linuxerlv/gonest/middleware/cors
go get -u github.com/linuxerlv/gonest/middleware/recovery
```

- [ ] **Step 2: Create CORS middleware wrapper**

Create `backend/shared/middleware/cors.go`:

```go
package middleware

import (
	"github.com/linuxerlv/gonest/middleware/cors"
	"github.com/linuxerlv/project-management/shared/config"
)

// NewCORS 创建CORS中间件
func NewCORS(cfg *config.CORSConfig) *cors.CorsMiddleware {
	return cors.New(&cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}
```

- [ ] **Step 3: Create recovery middleware wrapper**

Create `backend/shared/middleware/recovery.go`:

```go
package middleware

import (
	"github.com/linuxerlv/gonest/middleware/recovery"
)

// NewRecovery 创建Recovery中间件
func NewRecovery() *recovery.RecoveryMiddleware {
	return recovery.New(nil)
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/shared/middleware/
git commit -m "feat: add middleware wrappers"
```

---

### Task 1.7: Casbin RBAC Setup

**Files:**
- Create: `backend/model.conf`
- Create: `backend/policy.csv`
- Create: `backend/shared/auth/casbin.go`

- [ ] **Step 1: Create Casbin model**

Create `backend/model.conf`:

```conf
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
```

- [ ] **Step 2: Create initial policy**

Create `backend/policy.csv`:

```csv
p, admin, /api/v1/*, *
p, project_admin, /api/v1/projects/*, *
p, project_admin, /api/v1/tasks/*, *
p, project_member, /api/v1/projects/:id, GET
p, project_member, /api/v1/tasks, POST
p, project_member, /api/v1/tasks/:id, GET|PUT
p, project_member, /api/v1/tasks/:id/comments/*, *
```

- [ ] **Step 3: Install Casbin**

```bash
cd backend
go get -u github.com/casbin/casbin/v2
go get -u github.com/casbin/casbin/v2/persist/file-adapter
go get -u github.com/linuxerlv/gonest/middleware/casbin
```

- [ ] **Step 4: Create Casbin wrapper**

Create `backend/shared/auth/casbin.go`:

```go
package auth

import (
	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/linuxerlv/project-management/shared/config"
)

// NewCasbinEnforcer 创建Casbin执行器
func NewCasbinEnforcer(cfg *config.CasbinConfig) (*casbin.Enforcer, error) {
	adapter := fileadapter.NewAdapter(cfg.PolicyPath)
	enforcer, err := casbin.NewEnforcer(cfg.ModelPath, adapter)
	if err != nil {
		return nil, err
	}
	
	// 加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}
	
	return enforcer, nil
}
```

- [ ] **Step 5: Commit**

```bash
git add backend/model.conf backend/policy.csv backend/shared/auth/
git commit -m "feat: add Casbin RBAC setup"
```

---

### Task 1.8: Update main.go with Infrastructure

**Files:**
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Update main.go to use infrastructure**

Replace `backend/cmd/server/main.go`:

```go
package main

import (
	"fmt"
	"log"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/project-management/shared/auth"
	"github.com/linuxerlv/project-management/shared/config"
	"github.com/linuxerlv/project-management/shared/database"
	"github.com/linuxerlv/project-management/shared/logger"
	"github.com/linuxerlv/project-management/shared/middleware"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// 2. 初始化日志
	logInstance, err := logger.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	
	// 3. 初始化数据库
	db, err := database.NewDB(&database.Config{
		Type: cfg.Database.Type,
		Path: cfg.Database.Path,
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	
	// 4. 初始化Casbin
	enforcer, err := auth.NewCasbinEnforcer(&cfg.Casbin)
	if err != nil {
		log.Fatalf("Failed to create casbin enforcer: %v", err)
	}
	
	// 5. 创建应用
	builder := core.CreateBuilder()
	builder.Logger = logInstance
	
	// 注册服务到DI容器
	builder.Services.AddSingleton(db)
	builder.Services.AddSingleton(enforcer)
	
	app := builder.Build().(*core.WebApplication)
	
	// 6. 注册中间件
	app.Use(middleware.NewRecovery())
	app.Use(middleware.NewCORS(&cfg.CORS))
	
	// 7. 健康检查路由
	app.GET("/health", func(ctx abstract.ContextAbstract) error {
		return ctx.JSON(200, map[string]string{
			"status": "ok",
			"name":   cfg.Server.Name,
		})
	})
	
	// 8. 启动服务器
	logInstance.Info("Server starting...", logger.String("port", cfg.Server.Port))
	
	if err := app.Listen(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

- [ ] **Step 2: Add missing import**

Add import for `abstract`:

```go
import (
	// ... existing imports
	"github.com/linuxerlv/gonest/core/abstract"
)
```

- [ ] **Step 3: Verify it compiles**

Run: `cd backend && go mod tidy && go build ./cmd/server`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add backend/cmd/server/main.go backend/go.mod backend/go.sum
git commit -m "feat: wire up infrastructure in main.go"
```

---

## Chunk 2: User Module & Authentication

### Task 2.1: User Domain Entities

**Files:**
- Create: `backend/modules/user/domain/entity.go`
- Create: `backend/modules/user/domain/dto.go`
- Create: `backend/modules/user/domain/repository.go`

- [ ] **Step 1: Create user entity**

Create `backend/modules/user/domain/entity.go`:

```go
package domain

import (
	"time"

	"gorm.io/gorm"
)

// User 用户实体
type User struct {
	gorm.Model
	Username    string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email       string     `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password    string     `gorm:"size:255;not null" json:"-"`
	Avatar      string     `gorm:"size:255" json:"avatar"`
	Role        string     `gorm:"size:20;default:'member'" json:"role"`
	Status      string     `gorm:"size:20;default:'active'" json:"status"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsAdmin 是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// IsActive 是否活跃
func (u *User) IsActive() bool {
	return u.Status == "active"
}
```

- [ ] **Step 2: Create DTOs**

Create `backend/modules/user/domain/dto.go`:

```go
package domain

import "time"

// RegisterDTO 注册请求
type RegisterDTO struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginDTO 登录请求
type LoginDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateProfileDTO 更新资料请求
type UpdateProfileDTO struct {
	Username string `json:"username" validate:"omitempty,min=3,max=50"`
	Avatar   string `json:"avatar" validate:"omitempty,url"`
}

// ChangePasswordDTO 修改密码请求
type ChangePasswordDTO struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=6"`
}

// UserResponseDTO 用户响应
type UserResponseDTO struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Avatar      string    `json:"avatar"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ToResponse 转换为响应DTO
func (u *User) ToResponse() *UserResponseDTO {
	return &UserResponseDTO{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Avatar:      u.Avatar,
		Role:        u.Role,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}

// AuthResponseDTO 认证响应
type AuthResponseDTO struct {
	Token string          `json:"token"`
	User  *UserResponseDTO `json:"user"`
}
```

- [ ] **Step 3: Create repository interface**

Create `backend/modules/user/domain/repository.go`:

```go
package domain

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error
	FindByID(id uint) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByUsername(username string) (*User, error)
	FindAll(limit, offset int) ([]*User, error)
	Count() (int64, error)
	ExistsByEmail(email string) bool
	ExistsByUsername(username string) bool
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/modules/user/domain/
git commit -m "feat: add user domain entities and interfaces"
```

---

### Task 2.2: User Repository Implementation

**Files:**
- Create: `backend/modules/user/repository/user_repo.go`

- [ ] **Step 1: Create user repository**

Create `backend/modules/user/repository/user_repo.go`:

```go
package repository

import (
	"github.com/linuxerlv/project-management/modules/user/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(username string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (r *userRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Count(&count).Error
	return count, err
}

func (r *userRepository) ExistsByEmail(email string) bool {
	var count int64
	r.db.Model(&domain.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func (r *userRepository) ExistsByUsername(username string) bool {
	var count int64
	r.db.Model(&domain.User{}).Where("username = ?", username).Count(&count)
	return count > 0
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/modules/user/repository/
git commit -m "feat: add user repository implementation"
```

---

### Task 2.3: User Service

**Files:**
- Create: `backend/modules/user/service/user_service.go`

- [ ] **Step 1: Install bcrypt**

```bash
cd backend
go get -u golang.org/x/crypto/bcrypt
```

- [ ] **Step 2: Create user service**

Create `backend/modules/user/service/user_service.go`:

```go
package service

import (
	"errors"
	"time"

	"github.com/linuxerlv/gonest/middleware/auth"
	"github.com/linuxerlv/project-management/modules/user/domain"
	"github.com/linuxerlv/project-management/modules/user/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
)

type UserService struct {
	repo        repository.UserRepository
	jwtProvider *auth.JWTProvider
}

// NewUserService 创建用户服务
func NewUserService(repo repository.UserRepository, jwtProvider *auth.JWTProvider) *UserService {
	return &UserService{
		repo:        repo,
		jwtProvider: jwtProvider,
	}
}

// Register 用户注册
func (s *UserService) Register(dto *domain.RegisterDTO) (*domain.UserResponseDTO, error) {
	// 检查邮箱是否存在
	if s.repo.ExistsByEmail(dto.Email) {
		return nil, ErrEmailExists
	}
	
	// 检查用户名是否存在
	if s.repo.ExistsByUsername(dto.Username) {
		return nil, ErrUsernameExists
	}
	
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	
	// 创建用户
	user := &domain.User{
		Username: dto.Username,
		Email:    dto.Email,
		Password: string(hashedPassword),
		Role:     "member",
		Status:   "active",
	}
	
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	
	return user.ToResponse(), nil
}

// Login 用户登录
func (s *UserService) Login(dto *domain.LoginDTO) (*domain.AuthResponseDTO, error) {
	// 查找用户
	user, err := s.repo.FindByEmail(dto.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	
	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	
	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.New("user is inactive")
	}
	
	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	s.repo.Update(user)
	
	// 生成JWT Token
	token, err := s.jwtProvider.GenerateToken(&auth.Claims{
		UserID:   string(user.ID),
		Username: user.Username,
		Email:    user.Email,
		Roles:    []string{user.Role},
	})
	if err != nil {
		return nil, err
	}
	
	return &domain.AuthResponseDTO{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(id uint) (*domain.UserResponseDTO, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user.ToResponse(), nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(id uint, dto *domain.UpdateProfileDTO) (*domain.UserResponseDTO, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	
	if dto.Username != "" && dto.Username != user.Username {
		if s.repo.ExistsByUsername(dto.Username) {
			return nil, ErrUsernameExists
		}
		user.Username = dto.Username
	}
	
	if dto.Avatar != "" {
		user.Avatar = dto.Avatar
	}
	
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	
	return user.ToResponse(), nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(id uint, dto *domain.ChangePasswordDTO) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return ErrUserNotFound
	}
	
	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.OldPassword)); err != nil {
		return ErrInvalidCredentials
	}
	
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	user.Password = string(hashedPassword)
	return s.repo.Update(user)
}

// GetAll 获取所有用户（管理员）
func (s *UserService) GetAll(limit, offset int) ([]*domain.UserResponseDTO, int64, error) {
	users, err := s.repo.FindAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}
	
	count, err := s.repo.Count()
	if err != nil {
		return nil, 0, err
	}
	
	var responses []*domain.UserResponseDTO
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}
	
	return responses, count, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/modules/user/service/
git commit -m "feat: add user service with registration and login"
```

---

### Task 2.4: User Handler

**Files:**
- Create: `backend/modules/user/handler/auth_handler.go`
- Create: `backend/modules/user/handler/user_handler.go`
- Create: `backend/modules/user/handler/routes.go`

- [ ] **Step 1: Create auth handler**

Create `backend/modules/user/handler/auth_handler.go`:

```go
package handler

import (
	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/auth"
	"github.com/linuxerlv/project-management/modules/user/domain"
	"github.com/linuxerlv/project-management/modules/user/service"
	"github.com/linuxerlv/project-management/shared/response"
)

type AuthHandler struct {
	userService *service.UserService
	jwtProvider *auth.JWTProvider
}

func NewAuthHandler(userService *service.UserService, jwtProvider *auth.JWTProvider) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtProvider: jwtProvider,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(ctx abstract.ContextAbstract) error {
	var dto domain.RegisterDTO
	if err := ctx.Bind(&dto); err != nil {
		return response.BadRequest("invalid request data")
	}
	
	user, err := h.userService.Register(&dto)
	if err != nil {
		return response.BadRequest(err.Error())
	}
	
	return ctx.JSON(201, response.Success(user))
}

// Login 用户登录
func (h *AuthHandler) Login(ctx abstract.ContextAbstract) error {
	var dto domain.LoginDTO
	if err := ctx.Bind(&dto); err != nil {
		return response.BadRequest("invalid request data")
	}
	
	result, err := h.userService.Login(&dto)
	if err != nil {
		return response.Unauthorized(err.Error())
	}
	
	return ctx.JSON(200, response.Success(result))
}

// Profile 获取当前用户资料
func (h *AuthHandler) Profile(ctx abstract.ContextAbstract) error {
	userID := auth.GetUserID(ctx)
	if userID == "" {
		return response.Unauthorized("unauthorized")
	}
	
	profile, err := h.userService.GetProfile(parseUint(userID))
	if err != nil {
		return response.NotFound(err.Error())
	}
	
	return ctx.JSON(200, response.Success(profile))
}

// UpdateProfile 更新当前用户资料
func (h *AuthHandler) UpdateProfile(ctx abstract.ContextAbstract) error {
	userID := auth.GetUserID(ctx)
	if userID == "" {
		return response.Unauthorized("unauthorized")
	}
	
	var dto domain.UpdateProfileDTO
	if err := ctx.Bind(&dto); err != nil {
		return response.BadRequest("invalid request data")
	}
	
	profile, err := h.userService.UpdateProfile(parseUint(userID), &dto)
	if err != nil {
		return response.BadRequest(err.Error())
	}
	
	return ctx.JSON(200, response.Success(profile))
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(ctx abstract.ContextAbstract) error {
	userID := auth.GetUserID(ctx)
	if userID == "" {
		return response.Unauthorized("unauthorized")
	}
	
	var dto domain.ChangePasswordDTO
	if err := ctx.Bind(&dto); err != nil {
		return response.BadRequest("invalid request data")
	}
	
	if err := h.userService.ChangePassword(parseUint(userID), &dto); err != nil {
		return response.BadRequest(err.Error())
	}
	
	return ctx.JSON(200, response.Success(map[string]string{
		"message": "password changed successfully",
	}))
}

func parseUint(s string) uint {
	var id uint
	for _, c := range s {
		id = id*10 + uint(c-'0')
	}
	return id
}
```

- [ ] **Step 2: Create user handler**

Create `backend/modules/user/handler/user_handler.go`:

```go
package handler

import (
	"strconv"

	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/auth"
	"github.com/linuxerlv/project-management/modules/user/service"
	"github.com/linuxerlv/project-management/shared/response"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetAll 获取用户列表（管理员）
func (h *UserHandler) GetAll(ctx abstract.ContextAbstract) error {
	// 检查是否是管理员
	if !auth.HasRole(ctx, "admin") {
		return response.Forbidden("admin only")
	}
	
	limit, _ := strconv.Atoi(ctx.Query("limit"))
	if limit == 0 {
		limit = 20
	}
	
	offset, _ := strconv.Atoi(ctx.Query("offset"))
	
	users, total, err := h.userService.GetAll(limit, offset)
	if err != nil {
		return response.InternalError(err.Error())
	}
	
	return ctx.JSON(200, response.Success(map[string]interface{}{
		"users": users,
		"total": total,
	}))
}

// GetByID 获取用户详情
func (h *UserHandler) GetByID(ctx abstract.ContextAbstract) error {
	idStr := ctx.Param("id")
	id := parseUint(idStr)
	
	user, err := h.userService.GetProfile(id)
	if err != nil {
		return response.NotFound(err.Error())
	}
	
	return ctx.JSON(200, response.Success(user))
}
```

- [ ] **Step 3: Create routes**

Create `backend/modules/user/handler/routes.go`:

```go
package handler

import (
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/auth"
)

// RegisterRoutes 注册用户相关路由
func (h *AuthHandler) RegisterAuthRoutes(r abstract.RouterAbstract, authMiddleware abstract.MiddlewareAbstract) {
	// 公开路由
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	
	// 需要认证的路由
	authGroup := r.Group("")
	authGroup.Use(authMiddleware)
	
	authGroup.GET("/auth/profile", h.Profile)
	authGroup.PUT("/auth/profile", h.UpdateProfile)
	authGroup.PUT("/auth/password", h.ChangePassword)
}

func (h *UserHandler) RegisterUserRoutes(r abstract.RouterAbstract, authMiddleware abstract.MiddlewareAbstract) {
	userGroup := r.Group("/users")
	userGroup.Use(authMiddleware)
	
	userGroup.GET("", h.GetAll)
	userGroup.GET("/:id", h.GetByID)
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/modules/user/handler/
git commit -m "feat: add user and auth handlers"
```

---

### Task 2.5: User Module Registration

**Files:**
- Create: `backend/modules/user/module.go`
- Create: `backend/shared/auth/jwt.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Create JWT provider wrapper**

Create `backend/shared/auth/jwt.go`:

```go
package auth

import (
	"time"

	"github.com/linuxerlv/gonest/middleware/auth"
	"github.com/linuxerlv/project-management/shared/config"
)

// NewJWTProvider 创建JWT Provider
func NewJWTProvider(cfg *config.JWTConfig) *auth.JWTProvider {
	return auth.NewJWTProvider(&auth.JWTConfig{
		Secret:          cfg.Secret,
		AccessTokenTTL:  time.Duration(cfg.AccessTokenTTL) * time.Second,
		RefreshTokenTTL: time.Duration(cfg.RefreshTokenTTL) * time.Second,
		Issuer:          cfg.Issuer,
	}, nil)
}
```

- [ ] **Step 2: Create module registration**

Create `backend/modules/user/module.go`:

```go
package user

import (
	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/auth"
	"github.com/linuxerlv/project-management/modules/user/handler"
	"github.com/linuxerlv/project-management/modules/user/repository"
	"github.com/linuxerlv/project-management/modules/user/service"
	"gorm.io/gorm"
)

// Module 用户模块
type Module struct {
	DB          *gorm.DB
	JWTProvider *auth.JWTProvider
}

// Register 注册模块服务
func (m *Module) Register(services *core.ServiceCollection) {
	// 注册Repository
	services.AddScoped(func(s abstract.ServiceCollectionAbstract) repository.UserRepository {
		return repository.NewUserRepository(m.DB)
	})
	
	// 注册Service
	services.AddScoped(func(s abstract.ServiceCollectionAbstract) *service.UserService {
		repo := core.GetService[repository.UserRepository](s)
		return service.NewUserService(repo, m.JWTProvider)
	})
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(app *core.WebApplication, authMiddleware abstract.MiddlewareAbstract) {
	// 从DI获取服务
	userService := core.GetService[*service.UserService](app.Services)
	
	// 创建Handler
	authHandler := handler.NewAuthHandler(userService, m.JWTProvider)
	userHandler := handler.NewUserHandler(userService)
	
	// 注册路由
	api := app.Group("/api/v1")
	authHandler.RegisterAuthRoutes(api, authMiddleware)
	userHandler.RegisterUserRoutes(api, authMiddleware)
}
```

- [ ] **Step 3: Update main.go**

Modify `backend/cmd/server/main.go`:

```go
package main

import (
	"log"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/middleware/auth"
	"github.com/linuxerlv/project-management/modules/user"
	userDomain "github.com/linuxerlv/project-management/modules/user/domain"
	"github.com/linuxerlv/project-management/shared/auth"
	"github.com/linuxerlv/project-management/shared/config"
	"github.com/linuxerlv/project-management/shared/database"
	"github.com/linuxerlv/project-management/shared/logger"
	"github.com/linuxerlv/project-management/shared/middleware"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// 2. 初始化日志
	logInstance, err := logger.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	
	// 3. 初始化数据库
	db, err := database.NewDB(&database.Config{
		Type: cfg.Database.Type,
		Path: cfg.Database.Path,
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	
	// 4. 自动迁移
	if err := database.AutoMigrate(db, &userDomain.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	
	// 5. 初始化JWT
	jwtProvider := auth.NewJWTProvider(&cfg.JWT)
	
	// 6. 初始化Casbin
	enforcer, err := auth.NewCasbinEnforcer(&cfg.Casbin)
	if err != nil {
		log.Fatalf("Failed to create casbin enforcer: %v", err)
	}
	
	// 7. 创建应用
	builder := core.CreateBuilder()
	builder.Logger = logInstance
	
	// 注册共享服务
	builder.Services.AddSingleton(db)
	builder.Services.AddSingleton(enforcer)
	builder.Services.AddSingleton(jwtProvider)
	
	// 8. 注册用户模块
	userModule := &user.Module{
		DB:          db,
		JWTProvider: jwtProvider,
	}
	userModule.Register(builder.Services)
	
	app := builder.Build().(*core.WebApplication)
	
	// 9. 注册中间件
	app.Use(middleware.NewRecovery())
	app.Use(middleware.NewCORS(&cfg.CORS))
	
	// 10. 创建认证中间件
	authMiddleware := auth.New(jwtProvider, nil).AsMiddleware()
	
	// 11. 注册路由
	api := app.Group("/api/v1")
	
	// 健康检查
	app.GET("/health", func(ctx abstract.ContextAbstract) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})
	
	// 注册用户模块路由
	userModule.RegisterRoutes(app, authMiddleware)
	
	// 12. 启动服务器
	logInstance.Info("Server starting...", logger.String("port", cfg.Server.Port))
	
	if err := app.Listen(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

- [ ] **Step 4: Verify it compiles**

Run: `cd backend && go mod tidy && go build ./cmd/server`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add backend/modules/user/module.go backend/shared/auth/jwt.go backend/cmd/server/main.go
git commit -m "feat: integrate user module with main application"
```

---

### Task 2.6: Test User Module

**Files:**
- Create: `backend/modules/user/handler/auth_handler_test.go`

- [ ] **Step 1: Run the server**

Run: `cd backend && go run cmd/server/main.go`
Expected: Server starts on port 8080

- [ ] **Step 2: Test registration**

Run in another terminal:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'
```
Expected: `{"code":201,"message":"success","data":{...}}`

- [ ] **Step 3: Test login**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```
Expected: `{"code":200,"data":{"token":"...","user":{...}}}`

- [ ] **Step 4: Test profile with token**

```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer <token>"
```
Expected: User profile data

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "test: verify user module endpoints"
```

---

## Chunk 3: Team Module

### Task 3.1: Team Domain Entities

**Files:**
- Create: `backend/modules/team/domain/entity.go`
- Create: `backend/modules/team/domain/dto.go`
- Create: `backend/modules/team/domain/repository.go`

- [ ] **Step 1: Create team entity**

Create `backend/modules/team/domain/entity.go`:

```go
package domain

import (
	"time"

	"gorm.io/gorm"
)

// Team 团队实体
type Team struct {
	gorm.Model
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Logo        string    `gorm:"size:255" json:"logo"`
	OwnerID     uint      `gorm:"not null" json:"ownerId"`
	Members     []Member  `gorm:"foreignKey:TeamID" json:"members,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Member 团队成员
type Member struct {
	gorm.Model
	TeamID   uint      `gorm:"not null;uniqueIndex:idx_team_user" json:"teamId"`
	UserID   uint      `gorm:"not null;uniqueIndex:idx_team_user" json:"userId"`
	Role     string    `gorm:"size:20;default:'member'" json:"role"` // owner/admin/member
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joinedAt"`
}

// TableName 指定表名
func (Team) TableName() string {
	return "teams"
}

func (Member) TableName() string {
	return "team_members"
}
```

- [ ] **Step 2: Create DTOs**

Create `backend/modules/team/domain/dto.go`:

```go
package domain

import "time"

// CreateTeamDTO 创建团队请求
type CreateTeamDTO struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Logo        string `json:"logo" validate:"omitempty,url"`
}

// UpdateTeamDTO 更新团队请求
type UpdateTeamDTO struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=100"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
}

// AddMemberDTO 添加成员请求
type AddMemberDTO struct {
	UserID uint   `json:"userId" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=admin member"`
}

// UpdateMemberRoleDTO 更新成员角色
type UpdateMemberRoleDTO struct {
	Role string `json:"role" validate:"required,oneof=owner admin member"`
}

// TeamResponseDTO 团队响应
type TeamResponseDTO struct {
	ID           uint           `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Logo         string         `json:"logo"`
	OwnerID      uint           `json:"ownerId"`
	MemberCount  int            `json:"memberCount"`
	CreatedAt    time.Time      `json:"createdAt"`
}

// MemberResponseDTO 成员响应
type MemberResponseDTO struct {
	UserID   uint      `json:"userId"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
}

// TeamDetailResponseDTO 团队详情响应
type TeamDetailResponseDTO struct {
	*TeamResponseDTO
	Members []MemberResponseDTO `json:"members"`
}
```

- [ ] **Step 3: Create repository interface**

Create `backend/modules/team/domain/repository.go`:

```go
package domain

// TeamRepository 团队仓储接口
type TeamRepository interface {
	Create(team *Team) error
	Update(team *Team) error
	Delete(id uint) error
	FindByID(id uint) (*Team, error)
	FindByOwnerID(ownerID uint) ([]*Team, error)
	FindByUserID(userID uint) ([]*Team, error)
	AddMember(member *Member) error
	RemoveMember(teamID, userID uint) error
	UpdateMemberRole(teamID, userID uint, role string) error
	FindMembers(teamID uint) ([]Member, error)
	FindMemberByUserID(teamID, userID uint) (*Member, error)
	IsMember(teamID, userID uint) bool
	IsOwner(teamID, userID uint) bool
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/modules/team/domain/
git commit -m "feat: add team domain entities"
```

---

### Task 3.2: Team Repository Implementation

**Files:**
- Create: `backend/modules/team/repository/team_repo.go`

- [ ] **Step 1: Create team repository**

Create `backend/modules/team/repository/team_repo.go`:

```go
package repository

import (
	"github.com/linuxerlv/project-management/modules/team/domain"
	"gorm.io/gorm"
)

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) domain.TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(team *domain.Team) error {
	return r.db.Create(team).Error
}

func (r *teamRepository) Update(team *domain.Team) error {
	return r.db.Save(team).Error
}

func (r *teamRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Team{}, id).Error
}

func (r *teamRepository) FindByID(id uint) (*domain.Team, error) {
	var team domain.Team
	err := r.db.First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) FindByOwnerID(ownerID uint) ([]*domain.Team, error) {
	var teams []*domain.Team
	err := r.db.Where("owner_id = ?", ownerID).Find(&teams).Error
	return teams, err
}

func (r *teamRepository) FindByUserID(userID uint) ([]*domain.Team, error) {
	var teams []*domain.Team
	err := r.db.Table("teams").
		Select("teams.*").
		Joins("JOIN team_members ON team_members.team_id = teams.id").
		Where("team_members.user_id = ?", userID).
		Find(&teams).Error
	return teams, err
}

func (r *teamRepository) AddMember(member *domain.Member) error {
	return r.db.Create(member).Error
}

func (r *teamRepository) RemoveMember(teamID, userID uint) error {
	return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&domain.Member{}).Error
}

func (r *teamRepository) UpdateMemberRole(teamID, userID uint, role string) error {
	return r.db.Model(&domain.Member{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Update("role", role).Error
}

func (r *teamRepository) FindMembers(teamID uint) ([]domain.Member, error) {
	var members []domain.Member
	err := r.db.Where("team_id = ?", teamID).Find(&members).Error
	return members, err
}

func (r *teamRepository) FindMemberByUserID(teamID, userID uint) (*domain.Member, error) {
	var member domain.Member
	err := r.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *teamRepository) IsMember(teamID, userID uint) bool {
	var count int64
	r.db.Model(&domain.Member{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(&count)
	return count > 0
}

func (r *teamRepository) IsOwner(teamID, userID uint) bool {
	var team domain.Team
	err := r.db.Where("id = ? AND owner_id = ?", teamID, userID).First(&team).Error
	return err == nil
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/modules/team/repository/
git commit -m "feat: add team repository implementation"
```

---

### Task 3.3: Team Service & Handler

**Files:**
- Create: `backend/modules/team/service/team_service.go`
- Create: `backend/modules/team/handler/team_handler.go`
- Create: `backend/modules/team/handler/routes.go`
- Create: `backend/modules/team/module.go`

Due to length, I'll continue with abbreviated steps. Follow the same pattern as User module.

- [ ] **Step 1: Create team service**
- [ ] **Step 2: Create team handler**
- [ ] **Step 3: Create routes**
- [ ] **Step 4: Create module registration**
- [ ] **Step 5: Update main.go**
- [ ] **Step 6: Commit**

---

## Chunk 4: Project Module

Follow similar DDD structure for Project module with:
- Project entity (name, key, status, priority, team_id, owner_id)
- CRUD operations
- Project statistics
- Integration with Team

---

## Chunk 5: Task Module

The most complex module with:
- Task entity with status workflow
- Comments and attachments
- Subtasks support
- Kanban board API
- Real-time updates integration

---

## Chunk 6: Document, Activity, Notification Modules

Supporting modules:
- Document: Markdown documents per project
- Activity: Audit log of all actions
- Notification: User notifications

---

## Chunk 7: WebSocket Real-time Notifications

**Files:**
- Create: `backend/shared/websocket/hub.go`
- Create: `backend/shared/websocket/client.go`
- Create: `backend/shared/websocket/message.go`
- Create: `backend/shared/websocket/handler.go`

Implement WebSocket Hub for broadcasting task updates, comments, and notifications.

---

## Chunk 8: Frontend Monorepo Setup

### Task 8.1: Initialize Frontend

```bash
mkdir -p frontend
cd frontend
pnpm init
pnpm add -D typescript vite @types/node
```

### Task 8.2: Create Workspace

Create `pnpm-workspace.yaml`:
```yaml
packages:
  - 'packages/*'
```

### Task 8.3: Create Packages

- core: API client, types, utils
- ui: Element Plus components
- app: Main Vue application

---

## Chunk 9: Frontend Core Pages

- Login/Register
- Dashboard
- Project List/Detail
- Kanban Board
- Task Detail

---

## Execution Summary

**Estimated Effort:**
- Chunk 1-2: Backend foundation + User module (1-2 days)
- Chunk 3-6: Core business modules (2-3 days)
- Chunk 7: WebSocket integration (0.5 day)
- Chunk 8-9: Frontend (2-3 days)

**Total: ~6-8 days for MVP**

---

**Ready to execute?** Use `superpowers:subagent-driven-development` to delegate tasks to specialized agents.