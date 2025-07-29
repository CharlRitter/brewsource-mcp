package mcp

import (
	"encoding/json"
	"testing"
)

func TestMessage_Validation(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		expectErr bool
	}{
		{
			name:      "valid JSON-RPC 2.0 message",
			jsonData:  `{"jsonrpc":"2.0","method":"test","id":1}`,
			expectErr: false,
		},
		{
			name:      "invalid JSON",
			jsonData:  `{"jsonrpc":"2.0","method":"test","id":1`,
			expectErr: true,
		},
		{
			name:      "wrong JSON-RPC version",
			jsonData:  `{"jsonrpc":"1.0","method":"test","id":1}`,
			expectErr: true,
		},
		{
			name:      "missing JSON-RPC version",
			jsonData:  `{"method":"test","id":1}`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateMessage([]byte(tt.jsonData))
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateMessage() error = %v, expectErr = %v", err, tt.expectErr)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	err := &Error{
		Code:    InvalidParams,
		Message: "test error message",
		Data:    map[string]interface{}{"detail": "extra info"},
	}

	expected := "test error message"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestNewToolResult(t *testing.T) {
	text := "test result"
	result := NewToolResult(text)

	if len(result.Content) != 1 {
		t.Errorf("NewToolResult() content length = %d, want 1", len(result.Content))
	}

	if result.Content[0].Type != "text" {
		t.Errorf("NewToolResult() content type = %q, want %q", result.Content[0].Type, "text")
	}

	if result.Content[0].Text != text {
		t.Errorf("NewToolResult() content text = %q, want %q", result.Content[0].Text, text)
	}

	if result.IsError {
		t.Errorf("NewToolResult() IsError = true, want false")
	}
}

func TestNewErrorResult(t *testing.T) {
	message := "test error"
	result := NewErrorResult(message)

	if len(result.Content) != 1 {
		t.Errorf("NewErrorResult() content length = %d, want 1", len(result.Content))
	}

	if result.Content[0].Type != "text" {
		t.Errorf("NewErrorResult() content type = %q, want %q", result.Content[0].Type, "text")
	}

	if result.Content[0].Text != message {
		t.Errorf("NewErrorResult() content text = %q, want %q", result.Content[0].Text, message)
	}

	if !result.IsError {
		t.Errorf("NewErrorResult() IsError = false, want true")
	}
}

func TestNewMCPError(t *testing.T) {
	code := InvalidParams
	message := "test error"
	data := map[string]interface{}{"key": "value"}

	err := NewMCPError(code, message, data)

	if err.Code != code {
		t.Errorf("NewMCPError() Code = %d, want %d", err.Code, code)
	}

	if err.Message != message {
		t.Errorf("NewMCPError() Message = %q, want %q", err.Message, message)
	}

	if err.Data == nil {
		t.Errorf("NewMCPError() Data = nil, want %v", data)
	}
}

func TestNewMessage(t *testing.T) {
	method := "test_method"
	params := map[string]interface{}{"key": "value"}

	msg := NewMessage(method, params)

	if msg.JSONRPC != "2.0" {
		t.Errorf("NewMessage() JSONRPC = %q, want %q", msg.JSONRPC, "2.0")
	}

	if msg.Method != method {
		t.Errorf("NewMessage() Method = %q, want %q", msg.Method, method)
	}

	if msg.Params == nil {
		t.Errorf("NewMessage() Params = nil, want %v", params)
	}
}

func TestNewResponse(t *testing.T) {
	id := "test-id"
	result := map[string]interface{}{"status": "success"}

	msg := NewResponse(id, result)

	if msg.JSONRPC != "2.0" {
		t.Errorf("NewResponse() JSONRPC = %q, want %q", msg.JSONRPC, "2.0")
	}

	if msg.ID != id {
		t.Errorf("NewResponse() ID = %v, want %v", msg.ID, id)
	}

	if msg.Result == nil {
		t.Errorf("NewResponse() Result = nil, want %v", result)
	}
}

func TestNewErrorResponse(t *testing.T) {
	id := "test-id"
	err := &Error{Code: InternalError, Message: "internal error"}

	msg := NewErrorResponse(id, err)

	if msg.JSONRPC != "2.0" {
		t.Errorf("NewErrorResponse() JSONRPC = %q, want %q", msg.JSONRPC, "2.0")
	}

	if msg.ID != id {
		t.Errorf("NewErrorResponse() ID = %v, want %v", msg.ID, id)
	}

	if msg.Error != err {
		t.Errorf("NewErrorResponse() Error = %v, want %v", msg.Error, err)
	}
}

func TestStringSchema(t *testing.T) {
	description := "test description"
	schema := StringSchema(description, true)

	if schema["type"] != "string" {
		t.Errorf("StringSchema() type = %v, want %q", schema["type"], "string")
	}

	if schema["description"] != description {
		t.Errorf("StringSchema() description = %v, want %q", schema["description"], description)
	}
}

func TestObjectSchema(t *testing.T) {
	properties := map[string]interface{}{
		"name": StringSchema("Name field", true),
		"age":  map[string]interface{}{"type": "integer"},
	}
	required := []string{"name"}

	schema := ObjectSchema(properties, required)

	if schema["type"] != "object" {
		t.Errorf("ObjectSchema() type = %v, want %q", schema["type"], "object")
	}

	if schema["properties"] == nil {
		t.Errorf("ObjectSchema() properties = nil, want %v", properties)
	}

	if len(schema["required"].([]string)) != 1 || schema["required"].([]string)[0] != "name" {
		t.Errorf("ObjectSchema() required = %v, want %v", schema["required"], required)
	}
}

func TestMessage_JSONMarshaling(t *testing.T) {
	msg := &Message{
		JSONRPC: "2.0",
		ID:      "test-123",
		Method:  "test_method",
		Params:  map[string]interface{}{"key": "value"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Message
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.JSONRPC != msg.JSONRPC {
		t.Errorf("Unmarshaled JSONRPC = %q, want %q", unmarshaled.JSONRPC, msg.JSONRPC)
	}

	if unmarshaled.Method != msg.Method {
		t.Errorf("Unmarshaled Method = %q, want %q", unmarshaled.Method, msg.Method)
	}
}
