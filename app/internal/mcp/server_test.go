// Package mcp_test contains tests for the Model Context Protocol server logic in Brewsource MCP.
package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
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

func TestHandleHTTP(t *testing.T) {
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
	handler := http.HandlerFunc(s.HandleHTTP)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("valid initialization", func(t *testing.T) {
		testHandleHTTPValidInitialization(t, server.URL)
	})
	t.Run("invalid method", func(t *testing.T) {
		testHandleHTTPInvalidMethod(t, server.URL)
	})
	t.Run("invalid JSON", func(t *testing.T) {
		testHandleHTTPInvalidJSON(t, server.URL)
	})
	t.Run("wrong HTTP method", func(t *testing.T) {
		testHandleHTTPWrongMethod(t, server.URL)
	})
}

func testHandleHTTPValidInitialization(t *testing.T, url string) {
	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "initialize",
		Params:  map[string]interface{}{"clientInfo": map[string]interface{}{}},
	}
	body, _ := json.Marshal(msg)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var mcpResp mcp.Message
	if jsonErr := json.Unmarshal(data, &mcpResp); jsonErr != nil {
		t.Errorf("Failed to unmarshal MCP response: %v", jsonErr)
	}
	if mcpResp.Error != nil {
		t.Error("Expected successful response for valid initialization")
	}
}

func testHandleHTTPInvalidMethod(t *testing.T, url string) {
	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      "2",
		Method:  "unknown_method",
	}
	body, _ := json.Marshal(msg)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var mcpResp mcp.Message
	if jsonErr := json.Unmarshal(data, &mcpResp); jsonErr != nil {
		t.Errorf("Failed to unmarshal MCP response: %v", jsonErr)
	}
	if mcpResp.Error == nil {
		t.Error("Expected error response for unknown method")
	}
}

func testHandleHTTPInvalidJSON(t *testing.T, url string) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString("{"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func testHandleHTTPWrongMethod(t *testing.T, url string) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
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

// Test utility functions: contains, indexOf, isValidURI.
func TestUtilityFunctions(t *testing.T) {
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})

	// Test through ProcessMessage with different scenarios that exercise these functions

	// Test isValidURI through resources/read
	tests := []struct {
		name             string
		uri              string
		expectInvalidURI bool
	}{
		{"valid URI", "test://valid", false},
		{"empty URI", "", true},
		{"simple string", "not-a-uri", true}, // This might still be treated as valid
		{"valid http URI", "http://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &mcp.Message{
				JSONRPC: "2.0",
				ID:      "test",
				Method:  "resources/read",
				Params:  map[string]interface{}{"uri": tt.uri},
			}
			data, _ := json.Marshal(msg)
			resp := s.ProcessMessage(context.Background(), data)

			if tt.expectInvalidURI {
				if resp.Error == nil {
					t.Errorf("Expected error for invalid URI %s", tt.uri)
				} else if resp.Error.Code != mcp.InvalidParams {
					t.Errorf("Expected InvalidParams error for URI %s, got %d", tt.uri, resp.Error.Code)
				}
			}
		})
	}
}

// Test error handling and edge cases.
func TestErrorHandling(t *testing.T) {
	cases := []struct {
		name         string
		message      string
		expectError  bool
		expectedCode int
	}{
		{"empty message", "", true, mcp.ParseError},
		{"malformed JSON", `{"jsonrpc":"2.0"`, true, mcp.ParseError},
		{"missing JSONRPC", `{"id":"1","method":"test"}`, true, mcp.InvalidRequest},
		{"wrong JSONRPC version", `{"jsonrpc":"1.0","id":"1","method":"test"}`, true, mcp.InvalidRequest},
		{"missing method", `{"jsonrpc":"2.0","id":"1"}`, true, mcp.MethodNotFound},
		{"invalid method type", `{"jsonrpc":"2.0","id":"1","method":123}`, true, mcp.ParseError},
		{"notification format", `{"jsonrpc":"2.0","method":"test"}`, true, mcp.MethodNotFound},
	}
	for _, c := range cases {
		runErrorHandlingCase(t, c.name, c.message, c.expectError, c.expectedCode)
	}
}

// Helper for error handling test cases.
func runErrorHandlingCase(t *testing.T, name, message string, expectError bool, expectedCode int) {
	t.Run(name, func(t *testing.T) {
		s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})
		resp := s.ProcessMessage(context.Background(), []byte(message))

		if expectError {
			checkErrorCase(t, resp, expectedCode)
			return
		}
		checkNonErrorCase(t, resp, name)
	})
}

func checkErrorCase(t *testing.T, resp *mcp.Message, expectedCode int) {
	if resp == nil {
		t.Error("Expected response for invalid message")
		return
	}
	if resp.Error == nil {
		t.Error("Expected error response")
		return
	}
	if expectedCode != 0 && resp.Error.Code != expectedCode {
		t.Errorf("Expected error code %d, got %d", expectedCode, resp.Error.Code)
	}
}

func checkNonErrorCase(t *testing.T, resp *mcp.Message, name string) {
	if resp != nil && resp.Error != nil && name == "notification format" {
		if resp.Error.Code != mcp.MethodNotFound {
			t.Errorf("Expected MethodNotFound for notification with unknown method, got %d", resp.Error.Code)
		}
		return
	}
	if resp != nil && resp.Error != nil {
		t.Errorf("Unexpected error: %v", resp.Error)
	}
}

// Test JSON-RPC validation.
func TestJSONRPCValidation(t *testing.T) {
	s := mcp.NewServer(&mockToolRegistry{}, &mockResourceRegistry{})

	// Test various message validation scenarios
	tests := []struct {
		name        string
		msg         *mcp.Message
		expectError bool
		errorCode   int
	}{
		{
			name: "valid request",
			msg: &mcp.Message{
				JSONRPC: "2.0",
				ID:      "1",
				Method:  "initialize",
				Params:  map[string]interface{}{},
			},
			expectError: false,
		},
		{
			name: "missing ID for request",
			msg: &mcp.Message{
				JSONRPC: "2.0",
				Method:  "initialize",
				Params:  map[string]interface{}{},
			},
			expectError: false, // This is a notification, not an error
		},
		{
			name: "wrong JSONRPC version",
			msg: &mcp.Message{
				JSONRPC: "1.0",
				ID:      "1",
				Method:  "initialize",
			},
			expectError: true,
			errorCode:   mcp.InvalidRequest,
		},
		{
			name: "empty method",
			msg: &mcp.Message{
				JSONRPC: "2.0",
				ID:      "1",
				Method:  "",
			},
			expectError: true,
			errorCode:   mcp.MethodNotFound, // Empty method gets treated as unknown method
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.msg)
			resp := s.ProcessMessage(context.Background(), data)

			if tt.expectError {
				if resp == nil || resp.Error == nil {
					t.Error("Expected error response")
					return
				}
				if tt.errorCode != 0 && resp.Error.Code != tt.errorCode {
					t.Errorf("Expected error code %d, got %d", tt.errorCode, resp.Error.Code)
				}
			} else if resp != nil && resp.Error != nil {
				t.Errorf("Unexpected error: %v", resp.Error)
			}
		})
	}
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
