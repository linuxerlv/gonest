package core

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/linuxerlv/gonest/core/abstract"
)

type HttpContext struct {
	req           *http.Request
	res           http.ResponseWriter
	params        map[string]string
	query         map[string][]string
	body          []byte
	values        map[string]any
	headerWritten bool
	mu            sync.RWMutex
}

func NewContext(w http.ResponseWriter, r *http.Request) *HttpContext {
	return &HttpContext{
		req:    r,
		res:    w,
		params: make(map[string]string),
		values: make(map[string]any),
	}
}

func NewContextWithParams(w http.ResponseWriter, r *http.Request, params map[string]string) *HttpContext {
	return &HttpContext{
		req:    r,
		res:    w,
		params: params,
		values: make(map[string]any),
	}
}

func (c *HttpContext) Request() *http.Request              { return c.req }
func (c *HttpContext) ResponseWriter() http.ResponseWriter { return c.res }
func (c *HttpContext) SetRequest(r *http.Request)          { c.req = r }
func (c *HttpContext) Method() string                      { return c.req.Method }
func (c *HttpContext) Path() string                        { return c.req.URL.Path }
func (c *HttpContext) Param(name string) string            { return c.params[name] }
func (c *HttpContext) Header(name string) string           { return c.req.Header.Get(name) }

func (c *HttpContext) Query(name string) string {
	if c.query == nil {
		c.query = parseQuery(c.req.URL.RawQuery)
	}
	if values, ok := c.query[name]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

func (c *HttpContext) Body() []byte {
	if c.body == nil {
		body, _ := io.ReadAll(c.req.Body)
		c.body = body
	}
	return c.body
}

func (c *HttpContext) Bind(v any) error {
	if c.body == nil {
		body, err := io.ReadAll(c.req.Body)
		if err != nil {
			return err
		}
		c.body = body
	}
	return json.Unmarshal(c.body, v)
}

func (c *HttpContext) Status(code int) {
	if !c.headerWritten {
		c.res.WriteHeader(code)
		c.headerWritten = true
	}
}

func (c *HttpContext) JSON(code int, v any) error {
	c.res.Header().Set("Content-Type", "application/json")
	if !c.headerWritten {
		c.res.WriteHeader(code)
		c.headerWritten = true
	}
	return json.NewEncoder(c.res).Encode(v)
}

func (c *HttpContext) String(code int, s string) error {
	c.res.Header().Set("Content-Type", "text/plain")
	if !c.headerWritten {
		c.res.WriteHeader(code)
		c.headerWritten = true
	}
	_, err := c.res.Write([]byte(s))
	return err
}

func (c *HttpContext) Data(code int, contentType string, data []byte) error {
	c.res.Header().Set("Content-Type", contentType)
	if !c.headerWritten {
		c.res.WriteHeader(code)
		c.headerWritten = true
	}
	_, err := c.res.Write(data)
	return err
}

func (c *HttpContext) HeaderWritten() bool {
	return c.headerWritten
}

func (c *HttpContext) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

func (c *HttpContext) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[key]
}

func GetFromContext[T any](ctx abstract.Context, key string) T {
	v := ctx.Get(key)
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

func (c *HttpContext) Context() context.Context {
	return c.req.Context()
}

func parseQuery(rawQuery string) map[string][]string {
	result := make(map[string][]string)
	if rawQuery == "" {
		return result
	}
	for _, pair := range strings.Split(rawQuery, "&") {
		if pair == "" {
			continue
		}
		key, value := pair, ""
		if i := strings.IndexByte(pair, '='); i >= 0 {
			key, value = pair[:i], pair[i+1:]
		}
		result[key] = append(result[key], value)
	}
	return result
}

var _ abstract.Context = (*HttpContext)(nil)
