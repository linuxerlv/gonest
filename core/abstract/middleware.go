package abstract

// MiddlewareAbstract 中间件接口
type MiddlewareAbstract interface {
	Handle(ctx ContextAbstract, next func() error) error
}

// MiddlewareFuncAbstract 中间件函数类型
type MiddlewareFuncAbstract func(ctx ContextAbstract, next func() error) error

// Handle 实现 MiddlewareAbstract 接口
func (f MiddlewareFuncAbstract) Handle(ctx ContextAbstract, next func() error) error {
	return f(ctx, next)
}

// ChainableMiddlewareAbstract 可链式组合的中间件接口
type ChainableMiddlewareAbstract interface {
	MiddlewareAbstract
	Next(ctx ContextAbstract) error
}

// ConfigurableMiddlewareAbstract 可配置的中间件接口
type ConfigurableMiddlewareAbstract[T any] interface {
	Configure(config T) MiddlewareAbstract
}
