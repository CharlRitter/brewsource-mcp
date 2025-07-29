package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
)

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
