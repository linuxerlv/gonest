package abstract

import "net/http"

// HttpException HTTP异常接口
type HttpException interface {
	Error() string
	Status() int
	Message() string
}

// HttpErrorFactory HTTP错误工厂接口
type HttpErrorFactory interface {
	BadRequest(message string) error
	Unauthorized(message string) error
	Forbidden(message string) error
	NotFound(message string) error
	InternalError(message string) error
}

// HttpError HTTP异常结构体（同时满足接口）
type HttpError struct {
	Code int
	Msg  string
}

func (e *HttpError) Error() string   { return e.Msg }
func (e *HttpError) Status() int     { return e.Code }
func (e *HttpError) Message() string { return e.Msg }

// BadRequest 创建400错误
func BadRequest(message string) error {
	return &HttpError{Code: http.StatusBadRequest, Msg: message}
}

// Unauthorized 创建401错误
func Unauthorized(message string) error {
	return &HttpError{Code: http.StatusUnauthorized, Msg: message}
}

// Forbidden 创建403错误
func Forbidden(message string) error {
	return &HttpError{Code: http.StatusForbidden, Msg: message}
}

// NotFound 创建404错误
func NotFound(message string) error {
	return &HttpError{Code: http.StatusNotFound, Msg: message}
}

// InternalError 创建500错误
func InternalError(message string) error {
	return &HttpError{Code: http.StatusInternalServerError, Msg: message}
}

// NewHttpException 创建自定义HTTP异常
func NewHttpException(code int, message string) error {
	return &HttpError{Code: code, Msg: message}
}
