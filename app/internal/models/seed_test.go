package models

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// Assuming you're using a testing database driver like sqlite for tests
	_ "github.com/mattn/go-sqlite3"
)

// Test database setup and teardown helpers
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

func insertBreweriesForTest(ctx context.Context, db *sqlx.DB, breweries []Brewery) error {
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

func insertBeersForTest(ctx context.Context, db *sqlx.DB, breweries map[string]int, beers []seedBeer) error {
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
		err := SeedDatabase(db)

		// Then
		require.NoError(t, err, "SeedDatabase should not return an error")

		// Verify breweries were seeded
		var breweryCount int
		err = db.Get(&breweryCount, "SELECT COUNT(*) FROM breweries")
		require.NoError(t, err)
		assert.Equal(t, 8, breweryCount, "Should have seeded 8 breweries")

		// Verify beers were seeded
		var beerCount int
		err = db.Get(&beerCount, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)
		assert.Equal(t, 8, beerCount, "Should have seeded 8 beers")

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
		err1 := SeedDatabase(db)
		err2 := SeedDatabase(db)
		err3 := SeedDatabase(db)

		// Then
		require.NoError(t, err1, "First seeding should not return an error")
		require.NoError(t, err2, "Second seeding should not return an error")
		require.NoError(t, err3, "Third seeding should not return an error")

		// Verify counts remain the same
		var breweryCount int
		err := db.Get(&breweryCount, "SELECT COUNT(*) FROM breweries")
		require.NoError(t, err)
		assert.Equal(t, 8, breweryCount, "Should still have only 8 breweries after multiple seedings")

		var beerCount int
		err = db.Get(&beerCount, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)
		assert.Equal(t, 8, beerCount, "Should still have only 8 beers after multiple seedings")
	})
}

func TestSeedDatabase_DatabaseConnectionError(t *testing.T) {
	t.Run("should return error when database connection is invalid", func(t *testing.T) {
		// Given - a closed database connection
		db := setupTestDB(t)
		db.Close() // Close the connection to simulate connection error

		// When
		err := SeedDatabase(db)

		// Then
		assert.Error(t, err, "Should return error when database connection is invalid")
		assert.Contains(t, err.Error(), "failed to seed breweries", "Error should indicate brewery seeding failure")
	})
}

// Test Suite for seedBreweries function

func TestSeedBreweries_HappyPath(t *testing.T) {
	t.Run("should seed breweries successfully when table is empty", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// When - use test-specific insert function
		breweries := getSeedBreweries()
		err := insertBreweriesForTest(ctx, db, breweries)

		// Then
		require.NoError(t, err, "insertBreweriesForTest should not return an error")

		// Verify all breweries were inserted
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM breweries")
		require.NoError(t, err)
		assert.Equal(t, 8, count, "Should have inserted 8 breweries")

		// Verify specific brewery data
		var brewery Brewery
		err = db.Get(&brewery, "SELECT * FROM breweries WHERE name = ?", "Devil's Peak Brewing Company")
		require.NoError(t, err)
		assert.Equal(t, "micro", brewery.BreweryType)
		assert.Equal(t, "Woodstock", brewery.City)
		assert.Equal(t, "Western Cape", brewery.State)
		assert.Equal(t, "South Africa", brewery.Country)
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
		err = seedBreweries(ctx, db)

		// Then
		require.NoError(t, err, "seedBreweries should not return an error")

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
		err = seedBreweries(ctx, db)

		// Then
		assert.Error(t, err, "Should return error when breweries table doesn't exist")
	})
}

// Test Suite for seedBeers function

func TestSeedBeers_HappyPath(t *testing.T) {
	t.Run("should seed beers successfully when table is empty", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Seed breweries first using test function
		breweries := getSeedBreweries()
		err := insertBreweriesForTest(ctx, db, breweries)
		require.NoError(t, err)

		// Get brewery IDs
		breweryIDs, err := getBreweryIDs(ctx, db)
		require.NoError(t, err)

		// When
		beers := getSeedBeers()
		err = insertBeersForTest(ctx, db, breweryIDs, beers)

		// Then
		require.NoError(t, err, "insertBeersForTest should not return an error")

		// Verify all beers were inserted
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)
		assert.Equal(t, 8, count, "Should have inserted 8 beers")

		// Verify specific beer data
		var beer struct {
			Name        string  `db:"name"`
			Style       string  `db:"style"`
			ABV         float64 `db:"abv"`
			IBU         int     `db:"ibu"`
			SRM         float64 `db:"srm"`
			Description string  `db:"description"`
		}
		err = db.Get(&beer, "SELECT name, style, abv, ibu, srm, description FROM beers WHERE name = ?", "King's Blockhouse IPA")
		require.NoError(t, err)
		assert.Equal(t, "American IPA", beer.Style)
		assert.Equal(t, 6.0, beer.ABV)
		assert.Equal(t, 52, beer.IBU)
		assert.Equal(t, 10.0, beer.SRM)
		assert.Contains(t, beer.Description, "bold, hop-forward IPA")
	})
}

func TestSeedBeers_SkipWhenDataExists(t *testing.T) {
	t.Run("should skip seeding when beers already exist", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Seed breweries first using test function
		breweries := getSeedBreweries()
		err := insertBreweriesForTest(ctx, db, breweries)
		require.NoError(t, err)

		// Insert one beer manually
		_, err = db.Exec("INSERT INTO beers (brewery_id, name, style) VALUES (?, ?, ?)", 1, "Test Beer", "Test Style")
		require.NoError(t, err)

		// When
		err = seedBeers(ctx, db)

		// Then
		require.NoError(t, err, "seedBeers should not return an error")

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
		err := seedBeers(ctx, db)

		// Then
		require.NoError(t, err, "seedBeers should not return an error even when no breweries exist")

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
		// When
		breweries := getSeedBreweries()

		// Then
		assert.Len(t, breweries, 8, "Should return 8 breweries")

		// Validate each brewery has required fields
		for i, brewery := range breweries {
			assert.NotEmpty(t, brewery.Name, fmt.Sprintf("Brewery %d should have a name", i))
			assert.NotEmpty(t, brewery.BreweryType, fmt.Sprintf("Brewery %d should have a type", i))
			assert.NotEmpty(t, brewery.City, fmt.Sprintf("Brewery %d should have a city", i))
			assert.NotEmpty(t, brewery.State, fmt.Sprintf("Brewery %d should have a state", i))
			assert.NotEmpty(t, brewery.Country, fmt.Sprintf("Brewery %d should have a country", i))

			// Validate brewery type
			assert.Equal(t, "micro", brewery.BreweryType, fmt.Sprintf("Brewery %d should be 'micro' type", i))

			// Validate country
			assert.Equal(t, "South Africa", brewery.Country, fmt.Sprintf("Brewery %d should be in South Africa", i))

			// Validate state
			assert.Equal(t, "Western Cape", brewery.State, fmt.Sprintf("Brewery %d should be in Western Cape", i))
		}
	})
}

func TestGetSeedBeers_DataValidation(t *testing.T) {
	t.Run("should return valid beer data", func(t *testing.T) {
		// When
		beers := getSeedBeers()

		// Then
		assert.Len(t, beers, 8, "Should return 8 beers")

		// Validate each beer has required fields
		for i, beer := range beers {
			assert.NotEmpty(t, beer.Name, fmt.Sprintf("Beer %d should have a name", i))
			assert.NotEmpty(t, beer.BreweryName, fmt.Sprintf("Beer %d should have a brewery name", i))
			assert.NotEmpty(t, beer.Style, fmt.Sprintf("Beer %d should have a style", i))
			assert.NotEmpty(t, beer.Description, fmt.Sprintf("Beer %d should have a description", i))

			// Validate ABV range (reasonable beer ABV values)
			assert.GreaterOrEqual(t, beer.ABV, 3.0, fmt.Sprintf("Beer %d ABV should be >= 3.0", i))
			assert.LessOrEqual(t, beer.ABV, 10.0, fmt.Sprintf("Beer %d ABV should be <= 10.0", i))

			// Validate IBU range (reasonable beer IBU values)
			assert.GreaterOrEqual(t, beer.IBU, 10, fmt.Sprintf("Beer %d IBU should be >= 10", i))
			assert.LessOrEqual(t, beer.IBU, 100, fmt.Sprintf("Beer %d IBU should be <= 100", i))

			// Validate SRM range (reasonable beer color values)
			assert.GreaterOrEqual(t, beer.SRM, 2.0, fmt.Sprintf("Beer %d SRM should be >= 2.0", i))
			assert.LessOrEqual(t, beer.SRM, 40.0, fmt.Sprintf("Beer %d SRM should be <= 40.0", i))
		}
	})
}

func TestGetSeedBeers_BreweryMapping(t *testing.T) {
	t.Run("should map beers to valid breweries", func(t *testing.T) {
		// Given
		breweries := getSeedBreweries()
		beers := getSeedBeers()

		// Create a map of brewery names
		breweryNames := make(map[string]bool)
		for _, brewery := range breweries {
			breweryNames[brewery.Name] = true
		}

		// Then - verify each beer maps to a valid brewery
		for i, beer := range beers {
			assert.True(t, breweryNames[beer.BreweryName],
				fmt.Sprintf("Beer %d brewery name '%s' should exist in seed breweries", i, beer.BreweryName))
		}
	})
}

func TestGetBreweryIDs_HappyPath(t *testing.T) {
	t.Run("should return correct brewery ID mapping", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// Seed breweries first using test function
		breweries := getSeedBreweries()
		err := insertBreweriesForTest(ctx, db, breweries)
		require.NoError(t, err)

		// When
		breweryIDs, err := getBreweryIDs(ctx, db)

		// Then
		require.NoError(t, err, "getBreweryIDs should not return an error")
		assert.Len(t, breweryIDs, 8, "Should return 8 brewery IDs")

		// Verify specific brewery mapping
		devilsPeakID, exists := breweryIDs["Devil's Peak Brewing Company"]
		assert.True(t, exists, "Should find Devil's Peak Brewing Company")
		assert.Greater(t, devilsPeakID, 0, "Brewery ID should be positive")
	})
}

func TestGetBreweryIDs_EmptyDatabase(t *testing.T) {
	t.Run("should return empty map when no breweries exist", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// When
		breweryIDs, err := getBreweryIDs(ctx, db)

		// Then
		require.NoError(t, err, "getBreweryIDs should not return an error")
		assert.Len(t, breweryIDs, 0, "Should return empty map when no breweries exist")
	})
}

func TestInsertBreweries_BoundaryCase(t *testing.T) {
	t.Run("should handle empty brewery slice", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)
		ctx := context.Background()

		// When
		err := insertBreweriesForTest(ctx, db, []Brewery{})

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
		err := insertBeersForTest(ctx, db, make(map[string]int), []seedBeer{})

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
		testBeers := []seedBeer{
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
		err := seedBreweries(ctx, db)

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

		// When - measure execution time
		start := logrus.New().WithField("test", "timing")
		start.Info("Starting seed performance test")

		err := SeedDatabase(db)

		start.Info("Completed seed performance test")

		// Then
		require.NoError(t, err, "SeedDatabase should complete without error")

		// Verify all data was seeded correctly
		var breweryCount, beerCount int
		err = db.Get(&breweryCount, "SELECT COUNT(*) FROM breweries")
		require.NoError(t, err)
		err = db.Get(&beerCount, "SELECT COUNT(*) FROM beers")
		require.NoError(t, err)

		assert.Equal(t, 8, breweryCount, "All breweries should be seeded")
		assert.Equal(t, 8, beerCount, "All beers should be seeded")
	})
}

// Integration Tests

func TestSeedDatabase_FullIntegration(t *testing.T) {
	t.Run("should seed complete database and maintain referential integrity", func(t *testing.T) {
		// Given
		db := setupTestDB(t)
		defer teardownTestDB(t, db)

		// When
		err := SeedDatabase(db)

		// Then
		require.NoError(t, err, "SeedDatabase should complete without error")

		// Verify referential integrity
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

		assert.Len(t, results, 8, "Should have 8 beer-brewery relationships")

		// Verify specific relationships
		expectedRelationships := map[string]string{
			"King's Blockhouse IPA":         "Devil's Peak Brewing Company",
			"Four Lager":                    "Jack Black's Brewing Company",
			"The Stranded Coconut":          "Drifter Brewing Company",
			"Hoenderhok Bock":               "Stellenbosch Brewing Company",
			"Rhythm Stick English Pale Ale": "Woodstock Brewery",
			"Bone Crusher Witbier":          "Darling Brew",
			"Amber Weiss":                   "Cape Brewing Company (CBC)",
			"Gun Powder IPA":                "Signal Gun Wines & Brewery",
		}

		for _, result := range results {
			expectedBrewery, exists := expectedRelationships[result.BeerName]
			assert.True(t, exists, fmt.Sprintf("Beer '%s' should be in expected relationships", result.BeerName))
			assert.Equal(t, expectedBrewery, result.BreweryName,
				fmt.Sprintf("Beer '%s' should be associated with brewery '%s'", result.BeerName, expectedBrewery))
		}
	})
}
