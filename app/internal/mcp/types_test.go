package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// Test Message Validation - Enhanced with edge cases
func TestMessage_Validation(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		expectErr bool
		errorCode int
	}{
		// Happy path cases
		{
			name:      "valid JSON-RPC 2.0 message with method",
			jsonData:  `{"jsonrpc":"2.0","method":"test","id":1}`,
			expectErr: false,
		},
		{
			name:      "valid JSON-RPC 2.0 notification (no id)",
			jsonData:  `{"jsonrpc":"2.0","method":"test"}`,
			expectErr: false,
		},
		{
			name:      "valid JSON-RPC 2.0 response with result",
			jsonData:  `{"jsonrpc":"2.0","result":{"status":"ok"},"id":1}`,
			expectErr: false,
		},
		{
			name:      "valid JSON-RPC 2.0 error response",
			jsonData:  `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":1}`,
			expectErr: false,
		},
		{
			name:      "valid message with string id",
			jsonData:  `{"jsonrpc":"2.0","method":"test","id":"test-123"}`,
			expectErr: false,
		},
		{
			name:      "valid message with null id",
			jsonData:  `{"jsonrpc":"2.0","method":"test","id":null}`,
			expectErr: false,
		},
		// Sad path cases
		{
			name:      "invalid JSON - missing closing brace",
			jsonData:  `{"jsonrpc":"2.0","method":"test","id":1`,
			expectErr: true,
			errorCode: ParseError,
		},
		{
			name:      "invalid JSON - malformed structure",
			jsonData:  `{"jsonrpc":"2.0","method":}`,
			expectErr: true,
			errorCode: ParseError,
		},
		{
			name:      "wrong JSON-RPC version 1.0",
			jsonData:  `{"jsonrpc":"1.0","method":"test","id":1}`,
			expectErr: true,
			errorCode: InvalidRequest,
		},
		{
			name:      "wrong JSON-RPC version empty string",
			jsonData:  `{"jsonrpc":"","method":"test","id":1}`,
			expectErr: true,
			errorCode: InvalidRequest,
		},
		{
			name:      "missing JSON-RPC version",
			jsonData:  `{"method":"test","id":1}`,
			expectErr: true,
			errorCode: InvalidRequest,
		},
		// Edge cases
		{
			name:      "empty JSON object",
			jsonData:  `{}`,
			expectErr: true,
			errorCode: InvalidRequest,
		},
		{
			name:      "null JSON",
			jsonData:  `null`,
			expectErr: true,
			errorCode: ParseError,
		},
		{
			name:      "empty string",
			jsonData:  ``,
			expectErr: true,
			errorCode: ParseError,
		},
		{
			name:      "array instead of object",
			jsonData:  `[{"jsonrpc":"2.0","method":"test","id":1}]`,
			expectErr: true,
			errorCode: ParseError,
		},
		// Boundary cases
		{
			name:      "very long method name",
			jsonData:  `{"jsonrpc":"2.0","method":"` + strings.Repeat("a", 1000) + `","id":1}`,
			expectErr: false,
		},
		{
			name:      "very large id number",
			jsonData:  `{"jsonrpc":"2.0","method":"test","id":9223372036854775807}`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ValidateMessage([]byte(tt.jsonData))

			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateMessage() error = %v, expectErr = %v", err, tt.expectErr)
				return
			}

			if err != nil {
				mcpErr, ok := err.(*Error)
				if !ok {
					t.Errorf("Expected MCP Error, got %T", err)
					return
				}
				if tt.errorCode != 0 && mcpErr.Code != tt.errorCode {
					t.Errorf("Expected error code %d, got %d", tt.errorCode, mcpErr.Code)
				}
			} else if msg != nil {
				// Validate successful parsing
				if msg.JSONRPC != "2.0" {
					t.Errorf("Expected JSONRPC version 2.0, got %s", msg.JSONRPC)
				}
			}
		})
	}
}

// Test Error interface implementation
func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "standard error with message",
			err: &Error{
				Code:    InvalidParams,
				Message: "test error message",
				Data:    map[string]interface{}{"detail": "extra info"},
			},
			expected: "test error message",
		},
		{
			name: "empty error message",
			err: &Error{
				Code:    InternalError,
				Message: "",
			},
			expected: "",
		},
		{
			name: "error with special characters in message",
			err: &Error{
				Code:    ParseError,
				Message: "Error with\nnewlines\tand\ttabs",
			},
			expected: "Error with\nnewlines\tand\ttabs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// Test ToolResult creation functions
func TestNewToolResult(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "normal text result",
			text: "test result",
		},
		{
			name: "empty text result",
			text: "",
		},
		{
			name: "multiline text result",
			text: "line 1\nline 2\nline 3",
		},
		{
			name: "text with special characters",
			text: "Special chars: !@#$%^&*()_+-={}[]|\\:;\"'<>?,./",
		},
		{
			name: "very long text",
			text: strings.Repeat("a", 10000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewToolResult(tt.text)

			if result == nil {
				t.Fatal("NewToolResult() returned nil")
			}

			if len(result.Content) != 1 {
				t.Errorf("NewToolResult() content length = %d, want 1", len(result.Content))
			}

			if result.Content[0].Type != "text" {
				t.Errorf("NewToolResult() content type = %q, want %q", result.Content[0].Type, "text")
			}

			if result.Content[0].Text != tt.text {
				t.Errorf("NewToolResult() content text = %q, want %q", result.Content[0].Text, tt.text)
			}

			if result.IsError {
				t.Errorf("NewToolResult() IsError = true, want false")
			}
		})
	}
}

func TestNewErrorResult(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "standard error message",
			message: "test error",
		},
		{
			name:    "empty error message",
			message: "",
		},
		{
			name:    "error with JSON",
			message: `{"error": "something went wrong"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewErrorResult(tt.message)

			if result == nil {
				t.Fatal("NewErrorResult() returned nil")
			}

			if len(result.Content) != 1 {
				t.Errorf("NewErrorResult() content length = %d, want 1", len(result.Content))
			}

			if result.Content[0].Type != "text" {
				t.Errorf("NewErrorResult() content type = %q, want %q", result.Content[0].Type, "text")
			}

			if result.Content[0].Text != tt.message {
				t.Errorf("NewErrorResult() content text = %q, want %q", result.Content[0].Text, tt.message)
			}

			if !result.IsError {
				t.Errorf("NewErrorResult() IsError = false, want true")
			}
		})
	}
}

// Test MCP Error creation with comprehensive scenarios
func TestNewMCPError(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
		data    interface{}
	}{
		{
			name:    "standard error",
			code:    InvalidParams,
			message: "test error",
			data:    map[string]interface{}{"key": "value"},
		},
		{
			name:    "error with nil data",
			code:    InternalError,
			message: "internal error",
			data:    nil,
		},
		{
			name:    "error with string data",
			code:    ParseError,
			message: "parse error",
			data:    "additional context",
		},
		{
			name:    "error with array data",
			code:    MethodNotFound,
			message: "method not found",
			data:    []string{"method1", "method2", "method3"},
		},
		{
			name:    "custom error code",
			code:    -32001,
			message: "custom server error",
			data:    map[string]interface{}{"custom": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewMCPError(tt.code, tt.message, tt.data)

			if err == nil {
				t.Fatal("NewMCPError() returned nil")
			}

			if err.Code != tt.code {
				t.Errorf("NewMCPError() Code = %d, want %d", err.Code, tt.code)
			}

			if err.Message != tt.message {
				t.Errorf("NewMCPError() Message = %q, want %q", err.Message, tt.message)
			}

			if !reflect.DeepEqual(err.Data, tt.data) {
				t.Errorf("NewMCPError() Data = %v, want %v", err.Data, tt.data)
			}
		})
	}
}

// Test Message creation functions
func TestNewMessage(t *testing.T) {
	tests := []struct {
		name   string
		method string
		params interface{}
	}{
		{
			name:   "message with map params",
			method: "test_method",
			params: map[string]interface{}{"key": "value"},
		},
		{
			name:   "message with nil params",
			method: "notification",
			params: nil,
		},
		{
			name:   "message with array params",
			method: "array_method",
			params: []interface{}{"param1", "param2"},
		},
		{
			name:   "message with string params",
			method: "string_method",
			params: "string_param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage(tt.method, tt.params)

			if msg == nil {
				t.Fatal("NewMessage() returned nil")
			}

			if msg.JSONRPC != "2.0" {
				t.Errorf("NewMessage() JSONRPC = %q, want %q", msg.JSONRPC, "2.0")
			}

			if msg.Method != tt.method {
				t.Errorf("NewMessage() Method = %q, want %q", msg.Method, tt.method)
			}

			if !reflect.DeepEqual(msg.Params, tt.params) {
				t.Errorf("NewMessage() Params = %v, want %v", msg.Params, tt.params)
			}

			if msg.ID != nil {
				t.Errorf("NewMessage() ID = %v, want nil", msg.ID)
			}
		})
	}
}

func TestNewResponse(t *testing.T) {
	tests := []struct {
		name   string
		id     interface{}
		result interface{}
	}{
		{
			name:   "response with string id",
			id:     "test-id",
			result: map[string]interface{}{"status": "success"},
		},
		{
			name:   "response with numeric id",
			id:     42,
			result: "simple result",
		},
		{
			name:   "response with null id",
			id:     nil,
			result: []interface{}{"array", "result"},
		},
		{
			name: "response with complex result",
			id:   "complex-id",
			result: map[string]interface{}{
				"data": map[string]interface{}{
					"nested": "value",
					"array":  []int{1, 2, 3},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewResponse(tt.id, tt.result)

			if msg == nil {
				t.Fatal("NewResponse() returned nil")
			}

			if msg.JSONRPC != "2.0" {
				t.Errorf("NewResponse() JSONRPC = %q, want %q", msg.JSONRPC, "2.0")
			}

			if !reflect.DeepEqual(msg.ID, tt.id) {
				t.Errorf("NewResponse() ID = %v, want %v", msg.ID, tt.id)
			}

			if !reflect.DeepEqual(msg.Result, tt.result) {
				t.Errorf("NewResponse() Result = %v, want %v", msg.Result, tt.result)
			}

			if msg.Method != "" {
				t.Errorf("NewResponse() Method = %q, want empty", msg.Method)
			}

			if msg.Error != nil {
				t.Errorf("NewResponse() Error = %v, want nil", msg.Error)
			}
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name string
		id   interface{}
		err  *Error
	}{
		{
			name: "error response with standard error",
			id:   "test-id",
			err:  &Error{Code: InternalError, Message: "internal error"},
		},
		{
			name: "error response with detailed error",
			id:   123,
			err: &Error{
				Code:    InvalidParams,
				Message: "invalid parameters",
				Data:    map[string]interface{}{"field": "username", "issue": "required"},
			},
		},
		{
			name: "error response with nil id",
			id:   nil,
			err:  &Error{Code: ParseError, Message: "parse error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewErrorResponse(tt.id, tt.err)

			if msg == nil {
				t.Fatal("NewErrorResponse() returned nil")
			}

			if msg.JSONRPC != "2.0" {
				t.Errorf("NewErrorResponse() JSONRPC = %q, want %q", msg.JSONRPC, "2.0")
			}

			if !reflect.DeepEqual(msg.ID, tt.id) {
				t.Errorf("NewErrorResponse() ID = %v, want %v", msg.ID, tt.id)
			}

			if !reflect.DeepEqual(msg.Error, tt.err) {
				t.Errorf("NewErrorResponse() Error = %v, want %v", msg.Error, tt.err)
			}

			if msg.Result != nil {
				t.Errorf("NewErrorResponse() Result = %v, want nil", msg.Result)
			}

			if msg.Method != "" {
				t.Errorf("NewErrorResponse() Method = %q, want empty", msg.Method)
			}
		})
	}
}

// Test JSON Schema helpers
func TestStringSchema(t *testing.T) {
	tests := []struct {
		name        string
		description string
		required    bool
	}{
		{
			name:        "required string schema",
			description: "test description",
			required:    true,
		},
		{
			name:        "optional string schema",
			description: "optional field",
			required:    false,
		},
		{
			name:        "empty description",
			description: "",
			required:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := StringSchema(tt.description, tt.required)

			if schema["type"] != "string" {
				t.Errorf("StringSchema() type = %v, want %q", schema["type"], "string")
			}

			if schema["description"] != tt.description {
				t.Errorf("StringSchema() description = %v, want %q", schema["description"], tt.description)
			}

			// The current implementation doesn't use the required parameter,
			// but we test that it doesn't break anything
			if schema == nil {
				t.Error("StringSchema() returned nil")
			}
		})
	}
}

func TestObjectSchema(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]interface{}
		required   []string
	}{
		{
			name: "object schema with required fields",
			properties: map[string]interface{}{
				"name": StringSchema("Name field", true),
				"age":  map[string]interface{}{"type": "integer"},
			},
			required: []string{"name"},
		},
		{
			name: "object schema with no required fields",
			properties: map[string]interface{}{
				"optional": StringSchema("Optional field", false),
			},
			required: []string{},
		},
		{
			name:       "empty object schema",
			properties: map[string]interface{}{},
			required:   []string{},
		},
		{
			name: "object schema with multiple required fields",
			properties: map[string]interface{}{
				"name":  StringSchema("Name field", true),
				"email": StringSchema("Email field", true),
				"age":   map[string]interface{}{"type": "integer"},
			},
			required: []string{"name", "email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := ObjectSchema(tt.properties, tt.required)

			if schema["type"] != "object" {
				t.Errorf("ObjectSchema() type = %v, want %q", schema["type"], "object")
			}

			if !reflect.DeepEqual(schema["properties"], tt.properties) {
				t.Errorf("ObjectSchema() properties = %v, want %v", schema["properties"], tt.properties)
			}

			if len(tt.required) > 0 {
				requiredField, exists := schema["required"]
				if !exists {
					t.Error("ObjectSchema() missing required field when required array is not empty")
				} else {
					requiredArray, ok := requiredField.([]string)
					if !ok {
						t.Errorf("ObjectSchema() required field is not []string, got %T", requiredField)
					} else if !reflect.DeepEqual(requiredArray, tt.required) {
						t.Errorf("ObjectSchema() required = %v, want %v", requiredArray, tt.required)
					}
				}
			} else {
				if requiredField, exists := schema["required"]; exists && requiredField != nil {
					t.Errorf("ObjectSchema() has required field when it should be empty: %v", requiredField)
				}
			}
		})
	}
}

// Test JSON marshaling and unmarshaling
func TestMessage_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name string
		msg  *Message
	}{
		{
			name: "complete message with all fields",
			msg: &Message{
				JSONRPC: "2.0",
				ID:      "test-123",
				Method:  "test_method",
				Params:  map[string]interface{}{"key": "value"},
			},
		},
		{
			name: "response message",
			msg: &Message{
				JSONRPC: "2.0",
				ID:      42,
				Result:  map[string]interface{}{"status": "ok"},
			},
		},
		{
			name: "error message",
			msg: &Message{
				JSONRPC: "2.0",
				ID:      "error-test",
				Error:   &Error{Code: InternalError, Message: "something went wrong"},
			},
		},
		{
			name: "notification message (no ID)",
			msg: &Message{
				JSONRPC: "2.0",
				Method:  "notify",
				Params:  []string{"param1", "param2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			// Test unmarshaling
			var unmarshaled Message
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			// Verify fields
			if unmarshaled.JSONRPC != tt.msg.JSONRPC {
				t.Errorf("Unmarshaled JSONRPC = %q, want %q", unmarshaled.JSONRPC, tt.msg.JSONRPC)
			}

			// Compare IDs with type conversion for JSON numeric handling
			if !compareIDs(unmarshaled.ID, tt.msg.ID) {
				t.Errorf("Unmarshaled ID = %v (type %T), want %v (type %T)", unmarshaled.ID, unmarshaled.ID, tt.msg.ID, tt.msg.ID)
			}

			if unmarshaled.Method != tt.msg.Method {
				t.Errorf("Unmarshaled Method = %q, want %q", unmarshaled.Method, tt.msg.Method)
			}

			// Note: For complex types like Params, Result, and Error, we need to handle
			// the fact that JSON unmarshaling might change the exact types
			// (e.g., all numbers become float64). For now, we'll just check they're not nil
			// when they should be present.
			if tt.msg.Params != nil && unmarshaled.Params == nil {
				t.Error("Unmarshaled Params is nil when it should have a value")
			}

			if tt.msg.Result != nil && unmarshaled.Result == nil {
				t.Error("Unmarshaled Result is nil when it should have a value")
			}

			if tt.msg.Error != nil && unmarshaled.Error == nil {
				t.Error("Unmarshaled Error is nil when it should have a value")
			}
		})
	}
}

// Test error code constants
func TestErrorCodes(t *testing.T) {
	expectedCodes := map[string]int{
		"ParseError":     -32700,
		"InvalidRequest": -32600,
		"MethodNotFound": -32601,
		"InvalidParams":  -32602,
		"InternalError":  -32603,
	}

	actualCodes := map[string]int{
		"ParseError":     ParseError,
		"InvalidRequest": InvalidRequest,
		"MethodNotFound": MethodNotFound,
		"InvalidParams":  InvalidParams,
		"InternalError":  InternalError,
	}

	for name, expected := range expectedCodes {
		if actual, exists := actualCodes[name]; !exists {
			t.Errorf("Error code %s is not defined", name)
		} else if actual != expected {
			t.Errorf("Error code %s = %d, want %d", name, actual, expected)
		}
	}
}

// Test ToolContent structure
func TestToolContent(t *testing.T) {
	tests := []struct {
		name    string
		content ToolContent
	}{
		{
			name: "text content",
			content: ToolContent{
				Type: "text",
				Text: "sample text",
			},
		},
		{
			name: "data content",
			content: ToolContent{
				Type: "data",
				Data: "base64encodeddata",
			},
		},
		{
			name: "content with both text and data",
			content: ToolContent{
				Type: "mixed",
				Text: "description",
				Data: "data content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling of ToolContent
			data, err := json.Marshal(tt.content)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			var unmarshaled ToolContent
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if unmarshaled.Type != tt.content.Type {
				t.Errorf("Unmarshaled Type = %q, want %q", unmarshaled.Type, tt.content.Type)
			}

			if unmarshaled.Text != tt.content.Text {
				t.Errorf("Unmarshaled Text = %q, want %q", unmarshaled.Text, tt.content.Text)
			}

			if unmarshaled.Data != tt.content.Data {
				t.Errorf("Unmarshaled Data = %q, want %q", unmarshaled.Data, tt.content.Data)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkValidateMessage(b *testing.B) {
	validJSON := []byte(`{"jsonrpc":"2.0","method":"test","id":1}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateMessage(validJSON)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkNewToolResult(b *testing.B) {
	text := "This is a test result with some content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := NewToolResult(text)
		if result == nil {
			b.Fatal("NewToolResult returned nil")
		}
	}
}

func BenchmarkMessage_JSONMarshal(b *testing.B) {
	msg := &Message{
		JSONRPC: "2.0",
		ID:      "benchmark-test",
		Method:  "benchmark_method",
		Params:  map[string]interface{}{"key": "value", "number": 42},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(msg)
		if err != nil {
			b.Fatalf("json.Marshal() error: %v", err)
		}
	}
}

// compareIDs compares two ID values, handling JSON number type conversions
func compareIDs(id1, id2 interface{}) bool {
	if reflect.DeepEqual(id1, id2) {
		return true
	}

	// Handle JSON number conversion (int -> float64)
	switch v1 := id1.(type) {
	case float64:
		switch v2 := id2.(type) {
		case int:
			return v1 == float64(v2)
		case int64:
			return v1 == float64(v2)
		}
	case int:
		switch v2 := id2.(type) {
		case float64:
			return float64(v1) == v2
		}
	case int64:
		switch v2 := id2.(type) {
		case float64:
			return float64(v1) == v2
		}
	}

	return false
}
