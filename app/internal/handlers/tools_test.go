package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/internal/handlers"
	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
	"github.com/CharlRitter/brewsource-mcp/app/pkg/data"
)

// Test RegisterToolHandlers function.
func TestRegisterToolHandlers(t *testing.T) {
	bjcpData := &data.BJCPData{
		Styles: map[string]data.BJCPStyle{
			"21A": {Code: "21A", Name: "American IPA", Category: "IPA"},
		},
		Categories: []string{"IPA"},
		Metadata:   data.Metadata{Version: "2021", Source: "test"},
	}

	toolHandlers := handlers.NewToolHandlers(bjcpData, nil, nil)
	server := mcp.NewServer(toolHandlers, nil)

	// Call RegisterToolHandlers
	toolHandlers.RegisterToolHandlers(server)

	// Test that the handlers were registered by attempting to use them
	ctx := context.Background()

	// Test BJCP lookup tool
	toolCall := mcp.CallToolRequest{
		Name: "bjcp_lookup",
		Arguments: map[string]interface{}{
			"style_code": "21A",
		},
	}
	msg := mcp.NewMessage("tools/call", toolCall)
	msgData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal tool request: %v", err)
	}

	response := server.ProcessMessage(ctx, msgData)
	if response.Error != nil {
		t.Errorf("Expected successful tool call, got error: %v", response.Error)
	}
}

// Test BJCPLookup with empty style_name.
func TestBJCPLookup_EmptyStyleName(t *testing.T) {
	bjcpData := &data.BJCPData{
		Styles: map[string]data.BJCPStyle{
			"21A": {Code: "21A", Name: "American IPA", Category: "IPA"},
		},
		Categories: []string{"IPA"},
		Metadata:   data.Metadata{Version: "2021", Source: "test"},
	}

	toolHandlers := handlers.NewToolHandlers(bjcpData, nil, nil)
	server := mcp.NewServer(toolHandlers, nil)
	ctx := context.Background()

	// Test with empty style_name
	toolCall := mcp.CallToolRequest{
		Name: "bjcp_lookup",
		Arguments: map[string]interface{}{
			"style_name": "",
		},
	}
	msg := mcp.NewMessage("tools/call", toolCall)
	msgData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal tool request: %v", err)
	}

	response := server.ProcessMessage(ctx, msgData)
	if response.Error == nil {
		t.Error("Expected error for empty style_name, but got none")
	}
	if response.Error.Code != mcp.InvalidParams {
		t.Errorf("Expected InvalidParams error code, got %d", response.Error.Code)
	}
}

// Test parseLimit function indirectly through search_beers tool.
func TestParseLimit_DefaultCase(t *testing.T) {
	// Use existing mock beer service from this file
	bjcpData := &data.BJCPData{
		Styles:     map[string]data.BJCPStyle{},
		Categories: []string{},
		Metadata:   data.Metadata{Version: "2021", Source: "test"},
	}

	toolHandlers := handlers.NewToolHandlers(bjcpData, &mockBeerService{}, nil)
	server := mcp.NewServer(toolHandlers, nil)
	ctx := context.Background()

	// Test with invalid limit type (boolean)
	toolCall := mcp.CallToolRequest{
		Name: "search_beers",
		Arguments: map[string]interface{}{
			"name":  "test",
			"limit": true, // Invalid type - should trigger the default case
		},
	}
	msg := mcp.NewMessage("tools/call", toolCall)
	msgData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal tool request: %v", err)
	}

	response := server.ProcessMessage(ctx, msgData)
	if response.Error == nil {
		t.Error("Expected error for invalid limit type, but got none")
	}
	if response.Error.Code != mcp.InvalidParams {
		t.Errorf("Expected InvalidParams error code, got %d", response.Error.Code)
	}
}

func TestToolDefinitions(t *testing.T) {
	// Test that we can create tool definitions without database
	handlers := &handlers.ToolHandlers{}

	tools := handlers.GetToolDefinitions()

	expectedTools := []string{"bjcp_lookup", "search_beers", "find_breweries"}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true

		// Validate that each tool has required fields
		if tool.Name == "" {
			t.Error("Tool name should not be empty")
		}
		if tool.Description == "" {
			t.Error("Tool description should not be empty")
		}
		if tool.InputSchema == nil {
			t.Error("Tool input schema should not be nil")
		}
	}

	// Check that all expected tools are present
	for _, expectedTool := range expectedTools {
		if !toolNames[expectedTool] {
			t.Errorf("Expected tool %s not found", expectedTool)
		}
	}
}

func TestBJCPLookupTool_Definition(t *testing.T) {
	handlers := &handlers.ToolHandlers{}
	tools := handlers.GetToolDefinitions()

	var bjcpTool *mcp.Tool
	for _, tool := range tools {
		if tool.Name == "bjcp_lookup" {
			bjcpTool = &tool
			break
		}
	}

	if bjcpTool == nil {
		t.Fatal("bjcp_lookup tool not found")
	}

	if bjcpTool.Name != "bjcp_lookup" {
		t.Errorf("Expected name 'bjcp_lookup', got '%s'", bjcpTool.Name)
	}

	if !contains(bjcpTool.Description, "BJCP") {
		t.Error("Expected description to contain 'BJCP'")
	}

	// Validate input schema
	schema, ok := bjcpTool.InputSchema.(map[string]interface{})
	if !ok {
		t.Fatal("Expected input schema to be a map")
	}

	// Check style_code parameter definition
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected schema properties to be a map")
	}

	styleCode, ok := props["style_code"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected style_code definition to be a map")
	}

	if styleCode["type"] != "string" {
		t.Error("Expected style_code type to be string")
	}

	// Check style_name parameter definition
	styleName, ok := props["style_name"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected style_name definition to be a map")
	}

	if styleName["type"] != "string" {
		t.Error("Expected style_name type to be string")
	}
}

func TestSearchBeersTool_Definition(t *testing.T) {
	handlers := &handlers.ToolHandlers{}
	tools := handlers.GetToolDefinitions()

	var beerTool *mcp.Tool
	for _, tool := range tools {
		if tool.Name == "search_beers" {
			beerTool = &tool
			break
		}
	}

	if beerTool == nil {
		t.Fatal("search_beers tool not found")
	}

	if beerTool.Name != "search_beers" {
		t.Errorf("Expected name 'search_beers', got '%s'", beerTool.Name)
	}

	if !contains(beerTool.Description, "beer") {
		t.Error("Expected description to contain 'beer'")
	}

	// Validate input schema
	schema, ok := beerTool.InputSchema.(map[string]interface{})
	if !ok {
		t.Fatal("Expected input schema to be a map")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected schema properties to be a map")
	}

	// Check required parameters
	requiredParams := []string{"name", "style", "brewery", "location", "limit"}
	for _, param := range requiredParams {
		if _, exists := props[param]; !exists {
			t.Errorf("Missing required parameter: %s", param)
		}
	}

	// Validate limit parameter type and constraints
	limit, ok := props["limit"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected limit definition to be a map")
	}
	if limit["type"] != "integer" {
		t.Error("Expected limit type to be integer")
	}
}

// mockBeerService implements a mock for BeerService for testing.
type mockBeerService struct{}

func (m *mockBeerService) SearchBeers(
	_ context.Context,
	_ services.BeerSearchQuery,
) ([]*services.BeerSearchResult, error) {
	return []*services.BeerSearchResult{
		{
			Name:    "Test Beer",
			Brewery: "Test Brewery",
			Style:   "Test Style",
		},
	}, nil
}

// mockBeerServiceWithError implements a mock that returns errors for testing error paths.
type mockBeerServiceWithError struct{}

func (m *mockBeerServiceWithError) SearchBeers(
	_ context.Context,
	_ services.BeerSearchQuery,
) ([]*services.BeerSearchResult, error) {
	return nil, errors.New("database connection failed")
}

// mockBreweryService implements a mock for BreweryService for testing.
type mockBreweryService struct{}

func (m *mockBreweryService) SearchBreweries(
	_ context.Context,
	_ services.BrewerySearchQuery,
) ([]*services.BrewerySearchResult, error) {
	return []*services.BrewerySearchResult{
		{
			ID:          1,
			Name:        "Test Brewery",
			BreweryType: "micro",
			City:        "Test City",
			State:       "Test State",
			Country:     "Test Country",
		},
	}, nil
}

func TestSearchBeers_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		wantErr     bool
		errCode     int
		errContains string
	}{
		{
			name:        "empty parameters",
			args:        map[string]interface{}{},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "at least one search parameter is required",
		},
		{
			name: "all empty strings",
			args: map[string]interface{}{
				"name":     "",
				"style":    "",
				"brewery":  "",
				"location": "",
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "at least one search parameter is required",
		},
		{
			name: "invalid limit type",
			args: map[string]interface{}{
				"name":  "Test Beer",
				"limit": "invalid",
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "limit must be an integer",
		},
		{
			name: "negative limit",
			args: map[string]interface{}{
				"name":  "Test Beer",
				"limit": -1,
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "limit must be greater than zero",
		},
		{
			name: "excessive limit",
			args: map[string]interface{}{
				"name":  "Test Beer",
				"limit": 1000,
			},
			wantErr: false, // Should cap at max limit (100)
		},
		{
			name: "valid search with all parameters",
			args: map[string]interface{}{
				"name":     "IPA",
				"style":    "India Pale Ale",
				"brewery":  "Test Brewery",
				"location": "Test City",
				"limit":    10,
			},
			wantErr: false,
		},
	}

	// Create a mock service and properly initialize handlers
	mockService := &mockBeerService{}
	handlers := handlers.NewToolHandlers(nil, mockService, nil)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handlers.SearchBeers(ctx, tt.args)

			switch {
			case tt.wantErr:
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				mcpErr := &mcp.Error{}
				isValid := errors.As(err, &mcpErr)
				if !isValid {
					t.Errorf("Expected *mcp.Error but got %T", err)
					return
				}

				if mcpErr.Code != tt.errCode {
					t.Errorf("Expected error code %d but got %d", tt.errCode, mcpErr.Code)
				}

				if !contains(mcpErr.Message, tt.errContains) {
					t.Errorf("Expected error message containing %q but got %q", tt.errContains, mcpErr.Message)
				}
			case err != nil:
				t.Errorf("Unexpected error: %v", err)
			case result == nil:
				t.Error("Expected non-nil result")
			}
		})
	}
}

func TestFindBreweriesTool_Definition(t *testing.T) {
	handlers := &handlers.ToolHandlers{}
	tools := handlers.GetToolDefinitions()

	var breweryTool *mcp.Tool
	for _, tool := range tools {
		if tool.Name == "find_breweries" {
			breweryTool = &tool
			break
		}
	}

	if breweryTool == nil {
		t.Fatal("find_breweries tool not found")
	}

	if breweryTool.Name != "find_breweries" {
		t.Errorf("Expected name 'find_breweries', got '%s'", breweryTool.Name)
	}

	if !contains(breweryTool.Description, "brew") {
		t.Error("Expected description to contain 'brew'")
	}

	// Validate input schema
	schema, ok := breweryTool.InputSchema.(map[string]interface{})
	if !ok {
		t.Fatal("Expected input schema to be a map")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected schema properties to be a map")
	}

	// Check all location-related parameters
	locationParams := []string{"name", "location", "city", "state", "country", "limit"}
	for _, param := range locationParams {
		if _, exists := props[param]; !exists {
			t.Errorf("Missing parameter: %s", param)
		}
	}
}

func getFindBreweriesEdgeCaseTests() []struct {
	name        string
	args        map[string]interface{}
	wantErr     bool
	errCode     int
	errContains string
} {
	return []struct {
		name        string
		args        map[string]interface{}
		wantErr     bool
		errCode     int
		errContains string
	}{
		{
			name:        "empty parameters",
			args:        map[string]interface{}{},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "at least one search parameter is required",
		},
		{
			name: "all empty strings",
			args: map[string]interface{}{
				"name":     "",
				"location": "",
				"city":     "",
				"state":    "",
				"country":  "",
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "at least one search parameter is required",
		},
		{
			name: "invalid limit type",
			args: map[string]interface{}{
				"name":  "Test Brewery",
				"limit": "invalid",
			},
			wantErr: false, // Should use default limit
		},
		{
			name: "search by city only",
			args: map[string]interface{}{
				"city": "Portland",
			},
			wantErr: false,
		},
		{
			name: "search by state only",
			args: map[string]interface{}{
				"state": "Oregon",
			},
			wantErr: false,
		},
		{
			name: "search by country only",
			args: map[string]interface{}{
				"country": "United States",
			},
			wantErr: false,
		},
		{
			name: "combined location search",
			args: map[string]interface{}{
				"city":    "Portland",
				"state":   "Oregon",
				"country": "United States",
			},
			wantErr: false,
		},
		{
			name: "excessive limit",
			args: map[string]interface{}{
				"name":  "Test Brewery",
				"limit": 1000,
			},
			wantErr: false, // Should cap at max limit (100)
		},
	}
}

func runFindBreweriesTestCase(ctx context.Context, t *testing.T, handlers *handlers.ToolHandlers, tt struct {
	name        string
	args        map[string]interface{}
	wantErr     bool
	errCode     int
	errContains string
},
) {
	result, err := handlers.FindBreweries(ctx, tt.args)

	if tt.wantErr {
		validateFindBreweriesError(t, err, tt.errCode, tt.errContains)
		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func validateFindBreweriesError(t *testing.T, err error, expectedCode int, expectedContains string) {
	if err == nil {
		t.Error("Expected error but got none")
		return
	}

	mcpErr := &mcp.Error{}
	isValid := errors.As(err, &mcpErr)
	if !isValid {
		t.Errorf("Expected *mcp.Error but got %T", err)
		return
	}

	if mcpErr.Code != expectedCode {
		t.Errorf("Expected error code %d but got %d", expectedCode, mcpErr.Code)
	}

	if !contains(mcpErr.Message, expectedContains) {
		t.Errorf("Expected error message containing %q but got %q", expectedContains, mcpErr.Message)
	}
}

func TestFindBreweries_EdgeCases(t *testing.T) {
	tests := getFindBreweriesEdgeCaseTests()

	// Create mock services and properly initialize handlers
	mockBeerService := &mockBeerService{}
	mockBreweryService := &mockBreweryService{}
	handlers := handlers.NewToolHandlers(nil, mockBeerService, mockBreweryService)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runFindBreweriesTestCase(ctx, t, handlers, tt)
		})
	}
}

func getToolResponseFormattingTests() []struct {
	name     string
	content  []mcp.ToolContent
	validate func(t *testing.T, content string)
} {
	return []struct {
		name     string
		content  []mcp.ToolContent
		validate func(t *testing.T, content string)
	}{
		{
			name: "BJCP style info formatting",
			content: []mcp.ToolContent{{
				Type: "text",
				Text: "**BJCP Style 21A: American IPA**\n\n**Category:** American IPA",
			}},
			validate: validateBJCPStyleFormatting,
		},
		{
			name: "Beer search results formatting",
			content: []mcp.ToolContent{{
				Type: "text",
				Text: "**Found 2 beer(s):**\n\n**1. Test Beer**\n- **Brewery:** Test Brewery",
			}},
			validate: validateBeerSearchFormatting,
		},
		{
			name: "Brewery search results formatting",
			content: []mcp.ToolContent{{
				Type: "text",
				Text: "**Found 1 brewery(ies):**\n\n**1. Test Brewery**\n- **Location:** Portland, OR",
			}},
			validate: validateBrewerySearchFormatting,
		},
	}
}

func validateBJCPStyleFormatting(t *testing.T, content string) {
	if !strings.Contains(content, "**BJCP Style") {
		t.Error("Missing style header formatting")
	}
	if !strings.Contains(content, "**Category:**") {
		t.Error("Missing category section")
	}
}

func validateBeerSearchFormatting(t *testing.T, content string) {
	if !strings.Contains(content, "**Found") {
		t.Error("Missing results count")
	}
	if !strings.Contains(content, "**Brewery:**") {
		t.Error("Missing brewery information")
	}
}

func validateBrewerySearchFormatting(t *testing.T, content string) {
	if !strings.Contains(content, "**Found") {
		t.Error("Missing results count")
	}
	if !strings.Contains(content, "**Location:**") {
		t.Error("Missing location information")
	}
}

func runToolResponseFormattingTest(t *testing.T, tt struct {
	name     string
	content  []mcp.ToolContent
	validate func(t *testing.T, content string)
},
) {
	result := &mcp.ToolResult{
		Content: tt.content,
		IsError: false,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var unmarshaled mcp.ToolResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if len(unmarshaled.Content) != len(tt.content) {
		t.Errorf("Content length mismatch: got %d, want %d", len(unmarshaled.Content), len(tt.content))
	}

	tt.validate(t, unmarshaled.Content[0].Text)
}

func TestToolResponse_Formatting(t *testing.T) {
	tests := getToolResponseFormattingTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runToolResponseFormattingTest(t, tt)
		})
	}
}

func TestToolResult_JSONSerialization(t *testing.T) {
	result := &mcp.ToolResult{
		Content: []mcp.ToolContent{
			{
				Type: "text",
				Text: "Test content",
			},
		},
		IsError: false,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal ToolResult: %v", err)
	}

	var unmarshaled mcp.ToolResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ToolResult: %v", err)
	}

	if len(unmarshaled.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(unmarshaled.Content))
	}

	if unmarshaled.Content[0].Text != "Test content" {
		t.Errorf("Expected 'Test content', got %s", unmarshaled.Content[0].Text)
	}
}

func TestMCPError_Types(t *testing.T) {
	tests := []struct {
		name string
		code int
		want string
	}{
		{"ParseError", mcp.ParseError, "Parse error"},
		{"InvalidRequest", mcp.InvalidRequest, "Invalid request"},
		{"MethodNotFound", mcp.MethodNotFound, "Method not found"},
		{"InvalidParams", mcp.InvalidParams, "Invalid params"},
		{"InternalError", mcp.InternalError, "Internal error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mcp.NewMCPError(tt.code, tt.want, nil)
			if err.Code != tt.code {
				t.Errorf("Expected code %d, got %d", tt.code, err.Code)
			}
			if err.Message != tt.want {
				t.Errorf("Expected message '%s', got '%s'", tt.want, err.Message)
			}
		})
	}
}

func getBJCPLookupEdgeCaseTests() []struct {
	name        string
	args        map[string]interface{}
	wantErr     bool
	errCode     int
	errContains string
} {
	return []struct {
		name        string
		args        map[string]interface{}
		wantErr     bool
		errCode     int
		errContains string
	}{
		{
			name:        "empty parameters",
			args:        map[string]interface{}{},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "either 'style_code' or 'style_name' parameter is required",
		},
		{
			name: "empty style code",
			args: map[string]interface{}{
				"style_code": "",
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "invalid style_code format",
		},
		{
			name: "empty style name",
			args: map[string]interface{}{
				"style_name": "",
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "style_code or style_name cannot be empty",
		},
		{
			name: "style not found by code",
			args: map[string]interface{}{
				"style_code": "99Z",
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "BJCP style not found for: 99Z",
		},
		{
			name: "style not found by name",
			args: map[string]interface{}{
				"style_name": "Nonexistent Style",
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "BJCP style not found for: Nonexistent Style",
		},
		{
			name: "invalid type for style_code",
			args: map[string]interface{}{
				"style_code": 123,
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "either 'style_code' or 'style_name' parameter is required",
		},
		{
			name: "both parameters empty",
			args: map[string]interface{}{
				"style_code": "",
				"style_name": "",
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "invalid style_code format",
		},
		{
			name: "extremely long style code",
			args: map[string]interface{}{
				"style_code": "999ZZZ",
			},
			wantErr:     true,
			errCode:     mcp.InvalidParams,
			errContains: "invalid style_code format",
		},
	}
}

func runBJCPLookupTestCase(ctx context.Context, t *testing.T, handlers *handlers.ToolHandlers, tt struct {
	name        string
	args        map[string]interface{}
	wantErr     bool
	errCode     int
	errContains string
},
) {
	result, err := handlers.BJCPLookup(ctx, tt.args)

	if tt.wantErr {
		validateBJCPLookupError(t, err, tt.errCode, tt.errContains)
		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func validateBJCPLookupError(t *testing.T, err error, expectedCode int, expectedContains string) {
	if err == nil {
		t.Error("Expected error but got none")
		return
	}

	mcpErr := &mcp.Error{}
	isValid := errors.As(err, &mcpErr)
	if !isValid {
		t.Errorf("Expected *mcp.Error but got %T", err)
		return
	}

	if mcpErr.Code != expectedCode {
		t.Errorf("Expected error code %d but got %d", expectedCode, mcpErr.Code)
	}

	if !contains(mcpErr.Message, expectedContains) {
		t.Errorf("Expected error message containing %q but got %q", expectedContains, mcpErr.Message)
	}
}

func TestBJCPLookup_EdgeCases(t *testing.T) {
	tests := getBJCPLookupEdgeCaseTests()

	// Use mock BJCP data for testing
	bjcpData := &data.BJCPData{
		Styles: map[string]data.BJCPStyle{
			"21A": {Code: "21A", Name: "American IPA", Category: "IPA"},
		},
		Categories: []string{"IPA"},
		Metadata:   data.Metadata{Version: "2021", Source: "test"},
	}

	handlers := handlers.NewToolHandlers(bjcpData, nil, nil)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBJCPLookupTestCase(ctx, t, handlers, tt)
		})
	}
}

func TestArgumentExtraction_HelperFunctions(t *testing.T) {
	// Test argument extraction patterns that would be used in handlers
	args := map[string]interface{}{
		"string_field": "test_value",
		"int_field":    42,
		"float_field":  3.14,
		"bool_field":   true,
	}

	// Test string extraction
	if val, ok := args["string_field"].(string); !ok || val != "test_value" {
		t.Errorf("String extraction failed: got %v", val)
	}

	// Test int extraction with type conversion
	if val, ok := args["int_field"]; ok {
		switch v := val.(type) {
		case int:
			if v != 42 {
				t.Errorf("Int extraction failed: got %d", v)
			}
		case float64: // JSON unmarshaling often creates float64
			if int(v) != 42 {
				t.Errorf("Int extraction failed: got %f", v)
			}
		default:
			t.Errorf("Unexpected type for int field: %T", v)
		}
	}

	// Test missing field
	if _, ok := args["missing_field"]; ok {
		t.Error("Expected missing_field to be absent")
	}
}

func TestStringFormatting_BrewingInfo(t *testing.T) {
	// Test formatting functions that would be used in tool responses

	// Test ABV range formatting
	abvMin, abvMax := 5.5, 7.5
	abvRange := fmt.Sprintf("%.1f%% - %.1f%%", abvMin, abvMax)
	expected := "5.5% - 7.5%"
	if abvRange != expected {
		t.Errorf("ABV formatting: got %s, want %s", abvRange, expected)
	}

	// Test IBU range formatting
	ibuMin, ibuMax := 40, 70
	ibuRange := fmt.Sprintf("%d - %d IBU", ibuMin, ibuMax)
	expectedIBU := "40 - 70 IBU"
	if ibuRange != expectedIBU {
		t.Errorf("IBU formatting: got %s, want %s", ibuRange, expectedIBU)
	}

	// Test commercial examples formatting
	examples := []string{"Stone IPA", "Russian River Blind Pig", "Bell's Two Hearted"}
	formatted := strings.Join(examples, ", ")
	if !contains(formatted, "Stone IPA") || !contains(formatted, "Russian River") {
		t.Error("Commercial examples formatting failed")
	}
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Test SearchBeers error handling.
func TestSearchBeers_DatabaseError(t *testing.T) {
	bjcpData := &data.BJCPData{
		Styles: map[string]data.BJCPStyle{
			"21A": {Code: "21A", Name: "American IPA", Category: "IPA"},
		},
		Categories: []string{"IPA"},
		Metadata:   data.Metadata{Version: "2021", Source: "test"},
	}

	// Use the error mock to test error handling
	errorService := &mockBeerServiceWithError{}
	toolHandlers := handlers.NewToolHandlers(bjcpData, errorService, nil)

	ctx := context.Background()
	args := map[string]interface{}{
		"name": "Test Beer",
	}

	result, err := toolHandlers.SearchBeers(ctx, args)

	if result != nil {
		t.Error("Expected nil result when database error occurs")
	}

	if err == nil {
		t.Error("Expected error when database operation fails")
	}

	if !strings.Contains(err.Error(), "failed to search beers") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// Test parseLimit function indirectly through SearchBeers.
func TestParseLimit_StringError(t *testing.T) {
	toolHandlers := handlers.NewToolHandlers(nil, nil, nil)

	// Test through SearchBeers which uses parseLimit
	result, err := toolHandlers.SearchBeers(context.Background(), map[string]interface{}{
		"name":  "test",
		"limit": "not-a-number",
	})

	if result != nil {
		t.Error("Expected nil result for invalid limit")
	}

	if err == nil {
		t.Error("Expected error for invalid limit")
	}
}
