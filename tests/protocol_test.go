package tests

import (
	"testing"

	"github.com/linuxerlv/gonest/protocol"
)

func TestRouteBuilder_Use(t *testing.T) {
	builder := &protocol.RouteBuilder{}

	result := builder.Use("middleware1")

	if result != builder {
		t.Error("Expected builder to be returned")
	}
}

func TestRouteBuilder_Guard(t *testing.T) {
	builder := &protocol.RouteBuilder{}

	result := builder.Guard("guard1")

	if result != builder {
		t.Error("Expected builder to be returned")
	}
}

func TestRouteBuilder_Interceptor(t *testing.T) {
	builder := &protocol.RouteBuilder{}

	result := builder.Interceptor("interceptor1")

	if result != builder {
		t.Error("Expected builder to be returned")
	}
}

func TestNewRouteGroup(t *testing.T) {
	group := protocol.NewRouteGroup("/api", nil)

	if group == nil {
		t.Fatal("Expected group to be created")
	}

	if group.Prefix != "/api" {
		t.Errorf("Expected prefix '/api', got '%s'", group.Prefix)
	}
}

func TestMessageType_Constants(t *testing.T) {
	if protocol.TextMessage != 1 {
		t.Errorf("Expected TextMessage = 1, got %d", protocol.TextMessage)
	}

	if protocol.BinaryMessage != 2 {
		t.Errorf("Expected BinaryMessage = 2, got %d", protocol.BinaryMessage)
	}

	if protocol.CloseMessage != 8 {
		t.Errorf("Expected CloseMessage = 8, got %d", protocol.CloseMessage)
	}

	if protocol.PingMessage != 9 {
		t.Errorf("Expected PingMessage = 9, got %d", protocol.PingMessage)
	}

	if protocol.PongMessage != 10 {
		t.Errorf("Expected PongMessage = 10, got %d", protocol.PongMessage)
	}
}
