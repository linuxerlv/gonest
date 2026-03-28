package abstract

// ExceptionFilterAbstract 异常过滤器接口
type ExceptionFilterAbstract interface {
	Catch(ctx ContextAbstract, err error) error
}
