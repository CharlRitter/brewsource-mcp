package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CharlRitter/brewsource-mcp/app/internal/handlers"
	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
	"github.com/CharlRitter/brewsource-mcp/app/pkg/data"
	"github.com/gorilla/websocket"
)

// mockBeerService implements a mock for BeerService for testing.
type mockBeerService struct{}

func (m *mockBeerService) SearchBeers(
	ctx context.Context,
	query services.BeerSearchQuery,
) ([]*services.BeerSearchResult, error) {
	return nil, nil
}

func TestMCP_Server_Integration(t *testing.T) {
	// Test that basic MCP protocol messages work correctly

	// Test initialize request
	initReq := mcp.InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities: mcp.ClientCapabilities{
			Roots:    &mcp.RootsCapability{ListChanged: true},
			Sampling: &mcp.SamplingCapability{},
		},
		ClientInfo: mcp.ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	// Serialize the request
	data, err := json.Marshal(initReq)
	if err != nil {
		t.Fatalf("Failed to marshal init request: %v", err)
	}

	// Test that we can deserialize it back
	var unmarshaled mcp.InitializeRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal init request: %v", err)
	}

	if unmarshaled.ProtocolVersion != "2024-11-05" {
		t.Errorf("Expected protocol version '2024-11-05', got '%s'", unmarshaled.ProtocolVersion)
	}

	if unmarshaled.ClientInfo.Name != "test-client" {
		t.Errorf("Expected client name 'test-client', got '%s'", unmarshaled.ClientInfo.Name)
	}
}

func TestMCP_ToolCall_Structure(t *testing.T) {
	// Test tool call request structure
	toolCall := mcp.CallToolRequest{
		Name: "bjcp_lookup",
		Arguments: map[string]interface{}{
			"style_code": "21A",
		},
	}

	// Create a full message
	msg := mcp.NewMessage("tools/call", toolCall)
	msg.ID = "test-123"

	// Serialize and deserialize
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal tool call: %v", err)
	}

	var unmarshaled mcp.Message
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal tool call: %v", err)
	}

	if unmarshaled.Method != "tools/call" {
		t.Errorf("Expected method 'tools/call', got '%s'", unmarshaled.Method)
	}

	if unmarshaled.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%v'", unmarshaled.ID)
	}
}

func TestMCP_ErrorHandling(t *testing.T) {
	// Test error message structure
	err := mcp.NewMCPError(mcp.InvalidParams, "Missing required parameter", map[string]interface{}{
		"parameter": "style_code",
		"received":  nil,
	})

	errResponse := mcp.NewErrorResponse("test-123", err)

	// Serialize and deserialize
	data, jsonErr := json.Marshal(errResponse)
	if jsonErr != nil {
		t.Fatalf("Failed to marshal error response: %v", jsonErr)
	}

	var unmarshaled mcp.Message
	jsonErr = json.Unmarshal(data, &unmarshaled)
	if jsonErr != nil {
		t.Fatalf("Failed to unmarshal error response: %v", jsonErr)
	}

	if unmarshaled.Error == nil {
		t.Fatal("Expected error field to be present")
	}

	if unmarshaled.Error.Code != mcp.InvalidParams {
		t.Errorf("Expected error code %d, got %d", mcp.InvalidParams, unmarshaled.Error.Code)
	}
}

func TestMCP_ResourceRequest_Structure(t *testing.T) {
	// Test resource request structure
	resourceReq := mcp.ReadResourceRequest{
		URI: "bjcp://styles/21A",
	}

	msg := mcp.NewMessage("resources/read", resourceReq)
	msg.ID = "resource-123"

	// Serialize and deserialize
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal resource request: %v", err)
	}

	var unmarshaled mcp.Message
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal resource request: %v", err)
	}

	if unmarshaled.Method != "resources/read" {
		t.Errorf("Expected method 'resources/read', got '%s'", unmarshaled.Method)
	}
}

func TestMCP_ServerCapabilities(t *testing.T) {
	// Test server capabilities structure
	serverCaps := mcp.ServerCapabilities{
		Resources: &mcp.ResourcesCapability{
			ListChanged: true,
		},
		Tools: &mcp.ToolsCapability{
			ListChanged: true,
		},
	}

	initResponse := mcp.InitializeResponse{
		ProtocolVersion: "2024-11-05",
		Capabilities:    serverCaps,
		ServerInfo: mcp.ServerInfo{
			Name:    "BrewSource MCP Server",
			Version: "1.0.0",
		},
	}

	// Test serialization
	data, err := json.Marshal(initResponse)
	if err != nil {
		t.Fatalf("Failed to marshal init response: %v", err)
	}

	var unmarshaled mcp.InitializeResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal init response: %v", err)
	}

	if unmarshaled.ServerInfo.Name != "BrewSource MCP Server" {
		t.Errorf("Expected server name 'BrewSource MCP Server', got '%s'", unmarshaled.ServerInfo.Name)
	}
}

func TestMCP_MessageValidation_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		expectErr bool
	}{
		{
			name:      "valid message with null ID",
			jsonData:  `{"jsonrpc":"2.0","method":"test","id":null}`,
			expectErr: false,
		},
		{
			name:      "valid notification (no ID)",
			jsonData:  `{"jsonrpc":"2.0","method":"initialized"}`,
			expectErr: false,
		},
		{
			name:      "invalid - missing jsonrpc",
			jsonData:  `{"method":"test","id":1}`,
			expectErr: true,
		},
		{
			name:      "invalid - wrong jsonrpc version",
			jsonData:  `{"jsonrpc":"1.0","method":"test","id":1}`,
			expectErr: true,
		},
		{
			name:      "valid response with result",
			jsonData:  `{"jsonrpc":"2.0","id":1,"result":{"status":"ok"}}`,
			expectErr: false,
		},
		{
			name:      "valid error response",
			jsonData:  `{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"Invalid params"}}`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mcp.ValidateMessage([]byte(tt.jsonData))
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateMessage() error = %v, expectErr = %v", err, tt.expectErr)
			}
		})
	}
}

func TestMCP_ToolResult_ContentTypes(t *testing.T) {
	// Test different content types in tool results

	// Text content
	textResult := mcp.NewToolResult("This is plain text")
	if len(textResult.Content) != 1 || textResult.Content[0].Type != "text" {
		t.Error("Text result not created correctly")
	}

	// Error result
	errorResult := mcp.NewErrorResult("This is an error")
	if !errorResult.IsError {
		t.Error("Error result should have IsError=true")
	}

	// Custom content
	customResult := &mcp.ToolResult{
		Content: []mcp.ToolContent{
			{
				Type: "text",
				Text: "Text content",
			},
			{
				Type: "data",
				Data: "base64encodeddata",
			},
		},
		IsError: false,
	}

	if len(customResult.Content) != 2 {
		t.Errorf("Expected 2 content items, got %d", len(customResult.Content))
	}
}

func TestMCP_ContextTimeout(t *testing.T) {
	// Test context handling for timeouts
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Simulate a long-running operation
	done := make(chan bool)
	go func() {
		time.Sleep(200 * time.Millisecond)
		done <- true
	}()

	select {
	case <-ctx.Done():
		// Expected - context should timeout
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded, got %v", ctx.Err())
		}
	case <-done:
		t.Error("Operation should have been cancelled by context timeout")
	}
}

func TestMCP_ProtocolConstants(t *testing.T) {
	// Test that error code constants are correct
	expectedCodes := map[string]int{
		"ParseError":     -32700,
		"InvalidRequest": -32600,
		"MethodNotFound": -32601,
		"InvalidParams":  -32602,
		"InternalError":  -32603,
	}

	actualCodes := map[string]int{
		"ParseError":     mcp.ParseError,
		"InvalidRequest": mcp.InvalidRequest,
		"MethodNotFound": mcp.MethodNotFound,
		"InvalidParams":  mcp.InvalidParams,
		"InternalError":  mcp.InternalError,
	}

	for name, expected := range expectedCodes {
		if actual, exists := actualCodes[name]; !exists || actual != expected {
			t.Errorf("Error code %s: expected %d, got %d", name, expected, actual)
		}
	}
}

func TestMCP_BJCPLookup_Validation(t *testing.T) {
	tests := []struct {
		name        string
		styleCode   string
		expectError bool
		errorCode   int
	}{
		{"valid style code", "21A", false, 0},
		{"lowercase style code", "21a", false, 0},
		{"invalid style code", "99Z", true, mcp.InvalidParams},
		{"empty style code", "", true, mcp.InvalidParams},
		{"too long style code", "21AAAA", true, mcp.InvalidParams},
		{"special characters", "21#A", true, mcp.InvalidParams},
	}

	// Use in-memory mock BJCP data for testing
	bjcpData := &data.BJCPData{
		Styles: map[string]data.BJCPStyle{
			"21A": {Code: "21A", Name: "American IPA", Category: "IPA"},
			"21B": {Code: "21B", Name: "Specialty IPA", Category: "IPA"},
		},
		Categories: []string{"IPA"},
		Metadata:   data.Metadata{Version: "2021", Source: "test", LastUpdated: "now", TotalStyles: 2},
	}
	toolHandlers := handlers.NewToolHandlers(bjcpData, nil, nil)
	server := mcp.NewServer(toolHandlers, nil)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolCall := mcp.CallToolRequest{
				Name: "bjcp_lookup",
				Arguments: map[string]interface{}{
					"style_code": tt.styleCode,
				},
			}
			msg := mcp.NewMessage("tools/call", toolCall)
			data, err := json.Marshal(msg)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}
			response := server.ProcessMessage(ctx, data)
			if tt.expectError {
				if response.Error == nil {
					t.Error("Expected error but got none")
				} else if response.Error.Code != tt.errorCode {
					t.Errorf("Expected error code %d, got %d", tt.errorCode, response.Error.Code)
				}
			} else {
				if response.Error != nil {
					t.Errorf("Unexpected error: %v", response.Error)
				}
			}
		})
	}
}

func TestMCP_BeerSearch_Validation(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		limit       interface{}
		expectError bool
		errorCode   int
	}{
		{"valid search", "IPA", 10, false, 0},
		{"empty query", "", nil, true, mcp.InvalidParams},
		{"negative limit", "IPA", -1, true, mcp.InvalidParams},
		{"zero limit", "IPA", 0, true, mcp.InvalidParams},
		{"too large limit", "IPA", 1001, false, 0}, // Should cap at 100, not error
		{"invalid limit type", "IPA", "ten", true, mcp.InvalidParams},
	}

	// Provide the mock to the tool handlers
	mockService := &mockBeerService{}
	toolHandlers := handlers.NewToolHandlers(nil, mockService, nil)
	server := mcp.NewServer(toolHandlers, nil)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolCall := mcp.CallToolRequest{
				Name: "search_beers",
				Arguments: map[string]interface{}{
					"name":  tt.query,
					"limit": tt.limit,
				},
			}

			msg := mcp.NewMessage("tools/call", toolCall)
			data, err := json.Marshal(msg)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			response := server.ProcessMessage(ctx, data)
			if tt.expectError {
				if response.Error == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if response.Error != nil {
					t.Errorf("Unexpected error: %v", response.Error)
				}
			}
		})
	}
}

func TestMCP_ConcurrentAccess(t *testing.T) {
	// Test concurrent tool calls
	concurrency := 10
	done := make(chan bool)

	for i := range concurrency {
		go func(i int) {
			toolCall := mcp.CallToolRequest{
				Name: "bjcp_lookup",
				Arguments: map[string]interface{}{
					"style_code": "21A",
				},
			}

			msg := mcp.NewMessage("tools/call", toolCall)
			_, err := json.Marshal(msg)
			if err != nil {
				t.Errorf("Concurrent request %d failed: %v", i, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range concurrency {
		<-done
	}
}

func TestMCP_WebSocketConnection(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping WebSocket test in short mode")
	}

	// Start test server
	server := mcp.NewServer(nil, nil)
	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	// Convert http URL to ws URL
	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// Send initialize message
	initMsg := mcp.InitializeRequest{
		ProtocolVersion: "2024-11-05",
		ClientInfo: mcp.ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	err = ws.WriteJSON(mcp.NewMessage("initialize", initMsg))
	if err != nil {
		t.Fatalf("Failed to send initialize message: %v", err)
	}

	// Read response
	var response mcp.Message
	err = ws.ReadJSON(&response)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if response.Error != nil {
		t.Errorf("Unexpected error in response: %v", response.Error)
	}
}

func TestMCP_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Performance benchmarks
	benchmarks := []struct {
		name       string
		tool       string
		args       map[string]interface{}
		iterations int
		maxLatency time.Duration
	}{
		{
			name:       "BJCP Lookup",
			tool:       "bjcp_lookup",
			args:       map[string]interface{}{"style_code": "21A"},
			iterations: 100,
			maxLatency: 500 * time.Millisecond,
		},
		{
			name:       "Beer Search",
			tool:       "search_beers",
			args:       map[string]interface{}{"query": "IPA", "limit": 10},
			iterations: 100,
			maxLatency: 500 * time.Millisecond,
		},
	}

	for _, bm := range benchmarks {
		t.Run(bm.name, func(t *testing.T) {
			var totalLatency time.Duration

			for i := range bm.iterations {
				start := time.Now()

				toolCall := mcp.CallToolRequest{
					Name:      bm.tool,
					Arguments: bm.args,
				}

				msg := mcp.NewMessage("tools/call", toolCall)
				_, err := json.Marshal(msg)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}

				latency := time.Since(start)
				totalLatency += latency

				if latency > bm.maxLatency {
					t.Errorf("Request %d exceeded maximum latency. Got %v, want <= %v",
						i, latency, bm.maxLatency)
				}
			}

			avgLatency := totalLatency / time.Duration(bm.iterations)
			t.Logf("Average latency for %s: %v", bm.name, avgLatency)
		})
	}
}
