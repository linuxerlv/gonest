package sse

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/linuxerlv/gonest/protocol"
)

// ============================================================
//                    SSE Server
// ============================================================

type Server struct {
	clients map[string]*Client
	rooms   map[string]map[string]*Client
	mu      sync.RWMutex
	running bool
	server  *http.Server
	handler http.Handler
}

func NewServer() *Server {
	return &Server{
		clients: make(map[string]*Client),
		rooms:   make(map[string]map[string]*Client),
	}
}

func (s *Server) Name() string   { return "SSE" }
func (s *Server) Scheme() string { return "http" }
func (s *Server) Running() bool  { return s.running }

func (s *Server) Start(addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("SSE server already running")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleUpgrade)

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.running = true

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[SSE] Server error: %v", err)
		}
	}()

	log.Printf("[SSE] Server started on %s", addr)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	for _, client := range s.clients {
		client.Close()
	}

	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return err
		}
	}

	s.running = false
	log.Printf("[SSE] Server stopped")
	return nil
}

func (s *Server) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	client := NewClient(w, flusher)
	s.addClient(client)
	defer s.removeClient(client)

	notify := w.(http.CloseNotifier).CloseNotify()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-notify:
			return
		case <-ticker.C:
			if err := client.Comment("ping"); err != nil {
				return
			}
		case msg := <-client.send:
			if _, err := fmt.Fprint(w, msg); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (s *Server) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.id] = client
	log.Printf("[SSE] Client connected: %s", client.id)
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, client.id)

	for room := range client.rooms {
		if clients, ok := s.rooms[room]; ok {
			delete(clients, client.id)
		}
	}

	log.Printf("[SSE] Client disconnected: %s", client.id)
}

func (s *Server) GetClient(clientID string) protocol.SSEContext {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[clientID]
}

func (s *Server) GetRoomClients(room string) []protocol.SSEContext {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := s.rooms[room]
	result := make([]protocol.SSEContext, 0, len(clients))
	for _, c := range clients {
		result = append(result, c)
	}
	return result
}

func (s *Server) Broadcast(event string, data string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		client.Event(event, data)
	}
	return nil
}

func (s *Server) BroadcastTo(room string, event string, data string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := s.rooms[room]
	for _, client := range clients {
		client.Event(event, data)
	}
	return nil
}

func (s *Server) SendTo(clientID string, event string, data string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if client, ok := s.clients[clientID]; ok {
		return client.Event(event, data)
	}
	return fmt.Errorf("client %s not found", clientID)
}

func (s *Server) JoinRoom(clientID string, room string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, ok := s.clients[clientID]; ok {
		client.JoinRoom(room)
		if s.rooms[room] == nil {
			s.rooms[room] = make(map[string]*Client)
		}
		s.rooms[room][clientID] = client
	}
}

func (s *Server) LeaveRoom(clientID string, room string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, ok := s.clients[clientID]; ok {
		client.LeaveRoom(room)
	}
	if clients, ok := s.rooms[room]; ok {
		delete(clients, clientID)
	}
}

// ============================================================
//                    SSE Client
// ============================================================

type Client struct {
	id      string
	w       http.ResponseWriter
	flusher http.Flusher
	send    chan string
	rooms   map[string]bool
	mu      sync.RWMutex
	ctx     context.Context
	values  map[string]any
}

func NewClient(w http.ResponseWriter, flusher http.Flusher) *Client {
	return &Client{
		id:      generateClientID(),
		w:       w,
		flusher: flusher,
		send:    make(chan string, 100),
		rooms:   make(map[string]bool),
		values:  make(map[string]any),
		ctx:     context.Background(),
	}
}

func (c *Client) ClientID() string          { return c.id }
func (c *Client) Method() string            { return "GET" }
func (c *Client) Path() string              { return "/" }
func (c *Client) Header(name string) string { return "" }
func (c *Client) Context() context.Context  { return c.ctx }

func (c *Client) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

func (c *Client) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[key]
}

func (c *Client) Event(event string, data string) error {
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
	c.send <- msg
	return nil
}

func (c *Client) EventWithID(event string, id string, data string) error {
	msg := fmt.Sprintf("event: %s\nid: %s\ndata: %s\n\n", event, id, data)
	c.send <- msg
	return nil
}

func (c *Client) EventWithData(event string, id string, data string) error {
	return c.EventWithID(event, id, data)
}

func (c *Client) Comment(comment string) error {
	msg := fmt.Sprintf(": %s\n\n", comment)
	_, err := fmt.Fprint(c.w, msg)
	c.flusher.Flush()
	return err
}

func (c *Client) IsConnected() bool {
	select {
	case <-c.ctx.Done():
		return false
	default:
		return true
	}
}

func (c *Client) Room() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for room := range c.rooms {
		return room
	}
	return ""
}

func (c *Client) JoinRoom(room string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rooms[room] = true
}

func (c *Client) LeaveRoom(room string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.rooms, room)
}

func (c *Client) Close() {
	close(c.send)
}

func generateClientID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
