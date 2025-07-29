# BrewSource MCP Server Data Guide

This document provides a comprehensive overview of the data storage, format, validation, and seeding strategies used in the BrewSource MCP Server. It covers the hybrid storage model, BJCP JSON schema, validation workflow, and automatic database seeding for development and testing.

---

## Table of Contents

- [BrewSource MCP Server Data Guide](#brewsource-mcp-server-data-guide)
  - [Table of Contents](#table-of-contents)
  - [Data Storage Approach](#data-storage-approach)
    - [Overview](#overview)
    - [Why Hybrid Storage?](#why-hybrid-storage)
      - [JSON for BJCP Styles](#json-for-bjcp-styles)
      - [PostgreSQL for Beer \& Brewery Data](#postgresql-for-beer--brewery-data)
    - [Implementation Details](#implementation-details)
      - [BJCP Styles (JSON)](#bjcp-styles-json)
      - [Beer \& Brewery Data (PostgreSQL)](#beer--brewery-data-postgresql)
    - [File Structure](#file-structure)
    - [Benefits of This Approach](#benefits-of-this-approach)
      - [For BJCP Styles (JSON)](#for-bjcp-styles-json)
      - [For Beer/Brewery Data (Database)](#for-beerbrewery-data-database)
    - [Migration \& Future Work](#migration--future-work)
    - [Conclusion](#conclusion)
  - [BJCP JSON Format](#bjcp-json-format)
    - [File Location](#file-location)
    - [Top-Level Structure](#top-level-structure)
    - [Style Object Schema](#style-object-schema)
    - [Validation](#validation)
    - [Contribution Guidelines](#contribution-guidelines)
    - [What It Checks](#what-it-checks)
    - [How to Add New Styles](#how-to-add-new-styles)
    - [Troubleshooting](#troubleshooting)
  - [Database Seeding](#database-seeding)
    - [What is Seed Data?](#what-is-seed-data)
    - [When is Seeding Performed?](#when-is-seeding-performed)
    - [How Seeding Works](#how-seeding-works)
    - [What Gets Seeded?](#what-gets-seeded)
    - [Notes](#notes)

---

## Data Storage Approach

### Overview

The BrewSource MCP Server uses a hybrid data storage approach designed for both performance and maintainability. This approach leverages:

- **JSON files for BJCP style guidelines** (static, reference data)
- **PostgreSQL database for beer and brewery data** (dynamic, relational data)

This section describes the rationale, structure, and benefits of this approach.

---

### Why Hybrid Storage?

#### JSON for BJCP Styles
- **Static Reference**: BJCP styles are official guidelines that rarely change.
- **No Relationships**: Each style is self-contained.
- **Version Control**: JSON files are tracked in Git, making changes auditable and reviewable.
- **Performance**: Styles are loaded into memory for fast, in-memory lookups.
- **Portability**: No database required for style lookups; works offline and in any environment.
- **Simple Updates**: Update the JSON, commit, and deploy.

#### PostgreSQL for Beer & Brewery Data
- **Relational Data**: Beers belong to breweries; supports complex relationships.
- **Advanced Queries**: Enables search by location, style, and other criteria.
- **Geographic & Full-Text Search**: Supports distance-based and text-based queries.
- **Data Integrity**: Enforced with foreign keys and constraints.
- **Scalability**: Handles large, dynamic datasets efficiently.
- **Concurrent Access**: Supports multiple users and future community contributions.

---

### Implementation Details

#### BJCP Styles (JSON)
- Stored as JSON files in `app/data/` (e.g., `bjcp_styles.json`, `bjcp_categories.json`).
- Loaded at startup and served via a dedicated Go service (`pkg/data/bjcp.go`).
- Lookups and searches are performed in-memory for maximum speed.

**Example Usage:**
```go
// Lookup by style code
style := bjcpService.GetStyleByCode("21A")
// Search by name
style := bjcpService.GetStyleByName("American IPA")
// List by category
styles := bjcpService.GetStylesByCategory("IPA")
```

#### Beer & Brewery Data (PostgreSQL)
- Managed in a relational database with proper indexing and constraints.
- Accessed via Go services in `internal/services/` (e.g., `beer.go`, `brewery.go`).
- Supports complex queries, joins, and full-text search.

**Example Query:**
```sql
SELECT b.name, br.name, br.city
FROM beers b
JOIN breweries br ON b.brewery_id = br.id
WHERE b.style LIKE '%IPA%' AND br.state = 'CA';
```

---

### File Structure
```
app/
├── data/
│   ├── bjcp_styles.json      # Static reference data
│   └── bjcp_categories.json  # Category metadata
├── pkg/
│   └── data/
│       └── bjcp.go           # JSON-based BJCP service
└── internal/
    └── services/
        ├── beer.go           # Database-backed services
        └── brewery.go
```

---

### Benefits of This Approach

#### For BJCP Styles (JSON)
- **Fast lookups**: In-memory, no DB round-trips
- **Easy updates**: Edit JSON, commit, deploy
- **Portable**: No DB required for style lookups
- **Testable**: Mock data is just another JSON file
- **Versioned**: All changes tracked in Git

#### For Beer/Brewery Data (Database)
- **Data integrity**: Foreign keys, transactions
- **Scalable**: Handles large, growing datasets
- **Flexible queries**: Full-text, geo, and relational search
- **Concurrent access**: Safe for multiple users
- **Standard tools**: Leverages mature database ecosystem

---

### Migration & Future Work


- BJCP data has been migrated to multiple JSON files (`bjcp_2021_beer.json`, `bjcp_2015_mead.json`, `bjcp_2025_cider.json`, `bjcp_2015_special_ingredients.json`) and is no longer stored in the database.
- Handlers and services have been updated to use the JSON-based BJCP service, with plans for a unified datastore service to manage all style types in future phases.
- The database remains the source of truth for beer and brewery data, with ongoing optimization for search and indexing.
<!-- - Makefile targets support JSON validation (`validate-bjcp`). Database seeding is handled automatically during server startup. -->

---

### Conclusion

The hybrid storage model allows BrewSource MCP Server to use the best tool for each data type:

- **JSON for static, reference data** (BJCP styles): fast, portable, version-controlled
- **Database for dynamic, relational data** (beers, breweries): robust, scalable, queryable

This approach ensures high performance, easy maintenance, and a great developer experience.

---

## BJCP JSON Format

### File Location
- All BJCP JSON files are stored in `app/data/`.
- Example files: `bjcp_2021_beer.json`, `bjcp_2015_mead.json`, `bjcp_2025_cider.json`, etc.

### Top-Level Structure
Each file contains a top-level object with:
- `styles`: a dictionary of style objects, keyed by style code (e.g., "21A").
- `categories`: an array of category names.
- `metadata`: an object with versioning info.

Example:
```json
{
  "styles": {
    "21A": { ... },
    "M1A": { ... }
  },
  "categories": ["IPA", "Traditional Mead", ...],
  "metadata": {
    "version": "2021",
    "source": "bjcp.org",
    "last_updated": "2025-07-22",
    "total_styles": 100
  }
}
```

### Style Object Schema
Each style entry should include:
- `code` (string): Style code (e.g., "21A")
- `name` (string): Style name
- `category` (string): Category name
- `overall_impression` (string)
- `appearance` (string)
- `aroma` (string)
- `flavor` (string)
- `mouthfeel` (string)
- `comments` (string)
- `history` (string)
- `characteristic_ingredients` (string)
- `style_comparison` (string)
- `commercial_examples` (array of strings)
- `vitals` (object):
    - `abv_min`, `abv_max` (float)
    - `ibu_min`, `ibu_max` (int)
    - `srm_min`, `srm_max` (float)
    - `og_min`, `og_max` (float)
    - `fg_min`, `fg_max` (float)

Example style:
```json
{
  "code": "21A",
  "name": "American IPA",
  "category": "IPA",
  "overall_impression": "Bold hoppy flavor",
  "appearance": "Golden to amber",
  "aroma": "Citrus hop aroma",
  "flavor": "Hoppy bitterness",
  "mouthfeel": "Medium body",
  "comments": "Modern American style",
  "history": "Developed in America",
  "characteristic_ingredients": "American hops",
  "style_comparison": "More aggressive than English IPA",
  "commercial_examples": ["Stone IPA", "Russian River Blind Pig"],
  "vitals": {
    "abv_min": 5.5,
    "abv_max": 7.5,
    "ibu_min": 40,
    "ibu_max": 70,
    "srm_min": 6.0,
    "srm_max": 14.0,
    "og_min": 1.056,
    "og_max": 1.070,
    "fg_min": 1.008,
    "fg_max": 1.014
  }
}
```

### Validation
- Use `scripts/create_formatted_bjcp.py` to validate and format mead/cider files.
- Ensure all required fields are present and types are correct.
- For beer styles, use a similar schema and validation logic.

### Contribution Guidelines
- Add new styles by editing the appropriate JSON file in `app/data/`.
- Run the validation script before committing changes.
- Update `metadata.last_updated` and `metadata.total_styles` as needed.

---

### What It Checks
- Required fields for each style
- Correct data types (string, float, int, array)
- Value ranges for vitals (ABV, IBU, SRM, OG, FG)
- No duplicate style codes
- Valid JSON format

### How to Add New Styles
- Edit the appropriate JSON file in `app/data/`
<!-- - Run `make validate-bjcp` to check your changes -->
- Fix any errors before committing

### Troubleshooting
- If validation fails, the script will print errors and line numbers
- See [BJCP JSON Format](#bjcp-json-format) for schema details

---

## Database Seeding

### What is Seed Data?
Seed data is a set of sample breweries and beers that are inserted into the database to provide a working dataset for development, testing, and demonstration purposes. The seeding process is idempotent: it will not insert duplicate data if the tables are already populated.

### When is Seeding Performed?
Seeding is performed automatically during server startup. If the database is empty, initial data will be inserted. The process is idempotent and will not duplicate data.

### How Seeding Works
Seeding is invoked automatically by the server on startup. No manual action is required. For advanced scenarios, you may call the seeding logic programmatically in Go, but this is not needed for normal development.

### What Gets Seeded?
- **Breweries:** A set of well-known US breweries with full address and contact info.
- **Beers:** Popular beers from those breweries, with style, ABV, IBU, SRM, and descriptions.

### Notes
- The seeder checks for existing data and skips seeding if breweries or beers already exist.
- All seed data is for development and demonstration only.
- For production, use real data import workflows.

---

For more details, see the code in `app/internal/models/seed.go`.
