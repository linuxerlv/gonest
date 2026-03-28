package abstract

import "net/http"

type StatusWriter interface {
	Status(code int)
}

type JSONWriter interface {
	JSON(code int, v any) error
}

type StringWriter interface {
	String(code int, s string) error
}

type DataWriter interface {
	Data(code int, contentType string, data []byte) error
}

type HeaderWrittenChecker interface {
	HeaderWritten() bool
}

type RawResponseWriter interface {
	ResponseWriter() http.ResponseWriter
}

type FullResponseWriter interface {
	StatusWriter
	JSONWriter
	StringWriter
	DataWriter
	HeaderWrittenChecker
}
