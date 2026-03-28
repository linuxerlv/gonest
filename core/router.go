package core

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/linuxerlv/gonest/core/abstract"
)

type Route struct {
	method       string
	path         string
	handler      abstract.RouteHandler
	guards       []abstract.Guard
	interceptors []abstract.Interceptor
	pipes        []abstract.Pipe
}

func (r *Route) Method() string                 { return r.method }
func (r *Route) Path() string                   { return r.path }
func (r *Route) Handler() abstract.RouteHandler { return r.handler }

type RouteBuilder struct {
	route *Route
}

func (b *RouteBuilder) Guard(guard abstract.Guard) abstract.RouteBuilder {
	b.route.guards = append(b.route.guards, guard)
	return b
}

func (b *RouteBuilder) Interceptor(interceptor abstract.Interceptor) abstract.RouteBuilder {
	b.route.interceptors = append(b.route.interceptors, interceptor)
	return b
}

func (b *RouteBuilder) Pipe(pipe abstract.Pipe) abstract.RouteBuilder {
	b.route.pipes = append(b.route.pipes, pipe)
	return b
}

type RouteGroup struct {
	prefix string
	router *HttpRouter
}

func (g *RouteGroup) GET(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return g.router.GET(g.prefix+path, handler)
}

func (g *RouteGroup) POST(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return g.router.POST(g.prefix+path, handler)
}

func (g *RouteGroup) PUT(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return g.router.PUT(g.prefix+path, handler)
}

func (g *RouteGroup) DELETE(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return g.router.DELETE(g.prefix+path, handler)
}

func (g *RouteGroup) PATCH(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return g.router.PATCH(g.prefix+path, handler)
}

func (g *RouteGroup) OPTIONS(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return g.router.OPTIONS(g.prefix+path, handler)
}

type routeNode struct {
	children map[string]*routeNode
	routes   map[string]*Route
}

type HttpRouter struct {
	routes []*Route
	root   *routeNode
	mu     sync.RWMutex
}

func NewRouter() *HttpRouter {
	return &HttpRouter{
		routes: make([]*Route, 0),
		root:   &routeNode{children: make(map[string]*routeNode)},
	}
}

func (r *HttpRouter) GET(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return r.addRoute("GET", path, handler)
}

func (r *HttpRouter) POST(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return r.addRoute("POST", path, handler)
}

func (r *HttpRouter) PUT(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return r.addRoute("PUT", path, handler)
}

func (r *HttpRouter) DELETE(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return r.addRoute("DELETE", path, handler)
}

func (r *HttpRouter) PATCH(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return r.addRoute("PATCH", path, handler)
}

func (r *HttpRouter) OPTIONS(path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return r.addRoute("OPTIONS", path, handler)
}

func (r *HttpRouter) Group(prefix string) abstract.RouteGroup {
	return &RouteGroup{prefix: prefix, router: r}
}

func (r *HttpRouter) addRoute(method, path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	route := &Route{
		method:       method,
		path:         path,
		handler:      handler,
		guards:       make([]abstract.Guard, 0),
		interceptors: make([]abstract.Interceptor, 0),
		pipes:        make([]abstract.Pipe, 0),
	}

	r.mu.Lock()
	r.routes = append(r.routes, route)
	r.insertRoute(method, path, route)
	r.mu.Unlock()

	return &RouteBuilder{route: route}
}

func (r *HttpRouter) AddRoute(method, path string, handler abstract.RouteHandler) abstract.RouteBuilder {
	return r.addRoute(method, path, handler)
}

func (r *HttpRouter) insertRoute(method, path string, route *Route) {
	segments := splitPath(path)
	node := r.root

	for _, segment := range segments {
		key := segment
		if strings.HasPrefix(segment, ":") {
			key = ":param"
		}

		if node.children[key] == nil {
			node.children[key] = &routeNode{children: make(map[string]*routeNode), routes: make(map[string]*Route)}
		}
		node = node.children[key]
	}
	if node.routes == nil {
		node.routes = make(map[string]*Route)
	}
	node.routes[method] = route
}

func (r *HttpRouter) Match(req *http.Request) (abstract.Route, map[string]string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	segments := splitPath(req.URL.Path)
	node := r.root
	params := make(map[string]string)

	for _, segment := range segments {
		if node.children[segment] != nil {
			node = node.children[segment]
		} else if node.children[":param"] != nil {
			node = node.children[":param"]
		} else {
			return nil, nil
		}
	}

	if node.routes != nil {
		if route, ok := node.routes[req.Method]; ok {
			extractParams(segments, route.path, params)
			return route, params
		}
	}

	return nil, nil
}

func extractParams(segments []string, routePath string, params map[string]string) {
	routeSegments := splitPath(routePath)
	for i, seg := range routeSegments {
		if strings.HasPrefix(seg, ":") {
			paramName := seg[1:]
			if i < len(segments) {
				params[paramName] = segments[i]
			}
		}
	}
}

func (r *HttpRouter) ServeHTTP(w http.ResponseWriter, req *http.Request, app *Application) {
	route, params := r.Match(req)
	if route == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctx := NewContextWithParams(w, req, params)

	handler := route.Handler()

	// Guard wrapper - added FIRST so it runs AFTER middleware (innermost position)
	// This ensures middleware (like auth) can set context values before guards check them
	allGuards := append(app.guards, route.(*Route).guards...)
	if len(allGuards) > 0 {
		actualHandler := handler
		handler = func(c abstract.Context) error {
			for _, guard := range allGuards {
				if !guard.CanActivate(c) {
					return abstract.Forbidden("access denied")
				}
			}
			return actualHandler(c)
		}
	}

	// Middleware chain - wraps guard+handler
	// First middleware added (i=0) becomes outermost, runs first
	for i := len(app.middlewares) - 1; i >= 0; i-- {
		mw := app.middlewares[i]
		next := handler
		handler = func(c abstract.Context) error {
			return mw.Handle(c, func() error { return next(c) })
		}
	}

	allInterceptors := append(app.interceptors, route.(*Route).interceptors...)
	var finalHandler abstract.RouteHandler = handler
	for i := len(allInterceptors) - 1; i >= 0; i-- {
		interceptor := allInterceptors[i]
		next := finalHandler
		finalHandler = func(c abstract.Context) error {
			_, err := interceptor.Intercept(c, next)
			return err
		}
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			err := fmt.Errorf("%v", recovered)
			for _, filter := range app.filters {
				filter.Catch(ctx, err)
			}
			writeError(ctx, err)
		}
	}()

	if err := finalHandler(ctx); err != nil {
		for _, filter := range app.filters {
			filter.Catch(ctx, err)
		}
		writeError(ctx, err)
	}
}

func writeError(ctx abstract.Context, err error) {
	if ctx.HeaderWritten() {
		return
	}
	if httpErr, ok := err.(abstract.HttpException); ok {
		ctx.JSON(httpErr.Status(), map[string]interface{}{
			"code":    httpErr.Status(),
			"message": httpErr.Message(),
		})
	} else {
		ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
	}
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

var _ abstract.Router = (*HttpRouter)(nil)
var _ abstract.RouteBuilder = (*RouteBuilder)(nil)
var _ abstract.RouteGroup = (*RouteGroup)(nil)
var _ abstract.Route = (*Route)(nil)
