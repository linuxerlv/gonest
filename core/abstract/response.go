package abstract

import "net/http"

type StatusWriterAbstract interface {
	Status(code int)
}

type JSONWriterAbstract interface {
	JSON(code int, v any) error
}

type StringWriterAbstract interface {
	String(code int, s string) error
}

type DataWriterAbstract interface {
	Data(code int, contentType string, data []byte) error
}

type HeaderWrittenCheckerAbstract interface {
	HeaderWritten() bool
}

type RawResponseWriterAbstract interface {
	ResponseWriter() http.ResponseWriter
}

type FullResponseWriterAbstract interface {
	StatusWriterAbstract
	JSONWriterAbstract
	StringWriterAbstract
	DataWriterAbstract
	HeaderWrittenCheckerAbstract
}
