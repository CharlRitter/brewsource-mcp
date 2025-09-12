# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

**BrewSource MCP Server** is a Go-based Model Context Protocol (MCP) server that provides AI assistants with access to brewing knowledge and tools. It's currently in Phase 1 MVP, focusing on BJCP style lookups, beer search, and brewery discovery.

The project demonstrates best practices for MCP protocol implementation while serving the brewing community with accurate, accessible information through AI assistants.

## Development Commands

### Primary Development Workflow
```bash
# Start complete development environment (Kubernetes + Tilt)
make up

# Stop development environment (keeps cluster running)
make down

# Clean up everything (delete cluster)
make clean

# Open k9s cluster explorer
make k9s
```

### Building and Testing
```bash
# Build the Go application
make build

# Run all tests
make test

# Run linter
make lint

# Run linter with auto-fix
make lint-fix

# Run security scans
make security

# Format code
make format
```

### Database Operations
```bash
# The database is automatically migrated and seeded during startup
# Manual seeding (if needed): make seed-data

# Connect to database
psql-dev  # (alias set in .envrc)

# Connect to Redis
redis-cli-dev  # (alias set in .envrc)
```

## Architecture & Code Structure

### High-Level Architecture
This is a **layered MCP server architecture** with:
- **MCP Protocol Layer** (`internal/mcp/`) - Handles JSON-RPC 2.0 over WebSocket/stdio
- **Handler Layer** (`internal/handlers/`) - Tool and resource request handlers
- **Service Layer** (`internal/services/`) - Business logic for beer/brewery operations
- **Data Layer** (`internal/models/`, `pkg/data/`) - Database models and BJCP data

### Key Components

**MCP Server Core** (`internal/mcp/server.go`):
- Dual-mode server supporting both WebSocket and stdio connections
- Implements JSON-RPC 2.0 protocol with proper error handling
- Dynamic tool and resource handler registration system

**Tool Handlers** (`internal/handlers/tools.go`):
- `bjcp_lookup` - BJCP style information by code or name
- `search_beers` - Multi-criteria beer search with pagination
- `find_breweries` - Geographic brewery search

**Resource Handlers** (`internal/handlers/resources.go`):
- URI-based resource system (`bjcp://`, `beers://`, `breweries://`)
- JSON content with proper MIME type handling

**Data Management**:
- **Hybrid approach**: BJCP reference data in JSON files, dynamic data in PostgreSQL
- **BJCP Service** (`pkg/data/`) - In-memory BJCP style guide operations
- **Database Services** (`internal/services/`) - PostgreSQL operations with optional Redis caching

### Project Structure Patterns
```
app/
├── cmd/server/          # Application entry point
├── internal/            # Private application code
│   ├── mcp/            # MCP protocol implementation
│   ├── handlers/       # Tool and resource handlers
│   ├── models/         # Database models and migrations
│   └── services/       # Business logic services
├── pkg/                # Public/reusable packages
│   ├── data/          # BJCP data structures and operations
│   └── brewing/       # Brewing calculations (future phases)
└── bin/               # Compiled binaries
```

## Development Environment

### Technology Stack
- **Go 1.24+** - Main language with standard project layout
- **PostgreSQL** - Primary database for beers/breweries
- **Redis** - Optional caching layer
- **Kubernetes + Tilt** - Development environment with live reload
- **Kind** - Local Kubernetes cluster

### Environment Setup
The project uses **direnv** (`.envrc`) for automatic environment configuration:
- Database URLs and connection settings
- Kubernetes context and aliases
- Helpful command aliases (`k`, `kl`, `kg`, etc.)

### Key Environment Variables
```bash
DATABASE_URL="postgres://brewsource_user:brewsource_pass@localhost:5432/brewsource_dev?sslmode=disable"
REDIS_URL="redis://localhost:6379/0"
LOG_LEVEL="debug"
PORT="8080"
```

## Development Guidelines

### MCP Protocol Patterns
Always follow the MCP JSON-RPC 2.0 specification:
- Proper tool parameter validation with `mcp.InvalidParams` errors
- Structured tool results with `mcp.ToolResult` and `mcp.ToolContent`
- Resource handlers with URI pattern matching
- Context handling for timeouts and cancellation

### Error Handling Patterns
```go
// MCP-compliant error responses
return nil, &mcp.Error{
    Code:    mcp.InvalidParams,
    Message: "descriptive error message",
    Data:    map[string]interface{}{"context": "details"},
}
```

### Database Patterns
- Always use contexts for database operations
- Parameterized queries to prevent SQL injection
- Proper connection pooling with configured limits
- Wrap errors with context using `fmt.Errorf`

### Testing Approach
- Comprehensive unit tests in `*_test.go` files
- Integration tests for MCP protocol interactions
- Mock services for isolated testing
- Test coverage for error conditions and edge cases

## Brewing Domain Knowledge

### BJCP (Beer Judge Certification Program)
- Style codes format: `[Category][Subcategory]` (e.g., "21A" for American IPA)
- Complete style database with vitals, descriptions, and commercial examples
- Validation: Categories 1-34, subcategories A-Z

### Key Brewing Measurements
- **IBU** - International Bitterness Units
- **SRM** - Standard Reference Method (color)
- **ABV** - Alcohol By Volume percentage
- **OG/FG** - Original/Final Gravity

### Data Sources
- BJCP data stored as version-controlled JSON in `app/data/`
- Commercial beer/brewery data in PostgreSQL database
- Hybrid approach allows reliable reference data with dynamic search capability

## Common Development Tasks

### Adding New MCP Tools
1. Define tool schema in `handlers/tools.go` `GetToolDefinitions()`
2. Implement handler function with proper signature
3. Register handler in `RegisterToolHandlers()`
4. Add validation and error handling
5. Write unit tests covering success and error cases

### Database Schema Changes
1. Update models in `internal/models/`
2. Add migration logic to `MigrateDatabase()`
3. Update seed data if needed
4. Test with `make clean && make up`

### Adding New Resources
1. Define resource in `handlers/resources.go` `GetResourceDefinitions()`
2. Implement handler function for URI pattern
3. Register pattern in `RegisterResourceHandlers()`
4. Handle content type and format properly

## Server Modes

### WebSocket Mode (Default)
```bash
# Start server (default mode)
./app/bin/brewsource-mcp

# Or explicitly
./app/bin/brewsource-mcp -mode=websocket -port=8080
```
- Accessible at `ws://localhost:8080/mcp`
- Health check at `http://localhost:8080/health`
- Server info at `http://localhost:8080/`

### Stdio Mode
```bash
# For MCP clients using stdio
./app/bin/brewsource-mcp -mode=stdio
```
- Reads JSON-RPC messages from stdin
- Writes responses to stdout
- Used for local MCP integrations

## Phase 1 MVP Tools

### `bjcp_lookup`
Look up BJCP beer styles by code (e.g., "21A") or name (e.g., "American IPA").
- Returns complete style information including vitals, descriptions, examples
- Validates style codes against BJCP format
- Case-insensitive name matching

### `search_beers`
Search commercial beer database with filters:
- `name` - Beer name search
- `style` - Beer style filter
- `brewery` - Brewery name filter
- `location` - Geographic filter
- `limit` - Result limit (max 100, default 20)

### `find_breweries`
Search brewery database with filters:
- `name` - Brewery name search
- `location` - General location search
- `city`/`state`/`country` - Specific geographic filters
- `limit` - Result limit (max 100, default 20)

## Troubleshooting

### Common Issues
- **Database connection failed**: Check `DATABASE_URL` and ensure PostgreSQL is running
- **Kubernetes cluster issues**: Try `make clean && make up` to recreate
- **Build failures**: Ensure Go 1.24+ and check `go mod download`
- **Port conflicts**: Check if ports 8080, 5432, 6379 are available

### Useful Commands
```bash
# Check cluster status
kubectl get pods -n brewsource-dev

# View logs
kubectl logs -n brewsource-dev deployment/brewsource-mcp

# Database connection test
psql "${DATABASE_URL}" -c "SELECT 1;"

# Redis connection test
redis-cli -h localhost -p 6379 ping
```

### Development Aliases (from .envrc)
```bash
k      # kubectl -n brewsource-dev
kg     # kubectl get -n brewsource-dev
kl     # kubectl logs -n brewsource-dev
kd     # kubectl describe -n brewsource-dev
up     # tilt up
down   # tilt down
```
