package abstract

import "net/http"

// RequestReader 请求基本信息读取接口
type RequestReader interface {
	Method() string
	Path() string
	Header(name string) string
}

// RawRequest 原生请求对象访问接口
type RawRequest interface {
	Request() *http.Request
}

// RequestSetter 原生请求对象设置接口
type RequestSetter interface {
	SetRequest(r *http.Request)
}

// PathParamsReader 路径参数读取接口
type PathParamsReader interface {
	Param(name string) string
}

// QueryReader Query参数读取接口
type QueryReader interface {
	Query(name string) string
}

// BodyReader 请求体读取接口
type BodyReader interface {
	Body() []byte
}

// Binder 请求体绑定接口
type Binder interface {
	Bind(v any) error
}

// FullRequestReader 完整请求读取接口（组合）
type FullRequestReader interface {
	RequestReader
	RawRequest
	PathParamsReader
	QueryReader
	BodyReader
	Binder
}
