package gzip

import (
	"compress/gzip"
	"net/http"
	"sync"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
)

type Config struct {
	Level int
}

func DefaultConfig() *Config {
	return &Config{
		Level: 6,
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz      *gzip.Writer
	status  int
	written bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if !w.written {
		w.status = status
		w.written = true
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.gz.Write(b)
}

func (w *gzipResponseWriter) Close() error {
	return w.gz.Close()
}

func (w *gzipResponseWriter) Status() int {
	return w.status
}

func (w *gzipResponseWriter) Flush() {
	if w.gz != nil {
		w.gz.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func New(cfg *Config) abstract.MiddlewareAbstract {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	level := cfg.Level
	if level < 1 || level > 9 {
		level = 6
	}

	pool := sync.Pool{
		New: func() any {
			gz, _ := gzip.NewWriterLevel(nil, level)
			return gz
		},
	}

	return abstract.MiddlewareFuncAbstract(func(ctx abstract.ContextAbstract, next func() error) error {
		if ctx.Header("Accept-Encoding") != "gzip" {
			return next()
		}

		hc := ctx.(*core.HttpContext)
		w := hc.ResponseWriter()

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gz := pool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer func() {
			gz.Close()
			pool.Put(gz)
		}()

		_ = &gzipResponseWriter{
			ResponseWriter: w,
			gz:             gz,
		}

		return next()
	})
}
