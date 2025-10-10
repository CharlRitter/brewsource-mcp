// Package mcp implements the Model Context Protocol server logic for Brewsource MCP.
package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

type Server struct {
	tools     map[string]ToolHandler
	resources map[string]ResourceHandler
	mu        sync.RWMutex
}

type ToolHandlerRegistry interface {
	RegisterToolHandlers(server *Server)
	GetToolDefinitions() []Tool
}

type ResourceHandlerRegistry interface {
	RegisterResourceHandlers(server *Server)
	GetResourceDefinitions() []Resource
}

func NewServer(toolRegistry ToolHandlerRegistry, resourceRegistry ResourceHandlerRegistry) *Server {
	server := &Server{
		tools:     make(map[string]ToolHandler),
		resources: make(map[string]ResourceHandler),
	}

	// Register handlers if registries are provided
	if toolRegistry != nil {
		toolRegistry.RegisterToolHandlers(server)
	}
	if resourceRegistry != nil {
		resourceRegistry.RegisterResourceHandlers(server)
	}

	return server
}

// HandleHTTP handles MCP requests over HTTP POST.
func (s *Server) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	ctx := r.Context()
	var data []byte
	var err error
	data, err = io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Check for invalid JSON before processing
	var check map[string]interface{}
	if jsonErr := json.Unmarshal(data, &check); jsonErr != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response := s.ProcessMessage(ctx, data)
	w.Header().Set("Content-Type", "application/json")
	if response != nil {
		responseData, merr := json.Marshal(response)
		if merr != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseData)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) RegisterToolHandler(name string, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[name] = handler
	logrus.Debugf("Registered tool handler: %s", name)
}

func (s *Server) RegisterResourceHandler(pattern string, handler ResourceHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[pattern] = handler
	logrus.Debugf("Registered resource handler: %s", pattern)
}

func (s *Server) ProcessMessage(ctx context.Context, data []byte) *Message {
	logrus.Debugf("Processing message: %s", string(data))

	msg, err := ValidateMessage(data)
	if err != nil {
		mcpErr := func() *Error {
			target := &Error{}
			_ = errors.As(err, &target)
			return target
		}()
		return NewErrorResponse(nil, mcpErr)
	}

	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "tools/list":
		return s.handleToolsList(msg)
	case "tools/call":
		return s.handleToolsCall(ctx, msg)
	case "resources/list":
		return s.handleResourcesList(msg)
	case "resources/read":
		return s.handleResourcesRead(ctx, msg)
	default:
		return NewErrorResponse(msg.ID, NewMCPError(MethodNotFound, "Method not found", nil))
	}
}

func (s *Server) handleInitialize(msg *Message) *Message {
	var req InitializeRequest
	if msg.Params != nil {
		paramData, _ := json.Marshal(msg.Params)
		if err := json.Unmarshal(paramData, &req); err != nil {
			return NewErrorResponse(msg.ID, NewMCPError(InvalidParams, "Invalid initialize parameters", nil))
		}
	}

	logrus.Infof("Initialize request from client: %s v%s", req.ClientInfo.Name, req.ClientInfo.Version)

	response := InitializeResponse{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
			Resources: &ResourcesCapability{
				Subscribe:   false,
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "BrewSource MCP Server",
			Version: "1.0.0",
		},
	}

	return NewResponse(msg.ID, response)
}

func (s *Server) handleToolsList(msg *Message) *Message {
	// Get tool definitions from registered handlers - for now use the hardcoded ones
	// In the future, this could be made dynamic
	tools := []Tool{
		{
			Name:        "bjcp_lookup",
			Description: "Look up BJCP beer style information by style code or name",
			InputSchema: ObjectSchema(map[string]interface{}{
				"style_code": StringSchema("BJCP style code (e.g., '21A' for American IPA)", false),
				"style_name": StringSchema("BJCP style name (e.g., 'American IPA')", false),
			}, []string{}),
		},
		{
			Name:        "search_beers",
			Description: "Search for commercial beers by name, style, brewery, or location",
			InputSchema: ObjectSchema(map[string]interface{}{
				"name":     StringSchema("Beer name to search for", false),
				"style":    StringSchema("Beer style to filter by", false),
				"brewery":  StringSchema("Brewery name to filter by", false),
				"location": StringSchema("Location (city, state, country) to filter by", false),
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results (default: 20, max: 100)",
				},
			}, []string{}),
		},
		{
			Name:        "find_breweries",
			Description: "Find breweries by name, location, city, state, or country",
			InputSchema: ObjectSchema(map[string]interface{}{
				"name":     StringSchema("Brewery name to search for", false),
				"location": StringSchema("General location search (city, state, country)", false),
				"city":     StringSchema("City to filter by", false),
				"state":    StringSchema("State/province to filter by", false),
				"country":  StringSchema("Country to filter by", false),
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results (default: 20, max: 100)",
				},
			}, []string{}),
		},
	}

	return NewResponse(msg.ID, map[string]interface{}{
		"tools": tools,
	})
}

func (s *Server) handleToolsCall(ctx context.Context, msg *Message) *Message {
	var req CallToolRequest
	if msg.Params != nil {
		paramData, _ := json.Marshal(msg.Params)
		if err := json.Unmarshal(paramData, &req); err != nil {
			return NewErrorResponse(msg.ID, NewMCPError(InvalidParams, "Invalid tool call parameters", nil))
		}
	}

	// Check if tool name is provided
	if req.Name == "" {
		return NewErrorResponse(msg.ID, NewMCPError(InvalidParams, "Missing tool name", nil))
	}

	s.mu.RLock()
	handler, exists := s.tools[req.Name]
	s.mu.RUnlock()

	if !exists {
		return NewErrorResponse(msg.ID, NewMCPError(MethodNotFound, fmt.Sprintf("Tool not found: %s", req.Name), nil))
	}

	result, err := handler(ctx, req.Arguments)
	if err != nil {
		mcpErr := &Error{}
		if errors.As(err, &mcpErr) {
			return NewErrorResponse(msg.ID, mcpErr)
		}
		return NewErrorResponse(msg.ID, NewMCPError(InternalError, err.Error(), nil))
	}

	return NewResponse(msg.ID, result)
}

func (s *Server) handleResourcesList(msg *Message) *Message {
	resources := []Resource{
		{
			URI:         "bjcp://styles",
			Name:        "BJCP Beer Styles",
			Description: "Complete BJCP beer style guidelines database",
			MimeType:    "application/json",
		},
		{
			URI:         "breweries://directory",
			Name:        "Brewery Directory",
			Description: "Searchable directory of breweries",
			MimeType:    "application/json",
		},
		{
			URI:         "beers://catalog",
			Name:        "Beer Catalog",
			Description: "Commercial beer database",
			MimeType:    "application/json",
		},
		{
			URI:         "/version",
			Name:        "Service Version",
			Description: "Current version of the BrewSource MCP service",
			MimeType:    "application/json",
		},
	}

	return NewResponse(msg.ID, map[string]interface{}{
		"resources": resources,
	})
}

func (s *Server) handleResourcesRead(ctx context.Context, msg *Message) *Message {
	var req ReadResourceRequest
	if msg.Params != nil {
		paramData, _ := json.Marshal(msg.Params)
		if err := json.Unmarshal(paramData, &req); err != nil {
			return NewErrorResponse(msg.ID, NewMCPError(InvalidParams, "Invalid resource read parameters", nil))
		}
	}

	// Check if URI is provided
	if req.URI == "" {
		return NewErrorResponse(msg.ID, NewMCPError(InvalidParams, "Missing resource URI", nil))
	}

	// Basic URI validation - check if it contains a scheme
	if !isValidURI(req.URI) {
		return NewErrorResponse(msg.ID, NewMCPError(InvalidParams, "Malformed resource URI", nil))
	}

	// Simple pattern matching for resources
	var handler ResourceHandler
	s.mu.RLock()
	for pattern, h := range s.resources {
		if matchesPattern(pattern, req.URI) {
			handler = h
			break
		}
	}
	s.mu.RUnlock()

	if handler == nil {
		return NewErrorResponse(
			msg.ID,
			NewMCPError(MethodNotFound, fmt.Sprintf("Resource not found: %s", req.URI), nil),
		)
	}

	content, err := handler(ctx, req.URI)
	if err != nil {
		mcpErr := &Error{}
		if errors.As(err, &mcpErr) {
			return NewErrorResponse(msg.ID, mcpErr)
		}
		return NewErrorResponse(msg.ID, NewMCPError(InternalError, err.Error(), nil))
	}

	return NewResponse(msg.ID, map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"uri":      content.URI,
				"mimeType": content.MimeType,
				"text":     content.Text,
				"blob":     content.Blob,
			},
		},
	})
}

// Simple pattern matching - in production, use a proper router.
func matchesPattern(pattern, uri string) bool {
	if pattern == "*" {
		return true
	}

	// Simple prefix matching for now
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(uri) >= len(prefix) && uri[:len(prefix)] == prefix
	}

	return pattern == uri
}

// isValidURI checks if a URI has a basic valid format (contains scheme).
func isValidURI(uri string) bool {
	// Very basic URI validation - just check for scheme
	return len(uri) > 0 && len(uri) >= 3 &&
		((uri[0] >= 'a' && uri[0] <= 'z') || (uri[0] >= 'A' && uri[0] <= 'Z')) &&
		contains(uri, "://")
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(substr) <= len(s) && (substr == "" || indexOf(s, substr) >= 0)
}

// indexOf returns the index of the first instance of substr in s, or -1 if substr is not present in s.
func indexOf(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
