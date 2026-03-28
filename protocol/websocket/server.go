package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/linuxerlv/gonest/protocol"
)

// ============================================================
//                    WebSocket Server
// ============================================================

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	clients   map[string]*Client
	rooms     map[string]map[string]*Client
	handlers  map[string]protocol.WSHandler
	mu        sync.RWMutex
	running   bool
	server    *http.Server
	onConnect func(ctx protocol.WSContext) error
	onMessage func(ctx protocol.WSContext, msg []byte) error
	onClose   func(ctx protocol.WSContext) error
	onError   func(ctx protocol.WSContext, err error) error
}

func NewServer() *Server {
	return &Server{
		clients:  make(map[string]*Client),
		rooms:    make(map[string]map[string]*Client),
		handlers: make(map[string]protocol.WSHandler),
	}
}

func (s *Server) Name() string   { return "WebSocket" }
func (s *Server) Scheme() string { return "ws" }
func (s *Server) Running() bool  { return s.running }

func (s *Server) OnConnect(handler func(ctx protocol.WSContext) error) {
	s.onConnect = handler
}

func (s *Server) OnMessage(handler func(ctx protocol.WSContext, msg []byte) error) {
	s.onMessage = handler
}

func (s *Server) OnClose(handler func(ctx protocol.WSContext) error) {
	s.onClose = handler
}

func (s *Server) OnError(handler func(ctx protocol.WSContext, err error) error) {
	s.onError = handler
}

func (s *Server) WS(path string, handler protocol.WSHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

func (s *Server) Start(addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("WebSocket server already running")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleWS)

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.running = true

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[WS] Server error: %v", err)
		}
	}()

	log.Printf("[WS] Server started on %s", addr)
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
	log.Printf("[WS] Server stopped")
	return nil
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	s.mu.RLock()
	handler, ok := s.handlers[path]
	s.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}

	client := NewClient(conn, r)
	s.addClient(client)
	defer s.removeClient(client)

	if s.onConnect != nil {
		s.onConnect(client)
	}

	if handler != nil {
		go handler(client)
	}

	s.readLoop(client)
}

func (s *Server) readLoop(client *Client) {
	defer client.Close()

	for {
		messageType, data, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				if s.onError != nil {
					s.onError(client, err)
				}
			}
			break
		}

		client.mu.Lock()
		client.lastMessage = data
		client.lastMessageType = protocol.MessageType(messageType)
		client.mu.Unlock()

		if s.onMessage != nil {
			s.onMessage(client, data)
		}
	}

	if s.onClose != nil {
		s.onClose(client)
	}
}

func (s *Server) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.id] = client
	log.Printf("[WS] Client connected: %s", client.id)
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

	log.Printf("[WS] Client disconnected: %s", client.id)
}

func (s *Server) GetClient(clientID string) protocol.WSContext {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[clientID]
}

func (s *Server) GetRoom(room string) []protocol.WSContext {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := s.rooms[room]
	result := make([]protocol.WSContext, 0, len(clients))
	for _, c := range clients {
		result = append(result, c)
	}
	return result
}

func (s *Server) Broadcast(msg []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		client.SendBinary(msg)
	}
	return nil
}

func (s *Server) BroadcastText(msg string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		client.SendText(msg)
	}
	return nil
}

func (s *Server) BroadcastTo(room string, msg []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := s.rooms[room]
	for _, client := range clients {
		client.SendBinary(msg)
	}
	return nil
}

func (s *Server) BroadcastTextTo(room string, msg string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := s.rooms[room]
	for _, client := range clients {
		client.SendText(msg)
	}
	return nil
}

func (s *Server) SendTo(clientID string, msg []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if client, ok := s.clients[clientID]; ok {
		return client.SendBinary(msg)
	}
	return fmt.Errorf("client %s not found", clientID)
}

func (s *Server) SendTextTo(clientID string, msg string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if client, ok := s.clients[clientID]; ok {
		return client.SendText(msg)
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
//                    WebSocket Client
// ============================================================

type Client struct {
	id              string
	conn            *websocket.Conn
	rooms           map[string]bool
	mu              sync.RWMutex
	closed          bool
	ctx             context.Context
	values          map[string]any
	remoteAddr      string
	lastMessage     []byte
	lastMessageType protocol.MessageType
}

func NewClient(conn *websocket.Conn, r *http.Request) *Client {
	return &Client{
		id:         generateClientID(),
		conn:       conn,
		rooms:      make(map[string]bool),
		values:     make(map[string]any),
		ctx:        context.Background(),
		remoteAddr: r.RemoteAddr,
	}
}

func (c *Client) ClientID() string          { return c.id }
func (c *Client) Method() string            { return "WS" }
func (c *Client) Path() string              { return "/" }
func (c *Client) Header(name string) string { return "" }
func (c *Client) Context() context.Context  { return c.ctx }
func (c *Client) RemoteAddr() string        { return c.remoteAddr }

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

func (c *Client) SendText(msg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("connection closed")
	}
	return c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (c *Client) SendBinary(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("connection closed")
	}
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Client) SendJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.SendText(string(data))
}

func (c *Client) Receive() (protocol.MessageType, []byte, error) {
	messageType, data, err := c.conn.ReadMessage()
	if err != nil {
		return 0, nil, err
	}
	return protocol.MessageType(messageType), data, nil
}

func (c *Client) ReceiveJSON(v any) error {
	_, data, err := c.Receive()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.conn.Close()
}

func (c *Client) CloseWithStatus(code int, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(code, reason))
}

func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
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

func generateClientID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
