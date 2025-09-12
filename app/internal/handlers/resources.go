// Package handlers provides HTTP handlers for the Brewsource MCP server, including health checks and resource endpoints.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
	"github.com/CharlRitter/brewsource-mcp/app/pkg/data"
)

const (
	// resourceSampleLimit is the default number of sample items to return for resource catalogs.
	resourceSampleLimit = 10
)

// ResourceHandlers handles all MCP resource requests and implements ResourceHandlerRegistry.
type ResourceHandlers struct {
	bjcpService    *data.BJCPService
	beerService    *services.BeerService
	breweryService *services.BreweryService
}

// NewResourceHandlers creates a new instance of ResourceHandlers.
func NewResourceHandlers(
	bjcpData *data.BJCPData,
	beerService *services.BeerService,
	breweryService *services.BreweryService,
) *ResourceHandlers {
	bjcpService := data.NewBJCPServiceFromData(bjcpData)
	return &ResourceHandlers{
		bjcpService:    bjcpService,
		beerService:    beerService,
		breweryService: breweryService,
	}
}

// RegisterResourceHandlers implements ResourceHandlerRegistry interface.
func (h *ResourceHandlers) RegisterResourceHandlers(server *mcp.Server) {
	server.RegisterResourceHandler("bjcp://*", h.HandleBJCPResource)
	server.RegisterResourceHandler("beers://*", h.HandleBeerResource)
	server.RegisterResourceHandler("breweries://*", h.HandleBreweryResource)
}

// GetResourceDefinitions implements ResourceHandlerRegistry interface.
func (h *ResourceHandlers) GetResourceDefinitions() []mcp.Resource {
	return []mcp.Resource{
		{
			URI:         "bjcp://styles",
			Name:        "BJCP Beer Styles",
			Description: "Complete BJCP beer style guidelines database",
			MimeType:    "application/json",
		},
		{
			URI:         "bjcp://styles/{code}",
			Name:        "BJCP Style Details",
			Description: "Detailed information for a specific BJCP style",
			MimeType:    "application/json",
		},
		{
			URI:         "bjcp://categories",
			Name:        "BJCP Categories",
			Description: "List of all BJCP beer categories",
			MimeType:    "application/json",
		},
		{
			URI:         "beers://catalog",
			Name:        "Beer Catalog",
			Description: "Commercial beer database",
			MimeType:    "application/json",
		},
		{
			URI:         "breweries://directory",
			Name:        "Brewery Directory",
			Description: "Searchable directory of breweries",
			MimeType:    "application/json",
		},
	}
}

// HandleBJCPResource handles BJCP-related resource requests.
func (h *ResourceHandlers) HandleBJCPResource(ctx context.Context, uri string) (*mcp.ResourceContent, error) {
	switch {
	case uri == "bjcp://styles":
		return h.handleAllBJCPStyles(ctx)
	case uri == "bjcp://categories":
		return h.handleBJCPCategories(ctx)
	case strings.HasPrefix(uri, "bjcp://styles/"):
		styleCode := strings.TrimPrefix(uri, "bjcp://styles/")
		return h.handleBJCPStyleDetail(ctx, styleCode)
	default:
		return nil, mcp.NewMCPError(mcp.MethodNotFound, fmt.Sprintf("BJCP resource not found: %s", uri), nil)
	}
}

// HandleBeerResource handles beer-related resource requests.
func (h *ResourceHandlers) HandleBeerResource(ctx context.Context, uri string) (*mcp.ResourceContent, error) {
	switch uri {
	case "beers://catalog":
		return h.handleBeerCatalog(ctx)
	default:
		return nil, mcp.NewMCPError(mcp.MethodNotFound, fmt.Sprintf("Beer resource not found: %s", uri), nil)
	}
}

// HandleBreweryResource handles brewery-related resource requests.
func (h *ResourceHandlers) HandleBreweryResource(ctx context.Context, uri string) (*mcp.ResourceContent, error) {
	switch uri {
	case "breweries://directory":
		return h.handleBreweryDirectory(ctx)
	default:
		return nil, mcp.NewMCPError(mcp.MethodNotFound, fmt.Sprintf("Brewery resource not found: %s", uri), nil)
	}
}

func (h *ResourceHandlers) handleAllBJCPStyles(_ context.Context) (*mcp.ResourceContent, error) {
	// For now, return a summary of available styles
	categories := h.bjcpService.GetCategories()
	result := map[string]interface{}{
		"description":  "BJCP Beer Style Guidelines",
		"version":      h.bjcpService.GetMetadata().Version,
		"categories":   categories,
		"total_styles": len(h.bjcpService.GetAllStyles()),
		"usage": map[string]string{
			"lookup_by_code": "bjcp://styles/{code}",
			"example":        "bjcp://styles/21A",
		},
	}
	content, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal BJCP styles: %w", err)
	}
	return &mcp.ResourceContent{
		URI:      "bjcp://styles",
		MimeType: "application/json",
		Text:     string(content),
	}, nil
}

func (h *ResourceHandlers) handleBJCPCategories(_ context.Context) (*mcp.ResourceContent, error) {
	categories := h.bjcpService.GetCategories()
	result := map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
	}
	content, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal BJCP categories: %w", err)
	}
	return &mcp.ResourceContent{
		URI:      "bjcp://categories",
		MimeType: "application/json",
		Text:     string(content),
	}, nil
}

func (h *ResourceHandlers) handleBJCPStyleDetail(_ context.Context, styleCode string) (*mcp.ResourceContent, error) {
	style, err := h.bjcpService.GetStyleByCode(styleCode)
	if err != nil {
		return nil, mcp.NewMCPError(mcp.MethodNotFound, fmt.Sprintf("BJCP style not found: %s", styleCode), nil)
	}
	content, err := json.Marshal(style)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal BJCP style: %w", err)
	}
	return &mcp.ResourceContent{
		URI:      fmt.Sprintf("bjcp://styles/%s", styleCode),
		MimeType: "application/json",
		Text:     string(content),
	}, nil
}

func (h *ResourceHandlers) handleBeerCatalog(ctx context.Context) (*mcp.ResourceContent, error) {
	// Return a sample of beers to show the catalog structure
	query := services.BeerSearchQuery{
		Limit: resourceSampleLimit,
	}

	// Since we need at least one search parameter, let's search for a common style
	query.Style = "IPA"

	beers, err := h.beerService.SearchBeers(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get beer catalog sample: %w", err)
	}

	result := map[string]interface{}{
		"description":  "Commercial Beer Catalog",
		"sample_beers": beers,
		"usage": map[string]string{
			"search_tool": "Use the search_beers tool to query specific beers",
			"parameters":  "name, style, brewery, location",
		},
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal beer catalog: %w", err)
	}

	return &mcp.ResourceContent{
		URI:      "beers://catalog",
		MimeType: "application/json",
		Text:     string(content),
	}, nil
}

func (h *ResourceHandlers) handleBreweryDirectory(ctx context.Context) (*mcp.ResourceContent, error) {
	// Return a sample of breweries to show the directory structure
	query := services.BrewerySearchQuery{
		Limit: resourceSampleLimit,
	}

	// Search for breweries in a popular beer region
	query.State = "California"

	breweries, err := h.breweryService.SearchBreweries(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get brewery directory sample: %w", err)
	}
	// Ensure sample_breweries is always an array, not null
	if breweries == nil {
		breweries = []*services.BrewerySearchResult{}
	}
	result := map[string]interface{}{
		"description":      "Brewery Directory",
		"sample_breweries": breweries,
		"usage": map[string]string{
			"search_tool": "Use the find_breweries tool to query specific breweries",
			"parameters":  "name, location, city, state, country",
		},
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal brewery directory: %w", err)
	}

	return &mcp.ResourceContent{
		URI:      "breweries://directory",
		MimeType: "application/json",
		Text:     string(content),
	}, nil
}
