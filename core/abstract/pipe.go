package abstract

// PipeAbstract 管道接口
type PipeAbstract interface {
	Transform(value any, ctx ContextAbstract) (any, error)
}

// PipeFuncAbstract 管道函数类型
type PipeFuncAbstract func(value any, ctx ContextAbstract) (any, error)

// Transform 实现 PipeAbstract 接口
func (f PipeFuncAbstract) Transform(value any, ctx ContextAbstract) (any, error) {
	return f(value, ctx)
}
