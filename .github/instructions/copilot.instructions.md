# GitHub Copilot Instructions for BrewSource MCP Server

This document provides context and instructions for GitHub Copilot to assist with BrewSource MCP Server development.

## Project Overview

**BrewSource MCP Server** is an open-source Model Context Protocol (MCP) server that provides AI assistants with access to comprehensive brewing knowledge and tools. The project is built in Go and focuses on educational value and community contribution.

### Current Phase: Phase 1 MVP
- **BJCP Style Guide Integration** - Style lookup and information retrieval
- **Beer & Brewery Search** - Basic commercial beer and brewery databases
- **MCP Protocol Implementation** - HTTP connection modes
- **Public API Layer** - Three core tools: `bjcp_lookup`, `search_beers`, `find_breweries`

## Architecture & Technology Stack

### Core Technologies
- **Language**: Go 1.24+
- **Protocol**: Model Context Protocol (MCP) - JSON-RPC 2.0 over HTTP
- **Database**: PostgreSQL with optional Redis caching
- **Communication**: HTTP modes for different MCP clients

### Project Structure
```
brewsource-mcp/
├── cmd/server/           # Application entry point
├── internal/             # Private application code
│   ├── mcp/             # MCP protocol implementation
│   ├── handlers/        # Tool and resource handlers
│   ├── models/          # Database models
│   └── services/        # Business logic services
├── pkg/                 # Public library code
│   ├── bjcp/           # BJCP style guide utilities
│   └── brewing/        # Brewing calculations and formulas
├── docs/               # Comprehensive project documentation
└── .github/            # GitHub templates and workflows
```

## Code Style & Conventions

### Go Standards
- Follow standard Go conventions and idioms
- Use `gofmt` for consistent formatting
- Implement proper error handling with wrapped errors
- Use contexts for database operations and timeouts
- Write idiomatic Go code with clear naming

### MCP Protocol Patterns
```go
// Tool handler signature
func (h *ToolHandlers) ToolName(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error)

// Successful tool response
return &mcp.ToolResult{
    Content: []mcp.ToolContent{{
        Type: "text",
        Text: responseText,
    }},
}, nil

// Error response with proper MCP error format
return nil, &mcp.Error{
    Code:    mcp.InvalidParams,
    Message: "descriptive error message",
    Data:    map[string]interface{}{"details": "additional context"},
}
```

### Database Patterns
```go
// Use contexts and parameterized queries
func (r *Repository) GetByCode(ctx context.Context, code string) (*Model, error) {
    var model Model
    err := r.db.GetContext(ctx, &model, "SELECT * FROM table WHERE code = $1", code)
    if err != nil {
        return nil, fmt.Errorf("failed to get model by code %s: %w", code, err)
    }
    return &model, nil
}
```

### Brewing Domain Patterns
```go
// Document brewing formulas with references
// IBU calculation using Tinseth formula
// Reference: https://www.realbeer.com/hops/research.html
func (h *HopAddition) CalculateIBU(batchSize float64, originalGravity float64) float64 {
    utilizationFactor := calculateUtilization(h.BoilTime, originalGravity)
    return (h.AlphaAcid * h.Amount * utilizationFactor * 74.89) / batchSize
}
```

## Domain Knowledge

### Brewing Terminology
- **BJCP**: Beer Judge Certification Program - standardized beer style guidelines
- **IBU**: International Bitterness Units - measure of beer bitterness
- **SRM**: Standard Reference Method - beer color measurement
- **ABV**: Alcohol By Volume - alcohol content percentage
- **OG**: Original Gravity - sugar content before fermentation
- **FG**: Final Gravity - sugar content after fermentation

### BJCP Style Codes
- Format: `[Category][Subcategory]` (e.g., "21A" for American IPA)
- Categories 1-34 with subcategories A, B, C, etc.
- Special categories for historical and specialty beers

### Brewing Calculations
- **IBU**: Hop utilization based on boil time and gravity
- **SRM**: Color calculation from grain bill
- **ABV**: Calculated from OG and FG
- **Efficiency**: Brewhouse efficiency affects gravity calculations

## MCP Implementation Guidelines

### Tool Development
- **Input Validation**: Always validate tool parameters
- **Error Handling**: Return proper MCP error responses
- **Response Format**: Use consistent content structure
- **Documentation**: Include tool descriptions and parameter schemas

### Resource Implementation
- **URI Patterns**: Use consistent URI schemes (e.g., `bjcp://styles/21A`)
- **Content Types**: Properly set MIME types for resources
- **Caching**: Implement appropriate caching strategies
- **Error Responses**: Handle missing resources gracefully

### Connection Handling
- **HTTP**: For web-based MCP clients
- **Protocol Compliance**: Strict JSON-RPC 2.0 adherence
- **Connection Management**: Proper cleanup and error handling

## Testing Guidelines

### Unit Tests
```go
func TestBJCPStyle_IsValid(t *testing.T) {
    tests := []struct {
        name     string
        style    BJCPStyle
        expected bool
    }{
        {"valid IPA", BJCPStyle{Code: "21A", Name: "American IPA"}, true},
        {"empty code", BJCPStyle{Code: "", Name: "Test"}, false},
        {"invalid format", BJCPStyle{Code: "ABC", Name: "Test"}, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.style.IsValid(); got != tt.expected {
                t.Errorf("IsValid() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### Integration Tests
- Test MCP protocol interactions end-to-end
- Use test databases with cleanup
- Mock external API dependencies
- Test both HTTP modes

## Security Considerations

### Input Validation
- Validate all MCP tool parameters
- Sanitize database inputs
- Limit response sizes to prevent memory exhaustion
- Implement rate limiting for tool calls

### Database Security
- Use parameterized queries exclusively
- Implement connection pooling with limits
- Avoid exposing sensitive information in errors
- Regular dependency updates

## Development Priorities

### Phase 1 Focus Areas
1. **Core MCP Tools**: `bjcp_lookup`, `search_beers`, `find_breweries`
2. **Protocol Stability**: Robust HTTP implementations
3. **Data Accuracy**: Correct BJCP style information and brewing calculations
4. **Testing Coverage**: Comprehensive unit and integration tests
5. **Documentation**: Clear setup guides and API documentation

### Code Quality Priorities
1. **Error Handling**: Comprehensive error responses with context
2. **Performance**: Efficient database queries and caching
3. **Maintainability**: Clean, documented code following Go best practices
4. **Testing**: High test coverage with meaningful test cases
5. **Security**: Input validation and secure database operations

## Common Patterns to Suggest

### Error Handling
```go
// Wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to process BJCP lookup for style %s: %w", styleCode, err)
}

// MCP error responses
return nil, &mcp.Error{
    Code:    mcp.InvalidParams,
    Message: fmt.Sprintf("invalid style code format: %s", styleCode),
}
```

### Logging
```go
// Structured logging with context
log.WithFields(log.Fields{
    "style_code": styleCode,
    "client_id":  clientID,
}).Info("BJCP style lookup requested")
```

### Configuration
```go
// Environment-based configuration
type Config struct {
    DatabaseURL string `env:"DATABASE_URL,required"`
    RedisURL    string `env:"REDIS_URL"`
    Port        int    `env:"PORT" envDefault:"8080"`
}
```

## Documentation Standards

- **Code Comments**: Explain complex brewing calculations and business logic
- **Function Documentation**: Go doc format for all public functions
- **README Updates**: Keep installation and usage instructions current
- **API Documentation**: Document all MCP tools and resources
- **Examples**: Provide practical usage examples

## Contribution Guidelines

- **Incremental Changes**: Small, focused commits and pull requests
- **Test Coverage**: Include tests for all new functionality
- **Documentation**: Update relevant documentation for changes
- **Code Review**: All changes require review before merging
- **Breaking Changes**: Avoid in Phase 1, document thoroughly if necessary

---

**Instructions for Copilot**: When suggesting code for this project, prioritize:
1. **MCP Protocol Compliance** - Ensure all suggestions follow MCP standards
2. **Go Best Practices** - Idiomatic Go code with proper error handling
3. **Brewing Domain Accuracy** - Suggest accurate brewing calculations and terminology
4. **Security First** - Always include input validation and secure patterns
5. **Educational Value** - Code should be clear and well-documented for learning

Focus on Phase 1 MVP features and maintain consistency with existing patterns and architecture.
