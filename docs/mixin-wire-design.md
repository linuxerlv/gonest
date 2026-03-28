# Mixin + Wire 代码生成方案

## 设计理念

**目标**：用户写扩展语义代码，Wire 在编译时生成原生代码，调用方看起来就是原生 Go 实现，没有额外学习成本。

## 当前实现

### 1. Mixin 定义扩展点

```go
// MiddlewareUser 接口定义了扩展点
type MiddlewareUser interface {
    UseCORS() MiddlewareUser
    UseRecovery() MiddlewareUser
    UseLogging() MiddlewareUser
    // ...
    Application() WebApplication
}

// MiddlewareMixin 实现 MiddlewareUser 接口
type MiddlewareMixin struct {
    app      *WebApplication
    services *ServiceCollection
}
```

### 2. 组合结构体

```go
// WebAppWithMixin 组合 WebApplication 和 MiddlewareMixin
type WebAppWithMixin struct {
    app   *WebApplication
    mixin *MiddlewareMixin
}

// 实现 WebApplication 接口（委托给 app）
func (w *WebAppWithMixin) Services() ServiceCollection {
    return w.app.Services()
}

// 实现 MiddlewareUser 接口（委托给 mixin）
func (w *WebAppWithMixin) UseCORS() MiddlewareUser {
    return w.mixin.UseCORS()
}
```

### 3. 用户使用

```go
// 注入配置
builder.Services().AddCORS(config)
builder.Services().AddRecovery(nil)

// 构建应用
app := core.NewWebAppWithMixin(
    builder.Build().(*core.WebApplication),
    builder.Services().(*core.ServiceCollection),
)

// 使用中间件（看起来像原生方法）
app.UseCORS().UseRecovery().UseLogging()
```

## 未来方向：Wire 代码生成

### 目标效果

用户代码：
```go
app := builder.Build()
app.UseCORS().UseRecovery()  // 看起来像原生方法
```

Wire 生成的代码：
```go
// Wire 直接将 Mixin 方法生成到 WebApplication 中
func (a *WebApplication) UseCORS() *WebApplication {
    cfg := a.services.GetService(reflect.TypeOf("cors"))
    if mw := createCORSMiddleware(cfg); mw != nil {
        a.middlewares = append(a.middlewares, mw)
    }
    return a
}
```

### 实现步骤

1. **定义 Mixin 注解**
```go
//go:generate mixin WebApplication MiddlewareMixin
```

2. **Wire 生成器解析 Mixin**
- 识别 `Mixin` 标记的结构体
- 提取 Mixin 的方法签名
- 生成目标类型的方法实现

3. **生成合并代码**
- 将 Mixin 方法直接生成到目标类型
- 无运行时开销
- 用户无感知

## 优势

1. **零运行时开销**：编译时生成，无反射
2. **类型安全**：编译时检查
3. **用户友好**：看起来像原生代码
4. **可扩展**：通过 Mixin 定义新的扩展点

## 与其他方案对比

| 方案 | 运行时开销 | 类型安全 | 用户学习成本 |
|------|-----------|---------|-------------|
| 当前 Mixin | 低 | 是 | 中等 |
| Wire 生成 | 无 | 是 | 无 |
| 反射 | 高 | 否 | 低 |
| 接口组合 | 低 | 是 | 中等 |
