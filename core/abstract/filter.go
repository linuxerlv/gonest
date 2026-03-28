package abstract

// ExceptionFilter 异常过滤器接口
type ExceptionFilter interface {
	Catch(ctx Context, err error) error
}
