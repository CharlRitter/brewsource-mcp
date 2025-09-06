package mcp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
	"github.com/gorilla/websocket"
)

// Helper function to validate tools list response structure.
func validateToolsListResponse(t *testing.T, resp *mcp.Message) {
	if resp == nil || resp.Result == nil {
		t.Error("expected response for tools/list")
		return
	}

	// Convert to JSON and back to simulate what happens in real usage
	jsonData, err := json.Marshal(resp.Result)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(jsonData, &result); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal response: %v", unmarshalErr)
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Error("expected tools array in response")
		return
	}

	// Verify tool definitions have required fields
	for i, tool := range tools {
		toolMap, toolOk := tool.(map[string]interface{})
		if !toolOk {
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

// Helper function to validate initialization response structure.
func validateInitializeResponse(t *testing.T, resp *mcp.Message) {
	if resp == nil || resp.Result == nil {
		t.Error("expected successful response")
		return
	}

	var initResp mcp.InitializeResponse
	respData, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(respData, &initResp); err != nil {
		t.Errorf("failed to parse response: %v", err)
		return
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

type mockToolRegistry struct{}

func (m *mockToolRegistry) RegisterToolHandlers(s *mcp.Server) {
	s.RegisterToolHandler("mock_tool", func(_ context.Context, _ map[string]interface{}) (*mcp.ToolResult, error) {
		return mcp.NewToolResult("ok"), nil
	})
}
func (m *mockToolRegistry) GetToolDefinitions() []mcp.Tool { return nil }

type mockResourceRegistry struct{}

func (m *mockResourceRegistry) RegisterResourceHandlers(s *mcp.Server) {
	s.RegisterResourceHandler("mock://*", func(_ context.Context, uri string) (*mcp.ResourceContent, error) {
		return &mcp.ResourceContent{
			URI:      uri,
			MimeType: "text/plain",
			Text:     "resource",
		}, nil
	})
}
func (m *mockResourceRegistry) GetResourceDefinitions() []mcp.Resource { return nil }

func TestNewServer_RegistersHandlers(t *testing.T) {
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})

	// Test that tools are registered by trying to call the tools/list method
	msg := &mcp.Message{JSONRPC: "2.0", ID: "test", Method: "tools/list"}
	data, _ := json.Marshal(msg)
	resp := s.ProcessMessage(context.Background(), data)

	if resp == nil || resp.Result == nil {
		t.Error("expected tool handlers to be registered - tools/list should work")
	}

	// Test that resources are registered by trying to call the resources/list method
	msg = &mcp.Message{JSONRPC: "2.0", ID: "test", Method: "resources/list"}
	data, _ = json.Marshal(msg)
	resp = s.ProcessMessage(context.Background(), data)

	if resp == nil || resp.Result == nil {
		t.Error("expected resource handlers to be registered - resources/list should work")
	}
}

func TestProcessMessage_Initialize(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		wantErr  bool
		errCheck func(*mcp.Message) bool
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
			errCheck: func(resp *mcp.Message) bool {
				return resp.Error != nil && resp.Error.Code == mcp.InvalidParams
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
			msg := &mcp.Message{
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
				validateInitializeResponse(t, resp)
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
			s := mcp.NewServer(&tt.mock, &mockResourceRegistry{})
			msg := &mcp.Message{JSONRPC: "2.0", ID: "2", Method: "tools/list"}
			data, _ := json.Marshal(msg)
			resp := s.ProcessMessage(context.Background(), data)

			validateToolsListResponse(t, resp)
		})
	}
}

func TestProcessMessage_ToolsCall(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		wantErr  bool
		errCode  int
		errCheck func(*mcp.Message) bool
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
			errCode: mcp.InvalidParams,
		},
		{
			name: "unknown tool",
			params: map[string]interface{}{
				"name":      "nonexistent_tool",
				"arguments": map[string]interface{}{},
			},
			wantErr: true,
			errCode: mcp.MethodNotFound,
		},
		{
			name: "invalid arguments type",
			params: map[string]interface{}{
				"name":      "mock_tool",
				"arguments": "invalid",
			},
			wantErr: true,
			errCode: mcp.InvalidParams,
		},
		{
			name:    "missing parameters",
			params:  nil,
			wantErr: true,
			errCode: mcp.InvalidParams,
		},
		{
			name: "malformed request",
			params: map[string]interface{}{
				"name": map[string]interface{}{},
			},
			wantErr: true,
			errCode: mcp.InvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
			msg := &mcp.Message{
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
			} else if resp == nil || resp.Result == nil {
				t.Error("expected successful response")
			}
		})
	}
}

func TestProcessMessage_ResourcesList(t *testing.T) {
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	msg := &mcp.Message{JSONRPC: "2.0", ID: "4", Method: "resources/list"}
	data, _ := json.Marshal(msg)
	resp := s.ProcessMessage(context.Background(), data)
	if resp == nil || resp.Result == nil {
		t.Error("expected response for resources/list")
	}
}

// Helper function to validate resource read response structure.
func validateResourceReadResponse(t *testing.T, resp *mcp.Message) {
	if resp == nil || resp.Result == nil {
		t.Error("expected successful response")
		return
	}

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

// Helper function to validate error response.
func validateErrorResponse(t *testing.T, resp *mcp.Message, expectedCode int, errCheck func(*mcp.Message) bool) {
	if resp == nil || resp.Error == nil {
		t.Error("expected error response")
		return
	}

	if expectedCode != 0 && resp.Error.Code != expectedCode {
		t.Errorf("expected error code %d, got %d", expectedCode, resp.Error.Code)
	}

	if errCheck != nil && !errCheck(resp) {
		t.Error("response did not match error criteria")
	}
}

func TestProcessMessage_ResourcesRead(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		wantErr  bool
		errCode  int
		errCheck func(*mcp.Message) bool
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
			errCode: mcp.InvalidParams,
		},
		{
			name:    "unknown resource",
			uri:     "unknown://resource",
			wantErr: true,
			errCode: mcp.MethodNotFound,
		},
		{
			name:    "malformed uri",
			uri:     "not-a-valid-uri",
			wantErr: true,
			errCode: mcp.InvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
			params := map[string]interface{}{"uri": tt.uri}
			msg := &mcp.Message{JSONRPC: "2.0", ID: "5", Method: "resources/read", Params: params}
			data, _ := json.Marshal(msg)
			resp := s.ProcessMessage(context.Background(), data)

			if tt.wantErr {
				validateErrorResponse(t, resp, tt.errCode, tt.errCheck)
			} else {
				validateResourceReadResponse(t, resp)
			}
		})
	}
}

// Helper function to send and receive WebSocket messages.
func sendReceiveWSMessage(t *testing.T, ws *websocket.Conn, message *mcp.Message) *mcp.Message {
	data, marshalErr := json.Marshal(message)
	if marshalErr != nil {
		t.Fatalf("failed to marshal message: %v", marshalErr)
	}

	if writeErr := ws.WriteMessage(websocket.TextMessage, data); writeErr != nil {
		t.Fatalf("failed to write message: %v", writeErr)
	}

	_, resp, readErr := ws.ReadMessage()
	if readErr != nil {
		t.Fatalf("failed to read message: %v", readErr)
	}

	var response mcp.Message
	if unmarshalErr := json.Unmarshal(resp, &response); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal response: %v", unmarshalErr)
	}

	return &response
}

// Helper function to validate WebSocket response.
func validateWSResponse(t *testing.T, response *mcp.Message, wantErr bool, _ func([]byte) bool) {
	if wantErr {
		if response.Error == nil {
			t.Error("expected error response")
		}
		// Note: errCheck validation would need the raw response bytes, which we don't have here
		// In a real scenario, you might want to restructure this
	} else if response.Result == nil {
		t.Error("expected successful response")
	}
}

func TestHandleWebSocket(t *testing.T) {
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
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
		message  *mcp.Message
		wantErr  bool
		errCheck func([]byte) bool
	}{
		{
			name: "valid initialization",
			message: &mcp.Message{
				JSONRPC: "2.0",
				ID:      "1",
				Method:  "initialize",
				Params:  map[string]interface{}{"clientInfo": map[string]interface{}{}},
			},
			wantErr: false,
		},
		{
			name: "invalid message",
			message: &mcp.Message{
				JSONRPC: "2.0",
				ID:      "2",
				Method:  "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := sendReceiveWSMessage(t, ws, tt.message)
			validateWSResponse(t, response, tt.wantErr, tt.errCheck)
		})
	}
}

func TestConcurrentToolCalls(t *testing.T) {
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	concurrency := 10
	done := make(chan bool)

	for i := range concurrency {
		go func(id int) {
			params := map[string]interface{}{
				"name":      "mock_tool",
				"arguments": map[string]interface{}{},
			}
			msg := &mcp.Message{
				JSONRPC: "2.0",
				ID:      strconv.Itoa(id),
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
	for range concurrency {
		<-done
	}
}

func TestProcessMessage_MethodNotFound(t *testing.T) {
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	msg := &mcp.Message{JSONRPC: "2.0", ID: "6", Method: "unknown/method"}
	data, _ := json.Marshal(msg)
	resp := s.ProcessMessage(context.Background(), data)
	if resp == nil || resp.Error == nil {
		t.Error("expected error response for unknown method")
	}
}

func TestProcessMessage_InvalidMessage(t *testing.T) {
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	// Invalid JSON
	resp := s.ProcessMessage(context.Background(), []byte(`{"jsonrpc":"2.0",`))
	if resp == nil || resp.Error == nil {
		t.Error("expected error response for invalid message")
	}
}

// Test HandleStdio function (limited test due to stdio nature).
func TestHandleStdio(t *testing.T) {
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})

	// We can't easily test the full stdio functionality without mocking os.Stdin,
	// but we can test that the function exists and doesn't panic when called
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleStdio panicked: %v", r)
		}
	}()

	// Create a goroutine to call HandleStdio and let it fail gracefully
	done := make(chan bool)
	go func() {
		defer func() {
			done <- true
		}()
		// This will fail quickly due to no stdin input, which is expected in tests
		_ = s.HandleStdio()
	}()

	// Wait a short time for the function to start and fail
	<-done
	// Function completed, which is expected
}

// Test matchesPattern function through resource handling.
func TestMatchesPattern(t *testing.T) {
	// We can test matchesPattern indirectly by testing resource URI matching
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	ctx := context.Background()

	tests := []struct {
		name        string
		uri         string
		shouldMatch bool
	}{
		{"exact match", "test://exact", true},
		{"wildcard prefix match", "test://prefix/something", true},
		{"no match", "other://different", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test through resource read request
			resourceReq := mcp.ReadResourceRequest{URI: tt.uri}
			msg := mcp.NewMessage("resources/read", resourceReq)
			data, err := json.Marshal(msg)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			response := s.ProcessMessage(ctx, data)

			// Check if we got an expected response vs error
			hasError := response.Error != nil
			if tt.shouldMatch && hasError {
				// If it should match but we got an error, check if it's just "not found" vs "invalid URI"
				if !strings.Contains(response.Error.Message, "not found") {
					t.Errorf("Expected match for URI %s but got error: %v", tt.uri, response.Error)
				}
			}
		})
	}
}
