package abstract

import "context"

// ContextRunnerAbstract 上下文运行器接口
type ContextRunnerAbstract interface {
	Context() context.Context
}

// ContextAbstract 完整上下文接口（Handler使用）
type ContextAbstract interface {
	ContextRunnerAbstract
	FullRequestReaderAbstract
	FullResponseWriterAbstract
	RawResponseWriterAbstract
	RequestSetterAbstract
	ValueStoreAbstract
}

// ReadOnlyContextAbstract 只读上下文接口（中间件前置处理）
type ReadOnlyContextAbstract interface {
	ContextRunnerAbstract
	RawRequestAbstract
	RequestReaderAbstract
	PathParamsReaderAbstract
	QueryReaderAbstract
	ValueStoreAbstract
}

// WriteOnlyContextAbstract 只写上下文接口（错误处理）
type WriteOnlyContextAbstract interface {
	ContextRunnerAbstract
	FullResponseWriterAbstract
	RawResponseWriterAbstract
	ValueStoreAbstract
}
