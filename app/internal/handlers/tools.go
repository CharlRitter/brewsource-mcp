package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
	"github.com/CharlRitter/brewsource-mcp/app/pkg/data"
)

const (
	// defaultSearchLimit is the default number of results when no limit is specified.
	defaultSearchLimit = 20
	// maxSearchLimit is the maximum allowed number of results.
	maxSearchLimit = 100
)

// ToolHandlers handles all MCP tool requests and implements ToolHandlerRegistry.
type ToolHandlers struct {
	bjcpData       *data.BJCPData
	beerService    services.BeerServiceInterface
	breweryService services.BreweryServiceInterface
}

// NewToolHandlers creates a new instance of ToolHandlers.
func NewToolHandlers(
	bjcpData *data.BJCPData,
	beerService services.BeerServiceInterface,
	breweryService services.BreweryServiceInterface,
) *ToolHandlers {
	return &ToolHandlers{
		bjcpData:       bjcpData,
		beerService:    beerService,
		breweryService: breweryService,
	}
}

// RegisterToolHandlers implements ToolHandlerRegistry interface.
func (h *ToolHandlers) RegisterToolHandlers(server *mcp.Server) {
	server.RegisterToolHandler("bjcp_lookup", h.BJCPLookup)
	server.RegisterToolHandler("search_beers", h.SearchBeers)
	server.RegisterToolHandler("find_breweries", h.FindBreweries)
}

func (h *ToolHandlers) GetToolDefinitions() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "bjcp_lookup",
			Description: "Look up BJCP beer style information by style code or name",
			InputSchema: mcp.ObjectSchema(map[string]interface{}{
				"style_code": mcp.StringSchema("BJCP style code (e.g., '21A' for American IPA)", false),
				"style_name": mcp.StringSchema("BJCP style name (e.g., 'American IPA')", false),
			}, []string{}),
		},
		{
			Name:        "search_beers",
			Description: "Search for commercial beers by name, style, brewery, or location",
			InputSchema: mcp.ObjectSchema(map[string]interface{}{
				"name":     mcp.StringSchema("Beer name to search for", false),
				"style":    mcp.StringSchema("Beer style to filter by", false),
				"brewery":  mcp.StringSchema("Brewery name to filter by", false),
				"location": mcp.StringSchema("Location (city, state, country) to filter by", false),
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results (default: 20, max: 100)",
				},
			}, []string{}),
		},
		{
			Name:        "find_breweries",
			Description: "Find breweries by name, location, city, state, or country",
			InputSchema: mcp.ObjectSchema(map[string]interface{}{
				"name":     mcp.StringSchema("Brewery name to search for", false),
				"location": mcp.StringSchema("General location search (city, state, country)", false),
				"city":     mcp.StringSchema("City to filter by", false),
				"state":    mcp.StringSchema("State to filter by", false),
				"country":  mcp.StringSchema("Country to filter by", false),
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results (default: 20, max: 100)",
				},
			}, []string{}),
		},
	}
}

// isValidBJCPStyleCode validates BJCP style codes (e.g., 21A, 1B, 33C).
func isValidBJCPStyleCode(code string) bool {
	matched, _ := regexp.MatchString(`^[0-9]{1,2}[A-Z]$`, code)
	return matched
}

// BJCPLookup handles BJCP style lookup functionality.
func (h *ToolHandlers) BJCPLookup(_ context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
	styleCode, hasCode := args["style_code"].(string)
	styleName, hasName := args["style_name"].(string)

	if !hasCode && !hasName {
		return nil, &mcp.Error{
			Code:    mcp.InvalidParams,
			Message: "either 'style_code' or 'style_name' parameter is required",
			Data: map[string]interface{}{
				"provided_params": args,
			},
		}
	}

	var style *data.BJCPStyle
	var err error
	bjcpService := data.NewBJCPServiceFromData(h.bjcpData)
	switch {
	case hasCode:
		styleCode = strings.ToUpper(styleCode)
		if !isValidBJCPStyleCode(styleCode) {
			return nil, &mcp.Error{
				Code:    mcp.InvalidParams,
				Message: "invalid style_code format",
				Data:    map[string]interface{}{"style_code": styleCode},
			}
		}
		style, err = bjcpService.GetStyleByCode(styleCode)
	case hasName && styleName != "":
		style, err = bjcpService.GetStyleByName(styleName)
	default:
		return nil, &mcp.Error{
			Code:    mcp.InvalidParams,
			Message: "style_code or style_name cannot be empty",
		}
	}
	if err != nil {
		var lookupParam string
		if hasCode {
			lookupParam = styleCode
		}
		if hasName {
			lookupParam = styleName
		}
		return nil, &mcp.Error{
			Code:    mcp.InvalidParams,
			Message: fmt.Sprintf("BJCP style not found for: %s", lookupParam),
		}
	}
	v := style.Vitals
	response := fmt.Sprintf(`**BJCP Style %s: %s**

**Category:** %s

**Overall Impression:** %s
- **ABV:** %.1f - %.1f%%
- **IBU:** %d - %d
- **SRM:** %.1f - %.1f
- **OG:** %.3f - %.3f
- **FG:** %.3f - %.3f

**Appearance:** %s

**Aroma:** %s

**Flavor:** %s

**Mouthfeel:** %s

**Comments:** %s

**History:** %s

**Characteristic Ingredients:** %s

**Style Comparison:** %s

**Commercial Examples:** %s`,
		style.Code, style.Name,
		style.Category,
		style.OverallImpression,
		v.ABVMin, v.ABVMax,
		v.IBUMin, v.IBUMax,
		v.SRMMin, v.SRMMax,
		v.OGMin, v.OGMax,
		v.FGMin, v.FGMax,
		style.Appearance,
		style.Aroma,
		style.Flavor,
		style.Mouthfeel,
		style.Comments,
		style.History,
		style.CharacteristicIngredients,
		style.StyleComparison,
		strings.Join(style.CommercialExamples, ", "))
	return &mcp.ToolResult{
		Content: []mcp.ToolContent{{
			Type: "text",
			Text: response,
		}},
	}, nil
}

// SearchBeers handles beer search functionality.
func (h *ToolHandlers) SearchBeers(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
	query, err := h.parseBeerSearchQuery(args)
	if err != nil {
		return nil, err
	}

	if !h.hasAnyBeerSearchParam(query) {
		return nil, &mcp.Error{
			Code:    mcp.InvalidParams,
			Message: "at least one search parameter is required (name, style, brewery, or location)",
			Data: map[string]interface{}{
				"provided_params": args,
			},
		}
	}

	// Perform the search
	results, err := h.beerService.SearchBeers(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search beers: %w", err)
	}

	return h.formatBeerSearchResults(results)
}

// parseBeerSearchQuery extracts and validates search parameters for beer search.
func (h *ToolHandlers) parseBeerSearchQuery(args map[string]interface{}) (services.BeerSearchQuery, error) {
	query := services.BeerSearchQuery{}

	// Extract search parameters
	if name, ok := args["name"].(string); ok && name != "" {
		query.Name = name
	}
	if style, ok := args["style"].(string); ok && style != "" {
		query.Style = style
	}
	if brewery, ok := args["brewery"].(string); ok && brewery != "" {
		query.Brewery = brewery
	}
	if location, ok := args["location"].(string); ok && location != "" {
		query.Location = location
	}

	// Parse and validate limit
	limit, err := h.parseLimit(args)
	if err != nil {
		return query, err
	}
	query.Limit = limit

	return query, nil
}

// parseLimit extracts and validates the limit parameter from arguments.
func (h *ToolHandlers) parseLimit(args map[string]interface{}) (int, error) {
	var limit int
	var limitSet bool

	switch v := args["limit"].(type) {
	case float64:
		limit = int(v)
		limitSet = true
	case int:
		limit = v
		limitSet = true
	case string:
		parsedLimit, err := strconv.Atoi(v)
		if err != nil {
			return 0, &mcp.Error{
				Code:    mcp.InvalidParams,
				Message: "limit must be an integer",
			}
		}
		limit = parsedLimit
		limitSet = true
	case nil:
		// No limit provided, use default
	default:
		// Provided but not a valid type
		return 0, &mcp.Error{
			Code:    mcp.InvalidParams,
			Message: "limit must be an integer",
		}
	}

	// Default limit if not set
	if !limitSet {
		limit = defaultSearchLimit
	}

	// Validate limit range
	if limit <= 0 {
		return 0, &mcp.Error{
			Code:    mcp.InvalidParams,
			Message: "limit must be greater than zero",
		}
	}
	if limit > maxSearchLimit {
		// Cap at maximum limit instead of returning error
		limit = maxSearchLimit
	}

	return limit, nil
}

// hasAnyBeerSearchParam checks if any search criteria are provided.
func (h *ToolHandlers) hasAnyBeerSearchParam(query services.BeerSearchQuery) bool {
	return query.Name != "" || query.Style != "" || query.Brewery != "" || query.Location != ""
}

// formatBeerSearchResults formats the search results for display.
func (h *ToolHandlers) formatBeerSearchResults(results []*services.BeerSearchResult) (*mcp.ToolResult, error) {
	if len(results) == 0 {
		return &mcp.ToolResult{
			Content: []mcp.ToolContent{{
				Type: "text",
				Text: "No beers found matching your search criteria.",
			}},
		}, nil
	}

	// Format the response
	var response strings.Builder
	response.WriteString(fmt.Sprintf("**Found %d beer(s):**\n\n", len(results)))

	for i, beer := range results {
		response.WriteString(fmt.Sprintf("**%d. %s**\n", i+1, beer.Name))
		response.WriteString(fmt.Sprintf("- **Brewery:** %s\n", beer.Brewery))
		response.WriteString(fmt.Sprintf("- **Style:** %s\n", beer.Style))
		// No ABV, IBU, or Description fields in BeerSearchResult struct
		response.WriteString("\n")
	}

	return &mcp.ToolResult{
		Content: []mcp.ToolContent{{
			Type: "text",
			Text: response.String(),
		}},
	}, nil
}

// FindBreweries handles brewery search functionality.
func (h *ToolHandlers) FindBreweries(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
	query := parseBrewerySearchQuery(args)
	if !hasAnyBrewerySearchParam(query) {
		return nil, &mcp.Error{
			Code:    mcp.InvalidParams,
			Message: "at least one search parameter is required (name, location, city, state, or country)",
			Data: map[string]interface{}{
				"provided_params": args,
			},
		}
	}
	results, err := h.breweryService.SearchBreweries(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search breweries: %w", err)
	}
	if len(results) == 0 {
		return &mcp.ToolResult{
			Content: []mcp.ToolContent{{
				Type: "text",
				Text: "No breweries found matching your search criteria.",
			}},
		}, nil
	}
	return &mcp.ToolResult{
		Content: []mcp.ToolContent{{
			Type: "text",
			Text: formatBreweryResults(results),
		}},
	}, nil
}

func parseBrewerySearchQuery(args map[string]interface{}) services.BrewerySearchQuery {
	query := services.BrewerySearchQuery{}
	if name, ok := args["name"].(string); ok && name != "" {
		query.Name = name
	}
	if location, ok := args["location"].(string); ok && location != "" {
		query.Location = location
	}
	if city, ok := args["city"].(string); ok && city != "" {
		query.City = city
	}
	if state, ok := args["state"].(string); ok && state != "" {
		query.State = state
	}
	if country, ok := args["country"].(string); ok && country != "" {
		query.Country = country
	}
	if limitFloat, ok1 := args["limit"].(float64); ok1 {
		query.Limit = int(limitFloat)
	} else if limitStr, ok2 := args["limit"].(string); ok2 {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			query.Limit = limit
		}
	}
	if query.Limit <= 0 || query.Limit > maxSearchLimit {
		query.Limit = defaultSearchLimit
	}
	return query
}

func hasAnyBrewerySearchParam(query services.BrewerySearchQuery) bool {
	return query.Name != "" || query.Location != "" || query.City != "" || query.State != "" || query.Country != ""
}

func formatBreweryResults(results []*services.BrewerySearchResult) string {
	var response strings.Builder
	response.WriteString(fmt.Sprintf("**Found %d brewery(ies):**\n\n", len(results)))
	for i, brewery := range results {
		response.WriteString(fmt.Sprintf("**%d. %s**\n", i+1, brewery.Name))
		if brewery.BreweryType != "" {
			response.WriteString(fmt.Sprintf("- **Type:** %s\n", brewery.BreweryType))
		}
		if brewery.City != "" || brewery.State != "" || brewery.Country != "" {
			location := []string{}
			if brewery.City != "" {
				location = append(location, brewery.City)
			}
			if brewery.State != "" {
				location = append(location, brewery.State)
			}
			if brewery.Country != "" {
				location = append(location, brewery.Country)
			}
			response.WriteString(fmt.Sprintf("- **Location:** %s\n", strings.Join(location, ", ")))
		}
		if brewery.Website != "" {
			response.WriteString(fmt.Sprintf("- **Website:** %s\n", brewery.Website))
		}
		if brewery.Phone != "" {
			response.WriteString(fmt.Sprintf("- **Phone:** %s\n", brewery.Phone))
		}
		response.WriteString("\n")
	}
	return response.String()
}
