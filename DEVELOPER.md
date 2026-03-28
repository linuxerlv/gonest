# Gonest 框架贡献者指南

> 本文档面向框架贡献者，帮助你理解项目架构、编码规范、测试策略和扩展机制。

---

## 目录

1. [项目概述](#项目概述)
2. [架构设计](#架构设计)
3. [核心概念](#核心概念)
4. [开发环境设置](#开发环境设置)
5. [代码规范](#代码规范)
6. [测试指南](#测试指南)
7. [扩展框架](#扩展框架)
8. [提交代码](#提交代码)
9. [发布流程](#发布流程)

---

## 项目概述

### 什么是 Gonest？

Gonest 是一个受 ASP.NET Core 和 NestJS 启发的 Go Web 框架。它提供：

- **接口分离设计**：细粒度接口定义，便于测试和扩展
- **模块化架构**：中间件、协议、任务调度等均为独立包
- **依赖注入**：内置 DI 容器，支持 Singleton/Scoped/Transient 生命周期
- **多协议支持**：HTTP、WebSocket、SSE、gRPC、HTTP/3

### 项目定位

| 目标用户 | 使用场景 |
|---------|---------|
| 应用开发者 | 使用框架构建 Web 应用 |
| 框架贡献者 | 扩展框架核心功能、添加中间件 |

### 目录结构

```
gonest/
├── core/                    # 核心实现
│   ├── abstract/           # 接口定义（细粒度、可组合）
│   ├── application.go      # Application 实现
│   ├── builder.go          # WebApplicationBuilder 实现
│   ├── context.go          # HttpContext 实现
│   ├── router.go           # HttpRouter 实现
│   └── host.go             # HostBuilder 实现
│
├── config/                  # 配置模块（Koanf 实现）
├── logger/                  # 日志模块（Zap 实现）
│
├── middleware/              # 中间件扩展包
│   ├── auth/               # JWT/BasicAuth/APIKey 认证
│   ├── cors/               # CORS 跨域
│   ├── recovery/           # Panic 恢复
│   ├── ratelimit/          # 限流
│   ├── session/            # Session 管理
│   ├── casbin/             # Casbin 权限控制
│   └── ...                 # 其他中间件
│
├── protocol/                # 协议扩展包
│   ├── websocket/          # WebSocket
│   ├── sse/                # Server-Sent Events
│   ├── http3/              # HTTP/3 (QUIC)
│   └── grpc/               # gRPC
│
├── task/                    # 任务调度
│   ├── interface.go        # 接口定义
│   ├── memory.go           # 内存实现
│   ├── cron.go             # Cron 实现
│   └── asynq.go            # Asynq (Redis) 实现
│
├── ipc/                     # IPC 接口定义
├── cmd/gonest/              # CLI 工具
├── tests/                   # 测试文件
├── benchmark/               # 性能测试
└── gonest.go                # 向后兼容类型别名
```

---

## 架构设计

### 设计原则

1. **接口优先**：所有核心组件都有接口定义（`core/abstract/`）
2. **依赖倒置**：高层模块依赖抽象，不依赖具体实现
3. **单一职责**：每个接口只做一件事
4. **组合优于继承**：通过接口组合构建复杂类型

### 分层架构

```
┌─────────────────────────────────────────────────────────┐
│                    Application Layer                     │
│  (WebApplication, Application, WebApplicationBuilder)   │
├─────────────────────────────────────────────────────────┤
│                    Processing Layer                      │
│     (Middleware, Guard, Interceptor, Pipe, Filter)      │
├─────────────────────────────────────────────────────────┤
│                    Routing Layer                         │
│           (HttpRouter, Route, RouteGroup)               │
├─────────────────────────────────────────────────────────┤
│                    Context Layer                         │
│           (HttpContext, Request, Response)              │
├─────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                  │
│           (Config, Logger, DI Container)                │
└─────────────────────────────────────────────────────────┘
```

### 请求处理流程

```
HTTP Request
    │
    ▼
┌──────────────┐
│   Router     │ ← 路由匹配
└──────────────┘
    │
    ▼
┌──────────────┐
│ Middlewares  │ ← 洋葱模型：请求进入时正序执行
└──────────────┘
    │
    ▼
┌──────────────┐
│   Guards     │ ← 权限检查，返回 false 则拒绝请求
└──────────────┘
    │
    ▼
┌──────────────┐
│    Pipes     │ ← 数据转换/验证
└──────────────┘
    │
    ▼
┌──────────────┐
│ Interceptors │ ← 前置/后置处理
└──────────────┘
    │
    ▼
┌──────────────┐
│   Handler    │ ← 业务处理
└──────────────┘
    │
    ▼
┌──────────────┐
│   Filters    │ ← 异常捕获（发生错误时）
└──────────────┘
    │
    ▼
HTTP Response
```

### 模块依赖关系

```
gonest (入口)
    │
    ├── core (核心实现)
    │       │
    │       └── core/abstract (接口定义，无外部依赖)
    │
    ├── config (配置)
    │       └── 依赖 core/abstract
    │
    ├── logger (日志)
    │       └── 依赖 core/abstract
    │
    ├── middleware/* (中间件)
    │       └── 依赖 core, core/abstract
    │
    ├── protocol/* (协议)
    │       └── 依赖 core, core/abstract
    │
    └── task (任务调度)
            └── 依赖 core/abstract
```

**关键规则**：
- `core/abstract` 不依赖任何内部包
- 扩展包（middleware、protocol、task）只依赖 `core/abstract` 或 `core`
- `gonest.go` 提供向后兼容的类型别名

---

## 核心概念

### 1. Context（上下文）

Context 是请求处理的核心，包含请求信息和响应能力。

**接口设计**（细粒度组合）：

```go
// core/abstract/context.go

// 请求读取接口
type RequestReaderAbstract interface {
    Method() string
    Path() string
    Header(name string) string
}

// 响应写入接口
type ResponseWriterAbstract interface {
    Status(code int)
    JSON(code int, v any) error
    String(code int, s string) error
    Data(code int, contentType string, data []byte) error
}

// 值存储接口
type ValueStoreAbstract interface {
    Set(key string, value any)
    Get(key string) any
}

// 完整上下文接口（组合）
type ContextAbstract interface {
    ContextRunnerAbstract
    FullRequestReaderAbstract
    FullResponseWriterAbstract
    ValueStoreAbstract
}
```

**为什么这样设计？**

- 中间件可以只依赖需要的接口（如只读 `RequestReaderAbstract`）
- 测试时可以只 mock 需要的部分
- 便于扩展新的 Context 类型

**实现**：

```go
// core/context.go
type HttpContext struct {
    req    *http.Request
    res    http.ResponseWriter
    params map[string]string
    query  map[string][]string
    body   []byte
    values map[string]any
    mu     sync.RWMutex
}

// 接口验证
var _ abstract.ContextAbstract = (*HttpContext)(nil)
```

### 2. Router（路由器）

**接口**：

```go
// core/abstract/router.go

type RouterAbstract interface {
    RouteGetterAbstract    // GET, POST, PUT, DELETE, PATCH, OPTIONS
    GroupCreatorAbstract   // Group(prefix)
    RouteMatcherAbstract   // Match(req) (Route, params)
}

type RouteBuilderAbstract interface {
    Guard(guard GuardAbstract) RouteBuilderAbstract
    Interceptor(interceptor InterceptorAbstract) RouteBuilderAbstract
    Pipe(pipe PipeAbstract) RouteBuilderAbstract
}
```

**路由匹配**：

```go
// core/router.go
func (r *HttpRouter) Match(req *http.Request) (abstract.RouteAbstract, map[string]string) {
    // 1. 分割路径
    segments := splitPath(req.URL.Path)
    
    // 2. 遍历路由树
    node := r.root
    params := make(map[string]string)
    
    for _, segment := range segments {
        // 精确匹配优先
        if node.children[segment] != nil {
            node = node.children[segment]
        } else if node.children[":param"] != nil {
            // 参数匹配
            node = node.children[":param"]
        } else {
            return nil, nil
        }
    }
    
    // 3. 检查方法
    if route, ok := node.routes[req.Method]; ok {
        extractParams(segments, route.path, params)
        return route, params
    }
    
    return nil, nil
}
```

### 3. Middleware（中间件）

**接口**：

```go
// core/abstract/middleware.go

type MiddlewareAbstract interface {
    Handle(ctx ContextAbstract, next func() error) error
}

// 函数类型（便捷实现）
type MiddlewareFuncAbstract func(ctx ContextAbstract, next func() error) error

func (f MiddlewareFuncAbstract) Handle(ctx ContextAbstract, next func() error) error {
    return f(ctx, next)
}
```

**执行流程（洋葱模型）**：

```go
// 中间件链构建（core/router.go）
for i := len(app.middlewares) - 1; i >= 0; i-- {
    mw := app.middlewares[i]
    next := handler
    handler = func(c abstract.ContextAbstract) error {
        return mw.Handle(c, func() error { return next(c) })
    }
}
```

### 4. Guard（守卫）

**接口**：

```go
// core/abstract/guard.go

type GuardAbstract interface {
    CanActivate(ctx ContextAbstract) bool
}

type GuardFuncAbstract func(ctx ContextAbstract) bool
```

**用途**：权限控制，决定请求是否可以继续

**示例**：

```go
// 路由级 Guard
app.GET("/admin", handler).Guard(GuardFuncAbstract(func(ctx ContextAbstract) bool {
    return ctx.Header("X-Admin") == "true"
}))

// 全局 Guard
app.UseGlobalGuards(AuthGuard{})
```

### 5. Interceptor（拦截器）

**接口**：

```go
// core/abstract/interceptor.go

type InterceptorAbstract interface {
    Intercept(ctx ContextAbstract, next RouteHandlerAbstract) (any, error)
}
```

**用途**：在处理前后执行逻辑，可修改返回值

**与 Middleware 的区别**：

| 特性 | Middleware | Interceptor |
|-----|-----------|-------------|
| 作用范围 | 全局/路由级 | 全局/路由级 |
| 返回值 | error | (any, error) |
| 修改响应 | 间接 | 直接 |
| 主要用途 | 通用处理 | 业务相关 |

### 6. Pipe（管道）

**接口**：

```go
// core/abstract/pipe.go

type PipeAbstract interface {
    Transform(value any, ctx ContextAbstract) (any, error)
}
```

**用途**：数据转换和验证

### 7. ExceptionFilter（异常过滤器）

**接口**：

```go
// core/abstract/filter.go

type ExceptionFilterAbstract interface {
    Catch(ctx ContextAbstract, err error) error
}
```

**用途**：统一错误处理

### 8. 依赖注入

**接口**：

```go
// core/abstract/di.go

type ServiceCollectionAbstract interface {
    ServiceResolverAbstract   // GetService, GetRequiredService
    ServiceRegistrarAbstract  // AddSingleton, AddScoped, AddTransient
}
```

**生命周期**：

| 生命周期 | 说明 | 使用场景 |
|---------|------|---------|
| Singleton | 应用单例 | 数据库连接、配置 |
| Scoped | 每请求一个 | 请求上下文、事务 |
| Transient | 每次新建 | 无状态服务 |

**泛型方法**：

```go
// core/builder.go
func AddSingleton[T any](s *ServiceCollection, instance T) *ServiceCollection
func AddScoped[T any](s *ServiceCollection, factory func(abstract.ServiceCollectionAbstract) T) *ServiceCollection
func GetService[T any](s abstract.ServiceCollectionAbstract) T
```

---

## 开发环境设置

### 前置要求

- Go 1.24+
- Git
- Make（可选）

### 克隆项目

```bash
git clone https://github.com/linuxerlv/gonest.git
cd gonest
```

### 安装依赖

```bash
go mod download
```

### 运行测试

```bash
go test ./...
```

### 运行基准测试

```bash
go test -bench=. ./benchmark/...
```

---

## 代码规范

### 文件组织

```
middleware/auth/
├── auth.go          # 主要实现和公共 API
├── jwt.go           # JWT 相关实现
├── refresh.go       # Token 刷新逻辑
└── auth_test.go     # 测试文件
```

### 命名约定

| 类型 | 规范 | 示例 |
|-----|------|------|
| 包名 | 小写单词 | `cors`, `ratelimit` |
| 接口 | 名词 + Abstract | `GuardAbstract` |
| 实现 | 具体名称 | `AuthGuard`, `JWTProvider` |
| 函数类型 | 名称 + FuncAbstract | `GuardFuncAbstract` |
| 配置 | Config | `CORSConfig` |
| 错误 | 动词 + Error | `BadRequest`, `NotFound` |

### 接口设计规范

1. **细粒度接口**：每个接口只做一件事

```go
// 好的设计：细粒度，可组合
type RequestReaderAbstract interface {
    Method() string
    Path() string
}

type ResponseWriterAbstract interface {
    JSON(code int, v any) error
}

// 避免：大而全的接口
type Context interface {
    Method() string
    Path() string
    JSON(code int, v any) error
    // ... 20+ 方法
}
```

2. **接口验证**：在实现文件末尾添加验证

```go
var _ abstract.MiddlewareAbstract = (*CORSMiddleware)(nil)
```

3. **函数类型便捷方法**：提供函数类型以便快速实现

```go
type GuardFuncAbstract func(ctx ContextAbstract) bool

func (f GuardFuncAbstract) CanActivate(ctx ContextAbstract) bool {
    return f(ctx)
}
```

### 错误处理

使用 `core/abstract/errors.go` 中的标准错误：

```go
import "github.com/linuxerlv/gonest/core/abstract"

// 使用标准 HTTP 错误
if user == nil {
    return abstract.NotFound("用户不存在")
}

// 自定义错误
return abstract.NewHttpException(418, "我是一个茶壶")
```

### 日志规范

```go
import "github.com/linuxerlv/gonest/logger"

// 使用结构化日志
log.Info("用户登录",
    logger.String("user_id", "123"),
    logger.String("ip", "192.168.1.1"),
)

// 创建子 Logger
userLog := log.WithName("user-service")
```

### 配置规范

```go
// 中间件配置结构
type Config struct {
    // 必需参数不带默认值
    Secret string
    
    // 可选参数带默认值
    Timeout time.Duration
    Enabled bool
}

// 提供默认配置函数
func DefaultConfig() *Config {
    return &Config{
        Timeout: 30 * time.Second,
        Enabled: true,
    }
}

// New 函数处理 nil 配置
func New(cfg *Config) abstract.MiddlewareAbstract {
    if cfg == nil {
        cfg = DefaultConfig()
    }
    // ...
}
```

---

## 测试指南

### 测试组织

```
tests/
├── gonest_test.go          # 核心功能测试
├── builder_test.go         # Builder 模式测试
├── task_test.go            # 任务调度测试
├── gonest_bench_test.go    # 基准测试
└── integration/
    └── middleware_test.go  # 中间件集成测试

middleware/auth/
└── auth_test.go            # 单元测试
```

### 单元测试模式

```go
func TestMiddleware_Handle(t *testing.T) {
    // Arrange
    mw := New(nil)
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    w := httptest.NewRecorder()
    ctx := NewContext(w, req)
    
    // Act
    err := mw.Handle(ctx, func() error {
        return nil
    })
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### 表格驱动测试

```go
func TestRouter_Match(t *testing.T) {
    tests := []struct {
        name       string
        method     string
        path       string
        wantRoute  bool
        wantParams map[string]string
    }{
        {"exact match", "GET", "/users", true, nil},
        {"param match", "GET", "/users/123", true, map[string]string{"id": "123"}},
        {"not found", "GET", "/nonexistent", false, nil},
    }
    
    router := setupRouter()
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(tt.method, tt.path, nil)
            route, params := router.Match(req)
            
            if (route != nil) != tt.wantRoute {
                t.Errorf("want route %v, got %v", tt.wantRoute, route != nil)
            }
            
            if tt.wantParams != nil {
                for k, v := range tt.wantParams {
                    if params[k] != v {
                        t.Errorf("param %s: want %s, got %s", k, v, params[k])
                    }
                }
            }
        })
    }
}
```

### 集成测试

```go
func TestIntegration_FullStack(t *testing.T) {
    app := NewApplication()
    
    // 添加中间件
    app.Use(cors.New(nil))
    app.Use(recovery.New(nil))
    
    // 添加路由
    app.GET("/users", func(ctx Context) error {
        return ctx.JSON(200, []string{"user1", "user2"})
    })
    
    // 发送请求
    req := httptest.NewRequest(http.MethodGet, "/users", nil)
    w := httptest.NewRecorder()
    app.Router().ServeHTTP(w, req, app)
    
    // 验证响应
    if w.Code != 200 {
        t.Errorf("expected 200, got %d", w.Code)
    }
}
```

### 基准测试

```go
func BenchmarkMiddleware(b *testing.B) {
    mw := New(nil)
    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    ctx := NewContext(w, req)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        mw.Handle(ctx, func() error { return nil })
    }
}
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./middleware/auth/...

# 运行基准测试
go test -bench=. ./benchmark/...

# 查看覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 扩展框架

### 创建新中间件

**步骤 1：创建包结构**

```bash
mkdir middleware/mymiddleware
```

**步骤 2：实现接口**

```go
// middleware/mymiddleware/mymiddleware.go
package mymiddleware

import (
    "github.com/linuxerlv/gonest/core/abstract"
)

// Config 中间件配置
type Config struct {
    Enabled bool
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
    return &Config{
        Enabled: true,
    }
}

// New 创建中间件
func New(cfg *Config) abstract.MiddlewareAbstract {
    if cfg == nil {
        cfg = DefaultConfig()
    }
    
    return abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
        // 前置处理
        if !cfg.Enabled {
            return next()
        }
        
        // 调用下一个处理器
        err := next()
        
        // 后置处理
        
        return err
    })
}
```

**步骤 3：编写测试**

```go
// middleware/mymiddleware/mymiddleware_test.go
package mymiddleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func TestNew(t *testing.T) {
    mw := New(nil)
    if mw == nil {
        t.Fatal("expected middleware to be created")
    }
}

func TestMiddleware_Enabled(t *testing.T) {
    mw := New(&Config{Enabled: true})
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    w := httptest.NewRecorder()
    ctx := core.NewContext(w, req)
    
    called := false
    err := mw.Handle(ctx, func() error {
        called = true
        return nil
    })
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !called {
        t.Error("expected next to be called")
    }
}
```

### 创建新协议适配器

```go
// protocol/myprotocol/adapter.go
package myprotocol

import (
    "context"
    "github.com/linuxerlv/gonest/core/abstract"
)

type Adapter struct {
    addr    string
    running bool
}

func NewAdapter() *Adapter {
    return &Adapter{}
}

// 实现 abstract.ProtocolAdapterAbstract
func (a *Adapter) Name() string {
    return "myprotocol"
}

func (a *Adapter) Scheme() string {
    return "myproto"
}

func (a *Adapter) Start(addr string) error {
    a.addr = addr
    a.running = true
    return nil
}

func (a *Adapter) Stop(ctx context.Context) error {
    a.running = false
    return nil
}

func (a *Adapter) Running() bool {
    return a.running
}

// 接口验证
var _ abstract.ProtocolAdapterAbstract = (*Adapter)(nil)
```

### 创建新 Guard

```go
type MyGuard struct {
    requiredRole string
}

func (g *MyGuard) CanActivate(ctx abstract.ContextAbstract) bool {
    role, ok := ctx.Get("role").(string)
    if !ok {
        return false
    }
    return role == g.requiredRole
}

func NewRoleGuard(role string) abstract.GuardAbstract {
    return &MyGuard{requiredRole: role}
}
```

### 创建新 Interceptor

```go
type LoggingInterceptor struct {
    logger logger.Logger
}

func (i *LoggingInterceptor) Intercept(ctx abstract.ContextAbstract, next abstract.RouteHandlerAbstract) (any, error) {
    // 前置处理
    start := time.Now()
    i.logger.Info("请求开始", logger.String("path", ctx.Path()))
    
    // 执行处理器
    err := next(ctx)
    
    // 后置处理
    i.logger.Info("请求结束",
        logger.Duration("latency", time.Since(start)),
    )
    
    return nil, err
}
```

### 创建新 ExceptionFilter

```go
type JSONErrorFilter struct{}

func (f *JSONErrorFilter) Catch(ctx abstract.ContextAbstract, err error) error {
    if httpErr, ok := err.(*abstract.HttpException); ok {
        return ctx.JSON(httpErr.Status(), map[string]any{
            "error":  httpErr.Message(),
            "status": httpErr.Status(),
        })
    }
    
    return ctx.JSON(500, map[string]any{
        "error":  err.Error(),
        "status": 500,
    })
}
```

---

## 提交代码

### 提交前检查清单

- [ ] 代码通过 `go fmt` 格式化
- [ ] 代码通过 `go vet` 检查
- [ ] 所有测试通过
- [ ] 新功能有对应测试
- [ ] 文档已更新（如需要）

### Commit 消息格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型**：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具相关

**示例**：

```
feat(middleware): add rate limiting middleware

- Implement token bucket algorithm
- Support configurable limit and window
- Add tests and benchmarks

Closes #123
```

---

## 发布流程

### 版本号规范

遵循 [Semantic Versioning](https://semver.org/)：

- `MAJOR.MINOR.PATCH`
- 例如：`1.2.3`

### 发布步骤

1. **更新版本号**
   - `go.mod` 中的版本注释
   - `README.md` 中的版本徽章

2. **更新 CHANGELOG**
   ```markdown
   ## [1.2.0] - 2024-01-15
   
   ### Added
   - 新增 ratelimit 中间件
   
   ### Changed
   - 优化路由匹配性能
   
   ### Fixed
   - 修复 CORS 预检请求处理
   ```

3. **创建 Tag**
   ```bash
   git tag -a v1.2.0 -m "Release v1.2.0"
   git push origin v1.2.0
   ```

4. **发布到 GitHub**
   - 创建 Release Note
   - 附上二进制文件（如需要）

---

## 常见问题

### Q: 如何调试中间件执行顺序？

```go
app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
    log.Println("1: before")
    err := next()
    log.Println("1: after")
    return err
}))

app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
    log.Println("2: before")
    err := next()
    log.Println("2: after")
    return err
}))

// 输出：
// 1: before
// 2: before
// 2: after
// 1: after
```

### Q: 如何处理 Context 类型断言？

```go
// 安全的类型断言
if hc, ok := ctx.(*core.HttpContext); ok {
    w := hc.ResponseWriter()
    // ...
}
```

### Q: 如何在中间件中终止请求？

```go
app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
    if !isAuthorized(ctx) {
        // 不调用 next()，直接返回响应
        return ctx.String(401, "未授权")
    }
    return next()
}))
```

---

## 参考资料

- [API 参考](API_REFERENCE.md)
- [教程](TUTORIAL.md)
- [ASP.NET Core 文档](https://docs.microsoft.com/aspnet/core)
- [NestJS 文档](https://docs.nestjs.com)
- [Go 接口设计最佳实践](https://go.dev/doc/effective_go#interfaces)