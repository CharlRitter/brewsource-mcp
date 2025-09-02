package services

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock data for testing.
var (
	mockBeerRows = [][]driver.Value{
		{1, "King's Blockhouse IPA", "American IPA", "Devil's Peak Brewing Company", "South Africa", 6.0, 60},
		{2, "Hazy Pale Ale", "American Pale Ale", "Jack Black Brewing Co", "South Africa", 5.0, 35},
		{3, "Lager", "Pilsner", "Castle Lager", "South Africa", 4.5, 20},
	}
)

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "postgres")
	return sqlxDB, mock
}

func setupBeerService(db *sqlx.DB) *BeerService {
	redisClient := &redis.Client{} // Mock Redis client
	return NewBeerService(db, redisClient)
}

// Test NewBeerService constructor.
func TestNewBeerService(t *testing.T) {
	t.Run("Valid initialization", func(t *testing.T) {
		db := &sqlx.DB{}
		redisClient := &redis.Client{}
		svc := NewBeerService(db, redisClient)

		assert.NotNil(t, svc, "BeerService should not be nil")
		assert.Equal(t, db, svc.db, "Database should be set correctly")
		assert.Equal(t, redisClient, svc.redisClient, "Redis client should be set correctly")
	})

	t.Run("Nil database", func(t *testing.T) {
		redisClient := &redis.Client{}
		svc := NewBeerService(nil, redisClient)

		assert.NotNil(t, svc, "BeerService should not be nil even with nil DB")
		assert.Nil(t, svc.db, "Database should be nil")
	})

	t.Run("Nil redis client", func(t *testing.T) {
		db := &sqlx.DB{}
		svc := NewBeerService(db, nil)

		assert.NotNil(t, svc, "BeerService should not be nil even with nil Redis")
		assert.Nil(t, svc.redisClient, "Redis client should be nil")
	})
}

// Test struct field validation.
func TestBeerSearchQuery_Fields(t *testing.T) {
	t.Run("All fields set correctly", func(t *testing.T) {
		q := BeerSearchQuery{
			Name:     "King's Blockhouse IPA",
			Style:    "American IPA",
			Brewery:  "Devil's Peak Brewing Company",
			Location: "South Africa",
			Limit:    5,
		}

		assert.Equal(t, "King's Blockhouse IPA", q.Name)
		assert.Equal(t, "American IPA", q.Style)
		assert.Equal(t, "Devil's Peak Brewing Company", q.Brewery)
		assert.Equal(t, "South Africa", q.Location)
		assert.Equal(t, 5, q.Limit)
	})

	t.Run("Default values", func(t *testing.T) {
		q := BeerSearchQuery{}

		assert.Empty(t, q.Name)
		assert.Empty(t, q.Style)
		assert.Empty(t, q.Brewery)
		assert.Empty(t, q.Location)
		assert.Zero(t, q.Limit)
	})
}

func TestBeerSearchResult_Fields(t *testing.T) {
	t.Run("All fields set correctly", func(t *testing.T) {
		r := &BeerSearchResult{
			ID:      1,
			Name:    "King's Blockhouse IPA",
			Style:   "American IPA",
			Brewery: "Devil's Peak Brewing Company",
			Country: "South Africa",
		}

		assert.Equal(t, 1, r.ID)
		assert.Equal(t, "King's Blockhouse IPA", r.Name)
		assert.Equal(t, "American IPA", r.Style)
		assert.Equal(t, "Devil's Peak Brewing Company", r.Brewery)
		assert.Equal(t, "South Africa", r.Country)
	})
}

// Happy Path Tests.
func TestSearchBeers_HappyPath(t *testing.T) {
	t.Run("Search with name only", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1\s+AND b.name ILIKE \$1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"}).
			AddRow(mockBeerRows[0]...).
			AddRow(mockBeerRows[1]...)

		mock.ExpectQuery(expectedQuery).
			WithArgs("%IPA%").
			WillReturnRows(rows)

		query := BeerSearchQuery{Name: "IPA"}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "King's Blockhouse IPA", results[0].Name)
		assert.Equal(t, "Hazy Pale Ale", results[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Search with all filters", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1\s+AND b.name ILIKE \$1\s+AND b.style ILIKE \$2\s+AND br.name ILIKE \$3\s+AND br.city ILIKE \$4\s+LIMIT \$5`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"}).
			AddRow(mockBeerRows[0]...)

		mock.ExpectQuery(expectedQuery).
			WithArgs("%King%", "%IPA%", "%Devil%", "%Cape Town%", 5).
			WillReturnRows(rows)

		query := BeerSearchQuery{
			Name:     "King",
			Style:    "IPA",
			Brewery:  "Devil",
			Location: "Cape Town",
			Limit:    5,
		}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "King's Blockhouse IPA", results[0].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Search with limit only", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1\s+LIMIT \$1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"})
		for i := range 3 {
			rows.AddRow(mockBeerRows[i]...)
		}

		mock.ExpectQuery(expectedQuery).
			WithArgs(3).
			WillReturnRows(rows)

		query := BeerSearchQuery{Limit: 3}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.NoError(t, err)
		assert.Len(t, results, 3)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Search returns no results", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1\s+AND b.name ILIKE \$1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"})

		mock.ExpectQuery(expectedQuery).
			WithArgs("%NONEXISTENT%").
			WillReturnRows(rows)

		query := BeerSearchQuery{Name: "NONEXISTENT"}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.NoError(t, err)
		assert.Empty(t, results)
		assert.NotNil(t, results) // Should be empty slice, not nil
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Edge Cases and Boundary Testing.
func TestSearchBeers_EdgeCases(t *testing.T) {
	t.Run("Empty query parameters", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"})

		mock.ExpectQuery(expectedQuery).
			WithArgs().
			WillReturnRows(rows)

		query := BeerSearchQuery{
			Name:     "",
			Style:    "",
			Brewery:  "",
			Location: "",
			Limit:    0,
		}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Very large limit", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1\s+LIMIT \$1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"}).
			AddRow(mockBeerRows[0]...)

		mock.ExpectQuery(expectedQuery).
			WithArgs(999999).
			WillReturnRows(rows)

		query := BeerSearchQuery{Limit: 999999}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Negative limit", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		// Negative limit should not add LIMIT clause
		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"}).
			AddRow(mockBeerRows[0]...)

		mock.ExpectQuery(expectedQuery).
			WithArgs().
			WillReturnRows(rows)

		query := BeerSearchQuery{Limit: -5}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Unicode and special characters", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1\s+AND b.name ILIKE \$1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"}).
			AddRow(1, "Øl & Bière", "Lager", "Brewery café", "Norway", 5.0, 25)

		mock.ExpectQuery(expectedQuery).
			WithArgs("%Øl & Bière%").
			WillReturnRows(rows)

		query := BeerSearchQuery{Name: "Øl & Bière"}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Øl & Bière", results[0].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Very long search strings", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		longString := string(make([]byte, 1000))
		for i := range longString {
			longString = longString[:i] + "a" + longString[i+1:]
		}

		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1\s+AND b.name ILIKE \$1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"})

		mock.ExpectQuery(expectedQuery).
			WithArgs("%" + longString + "%").
			WillReturnRows(rows)

		query := BeerSearchQuery{Name: longString}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.NoError(t, err)
		assert.Empty(t, results)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Error Handling Tests.
func TestSearchBeers_ErrorHandling(t *testing.T) {
	t.Run("Database connection error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1\s+AND b.name ILIKE \$1`

		mock.ExpectQuery(expectedQuery).
			WithArgs("%IPA%").
			WillReturnError(sql.ErrConnDone)

		query := BeerSearchQuery{Name: "IPA"}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "sql: connection is already closed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Row scan error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b\.id, b\.name, b\.style, br\.name as brewery, br\.country, b\.abv, b\.ibu\s+FROM beers b\s+JOIN breweries br ON b\.brewery_id = br\.id\s+WHERE 1=1\s+AND b\.name ILIKE \$1`

		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"id", "name", "style"}).
			AddRow(1, "Test Beer", "IPA") // Missing brewery and country columns

		mock.ExpectQuery(expectedQuery).
			WithArgs("%IPA%").
			WillReturnRows(rows)

		query := BeerSearchQuery{Name: "IPA"}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.Error(t, err)
		assert.Nil(t, results)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Rows iteration error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b\.id, b\.name, b\.style, br\.name as brewery, br\.country, b\.abv, b\.ibu\s+FROM beers b\s+JOIN breweries br ON b\.brewery_id = br\.id\s+WHERE 1=1\s+AND b\.name ILIKE \$1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"}).
			AddRow(mockBeerRows[0]...).
			AddRow(1, "Second Beer", "IPA", "Test Brewery", "USA", 5.5, 45).
			RowError(1, errors.New("row iteration error"))

		mock.ExpectQuery(expectedQuery).
			WithArgs("%IPA%").
			WillReturnRows(rows)

		query := BeerSearchQuery{Name: "IPA"}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "row iteration error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Context and Timeout Tests.
func TestSearchBeers_Context(t *testing.T) {
	t.Run("Context cancellation", func(t *testing.T) {
		db, _ := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		query := BeerSearchQuery{Name: "IPA"}
		results, err := svc.SearchBeers(ctx, query)

		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("Context timeout", func(t *testing.T) {
		db, _ := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
		defer cancel()

		// Add small delay to ensure timeout
		time.Sleep(2 * time.Microsecond)

		query := BeerSearchQuery{Name: "IPA"}
		results, err := svc.SearchBeers(ctx, query)

		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

// Performance Tests.
func TestSearchBeers_Performance(t *testing.T) {
	t.Run("Large result set handling", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		expectedQuery := `SELECT b\.id, b\.name, b\.style, br\.name as brewery, br\.country, b\.abv, b\.ibu\s+FROM beers b\s+JOIN breweries br ON b\.brewery_id = br\.id\s+WHERE 1=1\s+LIMIT \$1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"})
		// Simulate 100 results instead of 1000 to avoid excessive output
		for i := range 100 {
			rows.AddRow(i, fmt.Sprintf("Beer %d", i), "Style", "Brewery", "Country", 5.0, 30)
		}

		mock.ExpectQuery(expectedQuery).
			WithArgs(100).
			WillReturnRows(rows)

		start := time.Now()
		query := BeerSearchQuery{Limit: 100}
		results, err := svc.SearchBeers(context.Background(), query)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Len(t, results, 100)
		assert.Less(t, duration, 1*time.Second, "Query should complete within 1 second")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// SQL Injection Prevention Tests.
func TestSearchBeers_SQLInjectionPrevention(t *testing.T) {
	t.Run("SQL injection attempts in name", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()
		svc := setupBeerService(db)

		maliciousInput := "'; DROP TABLE beers; --"
		expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1\s+AND b.name ILIKE \$1`

		rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"})

		mock.ExpectQuery(expectedQuery).
			WithArgs("%" + maliciousInput + "%").
			WillReturnRows(rows)

		query := BeerSearchQuery{Name: maliciousInput}
		results, err := svc.SearchBeers(context.Background(), query)

		assert.NoError(t, err)
		assert.Empty(t, results)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Integration Test (requires real database).
func TestSearchBeers_Integration(t *testing.T) {
	// This test can be enabled when a test database is available
	t.Skip("Integration test requires test database")

	t.Run("Real database integration", func(t *testing.T) {
		db, err := sqlx.Open("postgres", "user=brewsource_user dbname=brewsource_test sslmode=disable")
		if err != nil {
			t.Skip("Skipping integration test: could not connect to test database")
		}
		defer db.Close()

		redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
		svc := NewBeerService(db, redisClient)

		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		query := BeerSearchQuery{Limit: 5}
		results, err := svc.SearchBeers(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, results)
		// If database has data, we should get results
		if len(results) > 0 {
			assert.LessOrEqual(t, len(results), 5)
			for _, result := range results {
				assert.NotEmpty(t, result.Name)
				assert.NotEmpty(t, result.Brewery)
				assert.Positive(t, result.ID)
			}
		}
	})
}

// Benchmark Tests.
func BenchmarkSearchBeers(b *testing.B) {
	db, mock := setupMockDB(&testing.T{})
	defer db.Close()
	svc := setupBeerService(db)

	expectedQuery := `SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu\s+FROM beers b\s+JOIN breweries br ON b.brewery_id = br.id\s+WHERE 1=1\s+AND b.name ILIKE \$1`

	rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"}).
		AddRow(mockBeerRows[0]...)

	for range b.N {
		mock.ExpectQuery(expectedQuery).
			WithArgs("%IPA%").
			WillReturnRows(rows)
	}

	query := BeerSearchQuery{Name: "IPA"}

	b.ResetTimer()
	for range b.N {
		_, _ = svc.SearchBeers(context.Background(), query)
	}
}
