# Contributing to BrewSource MCP Server 🍺

Thank you for your interest in contributing to BrewSource MCP Server! This project aims to provide comprehensive brewing
 knowledge through the Model Context Protocol, and we welcome contributions from developers, brewers, and beer enthusiasts.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Code Guidelines](#code-guidelines)
- [Testing](#testing)
- [Documentation](#documentation)
- [Community](#community)

## Getting Started

### What We're Building

BrewSource MCP Server is an open-source project that provides AI assistants with access to brewing knowledge through the
 Model Context Protocol. We're currently in **Phase 1 MVP** focusing on:

- BJCP style guide integration
- Beer and brewery search capabilities
- Solid MCP protocol implementation
- Educational value for the community

### Types of Contributions We Welcome

- **🐛 Bug fixes** - Help us improve stability and reliability
- **✨ New features** - Implement tools and resources from our roadmap
- **📖 Documentation** - Improve guides, examples, and explanations
- **🧪 Testing** - Add test coverage and improve quality assurance
- **🔍 Code review** - Review pull requests and provide feedback
- **💡 Ideas** - Suggest improvements and new features
- **🍺 Brewing expertise** - Help ensure accuracy of brewing data and calculations

## Development Setup

### Prerequisites

Our development environment uses modern Kubernetes-native tools with automatic dependency management:

- **Nix** - For reproducible development environments (recommended)
- **direnv** - For automatic environment activation
- **Kind** - Local Kubernetes cluster (auto-installed via Nix)
- **Tilt** - Live development orchestration (auto-installed via Nix)
- **k9s** - Interactive Kubernetes CLI (auto-installed via Nix)
- **Git** - For version control

### Kubernetes-Native Development

1. **Fork and Clone**

   ```bash
   git clone https://github.com/YOUR_USERNAME/brewsource-mcp.git
   cd brewsource-mcp
   ```

2. **Start Development Environment**

   ```bash
   # All tools are automatically installed via Nix/direnv
   make up
   ```

3. **Access Development Services**
   - **Tilt Dashboard**: <http://localhost:10350>
   - **MCP Server**: <http://localhost:8080>
   - **PostgreSQL**: localhost:5432
   - **Redis**: localhost:6379

4. **Interactive Cluster Management**

   ```bash
   make k9s
   ```

5. **Run Tests**

   ```bash
   make test
   ```

6. **Stop Development Environment**

   ```bash
   make down
   ```

### Alternative: Manual Setup

If you prefer not to use Nix/direnv:

- **Go 1.21+** - Latest stable version recommended
- **Kind** - Local Kubernetes cluster
- **Tilt** - Development orchestration
- **k9s** - Interactive Kubernetes CLI
- **Docker** - For container builds

### Project Structure

```sh
brewsource-mcp/
├── app/                 # Application code
│   ├── cmd/server/      # Application entry point
│   ├── internal/        # Private application code
│   │   ├── mcp/         # MCP protocol implementation
│   │   ├── handlers/    # Tool and resource handlers
│   │   ├── models/      # Database models
│   │   └── services/    # Business logic
│   ├── pkg/             # Public library code
│   │   ├── bjcp/        # BJCP utilities
│   │   └── brewing/     # Brewing calculations
│   └── Dockerfile       # Container build
├── k8s/                 # Kubernetes manifests
├── docs/                # Documentation
├── .github/             # GitHub templates and workflows
├── Tiltfile             # Tilt configuration
└── shell.nix            # Nix development environment
```

## How to Contribute

### 1. Choose What to Work On

- **Good First Issues** - Look for issues labeled `good-first-issue`
- **Help Wanted** - Issues labeled `help-wanted` need community input
- **Phase 1 Features** - Check our [roadmap](README.md#roadmap) for MVP features
- **Documentation** - Always needs improvement and updates

### 2. Create an Issue (Optional but Recommended)

Before starting work, consider creating an issue to:

- Discuss your approach
- Get feedback from maintainers
- Avoid duplicate work
- Clarify requirements

### 3. Development Workflow

1. **Create a Branch**

   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/bug-description
   ```

2. **Make Your Changes**
   - Follow our [code guidelines](#code-guidelines)
   - Add tests for new functionality
   - Update documentation as needed

3. **Test Your Changes**

   ```bash
   # Run unit tests
   go test ./...

   # Run integration tests
   go test -tags=integration ./tests/...

   # Test MCP protocol compliance
   make test-mcp  # if we have this
   ```

4. **Commit Your Changes**

   ```bash
   git add .
   git commit -m "feat: add hop substitution tool"
   ```

5. **Push and Create Pull Request**

   ```bash
   git push origin feature/your-feature-name
   ```

### 4. Pull Request Process

- Fill out the PR template completely
- Link to related issues
- Ensure all tests pass
- Request review from maintainers
- Address feedback promptly

## Code Guidelines

### Go Standards

- Follow standard Go conventions and best practices
- Use `gofmt` for formatting (run `go fmt ./...`)
- Use `golint` and `go vet` for code quality
- Write idiomatic Go code

### Naming Conventions

```go
// Good
func GetBJCPStyle(styleCode string) (*BJCPStyle, error)
func (h *ToolHandlers) HandleBeerSearch(ctx context.Context, args map[string]interface{})

// Package names should be lowercase, single word
package bjcp
package brewing
```

### Error Handling

```go
// Always handle errors explicitly
result, err := service.GetStyle(styleCode)
if err != nil {
    return nil, fmt.Errorf("failed to get style %s: %w", styleCode, err)
}

// Use meaningful error messages
if styleCode == "" {
    return nil, errors.New("style code cannot be empty")
}
```

### MCP Protocol Compliance

```go
// Tool responses must follow MCP format
return &mcp.ToolResult{
    Content: []mcp.ToolContent{{
        Type: "text",
        Text: responseText,
    }},
}, nil

// Include error details in MCP error responses
return nil, &mcp.Error{
    Code:    mcp.InvalidParams,
    Message: "Invalid style code format",
    Data:    map[string]interface{}{"styleCode": styleCode},
}
```

### Database Interactions

```go
// Use contexts for database operations
func (r *StyleRepository) GetByCode(ctx context.Context, code string) (*Style, error) {
    var style Style
    err := r.db.GetContext(ctx, &style, "SELECT * FROM styles WHERE code = $1", code)
    return &style, err
}

// Use transactions for multi-step operations
tx, err := r.db.BeginTxx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

// ... operations ...

return tx.Commit()
```

### Brewing Calculations

```go
// Document brewing formulas with references
// IBU calculation using Tinseth formula
// Reference: https://www.realbeer.com/hops/research.html
func (h *HopAddition) CalculateIBU(batchSize float64, originalGravity float64) float64 {
    utilizationFactor := calculateUtilization(h.BoilTime, originalGravity)
    return (h.AlphaAcid * h.Amount * utilizationFactor * 74.89) / batchSize
}
```

## Testing

### Unit Tests

- Write tests for all new functions
- Use table-driven tests for multiple scenarios
- Mock external dependencies

```go
func TestBJCPStyle_IsValid(t *testing.T) {
    tests := []struct {
        name     string
        style    BJCPStyle
        expected bool
    }{
        {"valid IPA", BJCPStyle{Code: "21A", Name: "American IPA"}, true},
        {"empty code", BJCPStyle{Code: "", Name: "Test"}, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.style.IsValid()
            if got != tt.expected {
                t.Errorf("IsValid() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### Integration Tests

- Test MCP protocol interactions
- Test database operations
- Test external API integrations

```go
//go:build integration
// +build integration

func TestMCPServer_BJCPLookup(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()

    // Test BJCP lookup tool
    result, err := server.CallTool("bjcp_lookup", map[string]interface{}{
        "style_code": "21A",
    })

    require.NoError(t, err)
    assert.Contains(t, result.Content[0].Text, "American IPA")
}
```

### Test Data

- Use consistent test data for brewing calculations
- Mock external API responses
- Clean up test database after tests

## Documentation

### Code Comments

```go
// CalculateIBU calculates International Bitterness Units using the Tinseth formula.
// It takes into account hop utilization based on boil time and wort gravity.
//
// Parameters:
//   - batchSize: Final volume in gallons
//   - originalGravity: Specific gravity of wort (e.g., 1.050)
//
// Returns the calculated IBU value.
func (h *HopAddition) CalculateIBU(batchSize float64, originalGravity float64) float64 {
```

### Documentation Updates

When adding new features:

1. Update relevant documentation files
2. Add examples to README if applicable
3. Update API documentation
4. Add user stories if needed
5. Update test plan if needed

### Commit Messages

Use conventional commit format:

```
feat: add hop substitution tool
fix: resolve BJCP style lookup caching issue
docs: update installation guide for macOS
test: add integration tests for beer search
refactor: improve error handling in MCP handlers
```

## Community

### Getting Help

- **GitHub Issues** - For bugs and feature requests
- **Discussions** - For general questions and ideas
- **Documentation** - Check our comprehensive docs first

### Code Review

- All contributions require code review
- Be respectful and constructive in feedback
- Focus on code quality, not personal style
- Explain the "why" behind suggestions

### Recognition

Contributors are recognized in:

- Release notes for significant contributions
- README acknowledgments
- GitHub contributor graphs

## Development Phases

### Phase 1 (Current) - MVP Contributions

Priority areas for contributions:

- **BJCP Integration** - Improve style data accuracy
- **MCP Tools** - Enhance existing tools (`bjcp_lookup`, `search_beers`, `find_breweries`)
- **Testing** - Add comprehensive test coverage
- **Documentation** - Improve setup and usage guides

### Future Phases

We welcome discussions about future features, but current focus is on MVP stability.

## Questions?

- Create a GitHub issue with the `question` label
- Check existing issues and documentation first
- Be specific about what you're trying to accomplish

Thank you for contributing to BrewSource MCP Server! 🍺
