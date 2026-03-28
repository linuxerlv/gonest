package abstract

// InterceptorAbstract 拦截器接口
type InterceptorAbstract interface {
	Intercept(ctx ContextAbstract, next RouteHandlerAbstract) (any, error)
}

// InterceptorFuncAbstract 拦截器函数类型
type InterceptorFuncAbstract func(ctx ContextAbstract, next RouteHandlerAbstract) (any, error)

// Intercept 实现 InterceptorAbstract 接口
func (f InterceptorFuncAbstract) Intercept(ctx ContextAbstract, next RouteHandlerAbstract) (any, error) {
	return f(ctx, next)
}
