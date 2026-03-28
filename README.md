# Gonest

> Simple and elegant Go Web framework inspired by ASP.NET Core and NestJS

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/linuxerlv/gonest/workflows/CI/badge.svg)](https://github.com/linuxerlv/gonest/actions)

**English** | [中文](README.zh.md) | [日本語](README.ja.md)

---

## Core Features

| Feature | Description |
|---------|-------------|
| 🏗️ **Interface-First Architecture** | Interfaces in `core/abstract/`, implementations in `core/`, extensions depend on interfaces |
| 🔌 **Modular Design** | Middleware, protocols, and task scheduling are independent packages - use what you need |
| ⚡ **Fine-Grained Interfaces** | ContextAbstract = RequestReader + ResponseWriter + ValueStore - compose dependencies at will |
| 📦 **Production Ready** | Config (Koanf), logging (Zap), authentication, authorization - complete solution |

---

## Get Started in 5 Minutes

### Installation

```bash
go get github.com/linuxerlv/gonest
```

### Quick Start

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/extensions"
)

func main() {
    // Create WebApplication builder (ASP.NET Core style)
    builder := core.CreateBuilder()

    // Add services to DI container
    builder.Services().AddSingleton(&MyService{})

    // Build application
    app := builder.BuildWeb()

    // Use middleware (extension methods)
    extensions.UseRecovery(app, nil)
    extensions.UseCORS(app, nil)

    // Register routes
    app.MapGet("/hello", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"message": "Hello World"})
    })

    app.Run()
}
```

### Builder Pattern (Recommended with Dependency Injection)

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/config"
    "github.com/linuxerlv/gonest/extensions"
)

func main() {
    // Create WebApplication builder (ASP.NET Core style)
    builder := core.CreateBuilder()

    // Configure configuration and environment
    cfg := config.NewKoanfConfig(".")
    builder.UseConfig(cfg)
    builder.Environment().Set("APP_ENV", "production")

    // Register services to DI container
    builder.Services().AddSingleton(&MyService{})
    builder.Services().AddScoped(func(s abstract.ServiceCollectionAbstract) *DbContext {
        return &DbContext{DSN: builder.Environment().Get("DATABASE_URL")}
    })

    // Build application
    app := builder.BuildWeb()

    // Get service from DI container
    service := core.GetService[*MyService](app.Services())

    // Use middleware
    extensions.UseRecovery(app, nil)
    extensions.UseCORS(app, &extensions.CORSMiddlewareOptions{
        AllowOrigins: []string{"https://example.com"},
    })

    // Register routes
    app.MapGet("/hello", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"message": "Hello World"})
    })

    app.Run()
}
```

### Generic Application (Non-Web)

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    // Create generic Application builder (for non-web scenarios)
    builder := core.CreateApplicationBuilder()

    // Add services
    builder.Services().AddSingleton(&MyService{})
    builder.Services().AddTransient(func(s abstract.ServiceCollectionAbstract) *Worker {
        return &Worker{}
    })

    // Build application
    app := builder.Build()

    // Get service and run business logic
    service := core.GetService[*MyService](app.Services())
    service.DoWork()
}
```

---

## Project Structure

```
gonest/
├── core/
│   ├── abstract/              # Interface definitions (fine-grained, composable)
│   │   ├── context.go         # ContextAbstract, RequestReaderAbstract, ResponseWriterAbstract...
│   │   ├── router.go          # RouterAbstract, RouteBuilderAbstract...
│   │   ├── middleware.go      # MiddlewareAbstract...
│   │   ├── di.go              # ServiceCollectionAbstract...
│   │   ├── env.go             # EnvAbstract environment variables interface
│   │   └── ...                # Other interfaces
│   │
│   ├── context.go             # HttpContext implementation
│   ├── router.go              # HttpRouter implementation
│   ├── application.go         # Application implementation
│   ├── builder.go             # WebApplicationBuilder implementation
│   ├── env.go                 # Env environment variables implementation
│   └── ...                    # Other implementations
│
├── config/                    # Configuration module
│   ├── config.go              # Config interface (implements abstract.ConfigAbstract)
│   └── koanf.go               # Koanf implementation
│
├── logger/                    # Logger module
│   ├── logger.go              # Logger interface (implements abstract.LoggerAbstract)
│   └── zap.go                 # Zap implementation
│
├── middleware/                # Middleware extension packages
│   ├── auth/                  # JWT authentication
│   ├── session/               # Session management
│   ├── casbin/                # Casbin permission control
│   ├── cors/                  # CORS
│   ├── recovery/              # Panic recovery
│   ├── ratelimit/             # Rate limiting
│   ├── timeout/               # Timeout control
│   ├── gzip/                  # Gzip compression
│   ├── security/              # Security headers
│   ├── logger/                # Logger middleware
│   └── requestid/             # Request ID
│
├── protocol/                  # Protocol extension packages
│   ├── websocket/             # WebSocket
│   ├── sse/                   # Server-Sent Events
│   ├── http3/                 # HTTP/3
│   └── grpc/                  # gRPC
│
├── task/                      # Task scheduling
│   ├── interface.go           # TaskQueue, CronScheduler interfaces
│   ├── asynq.go               # Asynq implementation (Redis)
│   ├── cron.go                # Cron implementation
│   └── memory.go              # Memory implementation
│
├── ipc/                       # IPC interface (inter-process communication)
│   └── interface.go           # Endpoint, Publisher, Subscriber...
│
└── gonest.go                  # Backward compatible type aliases
```

---

## Interface Design

### Fine-Grained Interfaces (core/abstract/)

The framework adopts fine-grained interface design, allowing developers to depend on exactly what they need:

```go
// Request reading interface
type RequestReaderAbstract interface {
    Method() string
    Path() string
    Header(name string) string
}

// Response writing interface
type ResponseWriterAbstract interface {
    Status(code int)
    JSON(code int, v any) error
    String(code int, s string) error
}

// Complete context interface (composed)
type ContextAbstract interface {
    ContextRunnerAbstract
    FullRequestReaderAbstract
    FullResponseWriterAbstract
    ValueStoreAbstract
}
```

### Usage Patterns

```go
// 1. Use complete core package
import "github.com/linuxerlv/gonest/core"
app := core.CreateApplication()

// 2. Use only interface definitions (write extensions)
import "github.com/linuxerlv/gonest/core/abstract"
func MyMiddleware(ctx abstract.ContextAbstract) error { ... }

// 3. Use extension middleware
import "github.com/linuxerlv/gonest/middleware/auth"
app.Use(auth.New(provider, nil))

// 4. Backward compatibility (gonest package type aliases)
import "github.com/linuxerlv/gonest"
app := gonest.NewApplication()
```

---

## API Quick Reference

### Application Creation (ASP.NET Core Style)

```go
import "github.com/linuxerlv/gonest/core"

// Web Application (with HTTP server)
builder := core.CreateBuilder()           // Create WebApplication builder
app := builder.BuildWeb()                 // Build WebApplication
app.Run()                                 // Start server

// Generic Application (non-web scenarios)
builder := core.CreateApplicationBuilder() // Create Application builder
app := builder.Build()                     // Build Application

// Access services
services := builder.Services()             // Get ServiceCollection
cfg := builder.Configuration()             // Get Configuration
env := builder.Environment()               // Get Environment
```

### Route Registration

```go
import "github.com/linuxerlv/gonest/core/abstract"

// MapXXX methods (ASP.NET Core style)
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

// Route groups
api := app.Group("/api/v1")
api.MapGet("/users", listUsers)
```

### Middleware (Extension Methods)

```go
import "github.com/linuxerlv/gonest/extensions"

// UseXXX extension methods
extensions.UseRecovery(app, nil)
extensions.UseCORS(app, &extensions.CORSMiddlewareOptions{
    AllowOrigins: []string{"https://example.com"},
})
extensions.UseRateLimit(app, &extensions.RateLimitMiddlewareOptions{
    Limit:  100,
    Window: 60, // seconds
})
extensions.UseGzip(app, nil)
extensions.UseSecurityHeaders(app, nil)
extensions.UseRequestID(app, nil)
extensions.UseTimeout(app, &extensions.TimeoutMiddlewareOptions{
    Timeout: 30, // seconds
})

// Or use raw middleware
app.Use(middleware)
```

### Dependency Injection

```go
// Register services
builder.Services().AddSingleton(&MyService{})
builder.Services().AddScoped(func(s abstract.ServiceCollectionAbstract) *DbContext {
    return &DbContext{}
})
builder.Services().AddTransient(func(s abstract.ServiceCollectionAbstract) *CacheService {
    return &CacheService{}
})

// Retrieve services from application
service := core.GetService[*MyService](app.Services())
```

### Environment Variables

```go
// Read environment variables
dbUrl := builder.Environment().Get("DATABASE_URL")
port := builder.Environment().GetOrDefault("PORT", "8080")

// Check existence
if builder.Environment().Has("DEBUG") {
    // ...
}

// Get all
allEnv := builder.Env.All()
```

### Configuration Files

```go
import "github.com/linuxerlv/gonest/config"

// Load JSON configuration file
cfg := config.NewKoanfConfig(".")
cfg.Load(
    config.NewFileProvider("config.json", config.NewJSONParser()),
    config.NewJSONParser(),
)

// Load YAML configuration file
cfg.Load(
    config.NewFileProvider("config.yaml", config.NewYAMLParser()),
    config.NewYAMLParser(),
)

// Read configuration
port := cfg.GetString("server.port")
debug := cfg.GetBool("debug")

// Bind to struct
type ServerConfig struct {
    Port    string `koanf:"port"`
    Timeout int    `koanf:"timeout"`
}
var serverCfg ServerConfig
cfg.Unmarshal("server", &serverCfg)

// Set to builder
builder := core.CreateBuilder()
builder.Config = cfg
```

---

## Documentation Navigation

| Document | Audience | Content |
|----------|----------|---------|
| **[Tutorial](TUTORIAL.md)** | 🎓 Go Beginners | Progressive learning guide |
| **[API Reference](API_REFERENCE.md)** | 👨‍💻 Application Developers | Complete API documentation |
| **[Contributing Guide](DEVELOPER.md)** | 🛠️ Framework Contributors | Architecture design, coding standards, testing strategies, extension mechanisms |

---

## License

MIT License - See [LICENSE](LICENSE) file