// Package handlers_test contains tests for the HTTP handlers in Brewsource MCP.
package handlers_test

import (
	"context"
	"encoding/json"
	"os"
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

func TestGetVersionFromFile(t *testing.T) {
	// Create a temporary VERSION file to test version reading
	tempVersionFile := "VERSION.test"

	// Test case 1: VERSION file exists with content
	versionContent := "1.2.3\n"
	err := os.WriteFile(tempVersionFile, []byte(versionContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test VERSION file: %v", err)
	}
	defer func() {
		_ = os.Remove(tempVersionFile)
	}()

	// Temporarily rename the real VERSION file if it exists
	realVersionExists := false
	if _, statErr := os.Stat("VERSION"); statErr == nil {
		realVersionExists = true
		err = os.Rename("VERSION", "VERSION.backup")
		if err != nil {
			t.Fatalf("Failed to backup VERSION file: %v", err)
		}
		defer func() {
			_ = os.Rename("VERSION.backup", "VERSION")
		}()
	}

	// Rename our test file to VERSION
	err = os.Rename(tempVersionFile, "VERSION")
	if err != nil {
		t.Fatalf("Failed to rename test file: %v", err)
	}
	defer func() {
		_ = os.Remove("VERSION")
		if realVersionExists {
			_ = os.Rename("VERSION.backup", "VERSION")
		}
	}()

	// Now test both handlers to get coverage on getVersion with file
	healthHandler := handlers.HealthResourceHandler()
	versionHandler := handlers.VersionResourceHandler()
	ctx := context.Background()

	// Test health handler
	healthRes, err := healthHandler(ctx, "/health")
	if err != nil {
		t.Fatalf("unexpected error from health handler: %v", err)
	}
	var healthParsed map[string]string
	if err = json.Unmarshal([]byte(healthRes.Text), &healthParsed); err != nil {
		t.Fatalf("invalid JSON in health response: %v", err)
	}
	if healthParsed["version"] != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %s", healthParsed["version"])
	}

	// Test version handler
	versionRes, err := versionHandler(ctx, "/version")
	if err != nil {
		t.Fatalf("unexpected error from version handler: %v", err)
	}
	var versionParsed map[string]string
	if err = json.Unmarshal([]byte(versionRes.Text), &versionParsed); err != nil {
		t.Fatalf("invalid JSON in version response: %v", err)
	}
	if versionParsed["version"] != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %s", versionParsed["version"])
	}
}

func TestGetVersionWithoutFile(t *testing.T) {
	// Temporarily remove VERSION file if it exists
	if _, err := os.Stat("VERSION"); err == nil {
		err = os.Rename("VERSION", "VERSION.backup")
		if err != nil {
			t.Fatalf("Failed to backup VERSION file: %v", err)
		}
		defer func() {
			_ = os.Rename("VERSION.backup", "VERSION")
		}()
	}

	// Test handlers when no VERSION file exists (should default to "dev")
	healthHandler := handlers.HealthResourceHandler()
	versionHandler := handlers.VersionResourceHandler()
	ctx := context.Background()

	// Test health handler
	healthRes, err := healthHandler(ctx, "/health")
	if err != nil {
		t.Fatalf("unexpected error from health handler: %v", err)
	}
	var healthParsed map[string]string
	if err = json.Unmarshal([]byte(healthRes.Text), &healthParsed); err != nil {
		t.Fatalf("invalid JSON in health response: %v", err)
	}
	if healthParsed["version"] != "dev" {
		t.Errorf("expected version dev, got %s", healthParsed["version"])
	}

	// Test version handler
	versionRes, err := versionHandler(ctx, "/version")
	if err != nil {
		t.Fatalf("unexpected error from version handler: %v", err)
	}
	var versionParsed map[string]string
	if err = json.Unmarshal([]byte(versionRes.Text), &versionParsed); err != nil {
		t.Fatalf("invalid JSON in version response: %v", err)
	}
	if versionParsed["version"] != "dev" {
		t.Errorf("expected version dev, got %s", versionParsed["version"])
	}
}
