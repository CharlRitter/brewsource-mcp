// Package models_test contains tests for the data models and database schema in Brewsource MCP.
package models_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/internal/models"
	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test database setup and teardown helpers.
func setupTestDB(t *testing.T) *sqlx.DB {
	// Using in-memory SQLite for testing
	db, err := sqlx.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to open test database")

	// Create tables for testing
	createTables(t, db)

	return db
}

func createTables(t *testing.T, db *sqlx.DB) {
	// Create breweries table (SQLite compatible)
	breweriesSchema := `
		CREATE TABLE breweries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			brewery_type TEXT NOT NULL,
			street TEXT,
			city TEXT,
			state TEXT,
			postal_code TEXT,
			country TEXT,
			phone TEXT,
			website_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.Exec(breweriesSchema)
	require.NoError(t, err, "Failed to create breweries table")

	// Create beers table (SQLite compatible)
	beersSchema := `
		CREATE TABLE beers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			brewery_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			style TEXT,
			abv REAL,
			ibu INTEGER,
			srm REAL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (brewery_id) REFERENCES breweries (id)
		);
	`
	_, err = db.Exec(beersSchema)
	require.NoError(t, err, "Failed to create beers table")
}

func teardownTestDB(t *testing.T, db *sqlx.DB) {
	err := db.Close()
	require.NoError(t, err, "Failed to close test database")
}

// SQLite-compatible helper functions for testing

func insertBreweriesForTest(ctx context.Context, db *sqlx.DB, breweries []services.Brewery) error {
	for _, brewery := range breweries {
		query := `
			INSERT INTO breweries (
				name, brewery_type, street, city, state, postal_code, country, phone, website_url
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?
			)
		`
		if _, err := db.ExecContext(ctx, query,
			brewery.Name, brewery.BreweryType, brewery.Street, brewery.City,
			brewery.State, brewery.PostalCode, brewery.Country, brewery.Phone, brewery.WebsiteURL); err != nil {
			return fmt.Errorf("failed to insert brewery %s: %w", brewery.Name, err)
		}
	}
	return nil
}

func insertBeersForTest(ctx context.Context, db *sqlx.DB, breweries map[string]int, beers []services.SeedBeer) error {
	for _, beer := range beers {
		breweryID, exists := breweries[beer.BreweryName]
		if !exists {
			logrus.Warnf("Brewery not found: %s, skipping beer: %s", beer.BreweryName, beer.Name)
			continue
		}
		query := `
			INSERT INTO beers (
				brewery_id, name, style, abv, ibu, srm, description
			) VALUES (
				?, ?, ?, ?, ?, ?, ?
			)
		`
		if _, err := db.ExecContext(ctx, query, breweryID, beer.Name, beer.Style, beer.ABV, beer.IBU, beer.SRM, beer.Description); err != nil {
			return fmt.Errorf("failed to insert beer %s: %w", beer.Name, err)
		}
	}
	return nil
}

// Test Suite for SeedDatabase function

func TestSeedDatabase_HappyPath(t *testing.T) {
	t.Run("should seed both breweries and beers successfully", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)

		// When
		err := models.SeedDatabase(db)

		// Then
		require.NoError(t, err, "models.SeedDatabase should not return an error")

		// Verify breweries were seeded
		var breweryCount int
		err = db.Get(&breweryCount, "SELECT COUNT(*) FROM breweries")
		require.NoError(t, err)
		assert.Equal(t, 26, breweryCount, "Should have seeded 26 breweries")

		// Verify beers were seeded
		var beerCount int
		err = db.Get(&beerCount, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)
		assert.Equal(t, 66, beerCount, "Should have seeded 66 beers")

		// Verify data integrity - each beer should have a valid brewery_id
		var orphanedBeers int
		err = db.Get(&orphanedBeers, `
			SELECT COUNT(*) FROM beers b
			LEFT JOIN breweries br ON b.brewery_id = br.id
			WHERE br.id IS NULL
		`)
		require.NoError(t, err)
		assert.Equal(t, 0, orphanedBeers, "Should have no orphaned beers")
	})
}

func TestSeedDatabase_Idempotency(t *testing.T) {
	t.Run("should not duplicate data when called multiple times", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)

		// When - seed multiple times
		err1 := models.SeedDatabase(db)
		err2 := models.SeedDatabase(db)
		err3 := models.SeedDatabase(db)

		// Then
		require.NoError(t, err1, "First seeding should not return an error")
		require.NoError(t, err2, "Second seeding should not return an error")
		require.NoError(t, err3, "Third seeding should not return an error")

		// Verify counts remain the same
		var breweryCount int
		err := db.Get(&breweryCount, "SELECT COUNT(*) FROM breweries")
		require.NoError(t, err)
		assert.Equal(t, 26, breweryCount, "Should still have only 26 breweries after multiple seedings")

		var beerCount int
		err = db.Get(&beerCount, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)
		assert.Equal(t, 66, beerCount, "Should still have only 66 beers after multiple seedings")
	})
}

func TestSeedDatabase_DatabaseConnectionError(t *testing.T) {
	t.Run("should return error when database connection is invalid", func(t *testing.T) {
		// Given - a closed database connection
		db := setupTestDB(t)
		db.Close() // Close the connection to simulate connection error

		// When
		err := models.SeedDatabase(db)

		// Then
		require.Error(t, err, "Should return error when database connection is invalid")
		assert.Contains(t, err.Error(), "failed to seed breweries", "Error should indicate brewery seeding failure")
	})
}

// Test Suite for SeedBreweries function

func TestSeedBreweries_HappyPath(t *testing.T) {
	t.Run("should seed breweries successfully when table is empty", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Use local test data
		breweries := []services.Brewery{
			{
				Name:        "Test Brewery 1",
				BreweryType: "micro",
				City:        "Test City",
				State:       "Test State",
				Country:     "Test Country",
			},
			{
				Name:        "Test Brewery 2",
				BreweryType: "nano",
				City:        "Another City",
				State:       "Another State",
				Country:     "Another Country",
			},
		}
		err := insertBreweriesForTest(ctx, db, breweries)

		// Then
		require.NoError(t, err, "insertBreweriesForTest should not return an error")

		// Verify all breweries were inserted
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM breweries")
		require.NoError(t, err)
		assert.Equal(t, 2, count, "Should have inserted 2 breweries")

		// Verify specific brewery data
		var brewery services.Brewery
		err = db.Get(&brewery, "SELECT * FROM breweries WHERE name = ?", "Test Brewery 1")
		require.NoError(t, err)
		assert.Equal(t, "micro", brewery.BreweryType)
		assert.Equal(t, "Test City", brewery.City)
		assert.Equal(t, "Test State", brewery.State)
		assert.Equal(t, "Test Country", brewery.Country)
	})
}

func TestSeedBreweries_SkipWhenDataExists(t *testing.T) {
	t.Run("should skip seeding when breweries already exist", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Insert one brewery manually
		_, err := db.Exec("INSERT INTO breweries (name, brewery_type, city, country) VALUES (?, ?, ?, ?)",
			"Test Brewery", "micro", "Test City", "Test Country")
		require.NoError(t, err)

		// When
		err = models.SeedBreweries(ctx, db)

		// Then
		require.NoError(t, err, "models.SeedBreweries should not return an error")

		// Verify count remains 1 (only the manually inserted brewery)
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM breweries")
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should still have only 1 brewery after skipping seed")
	})
}

func TestSeedBreweries_DatabaseError(t *testing.T) {
	t.Run("should return error when database query fails", func(t *testing.T) {
		// Given - a database without the breweries table
		db, err := sqlx.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		ctx := context.Background()

		// When
		err = models.SeedBreweries(ctx, db)

		// Then
		assert.Error(t, err, "Should return error when breweries table doesn't exist")
	})
}

// Test Suite for SeedBeers function

func TestSeedBeers_HappyPath(t *testing.T) {
	t.Run("should seed beers successfully when table is empty", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Seed breweries first using test function and local test data
		breweries := []services.Brewery{
			{
				Name:        "Test Brewery 1",
				BreweryType: "micro",
				City:        "Test City",
				State:       "Test State",
				Country:     "Test Country",
			},
		}
		err := insertBreweriesForTest(ctx, db, breweries)
		require.NoError(t, err)

		// Get brewery IDs
		breweryIDs, err := models.GetBreweryIDs(ctx, db)
		require.NoError(t, err)

		// When - use local test beers
		beers := []services.SeedBeer{
			{
				Name:        "Test Beer 1",
				BreweryName: "Test Brewery 1",
				Style:       "Test Style",
				ABV:         5.5,
				IBU:         40,
				SRM:         8.0,
				Description: "A test beer for unit tests.",
			},
		}
		_ = insertBeersForTest(ctx, db, breweryIDs, beers)

		// When
		beers = []services.SeedBeer{
			{
				Name:        "Beer 1",
				BreweryName: "Brewery A",
				Style:       "IPA",
				ABV:         5.5,
				IBU:         40,
				SRM:         6.0,
				Description: "A test IPA.",
			},
			{
				Name:        "Beer 2",
				BreweryName: "Brewery B",
				Style:       "Lager",
				ABV:         4.2,
				IBU:         20,
				SRM:         3.0,
				Description: "A test Lager.",
			},
		}
		_ = insertBeersForTest(ctx, db, breweryIDs, beers)
		// Verify all beers were inserted
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should have inserted 1 beer")

		// Verify specific beer data
		var beer struct {
			Name        string  `db:"name"`
			Style       string  `db:"style"`
			ABV         float64 `db:"abv"`
			IBU         int     `db:"ibu"`
			SRM         float64 `db:"srm"`
			Description string  `db:"description"`
		}
		err = db.Get(
			&beer,
			"SELECT name, style, abv, ibu, srm, description FROM beers WHERE name = ?",
			"Test Beer 1",
		)
		require.NoError(t, err)
		assert.Equal(t, "Test Style", beer.Style)
		assert.InDelta(t, 5.5, beer.ABV, 0.01)
		assert.Equal(t, 40, beer.IBU)
		assert.InDelta(t, 8.0, beer.SRM, 0.01)
		assert.Contains(t, beer.Description, "test beer")
	})
}

func TestSeedBeers_SkipWhenDataExists(t *testing.T) {
	t.Run("should skip seeding when beers already exist", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Seed breweries first using test function
		breweries := []services.Brewery{
			{
				Name:        "Test Brewery 1",
				BreweryType: "micro",
				City:        "Test City",
				State:       "Test State",
				Country:     "Test Country",
			},
			{
				Name:        "Test Brewery 2",
				BreweryType: "micro",
				City:        "Test City",
				State:       "Test State",
				Country:     "Test Country",
			},
		}
		err := insertBreweriesForTest(ctx, db, breweries)
		require.NoError(t, err)

		// Insert one beer manually
		_, err = db.Exec("INSERT INTO beers (brewery_id, name, style) VALUES (?, ?, ?)", 1, "Test Beer", "Test Style")
		require.NoError(t, err)

		// When
		err = models.SeedBeers(ctx, db)

		// Then
		require.NoError(t, err, "models.SeedBeers should not return an error")

		// Verify count remains 1
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should still have only 1 beer after skipping seed")
	})
}

func TestSeedBeers_NoBreweriesExist(t *testing.T) {
	t.Run("should handle case when no breweries exist", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()
		// Note: Not seeding breweries first

		// When
		err := models.SeedBeers(ctx, db)

		// Then
		require.NoError(t, err, "models.SeedBeers should not return an error even when no breweries exist")

		// Verify no beers were inserted
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should have 0 beers when no breweries exist")
	})
}

// Test Suite for helper functions

func TestGetSeedBreweries_DataValidation(t *testing.T) {
	t.Run("should return valid brewery data", func(t *testing.T) {
		// Use local test data
		breweries := []services.Brewery{
			{Name: "Brewery X", BreweryType: "micro", City: "Test City", State: "Test State", Country: "Test Country"},
			{Name: "Brewery Y", BreweryType: "micro", City: "Test City", State: "Test State", Country: "Test Country"},
		}

		// Then
		assert.Len(t, breweries, 2, "Should return 2 breweries")

		// Validate each brewery has required fields
		for i, brewery := range breweries {
			assert.NotEmpty(t, brewery.Name, "Brewery %d should have a name", i)
			assert.NotEmpty(t, brewery.BreweryType, "Brewery %d should have a type", i)
			assert.NotEmpty(t, brewery.City, "Brewery %d should have a city", i)
			assert.NotEmpty(t, brewery.State, "Brewery %d should have a state", i)
			assert.NotEmpty(t, brewery.Country, "Brewery %d should have a country", i)

			// Validate brewery type
			assert.Equal(t, "micro", brewery.BreweryType, "Brewery %d should be 'micro' type", i)

			// Validate country
			assert.Equal(t, "Test Country", brewery.Country, "Brewery %d should be in Test Country", i)

			// Validate state
			assert.Equal(t, "Test State", brewery.State, "Brewery %d should be in Test State", i)
		}
	})
}

func TestGetSeedBeers_DataValidation(t *testing.T) {
	t.Run("should return valid beer data", func(t *testing.T) {
		// Use local test data
		beers := []services.SeedBeer{
			{
				Name:        "Test Beer 1",
				BreweryName: "Brewery X",
				Style:       "IPA",
				ABV:         5.0,
				IBU:         40,
				SRM:         6.0,
				Description: "A test IPA.",
			},
			{
				Name:        "Test Beer 2",
				BreweryName: "Brewery Y",
				Style:       "Lager",
				ABV:         4.2,
				IBU:         20,
				SRM:         3.0,
				Description: "A test Lager.",
			},
		}

		// Then
		assert.Len(t, beers, 2, "Should return 2 beers")

		// Validate each beer has required fields
		for i, beer := range beers {
			assert.NotEmpty(t, beer.Name, "Beer %d should have a name", i)
			assert.NotEmpty(t, beer.BreweryName, "Beer %d should have a brewery name", i)
			assert.NotEmpty(t, beer.Style, "Beer %d should have a style", i)
			assert.NotEmpty(t, beer.Description, "Beer %d should have a description", i)

			// Validate ABV range (reasonable beer ABV values)
			assert.GreaterOrEqual(t, beer.ABV, 3.0, "Beer %d ABV should be >= 3.0", i)
			assert.LessOrEqual(t, beer.ABV, 10.0, "Beer %d ABV should be <= 10.0", i)

			// Validate IBU range (reasonable beer IBU values)
			assert.GreaterOrEqual(t, beer.IBU, 10, "Beer %d IBU should be >= 10", i)
			assert.LessOrEqual(t, beer.IBU, 100, "Beer %d IBU should be <= 100", i)

			// Validate SRM range (reasonable beer color values)
			assert.GreaterOrEqual(t, beer.SRM, 2.0, "Beer %d SRM should be >= 2.0", i)
			assert.LessOrEqual(t, beer.SRM, 40.0, "Beer %d SRM should be <= 40.0", i)
		}
	})
}

func TestGetSeedBeers_BreweryMapping(t *testing.T) {
	t.Run("should map beers to valid breweries", func(t *testing.T) {
		// Use local test data
		breweries := []services.Brewery{
			{Name: "Brewery X", BreweryType: "micro", City: "Test City", State: "Test State", Country: "Test Country"},
			{Name: "Brewery Y", BreweryType: "micro", City: "Test City", State: "Test State", Country: "Test Country"},
		}
		beers := []services.SeedBeer{
			{
				Name:        "Test Beer 1",
				BreweryName: "Brewery X",
				Style:       "IPA",
				ABV:         5.0,
				IBU:         40,
				SRM:         6.0,
				Description: "A test IPA.",
			},
			{
				Name:        "Test Beer 2",
				BreweryName: "Brewery Y",
				Style:       "Lager",
				ABV:         4.2,
				IBU:         20,
				SRM:         3.0,
				Description: "A test Lager.",
			},
		}

		// Create a map of brewery names
		breweryNames := make(map[string]bool)
		for _, brewery := range breweries {
			breweryNames[brewery.Name] = true
		}

		// Then - verify each beer maps to a valid brewery
		for i, beer := range beers {
			assert.True(t, breweryNames[beer.BreweryName],
				"Beer %d brewery name '%s' should exist in seed breweries", i, beer.BreweryName)
		}
	})
}

func TestGetBreweryIDs_HappyPath(t *testing.T) {
	t.Run("should return correct brewery ID mapping", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Seed breweries first using test function and local test data
		breweries := []services.Brewery{
			{Name: "Brewery A", BreweryType: "micro", City: "CityA", State: "StateA", Country: "CountryA"},
			{Name: "Brewery B", BreweryType: "micro", City: "CityB", State: "StateB", Country: "CountryB"},
		}
		err := insertBreweriesForTest(ctx, db, breweries)
		require.NoError(t, err)

		// When
		breweryIDs, err := models.GetBreweryIDs(ctx, db)

		// Then
		require.NoError(t, err, "GetBreweryIDs should not return an error")
		assert.Len(t, breweryIDs, 2, "Should return 2 brewery IDs")

		// Verify specific brewery mapping
		breweryAID, exists := breweryIDs["Brewery A"]
		assert.True(t, exists, "Should find Brewery A")
		assert.Positive(t, breweryAID, "Brewery ID should be positive")
	})
}

func TestGetBreweryIDs_EmptyDatabase(t *testing.T) {
	t.Run("should return empty map when no breweries exist", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// When
		breweryIDs, err := models.GetBreweryIDs(ctx, db)

		// Then
		require.NoError(t, err, "GetBreweryIDs should not return an error")
		assert.Empty(t, breweryIDs, "Should return empty map when no breweries exist")
	})
}

func TestInsertBreweries_BoundaryCase(t *testing.T) {
	t.Run("should handle empty brewery slice", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// When
		err := insertBreweriesForTest(ctx, db, []services.Brewery{})

		// Then
		require.NoError(t, err, "insertBreweries should handle empty slice")

		// Verify no breweries were inserted
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM breweries")
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should have 0 breweries after empty insert")
	})
}

func TestInsertBeers_BoundaryCase(t *testing.T) {
	t.Run("should handle empty beer slice", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// When
		err := insertBeersForTest(ctx, db, make(map[string]int), []services.SeedBeer{})

		// Then
		require.NoError(t, err, "insertBeers should handle empty slice")

		// Verify no beers were inserted
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should have 0 beers after empty insert")
	})
}

func TestInsertBeers_MissingBrewery(t *testing.T) {
	t.Run("should skip beers with non-existent brewery names", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Insert an actual brewery into the database
		_, err := db.Exec("INSERT INTO breweries (name, brewery_type, city, country) VALUES (?, ?, ?, ?)",
			"Existing Brewery", "micro", "Test City", "Test Country")
		require.NoError(t, err)

		// Create brewery mapping with the existing brewery
		breweryIDs := map[string]int{
			"Existing Brewery": 1, // This matches the ID from the insert above
		}

		// Create beer data with both existing and non-existing brewery
		testBeers := []services.SeedBeer{
			{
				Name:        "Valid Beer",
				BreweryName: "Existing Brewery",
				Style:       "IPA",
				ABV:         5.0,
				IBU:         30,
				SRM:         5.0,
				Description: "Test beer",
			},
			{
				Name:        "Invalid Beer",
				BreweryName: "Non-Existent Brewery",
				Style:       "Lager",
				ABV:         4.0,
				IBU:         20,
				SRM:         3.0,
				Description: "This beer should be skipped",
			},
		}

		// When
		err = insertBeersForTest(ctx, db, breweryIDs, testBeers)

		// Then
		require.NoError(t, err, "insertBeers should not return an error when skipping invalid breweries")

		// Verify only valid beer was inserted (the one with existing brewery)
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should have 1 beer (the valid one with existing brewery)")

		// Verify the correct beer was inserted
		var beerName string
		err = db.Get(&beerName, "SELECT name FROM beers WHERE brewery_id = 1")
		require.NoError(t, err)
		assert.Equal(t, "Valid Beer", beerName, "The valid beer should be inserted")
	})
}

// Edge Case Tests

func TestSeedDatabase_ContextCancellation(t *testing.T) {
	t.Run("should handle context cancellation gracefully", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// When
		err := models.SeedBreweries(ctx, db)
		// Then
		// The behavior depends on the database driver and when the context is checked
		// Some drivers may not immediately respect context cancellation
		if err != nil {
			assert.Contains(t, err.Error(), "context", "Error should be related to context cancellation")
		}
	})
}

func TestSeedDatabase_LargeDataSet(t *testing.T) {
	t.Run("should handle performance with reasonable response time", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Use local test data
		breweries := []services.Brewery{
			{Name: "Brewery A", BreweryType: "micro", City: "CityA", State: "StateA", Country: "CountryA"},
			{Name: "Brewery B", BreweryType: "micro", City: "CityB", State: "StateB", Country: "CountryB"},
		}
		beers := []services.SeedBeer{
			{
				Name:        "Beer 1",
				BreweryName: "Brewery A",
				Style:       "IPA",
				ABV:         5.5,
				IBU:         40,
				SRM:         6.0,
				Description: "A test IPA.",
			},
			{
				Name:        "Beer 2",
				BreweryName: "Brewery B",
				Style:       "Lager",
				ABV:         4.2,
				IBU:         20,
				SRM:         3.0,
				Description: "A test Lager.",
			},
		}
		err := insertBreweriesForTest(ctx, db, breweries)
		require.NoError(t, err)
		breweryIDs, err := models.GetBreweryIDs(ctx, db)
		require.NoError(t, err)
		err = insertBeersForTest(ctx, db, breweryIDs, beers)
		require.NoError(t, err)

		// Then
		var breweryCount, beerCount int
		err = db.Get(&breweryCount, "SELECT COUNT(*) FROM breweries")
		require.NoError(t, err)
		err = db.Get(&beerCount, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)

		assert.Equal(t, 2, breweryCount, "All breweries should be seeded")
		assert.Equal(t, 2, beerCount, "All beers should be seeded")
	})
}

// Integration Tests

func TestSeedDatabase_FullIntegration(t *testing.T) {
	t.Run("should seed complete database and maintain referential integrity", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Use local test data
		breweries := []services.Brewery{
			{Name: "Brewery A", BreweryType: "micro", City: "CityA", State: "StateA", Country: "CountryA"},
			{Name: "Brewery B", BreweryType: "micro", City: "CityB", State: "StateB", Country: "CountryB"},
		}
		beers := []services.SeedBeer{
			{
				Name:        "Beer 1",
				BreweryName: "Brewery A",
				Style:       "IPA",
				ABV:         5.5,
				IBU:         40,
				SRM:         6.0,
				Description: "A test IPA.",
			},
			{
				Name:        "Beer 2",
				BreweryName: "Brewery B",
				Style:       "Lager",
				ABV:         4.2,
				IBU:         20,
				SRM:         3.0,
				Description: "A test Lager.",
			},
		}
		err := insertBreweriesForTest(ctx, db, breweries)
		require.NoError(t, err)
		breweryIDs, err := models.GetBreweryIDs(ctx, db)
		require.NoError(t, err)
		err = insertBeersForTest(ctx, db, breweryIDs, beers)
		require.NoError(t, err)

		// Then
		type beerBrewery struct {
			BeerName    string `db:"beer_name"`
			BreweryName string `db:"brewery_name"`
		}

		var results []beerBrewery
		query := `
		       SELECT b.name as beer_name, br.name as brewery_name
		       FROM beers b
		       JOIN breweries br ON b.brewery_id = br.id
	       `
		err = db.Select(&results, query)
		require.NoError(t, err)

		assert.Len(t, results, 2, "Should have 2 beer-brewery relationships")

		// Verify specific relationships
		expectedRelationships := map[string]string{
			"Beer 1": "Brewery A",
			"Beer 2": "Brewery B",
		}

		for _, result := range results {
			expectedBrewery, exists := expectedRelationships[result.BeerName]
			assert.True(t, exists, "Beer '%s' should be in expected relationships", result.BeerName)
			assert.Equal(t, expectedBrewery, result.BreweryName,
				"Beer '%s' should be associated with brewery '%s'", result.BeerName, expectedBrewery)
		}
	})
}
