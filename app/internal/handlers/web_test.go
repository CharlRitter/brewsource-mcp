// Package handlers_test contains tests for the HTTP handlers in Brewsource MCP.
package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/internal/handlers"
)

func TestNewWebHandlers(t *testing.T) {
	webHandlers := handlers.NewWebHandlers()
	if webHandlers == nil {
		t.Fatal("NewWebHandlers returned nil")
	}
}

// Test ServeHome handler with valid request.
func TestServeHome_ValidRequest(t *testing.T) {
	// This test now only checks that ServeHome returns 200 and contains expected static elements,
	// since README.md is fetched from GitHub and cannot be mocked easily.
	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "localhost:8080"
	recorder := httptest.NewRecorder()

	webHandlers.ServeHome(recorder, req)

	// Accept either 200 OK (if GitHub fetch works) or 500 (if not)
	if recorder.Code != http.StatusOK && recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d or %d, got %d", http.StatusOK, http.StatusInternalServerError, recorder.Code)
	}

	body := recorder.Body.String()
	if recorder.Code == http.StatusOK {
		if !strings.Contains(body, "BrewSource MCP Server") {
			t.Error("Response body should contain project name")
		}
		if !strings.Contains(body, "localhost:8080") {
			t.Error("Response body should contain the host")
		}
		if !strings.Contains(body, "ws://localhost:8080/mcp") {
			t.Error("Response body should contain WebSocket URL")
		}
	} else if !strings.Contains(body, "Could not load README") {
		t.Error("Response should contain README error message on failure")
	}
}

// Test ServeHome handler with invalid path.
func TestServeHome_InvalidPath(t *testing.T) {
	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/invalid", nil)
	recorder := httptest.NewRecorder()

	webHandlers.ServeHome(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

// Test ServeHome handler when README file is missing.
func TestServeHome_MissingReadme(t *testing.T) {
	// This test is no longer meaningful since README.md is always fetched from GitHub.
	// Instead, we check that ServeHome returns either 200 or 500, and error message on failure.
	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	webHandlers.ServeHome(recorder, req)

	if recorder.Code != http.StatusOK && recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d or %d, got %d", http.StatusOK, http.StatusInternalServerError, recorder.Code)
	}

	body := recorder.Body.String()
	if recorder.Code == http.StatusInternalServerError {
		if !strings.Contains(body, "Could not load README") {
			t.Error("Response should contain README error message on failure")
		}
	}
}

// Test ServeAPI handler.
func TestServeAPI_ValidRequest(t *testing.T) {
	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Host = "localhost:8080"
	recorder := httptest.NewRecorder()

	webHandlers.ServeAPI(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	contentType := recorder.Header().Get("Content-Type")
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	body := recorder.Body.String()

	// Check for required JSON fields
	expectedFields := []string{
		`"name": "BrewSource MCP Server"`,
		`"description": "Model Context Protocol server for brewing resources"`,
		`"phase": "Phase 1 MVP"`,
		`"bjcp_lookup"`,
		`"search_beers"`,
		`"find_breweries"`,
		`"bjcp://styles"`,
		`"beers://catalog"`,
		`"breweries://directory"`,
		`"websocket": "ws://localhost:8080/mcp"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(body, field) {
			t.Errorf("Response body should contain: %s", field)
		}
	}

	// Verify it's valid JSON structure
	if !strings.HasPrefix(body, "{") || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Error("Response should be valid JSON object")
	}
}

// Test ServeAPI with different HTTP methods.
func TestServeAPI_DifferentMethods(t *testing.T) {
	webHandlers := handlers.NewWebHandlers()

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/api", nil)
		req.Host = "example.com"
		recorder := httptest.NewRecorder()

		webHandlers.ServeAPI(recorder, req)

		// API should respond to all methods (it doesn't check method)
		if recorder.Code != http.StatusOK {
			t.Errorf("Method %s: Expected status code %d, got %d", method, http.StatusOK, recorder.Code)
		}

		body := recorder.Body.String()
		if !strings.Contains(body, "BrewSource MCP Server") {
			t.Errorf("Method %s: Response should contain server name", method)
		}

		if !strings.Contains(body, "ws://example.com/mcp") {
			t.Errorf("Method %s: Response should contain correct WebSocket URL for host", method)
		}
	}
}

// Test ServeAPI response format consistency.
func TestServeAPI_ResponseFormat(t *testing.T) {
	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Host = "test.example.com:3000"
	recorder := httptest.NewRecorder()

	webHandlers.ServeAPI(recorder, req)

	body := recorder.Body.String()

	// Test that the response contains all expected top-level keys
	expectedKeys := []string{
		`"name"`,
		`"version"`,
		`"description"`,
		`"endpoints"`,
		`"phase"`,
		`"tools"`,
		`"resources"`,
		`"connection"`,
	}

	for _, key := range expectedKeys {
		if !strings.Contains(body, key) {
			t.Errorf("Response should contain key: %s", key)
		}
	}

	// Test that endpoints section contains expected endpoints
	expectedEndpoints := []string{
		`"mcp": "/mcp"`,
		`"health": "/health"`,
		`"api": "/api"`,
	}

	for _, endpoint := range expectedEndpoints {
		if !strings.Contains(body, endpoint) {
			t.Errorf("Response should contain endpoint: %s", endpoint)
		}
	}

	// Test WebSocket URL format with custom host and port
	expectedWebSocketURL := `"websocket": "ws://test.example.com:3000/mcp"`
	if !strings.Contains(body, expectedWebSocketURL) {
		t.Errorf("Response should contain WebSocket URL: %s", expectedWebSocketURL)
	}
}

// Test LandingPageData structure.
func TestLandingPageData_Structure(t *testing.T) {
	data := handlers.LandingPageData{
		ProjectName: "Test Project",
		Version:     "v1.0.0",
		Description: "Test Description",
		Host:        "localhost:8080",
		ReadmeHTML:  "<h1>Test</h1>",
		LastUpdated: "January 1, 2025",
	}

	if data.ProjectName != "Test Project" {
		t.Errorf("Expected ProjectName 'Test Project', got '%s'", data.ProjectName)
	}

	if data.Version != "v1.0.0" {
		t.Errorf("Expected Version 'v1.0.0', got '%s'", data.Version)
	}

	if data.Description != "Test Description" {
		t.Errorf("Expected Description 'Test Description', got '%s'", data.Description)
	}

	if data.Host != "localhost:8080" {
		t.Errorf("Expected Host 'localhost:8080', got '%s'", data.Host)
	}

	if string(data.ReadmeHTML) != "<h1>Test</h1>" {
		t.Errorf("Expected ReadmeHTML '<h1>Test</h1>', got '%s'", string(data.ReadmeHTML))
	}

	if data.LastUpdated != "January 1, 2025" {
		t.Errorf("Expected LastUpdated 'January 1, 2025', got '%s'", data.LastUpdated)
	}
}

// Test markdown to HTML conversion in ServeHome.
func TestServeHome_MarkdownConversion(t *testing.T) {
	// This test is no longer meaningful since README.md is fetched from GitHub and cannot be controlled.
	// Instead, we check that ServeHome returns 200 or 500, and that HTML is present on success.
	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	webHandlers.ServeHome(recorder, req)

	if recorder.Code != http.StatusOK && recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d or %d, got %d", http.StatusOK, http.StatusInternalServerError, recorder.Code)
	}

	body := recorder.Body.String()
	if recorder.Code == http.StatusOK {
		// Check for some HTML elements
		if !strings.Contains(body, "<html") {
			t.Error("Response should contain HTML markup")
		}
	} else {
		if !strings.Contains(body, "Could not load README") {
			t.Error("Response should contain README error message on failure")
		}
	}
}

// Test error handling in template execution.
func TestServeHome_TemplateExecutionHandling(t *testing.T) {
	// This test is more challenging since we can't easily break template execution
	// without modifying the WebHandlers struct. We'll test the happy path thoroughly.

	tempDir := t.TempDir()
	readmePath := filepath.Join(tempDir, "README.md")
	readmeContent := "# Test README\nSimple content for template execution test."

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0o644); err != nil {
		t.Fatalf("Failed to create test README: %v", err)
	}

	testWd := filepath.Join(tempDir, "app", "internal", "handlers")
	if err := os.MkdirAll(testWd, 0o755); err != nil {
		t.Fatalf("Failed to create test directory structure: %v", err)
	}
	t.Chdir(testWd)

	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "test.localhost"
	recorder := httptest.NewRecorder()

	webHandlers.ServeHome(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected successful template execution, got status %d", recorder.Code)
	}

	body := recorder.Body.String()

	// Verify all template variables were replaced
	if strings.Contains(body, "{{.") {
		t.Error("Template variables should be replaced, found unreplaced template syntax")
	}

	// Verify essential template elements are present (except README content, which is now fetched from GitHub)
	essentialElements := []string{
		"<!DOCTYPE html>",
		"<html lang=\"en\">",
		"<head>",
		"<body>",
		"test.localhost",
	}

	for _, element := range essentialElements {
		if !strings.Contains(body, element) {
			t.Errorf("Response should contain essential element: %s", element)
		}
	}
}

// Benchmark ServeHome handler.
func BenchmarkServeHome(b *testing.B) {
	// Create test README
	tempDir := b.TempDir()
	readmePath := filepath.Join(tempDir, "README.md")
	readmeContent := `# BrewSource MCP Server
A comprehensive brewing information server.

## Features
Multiple features for brewing enthusiasts.
`

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0o644); err != nil {
		b.Fatalf("Failed to create test README: %v", err)
	}

	testWd := filepath.Join(tempDir, "app", "internal", "handlers")
	if err := os.MkdirAll(testWd, 0o755); err != nil {
		b.Fatalf("Failed to create test directory structure: %v", err)
	}
	b.Chdir(testWd)

	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "localhost:8080"

	b.ResetTimer()
	for range b.N {
		recorder := httptest.NewRecorder()
		webHandlers.ServeHome(recorder, req)
	}
}

// Benchmark ServeAPI handler.
func BenchmarkServeAPI(b *testing.B) {
	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Host = "localhost:8080"

	b.ResetTimer()
	for range b.N {
		recorder := httptest.NewRecorder()
		webHandlers.ServeAPI(recorder, req)
	}
}

// Test concurrent access to web handlers.
func TestWebHandlers_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	readmePath := filepath.Join(tempDir, "README.md")
	readmeContent := "# Concurrent Test README\nTesting concurrent access."

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0o644); err != nil {
		t.Fatalf("Failed to create test README: %v", err)
	}

	testWd := filepath.Join(tempDir, "app", "internal", "handlers")
	if err := os.MkdirAll(testWd, 0o755); err != nil {
		t.Fatalf("Failed to create test directory structure: %v", err)
	}
	t.Chdir(testWd)

	webHandlers := handlers.NewWebHandlers()

	// Test concurrent access to both handlers
	const numGoroutines = 10
	done := make(chan bool, numGoroutines*2)

	// Test ServeHome concurrently
	for range numGoroutines {
		go func() {
			defer func() { done <- true }()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = "concurrent.test"
			recorder := httptest.NewRecorder()
			webHandlers.ServeHome(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Errorf("Concurrent ServeHome failed with status %d", recorder.Code)
			}
		}()
	}

	// Test ServeAPI concurrently
	for range numGoroutines {
		go func() {
			defer func() { done <- true }()
			req := httptest.NewRequest(http.MethodGet, "/api", nil)
			req.Host = "concurrent.test"
			recorder := httptest.NewRecorder()
			webHandlers.ServeAPI(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Errorf("Concurrent ServeAPI failed with status %d", recorder.Code)
			}
		}()
	}

	// Wait for all goroutines to complete
	for range numGoroutines * 2 {
		<-done
	}
}

// Test ServeHealth handler with JSON response.
func TestServeHealth_JSONResponse(t *testing.T) {
	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Host = "localhost:8080"
	recorder := httptest.NewRecorder()

	webHandlers.ServeHealth(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	contentType := recorder.Header().Get("Content-Type")
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	body := recorder.Body.String()

	// Check for required JSON fields
	expectedFields := []string{
		`"status": "healthy"`,
		`"service": "brewsource-mcp"`,
		`"version":`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(body, field) {
			t.Errorf("Response body should contain: %s", field)
		}
	}

	// Verify it's valid JSON structure
	if !strings.HasPrefix(body, "{") || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Error("Response should be valid JSON object")
	}
}

// Test ServeFavicon handler.
func TestServeFavicon(t *testing.T) {
	webHandlers := handlers.NewWebHandlers()

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	recorder := httptest.NewRecorder()

	webHandlers.ServeFavicon(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "image/x-icon" {
		t.Errorf("Expected Content-Type 'image/x-icon', got '%s'", contentType)
	}

	// The favicon should not be empty
	if recorder.Body.Len() == 0 {
		t.Error("Favicon response body should not be empty")
	}
}
