package ipc

import (
	"errors"
	"time"
)

var (
	ErrNotConnected    = errors.New("ipc: not connected")
	ErrAlreadyBound    = errors.New("ipc: already bound")
	ErrTimeout         = errors.New("ipc: operation timeout")
	ErrConnectionLost  = errors.New("ipc: connection lost")
	ErrFactoryNotFound = errors.New("ipc: factory not found")
)

type MessageType int

const (
	MessageTypeBinary MessageType = iota
	MessageTypeJSON
)

type Message struct {
	Type  MessageType
	Data  []byte
	Topic string
}

type Config struct {
	Name              string
	Port              int
	Transport         string
	Timeout           time.Duration
	BufferSize        int
	ReconnectInterval time.Duration
}

func DefaultConfig(name string) Config {
	return Config{
		Name:              name,
		Port:              0,
		Transport:         "auto",
		Timeout:           5 * time.Second,
		BufferSize:        100,
		ReconnectInterval: 1 * time.Second,
	}
}

type SocketType int

const (
	SocketTypePair SocketType = iota
	SocketTypePub
	SocketTypeSub
	SocketTypeReq
	SocketTypeRep
	SocketTypePush
	SocketTypePull
)

type Endpoint interface {
	Bind() error
	Connect() error
	Send(msg *Message) error
	SendTimeout(msg *Message, timeout time.Duration) error
	Recv() (*Message, error)
	RecvTimeout(timeout time.Duration) (*Message, error)
	SendBytes(data []byte) error
	RecvBytes() ([]byte, error)
	Close() error
	Address() string
	IsConnected() bool
}

type Publisher interface {
	Endpoint
	Publish(topic string, data []byte) error
	PublishJSON(topic string, v any) error
}

type Subscriber interface {
	Endpoint
	Subscribe(topic string) error
	Unsubscribe(topic string) error
}

type Requester interface {
	Endpoint
	Request(data []byte) ([]byte, error)
	RequestTimeout(data []byte, timeout time.Duration) ([]byte, error)
}

type Replier interface {
	Endpoint
	RecvRequest() ([]byte, error)
	SendReply(data []byte) error
}

type Factory interface {
	NewPair(config Config) (Endpoint, error)
	NewPubSub(config Config) (Publisher, Subscriber, error)
	NewReqRep(config Config) (Requester, Replier, error)
	NewPushPull(config Config) (Endpoint, Endpoint, error)
	Name() string
}

type RequestMessage struct {
	ID        string            `json:"id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Query     map[string]string `json:"query"`
	Body      []byte            `json:"body"`
	Timestamp int64             `json:"timestamp"`
}

type ResponseMessage struct {
	ID        string            `json:"id"`
	Status    int               `json:"status"`
	Headers   map[string]string `json:"headers"`
	Body      []byte            `json:"body"`
	Error     string            `json:"error,omitempty"`
	Timestamp int64             `json:"timestamp"`
}

type EventMessage struct {
	Topic     string `json:"topic"`
	Data      []byte `json:"data"`
	Timestamp int64  `json:"timestamp"`
}
