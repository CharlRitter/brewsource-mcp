# BrewSource MCP Server 🍺

A comprehensive Model Context Protocol (MCP) server for brewing resources, built with Go.

## What is This?

**BrewSource MCP** is a specialized MCP server that gives AI assistants access to essential brewing knowledge and tools. Currently in **Phase 1 MVP**, it focuses on core public resources:

- 🔍 **Beer & Brewery Discovery** - Search basic commercial beer and brewery databases
- 📖 **BJCP Style Guide** - Complete beer style database with lookup capabilities
- 🔧 **Public API Layer** - Three core MCP tools for essential brewing queries

Future phases will expand with ingredient databases, personal analytics, recipe builders, and premium Brewfather integration.

## Understanding MCP (Model Context Protocol)

**MCP is a standardized way for AI assistants to access external tools and data.** Instead of being limited to their training data, AI models can:

- **Call external APIs** (like Brewfather, BJCP data)
- **Access databases** (brewery catalogs, ingredient databases)
- **Perform calculations** (brewing formulas, recipe scaling)
- **Retrieve real-time data** (current beer availability, events)

**Our MCP server** exposes **resources** (data) and **tools** (functions) that AI assistants can use to provide expert brewing assistance.

### MCP Resources (Data Access) - Phase 1 MVP
- `bjcp://styles` - Complete BJCP style guide (basic lookup)
- `breweries://directory` - Basic brewery database (name, location)
- `beers://catalog` - Basic commercial beer database (name, style, brewery)

*Note: Enhanced resources and ingredient databases will be added in future phases.*

### MCP Tools (Functions) - Phase 1 MVP
- `bjcp_lookup` - Get detailed BJCP style information by code
- `search_beers` - Search commercial beer catalog by name, style, brewery
- `find_breweries` - Find breweries by location or name

*Note: Additional tools will be released in future phases as outlined in the roadmap below.*

## Hybrid Data Storage Approach

BrewSource MCP uses a hybrid data storage strategy:

- **BJCP styles and reference data** are stored as version-controlled JSON files in `app/data/`.
- **Application data** (beers, breweries, users, etc.) is stored in a PostgreSQL database.

## Project Structure

```
brewsource-mcp/
├── cmd/server/           # Main application entry point
├── docs/                # Project documentation
│   ├── project/         # Project overview and architecture
│   └── testing/         # Test plans and user stories
├── internal/
│   ├── mcp/             # MCP protocol implementation
│   ├── handlers/        # Tool and resource handlers
│   ├── models/          # Database models
│   └── services/        # Business logic services
├── pkg/
│   ├── bjcp/           # BJCP style guide utilities
│   └── brewing/        # Brewing calculations and formulas
├── go.mod              # Go module definition
├── .envrc              # Development environment configuration
└── README.md           # This file
```

## 📚 Documentation


For comprehensive project documentation, see the **[docs/](docs/)** directory:

- **[Seeding Guide](docs/SEEDING.md)** - How to populate the database with sample breweries and beers

- **[Project Overview](docs/project/PROJECT_OVERVIEW.md)** - Vision, goals, and technical architecture
- **[User Stories](docs/testing/USER_STORIES.md)** - Detailed feature specifications in Gherkin syntax
- **[Test Plan](docs/testing/TEST_PLAN.md)** - Comprehensive testing strategy and requirements

## Quick Start

**The "git clone && make up" experience:**

```bash
git clone <repository-url>
cd brewsource-mcp
make up
```

That's it! This will:
- Create a local Kubernetes cluster with Kind
- Start all services (PostgreSQL, Redis, MCP server)
- Set up live-reload development with Tilt
- Forward ports for local access

### Prerequisites

**Option 1: Using Nix (Recommended)**
```bash
nix-shell  # Everything is included
```

**Option 2: Manual Installation**
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [Tilt](https://docs.tilt.dev/install.html)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [k9s](https://k9scli.io/topics/install/) (optional)
- [direnv](https://direnv.net/docs/installation.html) (optional)


### Development Workflow

```bash
# Start everything
make up

# Seed the database with sample data (breweries and beers)
make seed-data

# Explore cluster
make k9s

# Stop development (cluster stays)
make down

# Clean up everything
make clean
```

### Services

Once running, you'll have:
- **MCP Server**: http://localhost:8080
- **PostgreSQL**: localhost:5432 (user: brewsource_user, db: brewsource)
- **Redis**: localhost:6379
- **Tilt Dashboard**: http://localhost:10350

## MCP Integration

To add BrewSource MCP to your MCP configuration, include the following in your `mcp.json` or equivalent config:

```json
"brewsource": {
  "type": "stdio",
  "command": "node",
  "args": ["/home/charl/workspace/brewsource-mcp/bridge/index.js"]
}
```

This enables MCP clients to access BrewSource tools and resources via the MCP protocol.

### 1. Clone and Setup
```bash
git clone <repository-url>
cd brewsource-mcp
make up
```

### 2. Configure Environment
```bash
# Environment is automatically configured via .envrc and direnv
# No manual configuration needed for development
```

### 3. Create Database
```bash
# Create the database
createdb brewsource

# Or using psql
psql -c "CREATE DATABASE brewsource;"
```


### 4. Build and Run
```bash
# Build the server
make build

# Run development environment (Kubernetes + Tilt)
make up

# (Optional) Seed the database with sample data
make seed-data

# Access the Tilt dashboard at http://localhost:10350
# Use k9s for interactive cluster management
make k9s
```

### 5. Test the Server
```bash
# Health check
curl http://localhost:8080/health

# Server info
curl http://localhost:8080/

# WebSocket endpoint for MCP clients
# ws://localhost:8080/mcp
```

### 6. Run Tests
```bash
# Run all tests
make test
```

## Development Guide

### Adding New Tools

1. **Define the tool function** in `internal/handlers/tools.go`:
```go
func (h *ToolHandlers) MyNewTool(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
    // Your tool implementation
    return &mcp.ToolResult{
        Content: []mcp.ToolContent{{
            Type: "text",
            Text: "Tool result",
        }},
    }, nil
}
```

2. **Register the tool** in `RegisterToolHandlers()`:
```go
server.RegisterToolHandler("my_new_tool", h.MyNewTool)
```

3. **Add tool definition** in `getToolDefinition()` method in `internal/mcp/server.go`

### Adding New Resources

1. **Create resource handler** in `internal/handlers/resources.go`:
```go
func (h *ResourceHandlers) HandleMyResource(ctx context.Context, uri string) (*mcp.ResourceContent, error) {
    // Your resource implementation
    return &mcp.ResourceContent{
        URI:      uri,
        MimeType: "application/json",
        Text:     "resource data",
    }, nil
}
```

2. **Register the resource** in `RegisterResourceHandlers()`:
```go
server.RegisterResourceHandler("my://resource/*", h.HandleMyResource)
```

### Brewing Calculations

The `pkg/brewing` package contains all brewing formulas and calculations:

```go
// Calculate IBU from hop schedule
hopSchedule := brewing.HopSchedule{
    Additions: []brewing.HopAddition{
        {Name: "cascade", Amount: 1.0, Time: 60, AlphaAcid: 5.5},
    },
}
ibu := hopSchedule.CalculateIBU(5.0, 1.050) // 5 gallon batch

// Calculate beer color (SRM)
grainBill := brewing.GrainBill{
    Grains: []brewing.GrainEntry{
        {Name: "2-row", Amount: 8.0},
        {Name: "crystal_60", Amount: 1.0},
    },
}
srm := grainBill.CalculateSRM(5.0) // 5 gallon batch
```

### BJCP Style Guide

The `pkg/bjcp` package manages beer style data:

```go
// Load and search styles
styleGuide := bjcp.NewStyleGuide()
styleGuide.LoadFromJSON(bjcpData)

// Get specific style
style, err := styleGuide.GetStyle("21A") // American IPA

// Search styles
results := styleGuide.SearchStyles(bjcp.StyleSearchQuery{
    ABVMin: 5.0,
    ABVMax: 7.0,
    IBUMin: 40,
})
```

## Phase 1 Implementation Status ✅

**BrewSource MCP** is currently in Phase 1 MVP with the following implemented features:

### ✅ Core MCP Tools
- **`bjcp_lookup`** - Look up BJCP beer styles by code (e.g., "21A") or name
- **`search_beers`** - Search commercial beers by name, style, brewery, or location
- **`find_breweries`** - Find breweries by name, location, city, state, or country

### ✅ MCP Resources
- **`bjcp://styles`** - Complete BJCP style guidelines database
- **`bjcp://styles/{code}`** - Individual style details (e.g., bjcp://styles/21A)
- **`bjcp://categories`** - List of all BJCP categories
- **`beers://catalog`** - Commercial beer database
- **`breweries://directory`** - Brewery directory

### ✅ Infrastructure
- **WebSocket & Stdio Support** - Multiple connection modes for different MCP clients
- **PostgreSQL Database** - Persistent storage with proper indexing
- **Redis Caching** - Optional caching layer for improved performance
- **Seed Data** - Pre-populated with BJCP styles, breweries, and commercial beers
- **Comprehensive Testing** - Unit tests for brewing calculations and BJCP utilities

### ✅ Developer Experience
- **Makefile** - Common development tasks (`make help` to see all commands)
- **Environment Configuration** - Managed via .envrc and direnv
- **API Documentation** - Clear tool and resource schemas
- **Code Quality** - Proper error handling, logging, and Go best practices

## Architecture Deep Dive

### MCP Protocol Flow

1. **Client Connection**: MCP client connects via WebSocket or stdio
2. **Initialization**: Client and server exchange capabilities
3. **Resource/Tool Discovery**: Client can list available resources and tools
4. **Request/Response**: Client calls tools or requests resources
5. **JSON-RPC 2.0**: All communication uses JSON-RPC 2.0 format

### Database Design

The database schema supports:
- **Breweries** with location and contact information
- **Beers** linked to breweries with style classifications
- **BJCP Styles** with complete style guidelines
- **Ingredients** with type-specific properties (JSON fields)
- **Recipes** with ingredient lists and calculations

### Caching Strategy

- Redis caches frequently accessed data (BJCP styles, ingredient lookups)
- Database queries are optimized with proper indexes
- Static data (style guide) is loaded once at startup

## Troubleshooting

### Common Issues

**Database Connection Errors**
```bash
# Check if PostgreSQL is running
pg_isready

# Verify database exists
psql -l | grep brewsource

# Test connection string
psql "your-database-url-here"
```

**Port Already in Use**
```bash
# Check what's using port 8080
lsof -i :8080

# Run on different port
./bin/brewsource-mcp -mode=websocket -port=8081
```

**Missing Environment Variables**
```bash
# Check your environment variables (set via .envrc)
echo $DATABASE_URL
```

**Build Errors**
```bash
# Clean and rebuild
make clean
make build
```

### Getting Help

- Check the **[docs/](docs/)** directory for detailed documentation
- Review **[User Stories](docs/testing/USER_STORIES.md)** for feature specifications
- See **[Test Plan](docs/testing/TEST_PLAN.md)** for testing requirements
- Open an issue using our **[GitHub templates](.github/ISSUE_TEMPLATE/)**

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Expanding or Correcting Datasets

The core datasets for beers and breweries are defined as Go source files:

- [Beers dataset (`SeedBeer`)](app/internal/services/beer_schema.go)
- [Breweries dataset (`Brewery`)](app/internal/services/brewery_schema.go)
- [BJCP styles (beer, mead, cider, special ingredients) JSON](app/data/)

To expand the beer or brewery data (add new entries or fix errors), edit the relevant Go file and open a Pull Request with your changes. For BJCP style data, update the appropriate JSON file in `app/data/` and submit a PR. Please ensure your changes are well-formatted and include a clear description of the update.

### Code Standards

- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add comments for complex brewing calculations
- Include unit tests for new functions
- Update documentation for new features

## Roadmap

We follow an Agile development methodology with iterative releases and continuous feedback to ensure the platform evolves effectively to meet user needs.

**Overall Goal:** To build a comprehensive and evolving platform that serves as a central resource for beer enthusiasts, brewers, and the wider beer community.

**Guiding Principles:**
- **User-Centric Design:** Prioritise features that provide the most value to users
- **Iterative Development:** Release functional increments regularly to gather feedback
- **Scalability:** Design architecture to handle future growth and new features
- **Data Accuracy:** Ensure reliability and up-to-date nature of all beer-related data

### Phase 1: Minimum Viable Product (MVP) - Core Public Resources ✅

**Goal:** Launch a foundational public platform with essential beer knowledge and search capabilities to validate core concepts and attract initial users.

**MVP Definition:** An MCP server providing searchable BJCP style guidelines and a basic commercial beer catalogue.

**Key Features:**
- [x] **BJCP Style Guide Integration (Basic):** Style lookup by number or name (displaying basic characteristics like ABV, IBU, colour)
- [x] **Beer & Brewery Catalogue (Basic):** Searchable commercial beer database (name, style) and basic brewery directory (name, location)
- [x] **MCP Tools - Public Layer (Basic):** `bjcp_lookup`, `search_beers`, `find_breweries`
- [x] **Manual Data Input:** Limited initial set of styles, beers, and breweries
- [x] **WebSocket & Stdio Support:** Multiple connection modes for different MCP clients

**Outcome:** A functional, publicly accessible MCP Beer Server with core BJCP style lookup and basic beer/brewery search capabilities.

### Phase 2: Expanding Public Resources & Usability 🚧

**Goal:** Enhance the public resources with more detailed information and improved search capabilities, laying the groundwork for future features.

**Key Features:**
- [ ] **BJCP Style Guide Integration (Enhanced):**
  - Style comparison (2-3 styles side-by-side)
  - Detailed style search (by colour, ABV, hop character)
  - **Multi-source BJCP JSON support:** Load and query beer, mead, cider, and special ingredients from separate JSON files (e.g., `bjcp_2021_beer.json`, `bjcp_2015_mead.json`, `bjcp_2025_cider.json`, `bjcp_2015_special_ingredients.json`).
- [ ] **Brewing Ingredients Database (Basic):** Malt substitution charts, hop comparison (basic profiles), yeast strains database
- [ ] **Beer & Brewery Catalogue (Enhanced):** Beer-brewery linking, availability info (simple "available" flag)
- [ ] **MCP Tools - Public Layer (Expanded):** `bjcp_compare`, `ingredient_substitute`, `ingredient_compare`, `brewery_beers`

**Outcome:** A more robust public resource with expanded BJCP details, initial ingredient information (including yeast), and better interconnected beer/brewery data.

### Phase 3: Premium/Personal Features - Brewfather Integration & Personal Analytics

**Goal:** Introduce the first premium features, focusing on direct value for active brewers by integrating with Brewfather and offering basic personal analytics.

**Key Features:**
- [ ] **Brewfather Integration (Basic):** Inventory sync (pull fermentables, hops, yeast from Brewfather - read-only)
- [ ] **Personal Analytics (Basic):** Brewing trends (number of batches brewed, most used styles from synced data)
- [ ] **User Authentication & Profiles:** Secure user registration and login for premium features, basic user profile management

**Outcome:** Launch of the first premium tier, offering tangible value to brewers through Brewfather inventory synchronisation and initial personal analytics.

### Phase 4: Advanced Features - Interactive Recipe Builder & Community Engagement

**Goal:** Introduce an initial version of the interactive recipe builder and foster community engagement through recipe sharing and event listings.

**Key Features:**
- [ ] **Interactive Recipe Builder (Basic):** Style-guided creation, ingredient-based suggestions from Brewfather inventory
- [ ] **Community Features (Basic):** Brewery & beer festival events (manual input initially)
- [ ] **MCP Tools - New Endpoints:** `recipe_generate` (interactive recipe generation wizard)

**Outcome:** A nascent interactive recipe builder with inventory-aware suggestions and the beginning of community features.

### Phase 5: Continuous Improvement & Expansion (Future)

**Goal:** Iteratively enhance existing features, introduce more advanced functionalities, and respond to user feedback to continually grow the platform.

**Key Areas for Future Development:**
- **Advanced BJCP Features:** Fuzzy matching, historical style data, regional variations
- **Comprehensive Ingredients Database:** Detailed search by flavour/brewing properties, external data provider integration
- **Enhanced Beer Catalogue:** Local discovery, Untappd integration
- **Advanced Recipe Generation:** Water chemistry, equipment profiles, cost optimization, deep Brewfather integration
- **Premium Analytics:** Inventory optimization, detailed cost analysis, seasonal recommendations
- **Community Features:** Clone recipe database, food pairing suggestions, seasonal brewing recommendations

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [BJCP](https://www.bjcp.org/) for the comprehensive beer style guidelines
- [Open Brewery DB](https://openbrewerydb.org/) for brewery data
- [Model Context Protocol](https://modelcontextprotocol.io/) for the MCP specification
- The homebrewing community for inspiration and knowledge sharing


## Current ToDos
- Add more seed data (styles, beers, breweries)
- Move all seed data to `app/data/` JSON files
- Load seed data into cache on startup
---

**Happy Brewing!** 🍺
