# BrewSource MCP Server 🍺

A comprehensive Model Context Protocol (MCP) server for brewing resources, built with Go.

## What is This?

**BrewSource MCP** is a specialized MCP server that gives AI assistants access to essential brewing knowledge and tools.
 Currently in **Phase 1 MVP**, it focuses on core public resources:

- **Beer & Brewery Discovery** - Search basic commercial beer and brewery databases
- **BJCP Style Guide** - Complete beer style database with lookup capabilities
- **Public API Layer** - Three core MCP tools for essential brewing queries

Future phases will expand with ingredient databases, personal analytics, recipe builders, and premium Brewfather integration.

## Understanding MCP (Model Context Protocol)

**MCP is a standardized way for AI assistants to access external tools and data.** Instead of being limited to their training
 data, AI models can:

- **Call external APIs** (like Brewfather, BJCP data)
- **Access databases** (brewery catalogs, ingredient databases)
- **Perform calculations** (brewing formulas, recipe scaling)
- **Retrieve real-time data** (current beer availability, events)

**Our MCP server** exposes **resources** (data) and **tools** (functions) that AI assistants can use to provide expert
 brewing assistance.

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

```sh
brewsource-mcp/
├── app/                 # Application code
│   ├── cmd/server/      # Main application entry point
│   ├── data/            # BJCP style data (JSON files)
│   ├── internal/        # Internal application code
│   │   ├── handlers/    # HTTP and MCP handlers
│   │   ├── mcp/         # MCP protocol implementation
│   │   ├── models/      # Database models and seed data
│   │   └── services/    # Business logic services
│   └── pkg/data/        # BJCP data utilities
├── bridge/              # Node.js bridge for MCP integration
├── docs/                # Project documentation
├── k8s/                 # Kubernetes manifests
├── .github/             # GitHub workflow and templates
├── go.mod               # Go module definition
├── Makefile             # Build and development commands
├── Tiltfile             # Tilt development configuration
└── README.md            # This file
```

## 📚 Documentation

For comprehensive project documentation, see the **[docs/](docs/)** directory and the Kubernetes deployment guide:

- **[Data Storage Guide](docs/DATA.md)** – Data storage approach, BJCP JSON format, validation, and seeding
- **[Deployment Guide](docs/DEPLOYMENT.md)** – Docker and basic Kubernetes deployment instructions
- **[Kubernetes Production Deployment (k8s/README.md)](k8s/README.md)** – Step-by-step guide for secure, production-ready
 deployment on Oracle Cloud with K3s, Traefik, and Cert-Manager
- **[Testing Guide](docs/TESTING.md)** – Comprehensive testing strategy and user stories
- **[Project Overview](docs/PROJECT_OVERVIEW.md)** – Vision, goals, and technical architecture

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

# Explore cluster
make k9s

# Stop development (cluster stays)
make down

# Clean up everything
make clean
```

### Services

Once running, you'll have:

- **MCP Server**: <http://localhost:8080>
- **PostgreSQL**: localhost:5432 (user: brewsource_user, db: brewsource)
- **Redis**: localhost:6379
- **Tilt Dashboard**: <http://localhost:10350>

## MCP Client Configuration (HTTP Mode)

The BrewSource MCP server now supports HTTP-based MCP communication for both local development and production.

### Production/Remote Usage

To connect to the production MCP server, use the following configuration in your `mcp.json`:

```json
"brewsource": {
    "type": "http",
    "url": "https://brewsource.charlritter.com/mcp"
}
```

### Local Development Usage

To run and test the MCP server locally, use:

```json
"brewsource": {
    "type": "http",
    "url": "http://localhost:8080/mcp"
}
```

### Quick Start for Local Development

```bash
git clone <repository-url>
cd brewsource-mcp
make up
```

### 4. Build and Run

```bash
# Run development environment (Kubernetes + Tilt)
# Access the Tilt dashboard at http://localhost:10350
make up

# Use k9s for interactive cluster management
make k9s
```

### 5. Test the Server

```bash
# Health check
curl http://localhost:8080/health

# Current version
curl http://localhost:8080/version

# Server info
curl http://localhost:8080/api
```

### 6. Run Tests

```bash
# Run all tests
make test
```

## Development Guide

### Adding New Tools

1. **Define the tool function** in `app/internal/handlers/tools.go`:

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

1. **Register the tool** in `RegisterToolHandlers()`:

```go
server.RegisterToolHandler("my_new_tool", h.MyNewTool)
```

1. **Add tool definition** in `getToolDefinition()` method in `app/internal/mcp/server.go`

### Adding New Resources

1. **Create resource handler** in `app/internal/handlers/resources.go`:

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

1. **Register the resource** in `RegisterResourceHandlers()`:

```go
server.RegisterResourceHandler("my://resource/*", h.HandleMyResource)
```

### BJCP Style Guide

The `app/pkg/data` package manages beer style data:

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

## Phase 1 Implementation Status

**BrewSource MCP** is currently in Phase 1 MVP with the following implemented features:

### Core MCP Tools

- **`bjcp_lookup`** - Look up BJCP beer styles by code (e.g., "21A") or name
- **`search_beers`** - Search commercial beers by name, style, brewery, or location
- **`find_breweries`** - Find breweries by name, location, city, state, or country

### MCP Resources

- **`bjcp://styles`** - Complete BJCP style guidelines database
- **`bjcp://styles/{code}`** - Individual style details (e.g., bjcp://styles/21A)
- **`bjcp://categories`** - List of all BJCP categories
- **`beers://catalog`** - Commercial beer database
- **`breweries://directory`** - Brewery directory

### Infrastructure

- **PostgreSQL Database** - Persistent storage with proper indexing
- **Redis Caching** - Optional caching layer for improved performance
- **Seed Data** - Pre-populated with BJCP styles, breweries, and commercial beers
- **Comprehensive Testing** - Unit tests for brewing calculations and BJCP utilities

### Developer Experience

- **Makefile** - Common development tasks (`make help` to see all commands)
- **Environment Configuration** - Managed via .envrc and direnv
- **API Documentation** - Clear tool and resource schemas
- **Code Quality** - Proper error handling, logging, and Go best practices

## Architecture Deep Dive

### MCP Protocol Flow

1. **Client Connection**: MCP client connects via HTTP
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
./bin/brewsource-mcp -port=8081
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
- Review **[Testing Guide](docs/TESTING.md)** for user stories and test specifications
- See **[Deployment Guide](docs/DEPLOYMENT.md)** for deployment requirements
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
- [BJCP styles (beer, mead, cider) JSON](app/data/)

To expand the beer or brewery data (add new entries or fix errors), edit the relevant Go file and open a Pull Request with
 your changes. For BJCP style data, update the appropriate JSON file in `app/data/` and submit a PR. Please ensure your
 changes are well-formatted and include a clear description of the update.

### Code Standards

- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add comments for complex brewing calculations
- Include unit tests for new functions
- Update documentation for new features

## Roadmap

We follow an Agile development methodology with iterative releases and continuous feedback to ensure the platform evolves
 effectively to meet user needs.

**Overall Goal:** To build a comprehensive and evolving platform that serves as a central resource for beer enthusiasts,
 brewers, and the wider beer community.

**Guiding Principles:**

- **User-Centric Design:** Prioritise features that provide the most value to users
- **Iterative Development:** Release functional increments regularly to gather feedback
- **Scalability:** Design architecture to handle future growth and new features
- **Data Accuracy:** Ensure reliability and up-to-date nature of all beer-related data

### Phase 1: Minimum Viable Product (MVP) - Core Public Resources

**Goal:** Launch a foundational public platform with essential beer knowledge and search capabilities to validate core
 concepts and attract initial users.

**MVP Definition:** An MCP server providing searchable BJCP style guidelines and a basic commercial beer catalogue.

**Key Features:**

- [x] **BJCP Style Guide Integration (Basic):** Style lookup by number or name (displaying basic characteristics like ABV,
 IBU, colour)
- [x] **Beer & Brewery Catalogue (Basic):** Searchable commercial beer database (name, style) and basic brewery directory
 (name, location)
- [x] **MCP Tools - Public Layer (Basic):** `bjcp_lookup`, `search_beers`, `find_breweries`
- [x] **Manual Data Input:** Limited initial set of styles, beers, and breweries
- [x] **HTTP Support:** HTTP connection mode for the MCP client

**Outcome:** A functional, publicly accessible MCP Beer Server with core BJCP style lookup and basic beer/brewery search
 capabilities.

### Phase 2: Expanding Public Resources & Usability

**Goal:** Enhance the public resources with more detailed information and improved search capabilities, laying the groundwork
 for future features.

**Key Features:**

- [ ] **BJCP Style Guide Integration (Enhanced):**
  - Style comparison (2-3 styles side-by-side)
  - Detailed style search (by colour, ABV, hop character)
  - **Multi-source BJCP JSON support:** Load and query beer, mead, cider, and special ingredients from separate JSON files
 (e.g., `bjcp_2021_beer.json`, `bjcp_2015_mead.json`, `bjcp_2025_cider.json`, `bjcp_2015_special_ingredients.json`).
- [ ] **Brewing Ingredients Database (Basic):** Malt substitution charts, hop comparison (basic profiles), yeast strains
 database
- [ ] **Beer & Brewery Catalogue (Enhanced):** Beer-brewery linking, availability info (simple "available" flag)
- [ ] **MCP Tools - Public Layer (Expanded):** `bjcp_compare`, `ingredient_substitute`, `ingredient_compare`, `brewery_beers`

**Outcome:** A more robust public resource with expanded BJCP details, initial ingredient information (including yeast),
 and better interconnected beer/brewery data.

### Phase 3: Premium/Personal Features - Brewfather Integration & Personal Analytics

**Goal:** Introduce the first premium features, focusing on direct value for active brewers by integrating with Brewfather
 and offering basic personal analytics.

**Key Features:**

- [ ] **Brewfather Integration (Basic):** Inventory sync (pull fermentables, hops, yeast from Brewfather - read-only)
- [ ] **Personal Analytics (Basic):** Brewing trends (number of batches brewed, most used styles from synced data)
- [ ] **User Authentication & Profiles:** Secure user registration and login for premium features, basic user profile management

**Outcome:** Launch of the first premium tier, offering tangible value to brewers through Brewfather inventory synchronisation
 and initial personal analytics.

### Phase 4: Advanced Features - Interactive Recipe Builder & Community Engagement

**Goal:** Introduce an initial version of the interactive recipe builder and foster community engagement through recipe
 sharing and event listings.

**Key Features:**

- [ ] **Interactive Recipe Builder (Basic):** Style-guided creation, ingredient-based suggestions from Brewfather inventory
- [ ] **Community Features (Basic):** Brewery & beer festival events (manual input initially)
- [ ] **MCP Tools - New Endpoints:** `recipe_generate` (interactive recipe generation wizard)

**Outcome:** A nascent interactive recipe builder with inventory-aware suggestions and the beginning of community features.

### Phase 5: Continuous Improvement & Expansion (Future)

**Goal:** Iteratively enhance existing features, introduce more advanced functionalities, and respond to user feedback to
 continually grow the platform.

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
- Enhance MCP resource endpoints with better filtering
- Add more comprehensive error handling and logging
- Implement caching for frequently accessed data

---

**Happy Brewing!** 🍺
