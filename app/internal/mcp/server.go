package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Server struct {
	tools     map[string]ToolHandler
	resources map[string]ResourceHandler
	upgrader  websocket.Upgrader
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
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
	}

	// Register handlers
	toolRegistry.RegisterToolHandlers(server)
	resourceRegistry.RegisterResourceHandlers(server)

	return server
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

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	logrus.Info("New WebSocket connection established")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		_, data, rerr := conn.ReadMessage()
		if rerr != nil {
			if websocket.IsUnexpectedCloseError(rerr, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("WebSocket error: %v", rerr)
			}
			break
		}

		response := s.processMessage(ctx, data)
		if response != nil {
			responseData, merr := json.Marshal(response)
			if merr != nil {
				logrus.Errorf("Failed to marshal response: %v", merr)
				continue
			}

			if werr := conn.WriteMessage(websocket.TextMessage, responseData); werr != nil {
				logrus.Errorf("Failed to write WebSocket message: %v", werr)
				break
			}
		}
	}

	logrus.Info("WebSocket connection closed")
}

func (s *Server) HandleStdio() error {
	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()

	logrus.Info("Stdio server ready")

	for scanner.Scan() {
		data := scanner.Bytes()
		response := s.processMessage(ctx, data)

		if response != nil {
			responseData, err := json.Marshal(response)
			if err != nil {
				logrus.Errorf("Failed to marshal response: %v", err)
				continue
			}

			os.Stdout.Write(append(responseData, '\n'))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stdio scanner error: %w", err)
	}

	return nil
}

func (s *Server) processMessage(ctx context.Context, data []byte) *Message {
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
		"contents": []ResourceContent{*content},
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
