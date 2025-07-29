package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

type mockToolRegistry struct{}

func (m *mockToolRegistry) RegisterToolHandlers(s *Server) {
	s.RegisterToolHandler("mock_tool", func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
		return NewToolResult("ok"), nil
	})
}
func (m *mockToolRegistry) GetToolDefinitions() []Tool { return nil }

type mockResourceRegistry struct{}

func (m *mockResourceRegistry) RegisterResourceHandlers(s *Server) {
	s.RegisterResourceHandler("mock://*", func(ctx context.Context, uri string) (*ResourceContent, error) {
		return &ResourceContent{
			URI:      uri,
			MimeType: "text/plain",
			Text:     "resource",
		}, nil
	})
}
func (m *mockResourceRegistry) GetResourceDefinitions() []Resource { return nil }

func TestNewServer_RegistersHandlers(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	if len(s.tools) == 0 {
		t.Error("expected tool handlers to be registered")
	}
	if len(s.resources) == 0 {
		t.Error("expected resource handlers to be registered")
	}
}

func TestProcessMessage_Initialize(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	msg := &Message{JSONRPC: "2.0", ID: "1", Method: "initialize", Params: map[string]interface{}{"clientInfo": map[string]interface{}{}}}
	data, _ := json.Marshal(msg)
	resp := s.processMessage(context.Background(), data)
	if resp == nil || resp.Result == nil {
		t.Error("expected response for initialize")
	}
}

func TestProcessMessage_ToolsList(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	msg := &Message{JSONRPC: "2.0", ID: "2", Method: "tools/list"}
	data, _ := json.Marshal(msg)
	resp := s.processMessage(context.Background(), data)
	if resp == nil || resp.Result == nil {
		t.Error("expected response for tools/list")
	}
}

func TestProcessMessage_ToolsCall(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	params := map[string]interface{}{"name": "mock_tool", "arguments": map[string]interface{}{}}
	msg := &Message{JSONRPC: "2.0", ID: "3", Method: "tools/call", Params: params}
	data, _ := json.Marshal(msg)
	resp := s.processMessage(context.Background(), data)
	if resp == nil || resp.Result == nil {
		t.Error("expected response for tools/call")
	}
}

func TestProcessMessage_ResourcesList(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	msg := &Message{JSONRPC: "2.0", ID: "4", Method: "resources/list"}
	data, _ := json.Marshal(msg)
	resp := s.processMessage(context.Background(), data)
	if resp == nil || resp.Result == nil {
		t.Error("expected response for resources/list")
	}
}

func TestProcessMessage_ResourcesRead(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	params := map[string]interface{}{"uri": "mock://test"}
	msg := &Message{JSONRPC: "2.0", ID: "5", Method: "resources/read", Params: params}
	data, _ := json.Marshal(msg)
	resp := s.processMessage(context.Background(), data)
	if resp == nil || resp.Result == nil {
		t.Error("expected response for resources/read")
	}
}

func TestProcessMessage_MethodNotFound(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	msg := &Message{JSONRPC: "2.0", ID: "6", Method: "unknown/method"}
	data, _ := json.Marshal(msg)
	resp := s.processMessage(context.Background(), data)
	if resp == nil || resp.Error == nil {
		t.Error("expected error response for unknown method")
	}
}

func TestProcessMessage_InvalidMessage(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	// Invalid JSON
	resp := s.processMessage(context.Background(), []byte(`{"jsonrpc":"2.0",`))
	if resp == nil || resp.Error == nil {
		t.Error("expected error response for invalid message")
	}
}
