// Package handlers_test contains tests for the HTTP handlers in Brewsource MCP.
package handlers_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/internal/handlers"
)

func TestHealthResourceHandler(t *testing.T) {
	handler := handlers.HealthResourceHandler()
	ctx := context.Background()

	// Test valid /health URI
	res, err := handler(ctx, "/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.URI != "/health" {
		t.Errorf("expected URI /health, got %s", res.URI)
	}
	if res.MimeType != "application/json" {
		t.Errorf("expected application/json, got %s", res.MimeType)
	}
	// Check JSON content
	var parsed map[string]string
	if err = json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Fatalf("invalid JSON in response: %v", err)
	}
	if parsed["status"] != "healthy" {
		t.Errorf("expected status healthy, got %s", parsed["status"])
	}
	if parsed["service"] != "brewsource-mcp" {
		t.Errorf("expected service brewsource-mcp, got %s", parsed["service"])
	}
	if parsed["version"] == "" {
		t.Error("expected non-empty version")
	}

	// Test invalid URI
	_, err = handler(ctx, "/nothealth")
	if err == nil {
		t.Error("expected error for invalid URI, got nil")
	}
}

func TestVersionResourceHandler(t *testing.T) {
	handler := handlers.VersionResourceHandler()
	ctx := context.Background()

	// Test valid /version URI
	res, err := handler(ctx, "/version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.URI != "/version" {
		t.Errorf("expected URI /version, got %s", res.URI)
	}
	if res.MimeType != "application/json" {
		t.Errorf("expected application/json, got %s", res.MimeType)
	}
	// Check JSON content
	var parsed map[string]string
	if err = json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Fatalf("invalid JSON in response: %v", err)
	}
	if parsed["version"] == "" {
		t.Error("expected non-empty version")
	}

	// Test invalid URI
	_, err = handler(ctx, "/notversion")
	if err == nil {
		t.Error("expected error for invalid URI, got nil")
	}
}
