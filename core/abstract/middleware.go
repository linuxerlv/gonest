package abstract

// Middleware 中间件接口
type Middleware interface {
	Handle(ctx Context, next func() error) error
}

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(ctx Context, next func() error) error

// Handle 实现 Middleware 接口
func (f MiddlewareFunc) Handle(ctx Context, next func() error) error {
	return f(ctx, next)
}

// ChainableMiddleware 可链式组合的中间件接口
type ChainableMiddleware interface {
	Middleware
	Next(ctx Context) error
}

// ConfigurableMiddleware 可配置的中间件接口
type ConfigurableMiddleware[T any] interface {
	Configure(config T) Middleware
}
