# Gonest

> 简洁优雅的 Go Web 框架，灵感来自 ASP.NET Core 和 NestJS

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

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

### 基础使用

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/middleware/cors"
    "github.com/linuxerlv/gonest/middleware/recovery"
)

func main() {
    // 方式一：快速创建（适合简单应用）
    app := core.CreateApplication()
    app.Use(recovery.New(nil))
    app.Use(cors.New(nil))
    app.GET("/hello", func(ctx abstract.ContextAbstract) error {
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
)

func main() {
    // 创建 Builder
    builder := core.CreateBuilder()
    
    // 设置配置和环境变量（公开属性，直接访问）
    builder.Config = config.NewKoanfConfig(".")
    builder.Env.Set("APP_ENV", "production")
    
    // 注册服务到 DI 容器（Wire 注入友好）
    builder.Services.AddSingleton(&MyService{})
    builder.Services.AddScoped(func(s abstract.ServiceCollectionAbstract) *DbContext {
        return &DbContext{DSN: builder.Env.Get("DATABASE_URL")}
    })
    
    // 构建应用
    app := builder.Build().(*core.WebApplication)
    
    // 从 DI 容器获取服务
    service := core.GetService[*MyService](app.Services)
    
    // 注册路由
    app.MapGet("/hello", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"message": "Hello World"})
    })
    
    app.Run()
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

## API 速查

### 应用创建

```go
import "github.com/linuxerlv/gonest/core"

// 方式一：快速创建（适合简单应用）
app := core.CreateApplication()

// 方式二：Builder 模式（推荐，支持依赖注入）
builder := core.CreateBuilder()
builder.Config = cfg           // 设置配置
builder.Env.Set("KEY", "val")  // 设置环境变量
builder.Services.AddSingleton(&MyService{})  // 注册服务
app := builder.Build().(*core.WebApplication)

// 访问公开属性
cfg := app.Config
env := app.Env
services := app.Services
```

### 路由注册

```go
import "github.com/linuxerlv/gonest/core/abstract"

// Application 方式
app.GET("/users", func(ctx abstract.ContextAbstract) error {
    return ctx.JSON(200, users)
})

// WebApplication 方式（MapXXX 方法）
app.MapGet("/users", func(ctx abstract.ContextAbstract) error {
    return ctx.JSON(200, users)
})

app.MapPost("/users", func(ctx abstract.ContextAbstract) error {
    var user User
    ctx.Bind(&user)
    return ctx.JSON(201, user)
})

// 路由组
api := app.Group("/api/v1")
api.GET("/users", listUsers)
```

### 中间件

```go
import (
    "github.com/linuxerlv/gonest/middleware/cors"
    "github.com/linuxerlv/gonest/middleware/recovery"
    "github.com/linuxerlv/gonest/middleware/ratelimit"
)

app.Use(recovery.New(nil))
app.Use(cors.New(&cors.Config{
    AllowOrigins: []string{"https://example.com"},
}))
app.Use(ratelimit.New(&ratelimit.Config{
    Limit:  100,
    Window: time.Minute,
}))
```

### 依赖注入

```go
// 注册
builder.Services.AddSingleton(&MyService{})
builder.Services.AddScoped(func(s abstract.ServiceCollectionAbstract) *DbContext {
    return &DbContext{}
})

// 获取
service := core.GetService[*MyService](app.Services)
```

### 环境变量

```go
// 读取环境变量
dbUrl := builder.Env.Get("DATABASE_URL")
port := builder.Env.GetOrDefault("PORT", "8080")

// 检查存在
if builder.Env.Has("DEBUG") {
    // ...
}

// 获取所有
allEnv := builder.Env.All()
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

// 设置到 Builder
builder := core.CreateBuilder()
builder.Config = cfg
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