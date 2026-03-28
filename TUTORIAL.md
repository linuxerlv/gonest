# Gonest 渐进式教程

> 专为 Go 语言新手设计的框架学习指南

---

## 写给新手的话

如果你刚刚学会 Go 语言的基础语法（变量、函数、结构体等），对 Web 开发完全没有概念，这个教程就是为你准备的。

我们会从最简单的概念开始，一步一步教你用 Gonest 框架开发 Web 应用。每个知识点都有详细的解释和完整的代码示例。

**学习建议**：
- 按章节顺序学习，不要跳步
- 每个示例代码都动手运行一遍
- 遇到不懂的概念，先停下来思考或查阅资料

---

## 目录

1. [第一章：Hello World - 你的第一个 Web 应用](#第一章hello-world---你的第一个-web-应用)
2. [第二章：路由基础 - 定义网页地址](#第二章路由基础---定义网页地址)
3. [第三章：请求和响应 - 与用户交互](#第三章请求和响应---与用户交互)
4. [第四章：中间件 - 请求的把关人](#第四章中间件---请求的把关人)
5. [第五章：错误处理 - 程序出错了怎么办](#第五章错误处理---程序出错了怎么办)
6. [第六章：Guard、Interceptor、Pipe - 进阶功能](#第六章guardinterceptorpipe---进阶功能)
7. [第七章：Controller 模式 - 组织你的代码](#第七章controller-模式---组织你的代码)
8. [第八章：依赖注入 - 管理你的服务](#第八章依赖注入---管理你的服务)
9. [第九章：配置系统 - 让应用更灵活](#第九章配置系统---让应用更灵活)
10. [第十章：日志系统 - 记录应用行为](#第十章日志系统---记录应用行为)
11. [第十一章：任务调度 - 自动执行任务](#第十一章任务调度---自动执行任务)
12. [第十二章：认证和授权 - 保护你的应用](#第十二章认证和授权---保护你的应用)
13. [第十三章：完整项目示例](#第十三章完整项目示例)

---

## 第一章：Hello World - 你的第一个 Web 应用

### 1.1 什么是 Web 应用？

Web 应用就像一家餐厅：

- **顾客**：使用浏览器的人
- **服务员**：Web 应用，接收顾客的请求，提供响应
- **菜单**：路由，告诉顾客有哪些服务可用
- **厨房**：处理逻辑的代码

当顾客（浏览器）说"我要一份 Hello World"，服务员（Web 应用）就去厨房准备好，然后端给顾客。

### 1.2 安装 Gonest

首先，你需要有一个 Go 项目。在终端执行：

```bash
# 创建项目文件夹
mkdir myapp
cd myapp

# 初始化 Go 项目（这会创建 go.mod 文件）
go mod init myapp

# 安装 Gonest 框架
go get github.com/linuxerlv/gonest
```

### 1.3 创建第一个应用

创建一个 `main.go` 文件：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/extensions"
)

func main() {
    // 第一步：创建 WebApplication 构建器（ASP.NET Core 风格）
    builder := core.CreateBuilder()

    // 第二步：构建应用
    app := builder.Build()

    // 第三步：使用中间件（可选）
    extensions.UseRecovery(app, nil)

    // 第四步：定义一个路由（菜单上的一道菜）
    // 当有人访问 "/hello" 这个地址时，执行下面的函数
    app.MapGet("/hello", func(ctx abstract.ContextAbstract) error {
        // 返回一段文字给用户
        return ctx.String(200, "Hello World!")
    })

    // 第五步：启动应用，监听在 8080 端口
    app.Run()
}
```

### 1.4 运行并测试

```bash
# 运行你的应用
go run main.go
```

你会看到类似这样的输出：
```
Server starting on :8080
```

现在打开浏览器，访问 `http://localhost:8080/hello`，你会看到：
```
Hello World!
```

### 1.5 代码解释

让我们逐行理解：

```go
app := core.CreateApplication()
```
这行代码创建了一个 Web 应用实例。`core` 是框架的核心包，`CreateApplication` 是创建应用的函数。

```go
app.GET("/hello", func(ctx abstract.ContextAbstract) error {
    return ctx.String(200, "Hello World!")
})
```
- `GET`：表示这是一个 GET 请求（浏览器直接访问网页就是 GET 请求）
- `"/hello"`：定义网址路径
- `func(ctx abstract.ContextAbstract) error`：处理函数，`ctx` 是上下文，包含请求信息
- `ctx.String(200, "Hello World!")`：返回纯文本响应，200 是状态码（表示成功）

```go
app.Run()
```
启动应用，默认监听 8080 端口。

### 1.6 什么是状态码？

HTTP 状态码告诉浏览器请求的结果：

| 状态码 | 含义 | 说明 |
|-------|------|------|
| 200 | 成功 | 请求成功处理 |
| 201 | 已创建 | 创建新资源成功 |
| 400 | 请求错误 | 用户提交的数据有问题 |
| 401 | 未授权 | 需要登录 |
| 403 | 禁止访问 | 没有权限 |
| 404 | 未找到 | 网址不存在 |
| 500 | 服务器错误 | 代码出错了 |

---

## 第二章：路由基础 - 定义网页地址

### 2.1 什么是路由？

路由就是把网址（URL）和处理函数对应起来。比如：

- `/users` → 显示用户列表
- `/users/123` → 显示用户 123 的信息
- `/login` → 登录页面

### 2.2 不同的 HTTP 方法

除了 GET，还有其他 HTTP 方法：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // GET：获取数据（比如查看用户列表）
    app.MapGet("/users", func(ctx abstract.ContextAbstract) error {
        return ctx.String(200, "获取用户列表")
    })

    // POST：创建数据（比如注册新用户）
    app.MapPost("/users", func(ctx abstract.ContextAbstract) error {
        return ctx.String(201, "创建新用户")
    })

    // PUT：更新数据（比如修改用户信息）
    app.MapPut("/users/:id", func(ctx abstract.ContextAbstract) error {
        id := ctx.Param("id")
        return ctx.String(200, "更新用户 "+id)
    })

    // DELETE：删除数据（比如删除用户）
    app.MapDelete("/users/:id", func(ctx abstract.ContextAbstract) error {
        id := ctx.Param("id")
        return ctx.String(200, "删除用户 "+id)
    })

    // PATCH：部分更新数据
    app.MapPatch("/users/:id", func(ctx abstract.ContextAbstract) error {
        id := ctx.Param("id")
        return ctx.String(200, "部分更新用户 "+id)
    })
    
    app.Run()
}
```

### 2.3 路径参数（动态路径）

有时候网址的一部分是变化的，比如 `/users/123`、`/users/456`。我们可以用路径参数来捕获这些变化的部分：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // :id 是路径参数，可以匹配任何值
    app.MapGet("/users/:id", func(ctx abstract.ContextAbstract) error {
        // 用 Param 方法获取参数值
        id := ctx.Param("id")
        return ctx.String(200, "用户ID是: " + id)
    })

    // 多个路径参数
    app.MapGet("/users/:userId/posts/:postId", func(ctx abstract.ContextAbstract) error {
        userId := ctx.Param("userId")
        postId := ctx.Param("postId")
        return ctx.String(200, "用户 "+userId+" 的文章 "+postId)
    })

    app.Run()
}
```

测试：
- `http://localhost:8080/users/123` → 显示 "用户ID是: 123"
- `http://localhost:8080/users/abc` → 显示 "用户ID是: abc"
- `http://localhost:8080/users/42/posts/100` → 显示 "用户 42 的文章 100"

### 2.4 Query 参数（网址后面的参数）

Query 参数是网址 `?` 后面的部分，比如 `/search?q=golang&page=1`：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    app.MapGet("/search", func(ctx abstract.ContextAbstract) error {
        // 用 Query 方法获取参数
        q := ctx.Query("q")      // 搜索关键词
        page := ctx.Query("page") // 页码

        return ctx.String(200, "搜索: "+q+", 页码: "+page)
    })

    app.Run()
}
```

测试 `http://localhost:8080/search?q=golang&page=2`：
- 显示 "搜索: golang, 页码: 2"

### 2.5 路由组

当多个路由有相同的前缀时，可以用路由组：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 创建一个路由组，前缀是 "/api/v1"
    api := app.Group("/api/v1")

    // 所有路由都会自动加上 "/api/v1" 前缀
    api.MapGet("/users", func(ctx abstract.ContextAbstract) error {
        return ctx.String(200, "API v1 用户列表")
    })

    api.MapGet("/users/:id", func(ctx abstract.ContextAbstract) error {
        return ctx.String(200, "API v1 用户详情")
    })
    
    api.POST("/users", func(ctx abstract.ContextAbstract) error {
        return ctx.String(201, "API v1 创建用户")
    })
    
    // 实际路径是：
    // /api/v1/users
    // /api/v1/users/:id
    // /api/v1/users (POST)
    
    app.Run()
}
```

### 2.6 小结

| 功能 | 方法 | 示例 |
|------|------|------|
| 创建构建器 | `core.CreateBuilder()` | `builder := core.CreateBuilder()` |
| 构建应用 | `builder.Build()` | `app := builder.Build()` |
| 定义 GET 路由 | `app.MapGet()` | `app.MapGet("/hello", handler)` |
| 定义 POST 路由 | `app.MapPost()` | `app.MapPost("/users", handler)` |
| 获取路径参数 | `ctx.Param()` | `ctx.Param("id")` |
| 获取 Query 参数 | `ctx.Query()` | `ctx.Query("q")` |
| 创建路由组 | `app.Group()` | `api := app.Group("/api")` |

---

## 第三章：请求和响应 - 与用户交互

### 3.1 什么是 Context？

Context（上下文）就像服务员手中的托盘，里面包含：
- **请求信息**：顾客想要什么（请求方法、路径、参数、数据等）
- **响应能力**：服务员能给顾客什么（返回文字、JSON、HTML等）
- **存储空间**：可以在托盘上放东西，传递给下一个处理步骤

### 3.2 获取请求信息

```go
package main

import (
    "fmt"
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    app.MapGet("/info", func(ctx abstract.ContextAbstract) error {
        // 获取请求方法（GET、POST等）
        method := ctx.Method()

        // 获取请求路径
        path := ctx.Path()

        // 获取请求头（比如 Authorization）
        authHeader := ctx.Header("Authorization")

        // 获取 Query 参数
        name := ctx.Query("name")

        // 组合成响应信息
        info := fmt.Sprintf(
            "方法: %s\n路径: %s\nAuthorization: %s\n名字: %s",
            method, path, authHeader, name,
        )

        return ctx.String(200, info)
    })

    app.Run()
}
```

测试 `http://localhost:8080/info?name=test`：
```
方法: GET
路径: /info
Authorization: 
名字: test
```

### 3.3 获取请求体（POST 数据）

用户通过 POST 提交的数据在请求体中：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

// 定义一个结构体来接收数据
type User struct {
    Name string `json:"name"` // json:"name" 表示对应 JSON 中的 name 字段
    Age  int    `json:"age"`
}

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    app.MapPost("/users", func(ctx abstract.ContextAbstract) error {
        // 创建一个空的 User 结构体
        var user User

        // Bind 方法把 JSON 数据解析到 user 结构体中
        if err := ctx.Bind(&user); err != nil {
            return ctx.String(400, "数据格式错误: "+err.Error())
        }

        // 现可以使用 user.Name 和 user.Age
        return ctx.String(200, "收到: 名字="+user.Name+", 年龄="+string(user.Age))
    })

    app.Run()
}
```

测试：用 curl 或 Postman 发送 POST 请求：
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"张三","age":25}'
```

### 3.4 返回 JSON 响应

JSON 是一种数据格式，常用于前后端通信：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    app.MapGet("/user", func(ctx abstract.ContextAbstract) error {
        // 创建一个用户数据
        user := User{
            ID:   1,
            Name: "张三",
            Age:  25,
        }

        // 返回 JSON 格式
        return ctx.JSON(200, user)
    })

    // 返回数组
    app.MapGet("/users", func(ctx abstract.ContextAbstract) error {
        users := []User{
            {ID: 1, Name: "张三", Age: 25},
            {ID: 2, Name: "李四", Age: 30},
        }

        return ctx.JSON(200, users)
    })

    // 返回带消息的响应
    app.MapGet("/message", func(ctx abstract.ContextAbstract) error {
        response := map[string]any{
            "status":  "success",
            "message": "操作成功",
            "data":    User{ID: 1, Name: "张三", Age: 25},
        }

        return ctx.JSON(200, response)
    })

    app.Run()
}
```

访问 `/user` 返回：
```json
{"id":1,"name":"张三","age":25}
```

访问 `/users` 返回：
```json
[
  {"id":1,"name":"张三","age":25},
  {"id":2,"name":"李四","age":30}
]
```

### 3.5 在 Context 中存储数据

Context 可以存储数据，在中间件和处理函数之间传递：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 第一个处理函数存储数据
    app.MapGet("/test", func(ctx abstract.ContextAbstract) error {
        // 存储数据
        ctx.Set("user_id", "12345")
        ctx.Set("role", "admin")

        // 获取数据
        userId := ctx.Get("user_id")

        return ctx.String(200, "用户ID: "+userId.(string))
    })

    app.Run()
}
```

### 3.6 小结

| 功能 | 方法 | 示例 |
|------|------|------|
| 获取请求方法 | `ctx.Method()` | `"GET"` |
| 获取请求路径 | `ctx.Path()` | `"/users"` |
| 获取请求头 | `ctx.Header()` | `ctx.Header("Authorization")` |
| 获取 Query 参数 | `ctx.Query()` | `ctx.Query("name")` |
| 获取路径参数 | `ctx.Param()` | `ctx.Param("id")` |
| 解析请求体 | `ctx.Bind()` | `ctx.Bind(&user)` |
| 返回文字 | `ctx.String()` | `ctx.String(200, "hello")` |
| 返回 JSON | `ctx.JSON()` | `ctx.JSON(200, data)` |
| 存储数据 | `ctx.Set()` | `ctx.Set("key", value)` |
| 获取存储的数据 | `ctx.Get()` | `ctx.Get("key")` |

---

## 第四章：中间件 - 请求的把关人

### 4.1 什么是中间件？

中间件就像餐厅的安检员，在顾客进入餐厅前检查：
- 是否有预约
- 是否穿着得体
- 是否携带违禁物品

在 Web 应用中，中间件在请求到达处理函数之前执行，可以：
- 记录日志
- 检查用户是否登录
- 处理跨域请求
- 恢复程序崩溃

### 4.2 中间件的执行顺序

中间件按添加顺序执行，形成"洋葱"结构：

```
请求进入 → 中间件1 → 中间件2 → 处理函数 → 中间件2 → 中间件1 → 响应返回
```

```go
package main

import (
    "fmt"
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 第一个中间件
    app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
        fmt.Println("中间件1：开始")
        err := next()  // 调用下一个中间件或处理函数
        fmt.Println("中间件1：结束")
        return err
    }))

    // 第二个中间件
    app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
        fmt.Println("中间件2：开始")
        err := next()
        fmt.Println("中间件2：结束")
        return err
    }))

    // 处理函数
    app.MapGet("/test", func(ctx abstract.ContextAbstract) error {
        fmt.Println("处理函数执行")
        return ctx.String(200, "OK")
    })
    
    app.Run()
}
```

访问 `/test` 时，终端输出：
```
中间件1：开始
中间件2：开始
处理函数执行
中间件2：结束
中间件1：结束
```

### 4.3 使用内置中间件

Gonest 提供了多个内置中间件：

#### Recovery 中间件（恢复崩溃）

当程序崩溃（panic）时，自动恢复并返回错误响应：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/extensions"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 添加 Recovery 中间件（使用扩展方法）
    extensions.UseRecovery(app, nil)

    // 这个路由会崩溃，但 Recovery 会恢复
    app.MapGet("/panic", func(ctx abstract.ContextAbstract) error {
        panic("程序崩溃了！")
    })

    app.Run()
}
```

访问 `/panic` 时，不会让整个应用崩溃，而是返回 500 错误。

#### CORS 中间件（跨域处理）

当前端和后端不在同一个域名时，需要 CORS 中间件：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/extensions"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 配置 CORS（使用扩展方法）
    extensions.UseCORS(app, &extensions.CORSMiddlewareOptions{
        AllowOrigins:     []string{"http://localhost:3000"}, // 允许的前端地址
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        AllowCredentials: true,  // 允许携带 Cookie
        MaxAge:           86400, // 缓存时间（秒）
    })

    app.MapGet("/data", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"message": "跨域成功"})
    })

    app.Run()
}
```

#### RateLimit 中间件（限流）

防止用户请求太频繁：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/extensions"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 每分钟最多 10 次请求（使用扩展方法）
    extensions.UseRateLimit(app, &extensions.RateLimitMiddlewareOptions{
        Limit:  10,  // 限制次数
        Window: 60,  // 时间窗口（秒）
    })

    app.MapGet("/limited", func(ctx abstract.ContextAbstract) error {
        return ctx.String(200, "请求成功")
    })

    app.Run()
}
```

超过限制时返回 429 错误。

### 4.4 创建自定义中间件

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

// 创建一个记录请求时间的中间件
func TimingMiddleware() abstract.MiddlewareAbstract {
    return abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
        // 前置处理：记录开始时间
        startTime := ctx.Context().Value("start_time") // 这里简化了
        
        // 执行后续处理
        err := next()
        
        // 后置处理：可以在这里计算耗时并记录日志
        // elapsed := time.Since(startTime)
        
        return err
    })
}

// 创建一个检查 Token 的中间件
func AuthMiddleware() abstract.MiddlewareAbstract {
    return abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
        // 获取 Authorization 头
        token := ctx.Header("Authorization")
        
        // 如果没有 token，返回 401 错误
        if token == "" {
            return ctx.String(401, "请先登录")
        }
        
        // 验证 token（这里简化，实际应该验证）
        if token != "valid-token" {
            return ctx.String(401, "Token 无效")
        }
        
        // token 有效，继续执行
        return next()
    })
}

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 使用自定义中间件
    app.Use(TimingMiddleware())
    app.Use(AuthMiddleware())

    app.MapGet("/protected", func(ctx abstract.ContextAbstract) error {
        return ctx.String(200, "这是需要登录才能看到的内容")
    })

    app.Run()
}
```

### 4.5 中间件可以提前结束请求

中间件可以选择不调用 `next()`，直接返回响应：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 检查 API Key 的中间件
    app.Use(abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
        apiKey := ctx.Header("X-API-Key")

        // 如果 API Key 不正确，直接返回错误，不继续执行
        if apiKey != "my-secret-key" {
            return ctx.String(403, "API Key 错误")
        }
        
        // API Key 正确，继续执行
        return next()
    }))

    app.MapGet("/api/data", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"data": "敏感数据"})
    })

    app.Run()
}
```

只有携带正确 `X-API-Key` 头的请求才能访问 `/api/data`。

### 4.6 小结

| 中间件 | 功能 | 使用场景 | 扩展方法 |
|--------|------|----------|----------|
| Recovery | 恢复崩溃 | 防止 panic 导致应用崩溃 | `extensions.UseRecovery(app, nil)` |
| CORS | 跨域处理 | 前后端分离项目 | `extensions.UseCORS(app, options)` |
| RateLimit | 限流 | 防止恶意请求 | `extensions.UseRateLimit(app, options)` |
| Gzip | 压缩响应 | 减少传输大小 | `extensions.UseGzip(app, nil)` |
| Security | 安全头 | 增强安全性 | `extensions.UseSecurityHeaders(app, nil)` |
| RequestID | 请求ID | 追踪请求 | `extensions.UseRequestID(app, nil)` |
| Timeout | 超时控制 | 防止请求挂起 | `extensions.UseTimeout(app, options)` |
| 自定义 | 自由定义 | 验证、日志、计时等 | `app.Use(middleware)` |

---

## 第五章：错误处理 - 程序出错了怎么办

### 5.1 HTTP 错误

Gonet 提供了常用的 HTTP 错误函数：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 400 错误：用户提交的数据有问题
    app.MapGet("/bad-request", func(ctx abstract.ContextAbstract) error {
        return abstract.BadRequest("数据格式错误")
    })

    // 401 错误：需要登录
    app.MapGet("/unauthorized", func(ctx abstract.ContextAbstract) error {
        return abstract.Unauthorized("请先登录")
    })

    // 403 错误：没有权限
    app.MapGet("/forbidden", func(ctx abstract.ContextAbstract) error {
        return abstract.Forbidden("你没有权限访问")
    })

    // 404 错误：资源不存在
    app.MapGet("/not-found", func(ctx abstract.ContextAbstract) error {
        return abstract.NotFound("用户不存在")
    })

    // 500 错误：服务器内部错误
    app.MapGet("/internal-error", func(ctx abstract.ContextAbstract) error {
        return abstract.InternalError("服务器出错了")
    })

    // 自定义状态码错误
    app.MapGet("/custom-error", func(ctx abstract.ContextAbstract) error {
        return abstract.NewHttpException(418, "我是一个茶壶")
    })

    app.Run()
}
```

### 5.2 异常过滤器

异常过滤器可以统一处理错误，返回格式化的响应：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

// 定义一个异常过滤器
type MyExceptionFilter struct{}

// Catch 方法处理错误
func (f *MyExceptionFilter) Catch(ctx abstract.ContextAbstract, err error) error {
    // 判断是否是 HTTP 错误
    if httpErr, ok := err.(*abstract.HttpException); ok {
        // 返回 JSON 格式的错误响应
        return ctx.JSON(httpErr.Status(), map[string]string{
            "error":  httpErr.Message(),
            "status": "error",
        })
    }
    
    // 其他错误返回 500
    return ctx.JSON(500, map[string]string{
        "error":  err.Error(),
        "status": "error",
    })
}

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 添加全局异常过滤器
    app.UseGlobalFilters(&MyExceptionFilter{})

    // 这个路由会返回错误，过滤器会处理
    app.MapGet("/user/:id", func(ctx abstract.ContextAbstract) error {
        id := ctx.Param("id")
        
        // 模拟：如果 id 不是数字，返回错误
        if id == "0" {
            return abstract.NotFound("用户不存在")
        }
        
        return ctx.JSON(200, map[string]string{
            "id":   id,
            "name": "用户" + id,
        })
    })
    
    app.Run()
}
```

访问 `/user/0` 返回：
```json
{"error":"用户不存在","status":"error"}
```

### 5.3 实际应用示例：用户注册

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/middleware/recovery"
)

type RegisterRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Email    string `json:"email"`
}

func main() {
    app := core.CreateApplication()
    
    // 添加 Recovery 中间件
    app.Use(recovery.New(nil))
    
    app.POST("/register", func(ctx abstract.ContextAbstract) error {
        var req RegisterRequest
        
        // 解析请求体
        if err := ctx.Bind(&req); err != nil {
            return abstract.BadRequest("请求数据格式错误")
        }
        
        // 验证必填字段
        if req.Username == "" {
            return abstract.BadRequest("用户名不能为空")
        }
        if req.Password == "" {
            return abstract.BadRequest("密码不能为空")
        }
        if len(req.Password) < 6 {
            return abstract.BadRequest("密码长度至少6位")
        }
        
        // 模拟：检查用户名是否已存在
        if req.Username == "admin" {
            return abstract.BadRequest("用户名已存在")
        }
        
        // 注册成功
        return ctx.JSON(201, map[string]string{
            "message": "注册成功",
            "username": req.Username,
        })
    })
    
    app.Run()
}
```

### 5.4 小结

| 错误类型 | 状态码 | 使用场景 |
|----------|--------|----------|
| BadRequest | 400 | 用户提交的数据有问题 |
| Unauthorized | 401 | 用户没有登录 |
| Forbidden | 403 | 用户没有权限 |
| NotFound | 404 | 资源不存在 |
| InternalError | 500 | 服务器内部错误 |

---

## 第六章：Guard、Interceptor、Pipe - 进阶功能

这三个概念来自 ASP.NET Core 和 NestJS，用于更精细地控制请求处理流程。

### 6.1 Guard（守卫）- 决定是否允许访问

Guard 用于判断请求是否可以继续执行。如果 Guard 返回 `false`，请求会被拒绝。

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

// 定义一个 Guard：检查用户是否登录
type AuthGuard struct{}

// CanActivate 方法决定是否允许访问
func (g *AuthGuard) CanActivate(ctx abstract.ContextAbstract) bool {
    token := ctx.Header("Authorization")
    return token != "" && token == "valid-token"
}

// 用函数快速创建 Guard
func RoleGuard(requiredRole string) abstract.GuardAbstract {
    return abstract.GuardFuncAbstract(func(ctx abstract.ContextAbstract) bool {
        // 从 Context 中获取用户角色（假设登录中间件已经设置）
        role, _ := ctx.Get("role").(string)
        return role == requiredRole
    })
}

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 全局 Guard：所有路由都要检查
    app.UseGlobalGuards(&AuthGuard{})

    // 单个路由的 Guard
    app.MapGet("/admin", func(ctx abstract.ContextAbstract) error {
        return ctx.String(200, "管理员页面")
    }).Guard(RoleGuard("admin"))

    app.MapGet("/user", func(ctx abstract.ContextAbstract) error {
        return ctx.String(200, "用户页面")
    })

    app.Run()
}
```

### 6.2 Interceptor（拦截器）- 在处理前后执行

Interceptor 可以在处理函数执行前后做事情，甚至修改返回结果。

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

// 定义一个 Interceptor：记录处理时间
type TimingInterceptor struct{}

// Intercept 方法在处理前后执行
func (i *TimingInterceptor) Intercept(ctx abstract.ContextAbstract, next abstract.RouteHandlerAbstract) (any, error) {
    // 前置处理：可以记录开始时间
    ctx.Set("start_time", "记录开始")
    
    // 执行处理函数
    err := next(ctx)
    
    // 后置处理：可以计算耗时
    // elapsed := time.Since(startTime)
    
    // 可以修改返回结果
    // 这里返回 nil，表示不修改
    return nil, err
}

// 用函数快速创建 Interceptor
func LoggingInterceptor() abstract.InterceptorAbstract {
    return abstract.InterceptorFuncAbstract(func(ctx abstract.ContextAbstract, next abstract.RouteHandlerAbstract) (any, error) {
        // 记录请求信息
        println("请求:", ctx.Method(), ctx.Path())
        
        // 执行处理
        err := next(ctx)
        
        // 记录结果
        if err != nil {
            println("错误:", err.Error())
        } else {
            println("成功")
        }
        
        return nil, err
    })
}

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 全局 Interceptor
    app.UseGlobalInterceptors(LoggingInterceptor())

    app.MapGet("/test", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"message": "test"})
    })

    app.Run()
}
```

### 6.3 Pipe（管道）- 转换数据

Pipe 用于转换输入数据，比如验证、格式化。

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

// 定义一个 Pipe：验证数据
type ValidationPipe struct{}

// Transform 方法转换数据
func (p *ValidationPipe) Transform(value any, ctx abstract.ContextAbstract) (any, error) {
    // 这里可以对 value 进行验证或转换
    // 如果验证失败，返回错误
    return value, nil
}

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 全局 Pipe
    app.UseGlobalPipes(&ValidationPipe{})

    app.MapPost("/data", func(ctx abstract.ContextAbstract) error {
        var data map[string]any
        ctx.Bind(&data)
        return ctx.JSON(200, data)
    })

    app.Run()
}
```

### 6.4 三者的区别和关系

| 功能 | 作用时机 | 主要用途 |
|------|----------|----------|
| Guard | 处理前 | 判断是否允许访问 |
| Interceptor | 处理前后 | 记录日志、修改结果 |
| Pipe | 处理前 | 验证、转换数据 |

执行顺序：
```
请求 → 中间件 → Guard → Pipe → Interceptor(前) → 处理函数 → Interceptor(后) → 响应
```

---

## 第七章：Controller 模式 - 组织你的代码

### 7.1 什么是 Controller？

当应用变得复杂，把所有路由都写在 `main.go` 里会很乱。Controller 模式把相关路由组织在一起：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

// UserController：管理所有用户相关的路由
type UserController struct{}

// Routes 方法定义这个 Controller 的所有路由
func (c *UserController) Routes(r abstract.RouterAbstract) {
    r.GET("/users", c.List)           // 用户列表
    r.GET("/users/:id", c.Get)        // 用户详情
    r.POST("/users", c.Create)        // 创建用户
    r.PUT("/users/:id", c.Update)     // 更新用户
    r.DELETE("/users/:id", c.Delete)  // 删除用户
}

// 每个处理方法
func (c *UserController) List(ctx abstract.ContextAbstract) error {
    users := []map[string]string{
        {"id": "1", "name": "张三"},
        {"id": "2", "name": "李四"},
    }
    return ctx.JSON(200, users)
}

func (c *UserController) Get(ctx abstract.ContextAbstract) error {
    id := ctx.Param("id")
    return ctx.JSON(200, map[string]string{"id": id, "name": "用户" + id})
}

func (c *UserController) Create(ctx abstract.ContextAbstract) error {
    var user struct {
        Name string `json:"name"`
    }
    ctx.Bind(&user)
    return ctx.JSON(201, map[string]string{"id": "3", "name": user.Name})
}

func (c *UserController) Update(ctx abstract.ContextAbstract) error {
    id := ctx.Param("id")
    return ctx.JSON(200, map[string]string{"id": id, "message": "更新成功"})
}

func (c *UserController) Delete(ctx abstract.ContextAbstract) error {
    id := ctx.Param("id")
    return ctx.JSON(200, map[string]string{"id": id, "message": "删除成功"})
}

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 注册 Controller
    app.Controller(&UserController{})

    app.Run()
}
```

### 7.2 多个 Controller

可以创建多个 Controller 来组织不同的功能：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

// 用户 Controller
type UserController struct{}

func (c *UserController) Routes(r abstract.RouterAbstract) {
    r.GET("/users", c.List)
    r.GET("/users/:id", c.Get)
}

func (c *UserController) List(ctx abstract.ContextAbstract) error {
    return ctx.JSON(200, []map[string]string{{"id": "1", "name": "张三"}})
}

func (c *UserController) Get(ctx abstract.ContextAbstract) error {
    return ctx.JSON(200, map[string]string{"id": ctx.Param("id")})
}

// 文章 Controller
type PostController struct{}

func (c *PostController) Routes(r abstract.RouterAbstract) {
    r.GET("/posts", c.List)
    r.GET("/posts/:id", c.Get)
}

func (c *PostController) List(ctx abstract.ContextAbstract) error {
    return ctx.JSON(200, []map[string]string{{"id": "1", "title": "文章1"}})
}

func (c *PostController) Get(ctx abstract.ContextAbstract) error {
    return ctx.JSON(200, map[string]string{"id": ctx.Param("id")})
}

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 注册多个 Controller
    app.Controller(&UserController{})
    app.Controller(&PostController{})

    app.Run()
}
```

### 7.3 Controller 与路由组

Controller 可以和路由组配合，添加统一前缀：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

type APIController struct{}

func (c *APIController) Routes(r abstract.RouterAbstract) {
    // 创建路由组
    v1 := r.Group("/api/v1")
    
    // 所有路由都有 /api/v1 前缀
    v1.GET("/users", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"version": "v1"})
    })
}

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()
    app.Controller(&APIController{})
    app.Run()
}
```

---

## 第八章：依赖注入 - 管理你的服务

### 8.1 什么是依赖注入？

依赖注入（DI）是一种管理服务的方式。假设你需要一个数据库连接：

**不使用 DI**：
```go
// 每个函数都要自己创建数据库连接
func GetUser(ctx abstract.ContextAbstract) error {
    db := createDatabaseConnection()  // 重复创建
    user := db.FindUser(123)
    return ctx.JSON(200, user)
}
```

**使用 DI**：
```go
// 在应用启动时注册服务
// 处理函数直接获取已创建好的服务
func GetUser(ctx abstract.ContextAbstract) error {
    db := GetService[Database](services)  // 获取已注册的服务
    user := db.FindUser(123)
    return ctx.JSON(200, user)
}
```

### 8.2 三种生命周期

| 生命周期 | 说明 |
|----------|------|
| Singleton | 单例，整个应用只有一个实例 |
| Scoped | 每次请求创建一个实例 |
| Transient | 每次获取都创建新实例 |

### 8.3 注册和获取服务

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

// 定义一个服务接口
type Database interface {
    FindUser(id string) map[string]string
}

// 实现服务
type MemoryDatabase struct {
    users map[string]map[string]string
}

func NewMemoryDatabase() *MemoryDatabase {
    return &MemoryDatabase{
        users: map[string]map[string]string{
            "1": {"id": "1", "name": "张三"},
            "2": {"id": "2", "name": "李四"},
        },
    }
}

func (db *MemoryDatabase) FindUser(id string) map[string]string {
    return db.users[id]
}

func main() {
    // 使用 Builder 模式（ASP.NET Core 风格）
    builder := core.CreateBuilder()

    // 注册 Singleton 服务（单例）- 使用 Services() 方法
    builder.Services().AddSingleton(NewMemoryDatabase())

    // 构建应用
    app := builder.Build()

    // 使用服务 - 使用 Services() 方法
    app.MapGet("/users/:id", func(ctx abstract.ContextAbstract) error {
        // 获取服务
        db := core.GetService[Database](app.Services())

        // 使用服务
        user := db.FindUser(ctx.Param("id"))

        if user == nil {
            return abstract.NotFound("用户不存在")
        }

        return ctx.JSON(200, user)
    })

    app.Run()
}
```

### 8.4 使用工厂函数注册

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
)

type Logger interface {
    Log(message string)
}

type ConsoleLogger struct{}

func (l *ConsoleLogger) Log(message string) {
    println(message)
}

func main() {
    builder := core.CreateBuilder()

    // 使用工厂函数注册（每次获取都调用工厂）- 使用 Services() 方法
    builder.Services().AddTransient(func(s abstract.ServiceCollectionAbstract) Logger {
        return &ConsoleLogger{}
    })

    app := builder.Build()

    app.MapGet("/log", func(ctx abstract.ContextAbstract) error {
        // 每次调用都创建新的 Logger
        logger := core.GetService[Logger](app.Services())
        logger.Log("测试日志")
        return ctx.String(200, "OK")
    })

    app.Run()
}
```

---

## 第九章：配置系统 - 让应用更灵活

### 9.1 为什么需要配置？

不同环境需要不同配置：
- 开发环境：本地数据库、调试日志
- 生产环境：远程数据库、生产日志

配置系统让应用能读取配置文件或环境变量。

### 9.2 创建配置文件

**config.yaml**：
```yaml
server:
  port: "8080"
  name: "myapp"
  timeout: 30

database:
  host: "localhost"
  port: 5432
  name: "mydb"

debug: true
```

**config.json**：
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
  },
  "debug": true
}
```

### 9.3 加载配置文件

```go
package main

import (
    "github.com/linuxerlv/gonest/config"
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    // 创建配置实例
    cfg := config.NewKoanfConfig(".")

    // 加载 YAML 配置文件
    err := cfg.Load(
        config.NewFileProvider("config.yaml", config.NewYAMLParser()),
        config.NewYAMLParser(),
    )
    if err != nil {
        panic("加载配置失败: " + err.Error())
    }

    // 或加载 JSON 配置文件
    // cfg.Load(
    //     config.NewFileProvider("config.json", config.NewJSONParser()),
    //     config.NewJSONParser(),
    // )

    // 创建应用 - 使用 UseConfig 方法（ASP.NET Core 风格）
    builder := core.CreateBuilder()
    builder.UseConfig(cfg)
    app := builder.Build()

    // 读取配置
    app.MapGet("/config", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]any{
            "port":  cfg.GetString("server.port"),
            "name":  cfg.GetString("server.name"),
            "debug": cfg.GetBool("debug"),
        })
    })

    // 启动应用（会自动读取 server.port）
    app.Run()
}
```

### 9.4 结构体绑定配置

```go
package main

import (
    "github.com/linuxerlv/gonest/config"
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

// 定义配置结构体
type ServerConfig struct {
    Port    string `koanf:"port"`
    Name    string `koanf:"name"`
    Timeout int    `koanf:"timeout"`
}

type DatabaseConfig struct {
    Host string `koanf:"host"`
    Port int    `koanf:"port"`
    Name string `koanf:"name"`
}

type AppConfig struct {
    Server   ServerConfig   `koanf:"server"`
    Database DatabaseConfig `koanf:"database"`
    Debug    bool           `koanf:"debug"`
}

func main() {
    cfg := config.NewKoanfConfig(".")
    cfg.Load(
        config.NewFileProvider("config.yaml", config.NewYAMLParser()),
        config.NewYAMLParser(),
    )
    
    // 绑定到结构体
    var appCfg AppConfig
    cfg.Unmarshal("", &appCfg)  // 空字符串表示从根开始

    builder := core.CreateBuilder()
    builder.UseConfig(cfg)
    app := builder.Build()

    app.MapGet("/server-info", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, appCfg.Server)
    })

    app.MapGet("/database-info", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, appCfg.Database)
    })

    app.Run()
}
```

### 9.5 环境变量覆盖配置

环境变量可以覆盖配置文件中的值：

```go
package main

import (
    "github.com/linuxerlv/gonest/config"
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
)

func main() {
    cfg := config.NewKoanfConfig(".")
    
    // 1. 先加载配置文件
    cfg.Load(
        config.NewFileProvider("config.yaml", config.NewYAMLParser()),
        config.NewYAMLParser(),
    )
    
    // 2. 再加载环境变量（会覆盖同名配置）
    // 环境变量 APP_SERVER_PORT 会映射到 server.port
    cfg.Load(config.NewEnvProvider(config.WithEnvPrefix("APP_")), nil)

    builder := core.CreateBuilder()
    builder.UseConfig(cfg)

    // 也可以直接访问环境变量
    dbUrl := builder.Environment().Get("DATABASE_URL")

    app := builder.Build()

    app.MapGet("/config", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]any{
            "port":         cfg.GetString("server.port"),
            "database_url": dbUrl,
        })
    })

    app.Run()
}
```

设置环境变量运行：
```bash
# Linux/Mac
APP_SERVER_PORT=9000 DATABASE_URL=postgres://localhost/mydb go run main.go

# Windows PowerShell
$env:APP_SERVER_PORT="9000"
$env:DATABASE_URL="postgres://localhost/mydb"
go run main.go
```

### 9.6 多配置源优先级

```go
cfg := config.NewKoanfConfig(".")

// 优先级从低到高（后加载的覆盖前面的）

// 1. 默认值（最低优先级）
defaults := map[string]any{
    "server.port": "8080",
    "debug": false,
}
cfg.Load(config.NewMapProvider(defaults, "."), nil)

// 2. 配置文件
cfg.Load(config.NewFileProvider("config.yaml", config.NewYAMLParser()), config.NewYAMLParser())

// 3. 环境变量（最高优先级）
cfg.Load(config.NewEnvProvider(config.WithEnvPrefix("APP_")), nil)
```

### 9.7 常用配置方法

```go
// 基本类型
cfg.GetString("key")        // 字符串
cfg.GetInt("key")           // 整数
cfg.GetBool("key")          // 布尔值
cfg.GetFloat64("key")       // 浮点数
cfg.GetDuration("key")      // 时长

// 数组和 Map
cfg.GetStringSlice("key")   // 字符串数组
cfg.GetIntSlice("key")      // 整数数组
cfg.GetStringMap("key")     // map[string]any

// 检查和默认值
cfg.IsSet("key")            // 是否存在
cfg.GetDefault("key", "默认值")  // 不存在时返回默认值

// 结构体绑定
cfg.Unmarshal("key", &struct)  // 绑定到结构体
```

### 9.8 小结

| 功能 | 方法 |
|------|------|
| 加载 YAML 文件 | `cfg.Load(config.NewFileProvider("file.yaml", config.NewYAMLParser()), config.NewYAMLParser())` |
| 加载 JSON 文件 | `cfg.Load(config.NewFileProvider("file.json", config.NewJSONParser()), config.NewJSONParser())` |
| 加载环境变量 | `cfg.Load(config.NewEnvProvider(config.WithEnvPrefix("APP_")), nil)` |
| 读取配置 | `cfg.GetString("key")` |
| 绑定结构体 | `cfg.Unmarshal("key", &struct)` |
| 直接访问环境变量 | `builder.Env.Get("KEY")` |

---

## 第十章：日志系统 - 记录应用行为

### 10.1 什么是日志？

日志就像日记，记录应用运行时的信息：
- 谁访问了什么
- 发生了什么错误
- 性能数据

### 10.2 使用 Zap 日志

```go
package main

import (
    "github.com/linuxerlv/gonest/config"
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/logger"
)

func main() {
    // 创建日志配置
    logCfg := logger.DefaultConfig()
    // 开发模式：更易读的格式
    // logCfg = logger.DevelopmentConfig()
    // 生产模式：JSON 格式
    // logCfg = logger.ProductionConfig()
    
    // 创建日志实例
    log, _ := logger.NewZapLogger(logCfg)

    // 创建应用并使用日志 - 使用 UseLogger 方法（ASP.NET Core 风格）
    builder := core.CreateBuilder()
    builder.UseLogger(log)
    app := builder.Build()

    // 记录日志
    log.Info("应用启动")
    log.Info("监听端口",
        logger.String("port", "8080"),
        logger.Int("workers", 4),
    )

    app.MapGet("/test", func(ctx abstract.ContextAbstract) error {
        // 在处理函数中记录日志
        log.Info("收到请求",
            logger.String("method", ctx.Method()),
            logger.String("path", ctx.Path()),
        )

        return ctx.String(200, "OK")
    })

    app.Run()
}
```

### 10.3 日志级别

| 级别 | 说明 |
|------|------|
| Debug | 调试信息，开发时使用 |
| Info | 一般信息 |
| Warn | 警告信息 |
| Error | 错误信息 |
| Fatal | 严重错误，程序会退出 |

```go
log.Debug("调试信息")
log.Info("一般信息")
log.Warn("警告信息")
log.Error("错误信息", logger.Err(err))
log.Fatal("严重错误")  // 会退出程序
```

### 10.4 子 Logger

可以为不同模块创建子 Logger：

```go
userLog := log.WithName("user-service")
userLog.Info("用户登录", logger.String("username", "张三"))

orderLog := log.WithName("order-service")
orderLog.Info("创建订单", logger.String("orderId", "123"))
```

---

## 第十一章：任务调度 - 自动执行任务

### 11.1 什么是任务调度？

有些任务需要自动执行：
- 每天凌晨清理临时数据
- 每小时发送统计邮件
- 每分钟检查服务状态

### 11.2 CronScheduler（定时任务）

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/task"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 创建定时任务调度器
    scheduler := task.NewMemoryCronScheduler()

    // 添加定时任务：每分钟执行
    scheduler.AddIntervalJob(time.Minute, "cleanup", func(ctx context.Context) error {
        fmt.Println("执行清理任务:", time.Now())
        return nil
    })

    // 添加定时任务：每小时执行
    scheduler.AddIntervalJob(time.Hour, "stats", func(ctx context.Context) error {
        fmt.Println("发送统计邮件:", time.Now())
        return nil
    })

    // 启动调度器
    scheduler.Start()

    // 简单路由
    app.MapGet("/", func(ctx abstract.ContextAbstract) error {
        return ctx.String(200, "OK")
    })

    // 启动应用
    app.Run()

    // 程序退出时停止调度器
    scheduler.Stop(context.Background())
}
```

### 11.3 TaskQueue（后台任务队列）

用于处理耗时任务，不影响响应速度：

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/task"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 创建任务队列（5个工作线程）
    queue := task.NewMemoryTaskQueue("tasks", 5, 1000)

    // 注册任务处理器：发送邮件
    queue.RegisterHandler("send-email", func(ctx context.Context, t *task.QueueTask) error {
        // 解析任务数据
        var email struct {
            To      string `json:"to"`
            Subject string `json:"subject"`
            Body    string `json:"body"`
        }
        json.Unmarshal(t.Payload, &email)
        
        // 执行发送（这里只是打印）
        fmt.Printf("发送邮件给 %s: %s\n", email.To, email.Subject)
        
        return nil
    })
    
    // 启动队列
    queue.Start(context.Background())

    // 路由：添加发送邮件任务
    app.MapPost("/send-email", func(ctx abstract.ContextAbstract) error {
        var req struct {
            To      string `json:"to"`
            Subject string `json:"subject"`
            Body    string `json:"body"`
        }
        ctx.Bind(&req)

        // 创建任务
        payload, _ := json.Marshal(req)
        queue.Enqueue(&task.QueueTask{
            Type:    "send-email",
            Payload: payload,
        })

        // 立即返回，邮件在后台发送
        return ctx.JSON(200, map[string]string{
            "message": "邮件任务已添加",
        })
    })

    // 查看队列状态
    app.MapGet("/queue/stats", func(ctx abstract.ContextAbstract) error {
        stats := queue.Stats()
        return ctx.JSON(200, stats)
    })

    app.Run()

    // 退出时停止队列
    queue.Stop(context.Background())
}
```

### 11.4 任务选项

```go
// 设置重试次数和超时
queue.Enqueue(&task.QueueTask{Type: "send-email", Payload: payload},
    task.WithMaxRetry(3),      // 失败后最多重试3次
    task.WithTimeout(time.Minute), // 每次执行最多1分钟
)
```

---

## 第十二章：认证和授权 - 保护你的应用

### 12.1 认证和授权的区别

- **认证（Authentication）**：验证你是谁（登录）
- **授权（Authorization）**：验证你能做什么（权限）

### 12.2 JWT 认证

JWT（JSON Web Token）是一种常用的认证方式：

```go
package main

import (
    "time"
    
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/middleware/auth"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 创建 JWT 提供者
    jwtProvider := auth.NewJWTProvider(&auth.JWTConfig{
        Secret:          "my-secret-key",     // 密钥
        AccessTokenTTL:  time.Hour,           // Token 有效期
        RefreshTokenTTL: 24 * time.Hour,      // 刷新 Token 有效期
    }, nil)

    // 登录路由：发放 Token
    app.MapPost("/login", func(ctx abstract.ContextAbstract) error {
        var req struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }
        ctx.Bind(&req)

        // 验证用户名密码（这里简化）
        if req.Username != "admin" || req.Password != "123456" {
            return abstract.Unauthorized("用户名或密码错误")
        }

        // 生成 Token
        token, _ := jwtProvider.GenerateToken(&auth.Claims{
            UserID:   "1",
            Username: req.Username,
            Roles:    []string{"admin"},
        })

        return ctx.JSON(200, map[string]string{
            "token": token,
        })
    })

    // 使用认证中间件
    app.Use(auth.New(jwtProvider, nil).AsMiddleware())

    // 需要认证的路由
    app.MapGet("/profile", func(ctx abstract.ContextAbstract) error {
        // 获取用户信息
        userId := auth.GetUserID(ctx)
        username := auth.GetUsername(ctx)

        return ctx.JSON(200, map[string]string{
            "userId":   userId,
            "username": username,
        })
    })

    app.Run()
}
```

### 12.3 Basic Auth

最简单的认证方式，用户名密码在请求头中：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/middleware/auth"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 配置 Basic Auth
    app.Use(auth.NewBasicAuth(&auth.BasicAuthConfig{
        Users: map[string]string{
            "admin": "password123",
            "user":  "userpass",
        },
        Realm: "Restricted Area",
    }))

    app.MapGet("/protected", func(ctx abstract.ContextAbstract) error {
        return ctx.String(200, "认证成功")
    })

    app.Run()
}
```

测试：
```bash
curl -u admin:password123 http://localhost:8080/protected
```

### 12.4 API Key 认证

适用于 API 接口：

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/middleware/auth"
)

func main() {
    builder := core.CreateBuilder()
    app := builder.Build()

    // 配置 API Key
    app.Use(auth.NewAPIKey(&auth.APIKeyConfig{
        Keys:       []string{"key1", "key2", "key3"},
        HeaderName: "X-API-Key",
    }))

    app.MapGet("/api/data", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"data": "敏感数据"})
    })

    app.Run()
}
```

测试：
```bash
curl -H "X-API-Key: key1" http://localhost:8080/api/data
```

### 12.5 检查用户角色

```go
// 检查是否有某个角色
if auth.HasRole(ctx, "admin") {
    // 是管理员
}

// 检查是否有任意一个角色
if auth.HasAnyRole(ctx, "admin", "editor") {
    // 有 admin 或 editor 角色
}

// 检查是否有所有角色
if auth.HasAllRoles(ctx, "admin", "editor") {
    // 同时有 admin 和 editor 角色
}
```

---

## 第十三章：完整项目示例

### 13.1 项目结构

```
myapp/
├── main.go           # 入口文件
├── config.yaml       # 配置文件
├── controllers/      # 控制器
│   ├── user.go
│   └── post.go
├── services/         # 服务
│   ├── database.go
│   └── email.go
├── middleware/       # 自定义中间件
│   └── auth.go
└── models/           # 数据模型
│   ├── user.go
│   └── post.go
```

### 13.2 完整示例代码

```go
// main.go
package main

import (
    "context"
    "time"
    
    "github.com/linuxerlv/gonest/config"
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/logger"
    "github.com/linuxerlv/gonest/middleware/auth"
    "github.com/linuxerlv/gonest/middleware/cors"
    "github.com/linuxerlv/gonest/middleware/recovery"
    "github.com/linuxerlv/gonest/task"
)

// ==================== 数据模型 ====================

type User struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

type Post struct {
    ID      string `json:"id"`
    Title   string `json:"title"`
    Content string `json:"content"`
    Author  string `json:"author"`
}

// ==================== 服务 ====================

// 数据库服务（模拟）
type DatabaseService struct {
    users map[string]*User
    posts map[string]*Post
}

func NewDatabaseService() *DatabaseService {
    return &DatabaseService{
        users: map[string]*User{
            "1": {ID: "1", Username: "张三", Email: "zhang@example.com"},
            "2": {ID: "2", Username: "李四", Email: "li@example.com"},
        },
        posts: map[string]*Post{
            "1": {ID: "1", Title: "第一篇文章", Content: "内容...", Author: "1"},
            "2": {ID: "2", Title: "第二篇文章", Content: "内容...", Author: "2"},
        },
    }
}

func (db *DatabaseService) GetUser(id string) *User {
    return db.users[id]
}

func (db *DatabaseService) GetPost(id string) *Post {
    return db.posts[id]
}

func (db *DatabaseService) ListUsers() []*User {
    result := make([]*User, 0)
    for _, u := range db.users {
        result = append(result, u)
    }
    return result
}

func (db *DatabaseService) ListPosts() []*Post {
    result := make([]*Post, 0)
    for _, p := range db.posts {
        result = append(result, p)
    }
    return result
}

// ==================== 控制器 ====================

type UserController struct {
    db *DatabaseService
}

func (c *UserController) Routes(r abstract.RouterAbstract) {
    r.GET("/users", c.List)
    r.GET("/users/:id", c.Get)
    r.POST("/users", c.Create)
}

func (c *UserController) List(ctx abstract.ContextAbstract) error {
    users := c.db.ListUsers()
    return ctx.JSON(200, users)
}

func (c *UserController) Get(ctx abstract.ContextAbstract) error {
    id := ctx.Param("id")
    user := c.db.GetUser(id)
    if user == nil {
        return abstract.NotFound("用户不存在")
    }
    return ctx.JSON(200, user)
}

func (c *UserController) Create(ctx abstract.ContextAbstract) error {
    var user User
    ctx.Bind(&user)
    return ctx.JSON(201, user)
}

type PostController struct {
    db *DatabaseService
}

func (c *PostController) Routes(r abstract.RouterAbstract) {
    r.GET("/posts", c.List)
    r.GET("/posts/:id", c.Get)
}

func (c *PostController) List(ctx abstract.ContextAbstract) error {
    posts := c.db.ListPosts()
    return ctx.JSON(200, posts)
}

func (c *PostController) Get(ctx abstract.ContextAbstract) error {
    id := ctx.Param("id")
    post := c.db.GetPost(id)
    if post == nil {
        return abstract.NotFound("文章不存在")
    }
    return ctx.JSON(200, post)
}

// ==================== 异常过滤器 ====================

type ExceptionFilter struct{}

func (f *ExceptionFilter) Catch(ctx abstract.ContextAbstract, err error) error {
    if httpErr, ok := err.(*abstract.HttpException); ok {
        return ctx.JSON(httpErr.Status(), map[string]string{
            "error":   httpErr.Message(),
            "status":  "error",
            "code":    string(httpErr.Status()),
        })
    }
    return ctx.JSON(500, map[string]string{
        "error":  err.Error(),
        "status": "error",
    })
}

// ==================== 主函数 ====================

func main() {
    // 创建配置
    cfg := config.NewKoanfConfig(".")

    // 创建日志
    logCfg := logger.DevelopmentConfig()
    log, _ := logger.NewZapLogger(logCfg)

    // 使用 Builder 创建应用（ASP.NET Core 风格）
    builder := core.CreateBuilder()
    builder.UseConfig(cfg)
    builder.UseLogger(log)

    // 注册服务
    db := NewDatabaseService()
    builder.Services().AddSingleton(db)

    // 构建应用
    app := builder.Build()

    // 添加全局中间件（使用扩展方法）
    extensions.UseRecovery(app, nil)  // 恢复崩溃
    extensions.UseCORS(app, nil)      // 跨域处理

    // 添加异常过滤器
    app.UseGlobalFilters(&ExceptionFilter{})

    // 注册控制器
    app.Controller(&UserController{db: db})
    app.Controller(&PostController{db: db})

    // 创建任务调度器
    scheduler := task.NewMemoryCronScheduler()
    scheduler.AddIntervalJob(time.Hour, "cleanup", func(ctx context.Context) error {
        log.Info("执行清理任务")
        return nil
    })
    scheduler.Start()
    
    // 启动应用
    log.Info("应用启动")
    app.Run()
}
```

---

## 附录：常用代码片段

### 获取请求信息

```go
ctx.Method()              // 请求方法
ctx.Path()                // 请求路径
ctx.Header("Authorization") // 请求头
ctx.Query("name")         // Query 参数
ctx.Param("id")           // 路径参数
ctx.Body()                // 请求体（字节）
ctx.Bind(&data)           // 解析 JSON 到结构体
```

### 返回响应

```go
ctx.String(200, "文字")   // 返回文字
ctx.JSON(200, data)       // 返回 JSON
ctx.Data(200, "image/png", bytes) // 返回二进制数据
```

### HTTP 错误

```go
abstract.BadRequest("错误信息")    // 400
abstract.Unauthorized("请登录")    // 401
abstract.Forbidden("无权限")       // 403
abstract.NotFound("不存在")        // 404
abstract.InternalError("服务器错误") // 500
```

### 中间件（扩展方法）

```go
import "github.com/linuxerlv/gonest/extensions"

// 使用扩展方法添加中间件
extensions.UseRecovery(app, nil)
extensions.UseCORS(app, &extensions.CORSMiddlewareOptions{...})
extensions.UseRateLimit(app, &extensions.RateLimitMiddlewareOptions{...})
extensions.UseGzip(app, nil)
extensions.UseSecurityHeaders(app, nil)
extensions.UseRequestID(app, nil)
extensions.UseTimeout(app, &extensions.TimeoutMiddlewareOptions{...})

// 或者使用原始中间件
app.Use(middleware)
```

---

## 下一步

恭喜你完成了这个教程！现在你已经学会了 Gonest 框架的主要功能。

建议的下一步：
1. 尝试开发一个小项目（比如博客系统）
2. 阅读 [API 参考](API_REFERENCE.md) 了解更多细节
3. 探索更多内置中间件的功能

有问题可以查看框架的测试文件（`tests/` 目录），里面有更多使用示例。

---

**Happy Coding!**