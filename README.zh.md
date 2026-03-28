# Gonest

> 简洁优雅的 Go Web 框架，灵感来自 ASP.NET Core 和 NestJS

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/linuxerlv/gonest/workflows/CI/badge.svg)](https://github.com/linuxerlv/gonest/actions)

[English](README.md) | [日本語](README.ja.md)

---

## 核心特性

| 特性 | 说明 |
|-----|------|
| 🏗️ **接口分离设计** | `core/abstract/` 定义接口，`core/` 实现，扩展包面向接口开发 |
| 🔌 **模块化扩展** | 中间件、协议、任务调度均为独立包，按需引入 |
| ⚡ **细粒度接口** | ContextAbstract = RequestReader + ResponseWriter + ValueStore，按需依赖 |
| 📦 **开箱即用** | 配置（Koanf）、日志（Zap）、认证、授权全套解决方案 |

---

## 5 分钟快速开始

### 安装

```bash
go get github.com/linuxerlv/gonest
```

### 快速开始（ASP.NET Core 风格）

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/extensions"
)

func main() {
    // 创建 WebApplication Builder（ASP.NET Core 风格）
    builder := core.CreateBuilder()

    // 添加服务到 DI 容器
    builder.Services().AddSingleton(&MyService{})

    // 构建应用
    app := builder.Build()

    // 使用中间件（扩展方法 - 链式调用）
    app = extensions.Extend(app).
        UseRecovery(nil).
        UseCORS(&extensions.CORSMiddlewareOptions{
            AllowOrigins: []string{"https://example.com"},
        }).
        UseLogging(nil)

    // 注册路由
    app.MapGet("/hello", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"message": "Hello World"})
    })

    app.Run()
}
```

### Builder 模式（推荐，支持依赖注入）

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/config"
    "github.com/linuxerlv/gonest/extensions"
)

func main() {
    // 创建 WebApplication Builder（ASP.NET Core 风格）
    builder := core.CreateBuilder()

    // 配置配置和环境变量
    cfg := config.NewKoanfConfig(".")
    builder.UseConfig(cfg)
    builder.Environment().Set("APP_ENV", "production")

    // 注册服务到 DI 容器
    builder.Services().AddSingleton(&MyService{})
    builder.Services().AddScoped(func(s abstract.ServiceCollectionAbstract) *DbContext {
        return &DbContext{DSN: builder.Environment().Get("DATABASE_URL")}
    })

    // 构建应用
    app := builder.Build()

    // 从 DI 容器获取服务
    service := core.GetService[*MyService](app.Services())

    // 使用中间件（扩展方法 - 链式调用）
    app = extensions.Extend(app).
        UseRecovery(nil).
        UseCORS(&extensions.CORSMiddlewareOptions{
            AllowOrigins: []string{"https://example.com"},
        })

    // 注册路由
    app.MapGet("/hello", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"message": "Hello World"})
    })

    app.Run()
}
```

### 通用应用（非 Web 场景）

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    // 创建通用 Application Builder（用于非 Web 场景）
    builder := core.CreateApplicationBuilder()

    // 添加服务
    builder.Services().AddSingleton(&MyService{})
    builder.Services().AddTransient(func(s abstract.ServiceCollectionAbstract) *Worker {
        return &Worker{}
    })

    // 构建应用
    app := builder.Build()

    // 获取服务并运行业务逻辑
    service := core.GetService[*MyService](app.Services())
    service.DoWork()
}
```

---

## 项目结构

```
gonest/
├── core/
│   ├── abstract/              # 接口定义（细粒度、可组合）
│   │   ├── context.go         # ContextAbstract, RequestReaderAbstract, ResponseWriterAbstract...
│   │   ├── router.go          # RouterAbstract, RouteBuilderAbstract...
│   │   ├── middleware.go      # MiddlewareAbstract...
│   │   ├── di.go              # ServiceCollectionAbstract...
│   │   ├── env.go             # EnvAbstract 环境变量接口
│   │   └── ...                # 其他接口
│   │
│   ├── context.go             # HttpContext 实现
│   ├── router.go              # HttpRouter 实现
│   ├── application.go         # Application 实现
│   ├── builder.go             # WebApplicationBuilder 实现
│   ├── env.go                 # Env 环境变量实现
│   └── ...                    # 其他实现
│
├── config/                    # 配置模块
│   ├── config.go              # Config 接口（实现 abstract.ConfigAbstract）
│   └── koanf.go               # Koanf 实现
│
├── logger/                    # 日志模块
│   ├── logger.go              # Logger 接口（实现 abstract.LoggerAbstract）
│   └── zap.go                 # Zap 实现
│
├── middleware/                # 中间件扩展包
│   ├── auth/                  # JWT 认证
│   ├── session/               # Session 管理
│   ├── casbin/                # Casbin 权限控制
│   ├── cors/                  # CORS
│   ├── recovery/              # Panic 恢复
│   ├── ratelimit/             # 限流
│   ├── timeout/               # 超时控制
│   ├── gzip/                  # Gzip 压缩
│   ├── security/              # 安全头
│   ├── logger/                # 日志中间件
│   └── requestid/             # 请求 ID
│
├── protocol/                  # 协议扩展包
│   ├── websocket/             # WebSocket
│   ├── sse/                   # Server-Sent Events
│   ├── http3/                 # HTTP/3
│   └── grpc/                  # gRPC
│
├── task/                      # 任务调度
│   ├── interface.go           # TaskQueue, CronScheduler 接口
│   ├── asynq.go               # Asynq 实现（Redis）
│   ├── cron.go                # Cron 实现
│   └── memory.go              # 内存实现
│
├── ipc/                       # IPC 接口（进程间通信）
│   └── interface.go           # Endpoint, Publisher, Subscriber...
│
└── gonest.go                  # 向后兼容类型别名
```

---

## 接口设计

### 细粒度接口（core/abstract/）

框架采用细粒度接口设计，开发者可以按需依赖：

```go
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
}

// 完整上下文接口（组合）
type ContextAbstract interface {
    ContextRunnerAbstract
    FullRequestReaderAbstract
    FullResponseWriterAbstract
    ValueStoreAbstract
}
```

### 使用方式

```go
// 1. 使用完整核心包
import "github.com/linuxerlv/gonest/core"
app := core.CreateApplication()

// 2. 只用接口定义（编写扩展）
import "github.com/linuxerlv/gonest/core/abstract"
func MyMiddleware(ctx abstract.ContextAbstract) error { ... }

// 3. 使用扩展中间件
import "github.com/linuxerlv/gonest/middleware/auth"
app.Use(auth.New(provider, nil))

// 4. 向后兼容（gonest 包类型别名）
import "github.com/linuxerlv/gonest"
app := gonest.NewApplication()
```

---

## API 速查（ASP.NET Core 风格）

### 应用创建

```go
import "github.com/linuxerlv/gonest/core"

// Web 应用（带 HTTP 服务器）
builder := core.CreateBuilder()           // 创建 WebApplication Builder
app := builder.Build()                 // 构建 WebApplication
app.Run()                                 // 启动服务器

// 通用应用（非 Web 场景）
builder := core.CreateApplicationBuilder() // 创建 Application Builder
app := builder.Build()                     // 构建 Application

// 访问服务
services := builder.Services()             // 获取 ServiceCollection
cfg := builder.Configuration()             // 获取 Configuration
env := builder.Environment()               // 获取 Environment
```

### 路由注册

```go
import "github.com/linuxerlv/gonest/core/abstract"

// MapXXX 方法（ASP.NET Core 风格）
app.MapGet("/users", func(ctx abstract.ContextAbstract) error {
    return ctx.JSON(200, users)
})

app.MapPost("/users", func(ctx abstract.ContextAbstract) error {
    var user User
    ctx.Bind(&user)
    return ctx.JSON(201, user)
})

app.MapPut("/users/:id", func(ctx abstract.ContextAbstract) error {
    id := ctx.Param("id")
    return ctx.JSON(200, map[string]string{"id": id})
})

app.MapDelete("/users/:id", func(ctx abstract.ContextAbstract) error {
    return ctx.JSON(200, nil)
})

// 路由组
api := app.Group("/api/v1")
api.MapGet("/users", listUsers)
```

### 中间件（扩展方法）

```go
import "github.com/linuxerlv/gonest/extensions"

// 方式一：链式调用（推荐）
app = extensions.Extend(app).
    UseRecovery(nil).
    UseCORS(&extensions.CORSMiddlewareOptions{
        AllowOrigins: []string{"https://example.com"},
    }).
    UseRateLimit(&extensions.RateLimitMiddlewareOptions{
        Limit:  100,
        Window: 60, // 秒
    }).
    UseGzip(nil).
    UseSecurity(nil).
    UseRequestID(nil).
    UseTimeout(&extensions.TimeoutMiddlewareOptions{
        Timeout: 30, // 秒
    })

// 方式二：使用原始中间件
app.Use(middleware)
```

### 依赖注入

```go
// 注册服务
builder.Services().AddSingleton(&MyService{})
builder.Services().AddScoped(func(s abstract.ServiceCollectionAbstract) *DbContext {
    return &DbContext{}
})
builder.Services().AddTransient(func(s abstract.ServiceCollectionAbstract) *CacheService {
    return &CacheService{}
})

// 从应用获取服务
service := core.GetService[*MyService](app.Services())
```

### 环境变量

```go
// 读取环境变量
dbUrl := builder.Environment().Get("DATABASE_URL")
port := builder.Environment().GetOrDefault("PORT", "8080")

// 检查存在
if builder.Environment().Has("DEBUG") {
    // ...
}

// 获取所有
allEnv := builder.Environment().All()
```

### 配置文件

```go
import "github.com/linuxerlv/gonest/config"

// 加载 JSON 配置文件
cfg := config.NewKoanfConfig(".")
cfg.Load(
    config.NewFileProvider("config.json", config.NewJSONParser()),
    config.NewJSONParser(),
)

// 加载 YAML 配置文件
cfg.Load(
    config.NewFileProvider("config.yaml", config.NewYAMLParser()),
    config.NewYAMLParser(),
)

// 读取配置
port := cfg.GetString("server.port")
debug := cfg.GetBool("debug")

// 绑定到结构体
type ServerConfig struct {
    Port    string `koanf:"port"`
    Timeout int    `koanf:"timeout"`
}
var serverCfg ServerConfig
cfg.Unmarshal("server", &serverCfg)

// 设置到 Builder（使用 UseConfig 方法）
builder := core.CreateBuilder()
builder.UseConfig(cfg)
```

---

## 文档导航

| 文档 | 适合人群 | 内容 |
|-----|---------|------|
| **[教程](TUTORIAL.md)** | 🎓 Go 初学者 | 渐进式学习指南 |
| **[API 参考](API_REFERENCE.md)** | 👨‍💻 应用开发者 | 完整 API 文档 |
| **[贡献者指南](DEVELOPER.md)** | 🛠️ 框架贡献者 | 架构设计、编码规范、测试策略、扩展机制 |

---

## License

MIT License - 详见 [LICENSE](LICENSE) 文件
