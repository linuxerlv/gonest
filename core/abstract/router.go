package abstract

import "net/http"

type RouteHandlerAbstract func(ctx ContextAbstract) error

type RouteAdderAbstract interface {
	AddRoute(method, path string, handler RouteHandlerAbstract) RouteBuilderAbstract
}

type RouteGetterAbstract interface {
	GET(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
	POST(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
	PUT(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
	DELETE(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
	PATCH(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
	OPTIONS(path string, handler RouteHandlerAbstract) RouteBuilderAbstract
}

type GroupCreatorAbstract interface {
	Group(prefix string) RouteGroupAbstract
}

type RouteMatcherAbstract interface {
	Match(req *http.Request) (RouteAbstract, map[string]string)
}

type RouterAbstract interface {
	RouteGetterAbstract
	GroupCreatorAbstract
	RouteMatcherAbstract
}

type RouteAbstract interface {
	Method() string
	Path() string
	Handler() RouteHandlerAbstract
}

type RouteBuilderAbstract interface {
	Guard(guard GuardAbstract) RouteBuilderAbstract
	Interceptor(interceptor InterceptorAbstract) RouteBuilderAbstract
	Pipe(pipe PipeAbstract) RouteBuilderAbstract
}

type RouteGroupAbstract interface {
	RouteGetterAbstract
}
