package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
)

func TestToolDefinitions(t *testing.T) {
	// Test that we can create tool definitions without database
	handlers := &ToolHandlers{}

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
	handlers := &ToolHandlers{}
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
}

func TestSearchBeersTool_Definition(t *testing.T) {
	handlers := &ToolHandlers{}
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
}

func TestFindBreweriesTool_Definition(t *testing.T) {
	handlers := &ToolHandlers{}
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
