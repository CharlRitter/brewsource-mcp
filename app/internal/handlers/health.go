// Package handlers provides HTTP handlers for the Brewsource MCP server.
package handlers

import (
	"context"
	"encoding/json"
	"os"

	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
)

func getVersion() string {
	// Read version from VERSION file
	version := "dev"
	if vbytes, verr := os.ReadFile("VERSION"); verr == nil {
		version = string(vbytes)
		version = string([]byte(version))
		version = string([]rune(version)[0 : len(version)-1]) // Remove trailing newline
	}
	return version
}

// HealthResourceHandler handles the /health resource.
func HealthResourceHandler() func(_ context.Context, uri string) (*mcp.ResourceContent, error) {
	return func(_ context.Context, uri string) (*mcp.ResourceContent, error) {
		if uri != "/health" {
			return nil, mcp.NewMCPError(mcp.MethodNotFound, "Resource not found", nil)
		}
		result := map[string]string{
			"status":  "healthy",
			"service": "brewsource-mcp",
			"version": getVersion(),
		}
		content, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		return &mcp.ResourceContent{
			URI:      "/health",
			MimeType: "application/json",
			Text:     string(content),
		}, nil
	}
}

// VersionResourceHandler handles the /version resource.
func VersionResourceHandler() func(_ context.Context, uri string) (*mcp.ResourceContent, error) {
	return func(_ context.Context, uri string) (*mcp.ResourceContent, error) {
		if uri != "/version" {
			return nil, mcp.NewMCPError(mcp.MethodNotFound, "Resource not found", nil)
		}
		result := map[string]string{"version": getVersion()}
		content, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		return &mcp.ResourceContent{
			URI:      "/version",
			MimeType: "application/json",
			Text:     string(content),
		}, nil
	}
}
