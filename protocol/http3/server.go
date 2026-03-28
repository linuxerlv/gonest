package http3

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/linuxerlv/gonest/protocol"
	"github.com/quic-go/quic-go/http3"
)

// ============================================================
//                    HTTP/3 Server
// ============================================================

type Server struct {
	addr      string
	certFile  string
	keyFile   string
	tlsConfig *tls.Config
	handler   http.Handler
	server    *http3.Server
	mu        sync.RWMutex
	running   bool
	routes    map[string]http.HandlerFunc
}

func NewServer() *Server {
	return &Server{
		routes: make(map[string]http.HandlerFunc),
	}
}

func (s *Server) Name() string   { return "HTTP/3" }
func (s *Server) Scheme() string { return "https" }
func (s *Server) Running() bool  { return s.running }

func (s *Server) SetTLSConfig(certFile, keyFile string) error {
	s.certFile = certFile
	s.keyFile = keyFile

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	s.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3"},
	}

	return nil
}

func (s *Server) SetHandler(handler http.Handler) {
	s.handler = handler
}

func (s *Server) GET(path string, handler protocol.HTTPHandler) *protocol.RouteBuilder {
	s.routes[path] = s.wrapHandler(handler)
	return &protocol.RouteBuilder{}
}

func (s *Server) POST(path string, handler protocol.HTTPHandler) *protocol.RouteBuilder {
	s.routes[path] = s.wrapHandler(handler)
	return &protocol.RouteBuilder{}
}

func (s *Server) PUT(path string, handler protocol.HTTPHandler) *protocol.RouteBuilder {
	s.routes[path] = s.wrapHandler(handler)
	return &protocol.RouteBuilder{}
}

func (s *Server) DELETE(path string, handler protocol.HTTPHandler) *protocol.RouteBuilder {
	s.routes[path] = s.wrapHandler(handler)
	return &protocol.RouteBuilder{}
}

func (s *Server) PATCH(path string, handler protocol.HTTPHandler) *protocol.RouteBuilder {
	s.routes[path] = s.wrapHandler(handler)
	return &protocol.RouteBuilder{}
}

func (s *Server) OPTIONS(path string, handler protocol.HTTPHandler) *protocol.RouteBuilder {
	s.routes[path] = s.wrapHandler(handler)
	return &protocol.RouteBuilder{}
}

func (s *Server) wrapHandler(handler protocol.HTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)
		if err := handler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) Group(prefix string) *protocol.RouteGroup {
	return protocol.NewRouteGroup(prefix, s)
}

func (s *Server) Static(prefix string, root string) {
	http.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(root))))
}

func (s *Server) StaticFile(path string, file string) {
	s.routes[path] = func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, file)
	}
}

func (s *Server) Start(addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("HTTP/3 server already running")
	}

	if s.tlsConfig == nil {
		return fmt.Errorf("TLS configuration required for HTTP/3")
	}

	mux := http.NewServeMux()
	for path, handler := range s.routes {
		mux.HandleFunc(path, handler)
	}

	if s.handler != nil {
		mux.Handle("/", s.handler)
	}

	s.server = &http3.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: s.tlsConfig,
	}

	s.running = true

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[HTTP/3] Server error: %v", err)
		}
	}()

	log.Printf("[HTTP/3] Server started on %s", addr)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.server != nil {
		if err := s.server.Close(); err != nil {
			return err
		}
	}

	s.running = false
	log.Printf("[HTTP/3] Server stopped")
	return nil
}

// ============================================================
//                    HTTP/3 Context
// ============================================================

type Context struct {
	w      http.ResponseWriter
	r      *http.Request
	params map[string]string
	values map[string]any
	mu     sync.RWMutex
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		w:      w,
		r:      r,
		params: make(map[string]string),
		values: make(map[string]any),
	}
}

func (c *Context) Method() string            { return c.r.Method }
func (c *Context) Path() string              { return c.r.URL.Path }
func (c *Context) Header(name string) string { return c.r.Header.Get(name) }
func (c *Context) Context() context.Context  { return c.r.Context() }

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

func (c *Context) Request() *http.Request              { return c.r }
func (c *Context) ResponseWriter() http.ResponseWriter { return c.w }
func (c *Context) Param(name string) string            { return c.params[name] }
func (c *Context) Writer() http.ResponseWriter         { return c.w }

func (c *Context) Query(name string) string {
	return c.r.URL.Query().Get(name)
}

func (c *Context) Body() []byte {
	return []byte{}
}

func (c *Context) Bind(v any) error {
	return nil
}

func (c *Context) Status(code int) {
	c.w.WriteHeader(code)
}

func (c *Context) JSON(code int, v any) error {
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(code)
	return nil
}

func (c *Context) String(code int, s string) error {
	c.w.Header().Set("Content-Type", "text/plain")
	c.w.WriteHeader(code)
	_, err := c.w.Write([]byte(s))
	return err
}

func (c *Context) Data(code int, contentType string, data []byte) error {
	c.w.Header().Set("Content-Type", contentType)
	c.w.WriteHeader(code)
	_, err := c.w.Write(data)
	return err
}

func (c *Context) SetParam(name, value string) {
	c.params[name] = value
}

// ============================================================
//                    QUIC Configuration
// ============================================================

type QUICConfig struct {
	MaxIdleTimeout        time.Duration
	MaxIncomingStreams    int64
	MaxIncomingUniStreams int64
	KeepAlivePeriod       time.Duration
}

func DefaultQUICConfig() *QUICConfig {
	return &QUICConfig{
		MaxIncomingStreams:    100,
		MaxIncomingUniStreams: 100,
		KeepAlivePeriod:       30 * time.Second,
	}
}
