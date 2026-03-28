package abstract

// Interceptor 拦截器接口
type Interceptor interface {
	Intercept(ctx Context, next RouteHandler) (any, error)
}

// InterceptorFunc 拦截器函数类型
type InterceptorFunc func(ctx Context, next RouteHandler) (any, error)

// Intercept 实现 Interceptor 接口
func (f InterceptorFunc) Intercept(ctx Context, next RouteHandler) (any, error) {
	return f(ctx, next)
}
