# Gonest API 参考文档

> 本文档详细说明 Gonest 框架的所有公开 API

---

## 目录

1. [核心接口（core/abstract）](#核心接口coreabstract)
2. [核心实现（core）](#核心实现core)
3. [中间件扩展包](#中间件扩展包)
4. [配置系统](#配置系统)
5. [日志系统](#日志系统)
6. [任务调度](#任务调度)
7. [IPC 接口](#ipc-接口)

---

## 核心接口（core/abstract）

### Context 相关接口

```go
// 请求读取接口
type RequestReaderAbstract interface {
    Method() string
    Path() string
    Header(name string) string
}

// 原生请求访问接口
type RawRequestAbstract interface {
    Request() *http.Request
}

// 路径参数读取接口
type PathParamsReaderAbstract interface {
    Param(name string) string
}

// Query 参数读取接口
type QueryReaderAbstract interface {
    Query(name string) string
}

// 请求体读取接口
type BodyReaderAbstract interface {
    Body() []byte
    Bind(v any) error
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

// 上下文运行器接口
type ContextRunnerAbstract interface {
    Context() context.Context
}

// 完整上下文接口（组合）
type ContextAbstract interface {
    ContextRunnerAbstract
    FullRequestReaderAbstract
    FullResponseWriterAbstract
    ValueStoreAbstract
}
```

### 路由相关接口

```go
type RouteHandlerAbstract func(ctx ContextAbstract) error

type RouterAbstract interface {
    GET(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
    POST(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
    PUT(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
    DELETE(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
    PATCH(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
    OPTIONS(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
    Group(prefix string) RouteGroupAbstract
    Match(req *http.Request) (RouteAbstract, map[string]string)
}

type RouteBuilderAbstract interface {
    Guard(guard GuardAbstract) RouteBuilderAbstract
    Interceptor(interceptor InterceptorAbstract) RouteBuilderAbstract
    Pipe(pipe PipeAbstract) RouteBuilderAbstract
}
```

### 中间件接口

```go
type MiddlewareAbstract interface {
    Handle(ctx ContextAbstract, next func() error) error
}

type MiddlewareFuncAbstract func(ctx ContextAbstract, next func() error) error

func (f MiddlewareFuncAbstract) Handle(ctx ContextAbstract, next func() error) error
```

### Guard / Interceptor / Pipe / Filter

```go
type GuardAbstract interface {
    CanActivate(ctx ContextAbstract) bool
}

type InterceptorAbstract interface {
    Intercept(ctx ContextAbstract, next RouteHandlerAbstract) (any, error)
}

type PipeAbstract interface {
    Transform(value any, ctx ContextAbstract) (any, error)
}

type ExceptionFilterAbstract interface {
    Catch(ctx ContextAbstract, err error) error
}
```

### DI 接口

```go
type ServiceLifetimeAbstract int

const (
    Singleton ServiceLifetimeAbstract = iota
    Scoped
    Transient
)

type ServiceCollectionAbstract interface {
    GetService(serviceType reflect.Type) any
    GetRequiredService(serviceType reflect.Type) any
    AddSingleton(instance any) ServiceRegistrarAbstract
    AddScoped(serviceType reflect.Type, factory any) ServiceRegistrarAbstract
    AddTransient(serviceType reflect.Type, factory any) ServiceRegistrarAbstract
}

type ScopeAbstract interface {
    Dispose()
    IsDisposed() bool
}
```

### 环境变量接口

```go
type EnvAbstract interface {
    Get(key string) string
    GetOrDefault(key, defaultValue string) string
    Has(key string) bool
    All() map[string]string
    Set(key, value string)
    Unset(key string)
}
```

### 应用接口

```go
type ApplicationAbstract interface {
    RouterAbstract
    Use(middleware MiddlewareAbstract) ApplicationAbstract
    UseGlobalGuards(guards ...GuardAbstract) ApplicationAbstract
    UseGlobalInterceptors(interceptors ...InterceptorAbstract) ApplicationAbstract
    UseGlobalPipes(pipes ...PipeAbstract) ApplicationAbstract
    UseGlobalFilters(filters ...ExceptionFilterAbstract) ApplicationAbstract
    Controller(controller ControllerAbstract) ApplicationAbstract
    Listen(addr string) error
    Shutdown(ctx context.Context) error
}

type WebApplicationAbstract interface {
    ApplicationAbstract
    Configuration() ConfigAbstract
    Log() LoggerAbstract
    Run() error
    RunAsync() <-chan error
}

type WebApplicationBuilderAbstract interface {
    UseConfig(cfg ConfigAbstract) WebApplicationBuilderAbstract
    UseLogger(log LoggerAbstract) WebApplicationBuilderAbstract
    ConfigureServices(configure func(ServiceCollectionAbstract)) WebApplicationBuilderAbstract
    Build() WebApplicationAbstract
}
```

---

## 核心实现（core）

### 结构体定义

```go
// Application - 基础应用
type Application struct {
    Config   config.Config          // 公开属性 - 框架配置
    Env      abstract.EnvAbstract   // 公开属性 - 环境变量
    Services *ServiceCollection     // 公开属性 - DI 容器
    // ... 其他私有字段
}

// WebApplication - Web 应用
type WebApplication struct {
    *Application
    Config   config.Config          // 公开属性 - 框架配置
    Env      abstract.EnvAbstract   // 公开属性 - 环境变量
    Services *ServiceCollection     // 公开属性 - DI 容器
    Logger   logger.Logger          // 公开属性 - 日志
    // ... 其他私有字段
}

// WebApplicationBuilder - 应用构建器
type WebApplicationBuilder struct {
    Services *ServiceCollection     // 公开属性 - DI 容器（Wire 注入友好）
    Config   config.Config          // 公开属性 - 框架配置
    Env      abstract.EnvAbstract   // 公开属性 - 环境变量
    Logger   logger.Logger          // 公开属性 - 日志
    Host     *HostBuilder           // 公开属性 - 主机构建器
    // ... 其他私有字段
}
```

### 应用创建

```go
import "github.com/linuxerlv/gonest/core"

// 快速创建（适合简单应用）
func CreateApplication(args ...string) *WebApplication

// Builder 模式（推荐）
func CreateBuilder(args ...string) *WebApplicationBuilder
func NewWebApplicationBuilder() *WebApplicationBuilder

// 创建基础组件
func NewApplication() *Application
func NewRouter() *HttpRouter
func NewContext(w http.ResponseWriter, r *http.Request) *HttpContext
func NewContextWithParams(w http.ResponseWriter, r *http.Request, params map[string]string) *HttpContext
func NewServiceCollection() *ServiceCollection
func NewEnv() *Env
```

### 使用示例

```go
// Builder 模式
builder := core.CreateBuilder()
builder.Config = config.NewKoanfConfig(".")
builder.Env.Set("APP_ENV", "production")
builder.Services.AddSingleton(&MyService{})
app := builder.Build().(*core.WebApplication)

// 访问公开属性
cfg := app.Config
env := app.Env
services := app.Services

// 从 DI 获取服务
service := core.GetService[*MyService](services)
```

### HttpContext

```go
type HttpContext struct {
    // 实现 abstract.ContextAbstract
}

// 请求方法
func (c *HttpContext) Request() *http.Request
func (c *HttpContext) Method() string
func (c *HttpContext) Path() string
func (c *HttpContext) Param(name string) string
func (c *HttpContext) Query(name string) string
func (c *HttpContext) Header(name string) string
func (c *HttpContext) Body() []byte
func (c *HttpContext) Bind(v any) error

// 响应方法
func (c *HttpContext) Status(code int)
func (c *HttpContext) JSON(code int, v any) error
func (c *HttpContext) String(code int, s string) error
func (c *HttpContext) Data(code int, contentType string, data []byte) error

// 存储
func (c *HttpContext) Set(key string, value any)
func (c *HttpContext) Get(key string) any

// 上下文
func (c *HttpContext) Context() context.Context
```

### 依赖注入泛型方法

```go
// 注册服务
func AddSingleton[T any](s *ServiceCollection, instance T) *ServiceCollection
func AddSingletonFunc[T any](s *ServiceCollection, factory func(abstract.ServiceCollectionAbstract) T) *ServiceCollection
func AddScoped[T any](s *ServiceCollection, factory func(abstract.ServiceCollectionAbstract) T) *ServiceCollection
func AddTransient[T any](s *ServiceCollection, factory func(abstract.ServiceCollectionAbstract) T) *ServiceCollection

// 获取服务
func GetService[T any](s abstract.ServiceCollectionAbstract) T
func GetRequiredService[T any](s abstract.ServiceCollectionAbstract) T
```

### 环境变量

```go
import "github.com/linuxerlv/gonest/core"

env := core.NewEnv()

// 读取环境变量
dbUrl := env.Get("DATABASE_URL")
port := env.GetOrDefault("PORT", "8080")

// 检查存在
if env.Has("DEBUG") {
    // ...
}

// 获取所有
allEnv := env.All()

// 设置（用于测试）
env.Set("KEY", "value")
env.Unset("KEY")
```

### HTTP 错误

```go
import "github.com/linuxerlv/gonest/core/abstract"

func BadRequest(message string) error       // 400
func Unauthorized(message string) error     // 401
func Forbidden(message string) error        // 403
func NotFound(message string) error         // 404
func InternalError(message string) error    // 500
func NewHttpException(code int, message string) error
```

---

## 中间件扩展包

### 使用方式

```go
import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/middleware/cors"
    "github.com/linuxerlv/gonest/middleware/recovery"
    "github.com/linuxerlv/gonest/middleware/ratelimit"
)

app := core.CreateApplication()
app.Use(recovery.New(nil))
app.Use(cors.New(nil))
```

### 可用中间件

| 包 | 说明 | 配置结构 |
|---|------|---------|
| `middleware/cors` | CORS 跨域 | `cors.Config` |
| `middleware/recovery` | Panic 恢复 | `recovery.Config` |
| `middleware/ratelimit` | 限流 | `ratelimit.Config` |
| `middleware/timeout` | 超时控制 | `timeout.Config` |
| `middleware/gzip` | Gzip 压缩 | `gzip.Config` |
| `middleware/security` | 安全头 | `security.Config` |
| `middleware/requestid` | 请求 ID | `requestid.Config` |
| `middleware/logger` | 日志中间件 | `logger.Config` |
| `middleware/auth` | JWT 认证 | `auth.Config` |
| `middleware/session` | Session 管理 | `session.Config` |
| `middleware/casbin` | Casbin 权限 | `casbin.Config` |
| `middleware/oauth` | OAuth 认证 | `oauth.Config` |

### CORS 示例

```go
import "github.com/linuxerlv/gonest/middleware/cors"

app.Use(cors.New(&cors.Config{
    AllowOrigins:     []string{"https://example.com"},
    AllowMethods:     []string{"GET", "POST"},
    AllowHeaders:     []string{"Authorization"},
    AllowCredentials: true,
    MaxAge:           86400,
}))
```

### Auth 示例

```go
import "github.com/linuxerlv/gonest/middleware/auth"

jwtProvider := auth.NewJWTProvider(&auth.JWTConfig{
    Secret:          "my-secret",
    AccessTokenTTL:  time.Hour,
    RefreshTokenTTL: 24 * time.Hour,
}, nil)

app.Use(auth.New(jwtProvider, nil).AsMiddleware())
```

---

## 配置系统

### 加载配置文件

```go
import "github.com/linuxerlv/gonest/config"

cfg := config.NewKoanfConfig(".")

// 加载 JSON 文件
cfg.Load(
    config.NewFileProvider("config.json", config.NewJSONParser()),
    config.NewJSONParser(),
)

// 加载 YAML 文件
cfg.Load(
    config.NewFileProvider("config.yaml", config.NewYAMLParser()),
    config.NewYAMLParser(),
)

// 加载环境变量（前缀 APP_，映射为小写并用 . 分隔）
// APP_DATABASE_HOST -> database.host
cfg.Load(config.NewEnvProvider(config.WithEnvPrefix("APP_")), nil)

// 读取配置
port := cfg.GetString("server.port")
debug := cfg.GetBool("debug")

// 结构体绑定
type ServerConfig struct {
    Port    string `koanf:"port"`
    Timeout int    `koanf:"timeout"`
}
var serverCfg ServerConfig
cfg.Unmarshal("server", &serverCfg)
```

### 配置文件示例

**config.yaml**:
```yaml
server:
  port: "8080"
  name: "myapp"
  timeout: 30

database:
  host: "localhost"
  port: 5432
  name: "mydb"

log:
  level: "info"
```

**config.json**:
```json
{
  "server": {
    "port": "8080",
    "name": "myapp",
    "timeout": 30
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "mydb"
  }
}
```

### 多配置源合并

```go
cfg := config.NewKoanfConfig(".")

// 1. 默认配置
defaults := map[string]any{
    "server.port": "8080",
    "debug": false,
}
cfg.Load(config.NewMapProvider(defaults, "."), nil)

// 2. 配置文件（覆盖默认）
cfg.Load(config.NewFileProvider("config.yaml", config.NewYAMLParser()), config.NewYAMLParser())

// 3. 环境变量（覆盖文件）
cfg.Load(config.NewEnvProvider(config.WithEnvPrefix("APP_")), nil)

// 设置到应用
builder := core.CreateBuilder()
builder.Config = cfg
```

### 可用 Provider

| Provider | 说明 |
|----------|------|
| `NewFileProvider(path, parser)` | 加载配置文件 |
| `NewEnvProvider(opts...)` | 加载环境变量 |
| `NewMapProvider(data, delim)` | 加载 Map |
| `NewFlagProvider(flagSet, delim)` | 加载命令行参数 |
| `NewStructProvider(data, tag)` | 加载结构体 |

### 可用 Parser

| Parser | 说明 |
|--------|------|
| `NewJSONParser()` | JSON 解析 |
| `NewYAMLParser()` | YAML 解析 |

### 配置方法

```go
cfg.GetString("key")       // 获取字符串
cfg.GetInt("key")          // 获取整数
cfg.GetBool("key")         // 获取布尔值
cfg.GetFloat64("key")      // 获取浮点数
cfg.GetDuration("key")     // 获取时长
cfg.GetStringSlice("key")  // 获取字符串数组
cfg.IsSet("key")           // 判断是否存在
cfg.Unmarshal("key", &out) // 绑定到结构体
cfg.All()                  // 获取所有配置
```

---

## 日志系统

```go
import "github.com/linuxerlv/gonest/logger"

cfg := logger.DefaultConfig()
log, _ := logger.NewZapLogger(cfg)

log.Info("Server started", 
    logger.String("port", "8080"),
    logger.Int("workers", 4),
)

// 子 Logger
userLog := log.WithName("user-service")
```

---

## 任务调度

### CronScheduler

```go
import "github.com/linuxerlv/gonest/task"

scheduler := task.NewMemoryCronScheduler()

scheduler.AddJob("0 0 * * *", "daily-cleanup", func(ctx context.Context) error {
    return cleanupOldData()
})

scheduler.Start()
```

### TaskQueue

```go
import "github.com/linuxerlv/gonest/task"

queue := task.NewMemoryQueue("tasks", 5, 1000)

queue.RegisterHandler("send-email", func(ctx context.Context, t *task.QueueTask) error {
    return sendEmail(t.Payload)
})

queue.Start(context.Background())
queue.Enqueue(&task.QueueTask{Type: "send-email", Payload: data})
```

---

## IPC 接口

IPC 包只提供接口定义，用户自行实现：

```go
import "github.com/linuxerlv/gonest/ipc"

// Endpoint 基础端点接口
type Endpoint interface {
    Bind() error
    Connect() error
    Send(msg *Message) error
    Recv() (*Message, error)
    Close() error
    // ...
}

// Publisher / Subscriber 发布订阅
// Requester / Replier 请求回复
// Factory 工厂接口
```

---

## Wire 依赖注入

Gonest 的公开属性设计便于与 Wire 配合使用：

```go
// wire.go
//go:build wireinject

package main

import (
    "github.com/google/wire"
    "github.com/linuxerlv/gonest/core"
)

func InitializeApplication() (*core.WebApplication, error) {
    wire.Build(
        // 底层依赖
        ProvideConfig,    // → config.Config
        ProvideLogger,    // → logger.Logger
        ProvideEnv,       // → *core.Env
        
        // 服务注册
        ProvideServices,  // → *core.ServiceCollection
        
        // 最终组装
        ProvideWebApplication,  // → *core.WebApplication
    )
    return nil, nil
}

func ProvideConfig() *config.KoanfConfig {
    return config.NewKoanfConfig(".")
}

func ProvideLogger() logger.Logger {
    return logger.NewNopLogger()
}

func ProvideEnv() *core.Env {
    return core.NewEnv()
}

func ProvideServices(cfg config.Config, env *core.Env) *core.ServiceCollection {
    services := core.NewServiceCollection()
    
    // 注册服务，依赖 Config 和 Env
    core.AddSingleton(services, &MyService{})
    core.AddScoped(services, func(s abstract.ServiceCollectionAbstract) *DbContext {
        return &DbContext{DSN: env.Get("DATABASE_URL")}
    })
    
    return services
}

func ProvideWebApplication(
    cfg config.Config,
    log logger.Logger,
    env *core.Env,
    services *core.ServiceCollection,
) *core.WebApplication {
    builder := core.NewWebApplicationBuilder()
    builder.Config = cfg
    builder.Logger = log
    builder.Env = env
    builder.Services = services
    return builder.Build().(*core.WebApplication)
}
```

### 使用方式

```go
func main() {
    app, err := InitializeApplication()
    if err != nil {
        panic(err)
    }
    
    // Wire 已注入所有依赖
    app.MapGet("/hello", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"message": "Hello"})
    })
    
    app.Run()
}
```

---

## 向后兼容

`gonest` 包提供类型别名，现有代码可继续使用：

```go
import "github.com/linuxerlv/gonest"

// 类型别名
type Context = abstract.ContextAbstract
type Application = abstract.ApplicationAbstract
type Middleware = abstract.MiddlewareAbstract
// ...

// 函数别名
var NewApplication = core.NewApplication
var CreateApplication = core.CreateApplication
// ...
```