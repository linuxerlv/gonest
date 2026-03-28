package abstract

import "net/http"

type RouteHandler func(ctx Context) error

type RouteAdder interface {
	AddRoute(method, path string, handler RouteHandler) RouteBuilder
}

type RouteGetter interface {
	GET(path string, handler RouteHandler) RouteBuilder
	POST(path string, handler RouteHandler) RouteBuilder
	PUT(path string, handler RouteHandler) RouteBuilder
	DELETE(path string, handler RouteHandler) RouteBuilder
	PATCH(path string, handler RouteHandler) RouteBuilder
	OPTIONS(path string, handler RouteHandler) RouteBuilder
}

type GroupCreator interface {
	Group(prefix string) RouteGroup
}

type RouteMatcher interface {
	Match(req *http.Request) (Route, map[string]string)
}

type Router interface {
	RouteGetter
	GroupCreator
	RouteMatcher
}

type Route interface {
	Method() string
	Path() string
	Handler() RouteHandler
}

type RouteBuilder interface {
	Guard(guard Guard) RouteBuilder
	Interceptor(interceptor Interceptor) RouteBuilder
	Pipe(pipe Pipe) RouteBuilder
}

type RouteGroup interface {
	RouteGetter
}
