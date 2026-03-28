package protocol

import (
	"net/http"

	"github.com/linuxerlv/gonest/core/abstract"
)

type Context = abstract.ContextRunner

type MessageType = abstract.MessageType

type WSContext = abstract.WSContext

type WSHandler = abstract.WSHandler

type SSEContext = abstract.SSEContext

type SSEHandler = abstract.SSEHandler

type GRPCContext = abstract.GRPCContext

type ProtocolAdapter = abstract.ProtocolAdapter

type HTTPContext interface {
	Context
	Request() *http.Request
	ResponseWriter() http.ResponseWriter
	Param(name string) string
	Query(name string) string
	Body() []byte
	Bind(v any) error

	Status(code int)
	JSON(code int, v any) error
	String(code int, s string) error
	Data(code int, contentType string, data []byte) error
	Writer() http.ResponseWriter
}

type HTTPHandler func(ctx HTTPContext) error

type HTTPServer interface {
	ProtocolAdapter
	GET(path string, handler HTTPHandler) *RouteBuilder
	POST(path string, handler HTTPHandler) *RouteBuilder
	PUT(path string, handler HTTPHandler) *RouteBuilder
	DELETE(path string, handler HTTPHandler) *RouteBuilder
	PATCH(path string, handler HTTPHandler) *RouteBuilder
	OPTIONS(path string, handler HTTPHandler) *RouteBuilder
	Group(prefix string) *RouteGroup
	Static(prefix string, root string)
	StaticFile(path string, file string)
}

type WSServer = abstract.WSServer
type SSEServer = abstract.SSEServer
type GRPCServer = abstract.GRPCServer

type RouteBuilder struct {
	middlewares  []any
	guards       []any
	interceptors []any
}

func (b *RouteBuilder) Use(middleware any) *RouteBuilder {
	b.middlewares = append(b.middlewares, middleware)
	return b
}

func (b *RouteBuilder) Guard(guard any) *RouteBuilder {
	b.guards = append(b.guards, guard)
	return b
}

func (b *RouteBuilder) Interceptor(interceptor any) *RouteBuilder {
	b.interceptors = append(b.interceptors, interceptor)
	return b
}

type RouteGroup struct {
	Prefix string
	Server HTTPServer
}

func NewRouteGroup(prefix string, server HTTPServer) *RouteGroup {
	return &RouteGroup{Prefix: prefix, Server: server}
}

func (g *RouteGroup) GET(path string, handler HTTPHandler) *RouteBuilder {
	return g.Server.GET(g.Prefix+path, handler)
}

func (g *RouteGroup) POST(path string, handler HTTPHandler) *RouteBuilder {
	return g.Server.POST(g.Prefix+path, handler)
}

func (g *RouteGroup) PUT(path string, handler HTTPHandler) *RouteBuilder {
	return g.Server.PUT(g.Prefix+path, handler)
}

func (g *RouteGroup) DELETE(path string, handler HTTPHandler) *RouteBuilder {
	return g.Server.DELETE(g.Prefix+path, handler)
}

func (g *RouteGroup) PATCH(path string, handler HTTPHandler) *RouteBuilder {
	return g.Server.PATCH(g.Prefix+path, handler)
}

func (g *RouteGroup) OPTIONS(path string, handler HTTPHandler) *RouteBuilder {
	return g.Server.OPTIONS(g.Prefix+path, handler)
}

const (
	TextMessage   MessageType = abstract.TextMessage
	BinaryMessage MessageType = abstract.BinaryMessage
	CloseMessage  MessageType = abstract.CloseMessage
	PingMessage   MessageType = abstract.PingMessage
	PongMessage   MessageType = abstract.PongMessage
)
