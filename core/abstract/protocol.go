package abstract

import "context"

// MessageType WebSocket消息类型
type MessageType int

const (
	TextMessage   MessageType = 1
	BinaryMessage MessageType = 2
	CloseMessage  MessageType = 8
	PingMessage   MessageType = 9
	PongMessage   MessageType = 10
)

// WSContext WebSocket上下文接口
type WSContext interface {
	ContextRunner
	// 连接信息
	ClientID() string
	RemoteAddr() string
	// 发送消息
	SendText(msg string) error
	SendBinary(data []byte) error
	SendJSON(v any) error
	// 接收消息
	Receive() (MessageType, []byte, error)
	ReceiveJSON(v any) error
	// 连接管理
	Close() error
	CloseWithStatus(code int, reason string) error
	IsClosed() bool
	// 房间
	Room() string
	JoinRoom(room string)
	LeaveRoom(room string)
}

// WSHandler WebSocket处理函数类型
type WSHandler func(ctx WSContext) error

// SSEContext SSE上下文接口
type SSEContext interface {
	ContextRunner
	// 客户端信息
	ClientID() string
	// 发送事件
	Event(event string, data string) error
	EventWithID(event string, id string, data string) error
	// 注释（用于心跳）
	Comment(comment string) error
	// 状态
	IsConnected() bool
	// 房间
	Room() string
	JoinRoom(room string)
	LeaveRoom(room string)
}

// SSEHandler SSE处理函数类型
type SSEHandler func(ctx SSEContext) error

// GRPCContext gRPC上下文接口
type GRPCContext interface {
	ContextRunner
	// 请求信息
	Method() string
	Service() string
	// 消息
	RequestMessage() any
	ResponseMessage() any
	// 元数据
	Metadata(key string) string
	SetHeader(key, value string)
	SetTrailer(key, value string)
	// 错误
	SetError(err error)
}

// ProtocolAdapter 协议适配器接口
type ProtocolAdapter interface {
	Name() string
	Scheme() string
	Start(addr string) error
	Stop(ctx context.Context) error
	Running() bool
}

// HTTPServer HTTP服务器接口
type HTTPServer interface {
	ProtocolAdapter
	RouteGetter
	GroupCreator
	Static(prefix string, root string)
	StaticFile(path string, file string)
}

// WSServer WebSocket服务器接口
type WSServer interface {
	ProtocolAdapter
	WS(path string, handler WSHandler)
	Broadcast(msg []byte) error
	BroadcastText(msg string) error
	BroadcastTo(room string, msg []byte) error
	SendTo(clientID string, msg []byte) error
	GetRoom(room string) []WSContext
	GetClient(clientID string) WSContext
}

// SSEServer SSE服务器接口
type SSEServer interface {
	ProtocolAdapter
	SSE(path string, handler SSEHandler)
	Broadcast(event string, data string) error
	BroadcastTo(room string, event string, data string) error
	SendTo(clientID string, event string, data string) error
	GetClient(clientID string) SSEContext
}

// GRPCServer gRPC服务器接口
type GRPCServer interface {
	ProtocolAdapter
	RegisterService(desc any, impl any)
	Use(interceptor any)
}
