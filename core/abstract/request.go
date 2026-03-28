package abstract

import "net/http"

// RequestReaderAbstract 请求基本信息读取接口
type RequestReaderAbstract interface {
	Method() string
	Path() string
	Header(name string) string
}

// RawRequestAbstract 原生请求对象访问接口
type RawRequestAbstract interface {
	Request() *http.Request
}

// RequestSetterAbstract 原生请求对象设置接口
type RequestSetterAbstract interface {
	SetRequest(r *http.Request)
}

// PathParamsReaderAbstract 路径参数读取接口
type PathParamsReaderAbstract interface {
	Param(name string) string
}

// QueryReaderAbstract Query参数读取接口
type QueryReaderAbstract interface {
	Query(name string) string
}

// BodyReaderAbstract 请求体读取接口
type BodyReaderAbstract interface {
	Body() []byte
}

// BinderAbstract 请求体绑定接口
type BinderAbstract interface {
	Bind(v any) error
}

// FullRequestReaderAbstract 完整请求读取接口（组合）
type FullRequestReaderAbstract interface {
	RequestReaderAbstract
	RawRequestAbstract
	PathParamsReaderAbstract
	QueryReaderAbstract
	BodyReaderAbstract
	BinderAbstract
}
