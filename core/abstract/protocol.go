package abstract

import "context"

// MessageTypeAbstract WebSocket消息类型
type MessageTypeAbstract int

const (
	TextMessageAbstract   MessageTypeAbstract = 1
	BinaryMessageAbstract MessageTypeAbstract = 2
	CloseMessageAbstract  MessageTypeAbstract = 8
	PingMessageAbstract   MessageTypeAbstract = 9
	PongMessageAbstract   MessageTypeAbstract = 10
)

// WSContextAbstract WebSocket上下文接口
type WSContextAbstract interface {
	ContextRunnerAbstract
	// 连接信息
	ClientID() string
	RemoteAddr() string
	// 发送消息
	SendText(msg string) error
	SendBinary(data []byte) error
	SendJSON(v any) error
	// 接收消息
	Receive() (MessageTypeAbstract, []byte, error)
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

// WSHandlerAbstract WebSocket处理函数类型
type WSHandlerAbstract func(ctx WSContextAbstract) error

// SSEContextAbstract SSE上下文接口
type SSEContextAbstract interface {
	ContextRunnerAbstract
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

// SSEHandlerAbstract SSE处理函数类型
type SSEHandlerAbstract func(ctx SSEContextAbstract) error

// GRPCContextAbstract gRPC上下文接口
type GRPCContextAbstract interface {
	ContextRunnerAbstract
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

// ProtocolAdapterAbstract 协议适配器接口
type ProtocolAdapterAbstract interface {
	Name() string
	Scheme() string
	Start(addr string) error
	Stop(ctx context.Context) error
	Running() bool
}

// HTTPServerAbstract HTTP服务器接口
type HTTPServerAbstract interface {
	ProtocolAdapterAbstract
	RouteGetterAbstract
	GroupCreatorAbstract
	Static(prefix string, root string)
	StaticFile(path string, file string)
}

// WSServerAbstract WebSocket服务器接口
type WSServerAbstract interface {
	ProtocolAdapterAbstract
	WS(path string, handler WSHandlerAbstract)
	Broadcast(msg []byte) error
	BroadcastText(msg string) error
	BroadcastTo(room string, msg []byte) error
	SendTo(clientID string, msg []byte) error
	GetRoom(room string) []WSContextAbstract
	GetClient(clientID string) WSContextAbstract
}

// SSEServerAbstract SSE服务器接口
type SSEServerAbstract interface {
	ProtocolAdapterAbstract
	SSE(path string, handler SSEHandlerAbstract)
	Broadcast(event string, data string) error
	BroadcastTo(room string, event string, data string) error
	SendTo(clientID string, event string, data string) error
	GetClient(clientID string) SSEContextAbstract
}

// GRPCServerAbstract gRPC服务器接口
type GRPCServerAbstract interface {
	ProtocolAdapterAbstract
	RegisterService(desc any, impl any)
	Use(interceptor any)
}
