package abstract

// Pipe 管道接口
type Pipe interface {
	Transform(value any, ctx Context) (any, error)
}

// PipeFunc 管道函数类型
type PipeFunc func(value any, ctx Context) (any, error)

// Transform 实现 Pipe 接口
func (f PipeFunc) Transform(value any, ctx Context) (any, error) {
	return f(value, ctx)
}
