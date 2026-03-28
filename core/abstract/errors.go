package abstract

import "net/http"

// HttpExceptionAbstract HTTP异常接口
type HttpExceptionAbstract interface {
	Error() string
	Status() int
	Message() string
}

// HttpErrorFactoryAbstract HTTP错误工厂接口
type HttpErrorFactoryAbstract interface {
	BadRequest(message string) error
	Unauthorized(message string) error
	Forbidden(message string) error
	NotFound(message string) error
	InternalError(message string) error
}

// HttpException HTTP异常结构体（同时满足接口）
type HttpException struct {
	Code int
	Msg  string
}

func (e *HttpException) Error() string   { return e.Msg }
func (e *HttpException) Status() int     { return e.Code }
func (e *HttpException) Message() string { return e.Msg }

// BadRequest 创建400错误
func BadRequest(message string) error {
	return &HttpException{Code: http.StatusBadRequest, Msg: message}
}

// Unauthorized 创建401错误
func Unauthorized(message string) error {
	return &HttpException{Code: http.StatusUnauthorized, Msg: message}
}

// Forbidden 创建403错误
func Forbidden(message string) error {
	return &HttpException{Code: http.StatusForbidden, Msg: message}
}

// NotFound 创建404错误
func NotFound(message string) error {
	return &HttpException{Code: http.StatusNotFound, Msg: message}
}

// InternalError 创建500错误
func InternalError(message string) error {
	return &HttpException{Code: http.StatusInternalServerError, Msg: message}
}

// NewHttpException 创建自定义HTTP异常
func NewHttpException(code int, message string) error {
	return &HttpException{Code: code, Msg: message}
}
