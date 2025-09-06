package mcp

import (
	"context"
	"encoding/json"
)

// MCP Protocol Types
// Based on Model Context Protocol specification

type Message struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewMCPError constructs a new MCP Error.
func NewMCPError(code int, message string, data interface{}) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Error codes from JSON-RPC 2.0 specification.
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// MCP-specific message types

type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

type InitializeResponse struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

type ServerCapabilities struct {
	Logging   *LoggingCapability   `json:"logging,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Tools     *ToolsCapability     `json:"tools,omitempty"`
}

type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type SamplingCapability struct{}

type LoggingCapability struct{}

type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool definitions

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
}

type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// Resource definitions

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

type ReadResourceRequest struct {
	URI string `json:"uri"`
}

// Handler function types

type (
	ToolHandler     func(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
	ResourceHandler func(ctx context.Context, uri string) (*ResourceContent, error)
)

// Helper functions for creating responses

func NewToolResult(text string) *ToolResult {
	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: text,
		}},
	}
}

func NewErrorResult(message string) *ToolResult {
	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: message,
		}},
		IsError: true,
	}
}

// Message helpers

func NewMessage(method string, params interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

func NewResponse(id interface{}, result interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func NewErrorResponse(id interface{}, err *Error) *Message {
	return &Message{
		JSONRPC: "2.0",
		ID:      id,
		Error:   err,
	}
}

// JSON Schema helpers for tool input validation

func StringSchema(description string, _ bool) map[string]interface{} {
	schema := map[string]interface{}{
		"type":        "string",
		"description": description,
	}
	return schema
}

func ObjectSchema(properties map[string]interface{}, required []string) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// Validation helpers

func ValidateMessage(data []byte) (*Message, error) {
	// Check for null input first
	if len(data) == 0 || string(data) == "null" {
		return nil, NewMCPError(ParseError, "Invalid JSON", nil)
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, NewMCPError(ParseError, "Invalid JSON", nil)
	}

	if msg.JSONRPC != "2.0" {
		return nil, NewMCPError(InvalidRequest, "Invalid JSON-RPC version", nil)
	}

	return &msg, nil
}
