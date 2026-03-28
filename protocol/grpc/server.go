package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

// ============================================================
//                    gRPC Server
// ============================================================

type Server struct {
	server   *grpc.Server
	listener net.Listener
	addr     string
	mu       sync.RWMutex
	running  bool
	services map[string]any
	opts     []grpc.ServerOption
}

func NewServer(opts ...grpc.ServerOption) *Server {
	return &Server{
		services: make(map[string]any),
		opts:     opts,
	}
}

func (s *Server) Name() string   { return "gRPC" }
func (s *Server) Scheme() string { return "grpc" }
func (s *Server) Running() bool  { return s.running }

func (s *Server) RegisterService(desc any, impl any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch d := desc.(type) {
	case *grpc.ServiceDesc:
		if s.server != nil {
			s.server.RegisterService(d, impl)
		}
		s.services[d.ServiceName] = impl
	default:
		log.Printf("[gRPC] Unknown service descriptor type: %T", desc)
	}
}

func (s *Server) Use(interceptor any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch i := interceptor.(type) {
	case grpc.UnaryServerInterceptor:
		s.opts = append(s.opts, grpc.UnaryInterceptor(i))
	case grpc.StreamServerInterceptor:
		s.opts = append(s.opts, grpc.StreamInterceptor(i))
	case []grpc.UnaryServerInterceptor:
		s.opts = append(s.opts, grpc.ChainUnaryInterceptor(i...))
	}
}

func (s *Server) UseUnary(interceptor grpc.UnaryServerInterceptor) {
	s.opts = append(s.opts, grpc.UnaryInterceptor(interceptor))
}

func (s *Server) UseStream(interceptor grpc.StreamServerInterceptor) {
	s.opts = append(s.opts, grpc.StreamInterceptor(interceptor))
}

func (s *Server) ChainUnary(interceptors ...grpc.UnaryServerInterceptor) {
	s.opts = append(s.opts, grpc.ChainUnaryInterceptor(interceptors...))
}

func (s *Server) ChainStream(interceptors ...grpc.StreamServerInterceptor) {
	s.opts = append(s.opts, grpc.ChainStreamInterceptor(interceptors...))
}

func (s *Server) Start(addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("gRPC server already running")
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer(s.opts...)
	s.listener = listener
	s.addr = addr
	s.running = true

	go func() {
		if err := s.server.Serve(listener); err != nil {
			log.Printf("[gRPC] Server error: %v", err)
		}
	}()

	log.Printf("[gRPC] Server started on %s", addr)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.running = false
		log.Printf("[gRPC] Server stopped")
		return nil
	case <-ctx.Done():
		s.server.Stop()
		s.running = false
		return fmt.Errorf("forced shutdown")
	}
}

func (s *Server) Server() *grpc.Server {
	return s.server
}

// ============================================================
//                    gRPC Context
// ============================================================

type Context struct {
	ctx      context.Context
	method   string
	service  string
	req      any
	resp     any
	metadata map[string]string
	header   map[string]string
	trailer  map[string]string
	err      error
	mu       sync.RWMutex
	values   map[string]any
}

func NewContext(ctx context.Context) *Context {
	return &Context{
		ctx:      ctx,
		metadata: make(map[string]string),
		header:   make(map[string]string),
		trailer:  make(map[string]string),
		values:   make(map[string]any),
	}
}

func (c *Context) Method() string            { return c.method }
func (c *Context) Path() string              { return c.service + "/" + c.method }
func (c *Context) Header(name string) string { return c.metadata[name] }
func (c *Context) Context() context.Context  { return c.ctx }

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

func (c *Context) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[key]
}

func (c *Context) Service() string { return c.service }
func (c *Context) Request() any    { return c.req }
func (c *Context) Response() any   { return c.resp }

func (c *Context) Metadata(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metadata[key]
}

func (c *Context) SetHeader(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.header[key] = value
}

func (c *Context) SetTrailer(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.trailer[key] = value
}

func (c *Context) Error(err error) {
	c.err = err
}

func (c *Context) SetMethod(method string) {
	c.method = method
}

func (c *Context) SetService(service string) {
	c.service = service
}

func (c *Context) SetRequest(req any) {
	c.req = req
}

func (c *Context) SetResponse(resp any) {
	c.resp = resp
}

// ============================================================
//                    Interceptors
// ============================================================

type UnaryInterceptor func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error)
type StreamInterceptor func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log.Printf("[gRPC] %s called", info.FullMethod)
		resp, err := handler(ctx, req)
		if err != nil {
			log.Printf("[gRPC] %s error: %v", info.FullMethod, err)
		}
		return resp, err
	}
}

func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[gRPC] Panic in %s: %v", info.FullMethod, r)
				err = fmt.Errorf("internal error: %v", r)
			}
		}()
		return handler(ctx, req)
	}
}

func AuthInterceptor(authFunc func(ctx context.Context) (context.Context, error)) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		newCtx, err := authFunc(ctx)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}
