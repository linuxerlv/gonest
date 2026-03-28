package abstract

import "context"

// ContextRunner 上下文运行器接口
type ContextRunner interface {
	Context() context.Context
}

// Context 完整上下文接口（Handler使用）
type Context interface {
	ContextRunner
	FullRequestReader
	FullResponseWriter
	RawResponseWriter
	RequestSetter
	ValueStore
}

// ReadOnlyContext 只读上下文接口（中间件前置处理）
type ReadOnlyContext interface {
	ContextRunner
	RawRequest
	RequestReader
	PathParamsReader
	QueryReader
	ValueStore
}

// WriteOnlyContext 只写上下文接口（错误处理）
type WriteOnlyContext interface {
	ContextRunner
	FullResponseWriter
	RawResponseWriter
	ValueStore
}
