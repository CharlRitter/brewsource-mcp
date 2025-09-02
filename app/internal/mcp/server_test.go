package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
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
	tests := []struct {
		name     string
		params   map[string]interface{}
		wantErr  bool
		errCheck func(*Message) bool
	}{
		{
			name: "happy path - valid initialization",
			params: map[string]interface{}{
				"clientInfo": map[string]interface{}{
					"name":    "TestClient",
					"version": "1.0.0",
				},
			},
			wantErr: false,
		},
		{
			name:    "missing client info",
			params:  nil,
			wantErr: false, // Should still initialize with default values
		},
		{
			name: "malformed client info",
			params: map[string]interface{}{
				"clientInfo": "invalid",
			},
			wantErr: true,
			errCheck: func(resp *Message) bool {
				return resp.Error != nil && resp.Error.Code == InvalidParams
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
			msg := &Message{
				JSONRPC: "2.0",
				ID:      "1",
				Method:  "initialize",
				Params:  tt.params,
			}
			data, _ := json.Marshal(msg)
			resp := s.ProcessMessage(context.Background(), data)

			if tt.wantErr {
				if resp == nil || resp.Error == nil {
					t.Error("expected error response")
				}
				if tt.errCheck != nil && !tt.errCheck(resp) {
					t.Error("response did not match error criteria")
				}
			} else {
				if resp == nil || resp.Result == nil {
					t.Error("expected successful response")
				}

				// Verify response structure for successful cases
				if resp != nil && resp.Result != nil {
					var initResp InitializeResponse
					respData, _ := json.Marshal(resp.Result)
					if err := json.Unmarshal(respData, &initResp); err != nil {
						t.Errorf("failed to parse response: %v", err)
					}

					// Verify required fields
					if initResp.ProtocolVersion == "" {
						t.Error("missing protocol version")
					}
					if initResp.ServerInfo.Name == "" {
						t.Error("missing server name")
					}
					if initResp.ServerInfo.Version == "" {
						t.Error("missing server version")
					}
				}
			}
		})
	}
}

func TestProcessMessage_ToolsList(t *testing.T) {
	tests := []struct {
		name string
		mock mockToolRegistry
	}{
		{
			name: "list default tools",
			mock: mockToolRegistry{},
		},
		// Could add more test cases if we had different mock implementations
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(&tt.mock, &mockResourceRegistry{})
			msg := &Message{JSONRPC: "2.0", ID: "2", Method: "tools/list"}
			data, _ := json.Marshal(msg)
			resp := s.ProcessMessage(context.Background(), data)

			if resp == nil || resp.Result == nil {
				t.Error("expected response for tools/list")
			}

			// Verify response structure - simulate JSON marshaling/unmarshaling
			if resp != nil && resp.Result != nil {
				// Convert to JSON and back to simulate what happens in real usage
				jsonData, err := json.Marshal(resp.Result)
				if err != nil {
					t.Fatalf("failed to marshal response: %v", err)
				}

				var result map[string]interface{}
				if err := json.Unmarshal(jsonData, &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				tools, ok := result["tools"].([]interface{})
				if !ok {
					t.Error("expected tools array in response")
					return
				}

				// Verify tool definitions have required fields
				for i, tool := range tools {
					toolMap, ok := tool.(map[string]interface{})
					if !ok {
						t.Errorf("tool %d: expected map type", i)
						continue
					}

					required := []string{"name", "description", "inputSchema"}
					for _, field := range required {
						if _, exists := toolMap[field]; !exists {
							t.Errorf("tool %d: missing required field %s", i, field)
						}
					}
				}
			}
		})
	}
}

func TestProcessMessage_ToolsCall(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		wantErr  bool
		errCode  int
		errCheck func(*Message) bool
	}{
		{
			name: "happy path - valid tool call",
			params: map[string]interface{}{
				"name":      "mock_tool",
				"arguments": map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "missing tool name",
			params: map[string]interface{}{
				"arguments": map[string]interface{}{},
			},
			wantErr: true,
			errCode: InvalidParams,
		},
		{
			name: "unknown tool",
			params: map[string]interface{}{
				"name":      "nonexistent_tool",
				"arguments": map[string]interface{}{},
			},
			wantErr: true,
			errCode: MethodNotFound,
		},
		{
			name: "invalid arguments type",
			params: map[string]interface{}{
				"name":      "mock_tool",
				"arguments": "invalid",
			},
			wantErr: true,
			errCode: InvalidParams,
		},
		{
			name:    "missing parameters",
			params:  nil,
			wantErr: true,
			errCode: InvalidParams,
		},
		{
			name: "malformed request",
			params: map[string]interface{}{
				"name": map[string]interface{}{},
			},
			wantErr: true,
			errCode: InvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
			msg := &Message{
				JSONRPC: "2.0",
				ID:      "3",
				Method:  "tools/call",
				Params:  tt.params,
			}
			data, _ := json.Marshal(msg)
			resp := s.ProcessMessage(context.Background(), data)

			if tt.wantErr {
				if resp == nil || resp.Error == nil {
					t.Error("expected error response")
				} else if resp.Error.Code != tt.errCode {
					t.Errorf("expected error code %d, got %d", tt.errCode, resp.Error.Code)
				}
				if tt.errCheck != nil && !tt.errCheck(resp) {
					t.Error("response did not match error criteria")
				}
			} else {
				if resp == nil || resp.Result == nil {
					t.Error("expected successful response")
				}
			}
		})
	}
}

func TestProcessMessage_ResourcesList(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	msg := &Message{JSONRPC: "2.0", ID: "4", Method: "resources/list"}
	data, _ := json.Marshal(msg)
	resp := s.ProcessMessage(context.Background(), data)
	if resp == nil || resp.Result == nil {
		t.Error("expected response for resources/list")
	}
}

func TestProcessMessage_ResourcesRead(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		wantErr  bool
		errCode  int
		errCheck func(*Message) bool
	}{
		{
			name:    "happy path - valid resource",
			uri:     "mock://test",
			wantErr: false,
		},
		{
			name:    "missing uri",
			uri:     "",
			wantErr: true,
			errCode: InvalidParams,
		},
		{
			name:    "unknown resource",
			uri:     "unknown://resource",
			wantErr: true,
			errCode: MethodNotFound,
		},
		{
			name:    "malformed uri",
			uri:     "not-a-valid-uri",
			wantErr: true,
			errCode: InvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
			params := map[string]interface{}{"uri": tt.uri}
			msg := &Message{JSONRPC: "2.0", ID: "5", Method: "resources/read", Params: params}
			data, _ := json.Marshal(msg)
			resp := s.ProcessMessage(context.Background(), data)

			if tt.wantErr {
				if resp == nil || resp.Error == nil {
					t.Error("expected error response")
				} else if tt.errCode != 0 && resp.Error.Code != tt.errCode {
					t.Errorf("expected error code %d, got %d", tt.errCode, resp.Error.Code)
				}
				if tt.errCheck != nil && !tt.errCheck(resp) {
					t.Error("response did not match error criteria")
				}
			} else {
				if resp == nil || resp.Result == nil {
					t.Error("expected successful response")
				}

				// Verify response structure
				if resp != nil && resp.Result != nil {
					result, ok := resp.Result.(map[string]interface{})
					if !ok {
						t.Error("expected map response type")
						return
					}

					contents, ok := result["contents"].([]interface{})
					if !ok {
						t.Error("expected contents array in response")
						return
					}

					if len(contents) == 0 {
						t.Error("expected at least one content item")
						return
					}

					// Check first content item
					content, ok := contents[0].(map[string]interface{})
					if !ok {
						t.Error("expected content item to be a map")
						return
					}

					required := []string{"uri", "mimeType"}
					for _, field := range required {
						if _, exists := content[field]; !exists {
							t.Errorf("content missing required field %s", field)
						}
					}
				}
			}
		})
	}
}

func TestHandleWebSocket(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	server := httptest.NewServer(http.HandlerFunc(s.HandleWebSocket))
	defer server.Close()

	// Convert http to ws
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to the test server
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open websocket connection: %v", err)
	}
	defer ws.Close()

	tests := []struct {
		name     string
		message  *Message
		wantErr  bool
		errCheck func([]byte) bool
	}{
		{
			name: "valid initialization",
			message: &Message{
				JSONRPC: "2.0",
				ID:      "1",
				Method:  "initialize",
				Params:  map[string]interface{}{"clientInfo": map[string]interface{}{}},
			},
			wantErr: false,
		},
		{
			name: "invalid message",
			message: &Message{
				JSONRPC: "2.0",
				ID:      "2",
				Method:  "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("failed to marshal message: %v", err)
			}

			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				t.Fatalf("failed to write message: %v", err)
			}

			_, resp, err := ws.ReadMessage()
			if err != nil {
				t.Fatalf("failed to read message: %v", err)
			}

			var response Message
			if err := json.Unmarshal(resp, &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if tt.wantErr {
				if response.Error == nil {
					t.Error("expected error response")
				}
				if tt.errCheck != nil && !tt.errCheck(resp) {
					t.Error("response did not match error criteria")
				}
			} else {
				if response.Result == nil {
					t.Error("expected successful response")
				}
			}
		})
	}
}

func TestConcurrentToolCalls(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	concurrency := 10
	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			params := map[string]interface{}{
				"name":      "mock_tool",
				"arguments": map[string]interface{}{},
			}
			msg := &Message{
				JSONRPC: "2.0",
				ID:      fmt.Sprintf("%d", id),
				Method:  "tools/call",
				Params:  params,
			}
			data, _ := json.Marshal(msg)
			resp := s.ProcessMessage(context.Background(), data)
			if resp == nil || resp.Result == nil {
				t.Error("failed concurrent tool call")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

func TestProcessMessage_MethodNotFound(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	msg := &Message{JSONRPC: "2.0", ID: "6", Method: "unknown/method"}
	data, _ := json.Marshal(msg)
	resp := s.ProcessMessage(context.Background(), data)
	if resp == nil || resp.Error == nil {
		t.Error("expected error response for unknown method")
	}
}

func TestProcessMessage_InvalidMessage(t *testing.T) {
	s := NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	// Invalid JSON
	resp := s.ProcessMessage(context.Background(), []byte(`{"jsonrpc":"2.0",`))
	if resp == nil || resp.Error == nil {
		t.Error("expected error response for invalid message")
	}
}
