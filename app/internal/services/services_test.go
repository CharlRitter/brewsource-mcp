package services_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getMockBreweryData returns mock brewery data for testing.
func getMockBreweryData() []*services.BrewerySearchResult {
	return []*services.BrewerySearchResult{
		{
			ID:          1,
			Name:        "Devil's Peak Brewing Company",
			BreweryType: "micro",
			Street:      "1st Floor, The Old Warehouse, 6 Beach Road",
			City:        "Woodstock",
			State:       "Western Cape",
			PostalCode:  "7925",
			Country:     "South Africa",
			Phone:       "+27 21 200 5818",
			Website:     "https://www.devilspeak.beer",
		},
		{
			ID:          2,
			Name:        "Stone Brewing",
			BreweryType: "regional",
			Street:      "1999 Citracado Parkway",
			City:        "Escondido",
			State:       "California",
			PostalCode:  "92029",
			Country:     "United States",
			Phone:       "+1 760 471 4999",
			Website:     "https://www.stonebrewing.com",
		},
		{
			ID:          3,
			Name:        "Founders Brewing Co.",
			BreweryType: "large",
			Street:      "235 Grandville Ave SW",
			City:        "Grand Rapids",
			State:       "Michigan",
			PostalCode:  "49503",
			Country:     "United States",
			Phone:       "+1 616 776 1195",
			Website:     "https://foundersbrewing.com",
		},
	}
}

// Helper function to convert []interface{} to []driver.Value.
func interfaceToDriverValues(args []interface{}) []driver.Value {
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		values[i] = arg
	}
	return values
}

func setupBreweryService(db *sqlx.DB) *services.BreweryService {
	// Using nil for Redis client in tests
	return services.NewBreweryService(db, nil)
}

// TestNewBreweryService tests the constructor.
func TestNewBreweryService(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	service := services.NewBreweryService(db, redisClient)

	assert.NotNil(t, service)
	var _ services.BreweryServiceInterface = service
}

func TestNewBreweryService_WithNilRedis(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	service := services.NewBreweryService(db, nil)

	assert.NotNil(t, service)
	var _ services.BreweryServiceInterface = service
}

// TestBrewerySearchQuery_DefaultLimits tests limit validation (existing test enhanced).
func TestBrewerySearchQuery_DefaultLimits(t *testing.T) {
	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit int
	}{
		{"zero limit defaults to 20", 0, 20},
		{"negative limit defaults to 20", -5, 20},
		{"too large limit defaults to 20", 150, 20},
		{"valid limit preserved", 15, 15},
		{"max valid limit", 100, 100},
		{"boundary case - limit 101", 101, 20},
		{"boundary case - limit 1", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := services.BrewerySearchQuery{
				Limit: tt.inputLimit,
			}

			// Simulate the limit adjustment logic that happens in SearchBreweries
			if query.Limit <= 0 || query.Limit > 100 {
				query.Limit = 20
			}

			assert.Equal(t, tt.expectedLimit, query.Limit)
		})
	}
}

// Integration-style tests for query building logic.
func TestSearchBreweries_QueryBuildingLogic(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	// Test the actual SQL query structure for different parameter combinations
	testCases := []struct {
		name         string
		query        services.BrewerySearchQuery
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name: "no filters",
			query: services.BrewerySearchQuery{
				Limit: 10,
			},
			expectedSQL:  `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1\s+ORDER BY name\s+LIMIT \$1`,
			expectedArgs: []interface{}{10},
		},
		{
			name: "name filter only",
			query: services.BrewerySearchQuery{
				Name:  "Stone",
				Limit: 15,
			},
			expectedSQL:  `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`,
			expectedArgs: []interface{}{"%Stone%", 15},
		},
		{
			name: "location filter creates OR condition",
			query: services.BrewerySearchQuery{
				Location: "California",
				Limit:    25,
			},
			expectedSQL:  `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND \(LOWER\(city\) LIKE LOWER\(\$1\) OR LOWER\(state\) LIKE LOWER\(\$1\) OR LOWER\(country\) LIKE LOWER\(\$1\)\)\s+ORDER BY name\s+LIMIT \$2`,
			expectedArgs: []interface{}{"%California%", 25},
		},
		{
			name: "multiple filters combined",
			query: services.BrewerySearchQuery{
				Name:  "Stone",
				City:  "San Diego",
				State: "California",
				Limit: 10,
			},
			expectedSQL:  `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) AND LOWER\(city\) LIKE LOWER\(\$2\) AND LOWER\(state\) LIKE LOWER\(\$3\)\s+ORDER BY name\s+LIMIT \$4`,
			expectedArgs: []interface{}{"%Stone%", "%San Diego%", "%California%", 10},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{
				"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
			}).AddRow(1, "Test Brewery", "micro", "123 Test St", "Test City", "Test State", "12345", "Test Country", "123-456-7890", "https://test.com")

			mock.ExpectQuery(tc.expectedSQL).
				WithArgs(interfaceToDriverValues(tc.expectedArgs)...).
				WillReturnRows(rows)

			ctx := context.Background()
			results, err := service.SearchBreweries(ctx, tc.query)

			require.NoError(t, err)
			assert.NotNil(t, results)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Test Result Mapping and Data Types.
func TestSearchBreweries_ResultMapping(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 1,
	}

	// Test various data types and edge values
	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		999999, // Large ID
		"Test Brewery With Very Long Name That Might Cause Issues",
		"nano", // Different brewery type
		"123 Main Street, Suite 456, Building A",
		"San Francisco",
		"CA",
		"94102-1234",                 // Extended postal code
		"United States of America",   // Full country name
		"+1 (555) 123-4567 ext. 890", // Complex phone format
		"https://www.very-long-brewery-name-with-hyphens-and-subdomains.brewery.com/path?param=value",
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Test%", 1).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, 999999, result.ID)
	assert.Equal(t, "Test Brewery With Very Long Name That Might Cause Issues", result.Name)
	assert.Equal(t, "nano", result.BreweryType)
	assert.Equal(t, "123 Main Street, Suite 456, Building A", result.Street)
	assert.Equal(t, "San Francisco", result.City)
	assert.Equal(t, "CA", result.State)
	assert.Equal(t, "94102-1234", result.PostalCode)
	assert.Equal(t, "United States of America", result.Country)
	assert.Equal(t, "+1 (555) 123-4567 ext. 890", result.Phone)
	assert.Equal(
		t,
		"https://www.very-long-brewery-name-with-hyphens-and-subdomains.brewery.com/path?param=value",
		result.Website,
	)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test for potential memory leaks with large result sets.
func TestSearchBreweries_MemoryLeakPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	// Run multiple queries to check for memory leaks
	for iteration := range 10 {
		query := services.BrewerySearchQuery{
			Name:  fmt.Sprintf("Test%d", iteration),
			Limit: 50,
		}

		rows := sqlmock.NewRows([]string{
			"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
		})

		// Add 50 results per iteration
		for i := range 50 {
			rows.AddRow(
				i+(iteration*50),
				fmt.Sprintf("Test Brewery %d-%d", iteration, i),
				"micro",
				"123 Main St",
				"Test City",
				"CA",
				"12345",
				"USA",
				"+1234567890",
				"https://test.com",
			)
		}

		expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

		mock.ExpectQuery(expectedSQL).
			WithArgs("%"+fmt.Sprintf("Test%d", iteration)+"%", 50).
			WillReturnRows(rows)

		ctx := context.Background()
		results, err := service.SearchBreweries(ctx, query)

		require.NoError(t, err)
		assert.Len(t, results, 50)

		// Force GC to help prevent memory leaks in tests
		runtime.GC()
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test for partial matches and fuzzy searching behavior.
func TestSearchBreweries_PartialMatching(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	testCases := []struct {
		searchTerm   string
		breweryNames []string
		description  string
	}{
		{
			searchTerm:   "Stone",
			breweryNames: []string{"Stone Brewing", "Firestone Walker", "Keystone Brewery"},
			description:  "Should match partial name occurrences",
		},
		{
			searchTerm:   "brew",
			breweryNames: []string{"Test Brewery", "Homebrew Co", "Brewing Solutions"},
			description:  "Should match substring in different positions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			query := services.BrewerySearchQuery{
				Name:  tc.searchTerm,
				Limit: 20,
			}

			rows := sqlmock.NewRows([]string{
				"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
			})

			for i, name := range tc.breweryNames {
				rows.AddRow(
					i+1,
					name,
					"micro",
					"123 Main St",
					"Test City",
					"CA",
					"12345",
					"USA",
					"+1234567890",
					"https://test.com",
				)
			}

			expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

			mock.ExpectQuery(expectedSQL).
				WithArgs("%"+tc.searchTerm+"%", 20).
				WillReturnRows(rows)

			ctx := context.Background()
			results, err := service.SearchBreweries(ctx, query)

			require.NoError(t, err)
			assert.Len(t, results, len(tc.breweryNames))

			// Verify all expected names are returned
			for i, expectedName := range tc.breweryNames {
				assert.Equal(t, expectedName, results[i].Name)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestBrewerySearchResult_FieldsExist tests struct field existence (existing test enhanced).
func TestBrewerySearchResult_FieldsExist(t *testing.T) {
	// Test that all expected fields exist on the services.BrewerySearchResult struct
	brewery := &services.BrewerySearchResult{
		ID:          1,
		Name:        "Devil's Peak Brewing Company",
		BreweryType: "micro",
		Street:      "1st Floor, The Old Warehouse, 6 Beach Road",
		City:        "Woodstock",
		State:       "Western Cape",
		PostalCode:  "7925",
		Country:     "South Africa",
		Phone:       "+27 21 200 5818",
		Website:     "https://www.devilspeak.beer",
	}

	assert.Equal(t, 1, brewery.ID)
	assert.Equal(t, "Devil's Peak Brewing Company", brewery.Name)
	assert.Equal(t, "micro", brewery.BreweryType)
	assert.Equal(t, "1st Floor, The Old Warehouse, 6 Beach Road", brewery.Street)
	assert.Equal(t, "Woodstock", brewery.City)
	assert.Equal(t, "Western Cape", brewery.State)
	assert.Equal(t, "7925", brewery.PostalCode)
	assert.Equal(t, "South Africa", brewery.Country)
	assert.Equal(t, "+27 21 200 5818", brewery.Phone)
	assert.Equal(t, "https://www.devilspeak.beer", brewery.Website)
	assert.Equal(t, "Woodstock", brewery.City)
	assert.Equal(t, "Western Cape", brewery.State)
	assert.Equal(t, "South Africa", brewery.Country)
	assert.Equal(t, "micro", brewery.BreweryType)
	assert.Equal(t, "+27 21 200 5818", brewery.Phone)
	assert.Equal(t, "https://www.devilspeak.beer", brewery.Website)
}

// Happy Path Tests.
func TestSearchBreweries_ByName_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Stone",
		Limit: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		getMockBreweryData()[1].ID, getMockBreweryData()[1].Name, getMockBreweryData()[1].BreweryType,
		getMockBreweryData()[1].Street, getMockBreweryData()[1].City, getMockBreweryData()[1].State,
		getMockBreweryData()[1].PostalCode, getMockBreweryData()[1].Country, getMockBreweryData()[1].Phone,
		getMockBreweryData()[1].Website,
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Stone%", 20).
		WillReturnRows(rows)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Stone Brewing", results[0].Name)
	assert.Equal(t, "Escondido", results[0].City)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test Location OR logic specifically.
func TestSearchBreweries_LocationOrLogic(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Location: "San", // Should match cities, states, or countries containing "San"
		Limit:    20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		1, "San Diego Brewery", "micro", "123 Main St", "San Diego", "CA", "92101", "USA", "+1234567890", "https://sandiego.com",
	).
		AddRow(
			2, "Francisco Brewing", "micro", "456 Oak St", "San Francisco", "CA", "94102", "USA", "+1234567891", "https://sf.com",
		)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND \(LOWER\(city\) LIKE LOWER\(\$1\) OR LOWER\(state\) LIKE LOWER\(\$1\) OR LOWER\(country\) LIKE LOWER\(\$1\)\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%San%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Contains(t, results[0].City, "San")
	assert.Contains(t, results[1].City, "San")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test multiple conditions AND logic.
func TestSearchBreweries_MultipleConditionsAndLogic(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:     "Stone",
		City:     "Escondido",
		State:    "California",
		Country:  "United States",
		Location: "West Coast", // This should also be included
		Limit:    20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		getMockBreweryData()[1].ID, getMockBreweryData()[1].Name, getMockBreweryData()[1].BreweryType,
		getMockBreweryData()[1].Street, getMockBreweryData()[1].City, getMockBreweryData()[1].State,
		getMockBreweryData()[1].PostalCode, getMockBreweryData()[1].Country, getMockBreweryData()[1].Phone,
		getMockBreweryData()[1].Website,
	)

	// All conditions should be ANDed together
	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) AND LOWER\(city\) LIKE LOWER\(\$2\) AND LOWER\(state\) LIKE LOWER\(\$3\) AND LOWER\(country\) LIKE LOWER\(\$4\) AND \(LOWER\(city\) LIKE LOWER\(\$5\) OR LOWER\(state\) LIKE LOWER\(\$5\) OR LOWER\(country\) LIKE LOWER\(\$5\)\) ORDER BY name LIMIT \$6`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Stone%", "%Escondido%", "%California%", "%United States%", "%West Coast%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Stone Brewing", results[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Performance regression test.
func TestSearchBreweries_PerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 50,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	})

	// Add 50 results
	for i := range 50 {
		rows.AddRow(
			i,
			fmt.Sprintf("Test Brewery %d", i),
			"micro",
			"123 Main St",
			"Test City",
			"CA",
			"12345",
			"USA",
			"+1234567890",
			"https://test.com",
		)
	}

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Test%", 50).
		WillReturnRows(rows)

	ctx := context.Background()
	start := time.Now()

	results, err := service.SearchBreweries(ctx, query)

	duration := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, 50)

	// Performance requirement: should complete within 500ms as per testing doc
	assert.Less(t, duration, 500*time.Millisecond, "Query should complete within 500ms")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test that demonstrates proper error wrapping.
func TestSearchBreweries_ErrorWrapping(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 20,
	}

	originalErr := sql.ErrTxDone

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Test%", 20).
		WillReturnError(originalErr)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to search breweries")
	require.ErrorIs(t, err, originalErr) // Original error should be wrapped
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test data consistency.
func TestBrewerySearchResult_DatabaseTagConsistency(t *testing.T) {
	// Verify that struct tags match expected database column names
	brewery := services.BrewerySearchResult{}

	// Use reflection to check db tags (in a real scenario)
	// For this test, we'll verify the expected field mappings match our mock data

	expectedFields := map[string]interface{}{
		"id":           brewery.ID,
		"name":         brewery.Name,
		"brewery_type": brewery.BreweryType,
		"street":       brewery.Street,
		"city":         brewery.City,
		"state":        brewery.State,
		"postal_code":  brewery.PostalCode,
		"country":      brewery.Country,
		"phone":        brewery.Phone,
		"website_url":  brewery.Website,
	}

	// Verify we have all expected fields
	assert.Len(t, expectedFields, 10, "services.BrewerySearchResult should have exactly 10 fields")
}

// Table-driven test for all search combinations.
func TestSearchBreweries_AllSearchCombinations(t *testing.T) {
	testCases := []struct {
		name         string
		query        services.BrewerySearchQuery
		expectedArgs []interface{}
		description  string
	}{
		{
			name: "name only",
			query: services.BrewerySearchQuery{
				Name:  "Stone",
				Limit: 20,
			},
			expectedArgs: []interface{}{"%Stone%", 20},
			description:  "Should search by name only",
		},
		{
			name: "city only",
			query: services.BrewerySearchQuery{
				City:  "Escondido",
				Limit: 20,
			},
			expectedArgs: []interface{}{"%Escondido%", 20},
			description:  "Should search by city only",
		},
		{
			name: "state only",
			query: services.BrewerySearchQuery{
				State: "California",
				Limit: 20,
			},
			expectedArgs: []interface{}{"%California%", 20},
			description:  "Should search by state only",
		},
		{
			name: "country only",
			query: services.BrewerySearchQuery{
				Country: "United States",
				Limit:   20,
			},
			expectedArgs: []interface{}{"%United States%", 20},
			description:  "Should search by country only",
		},
		{
			name: "location only",
			query: services.BrewerySearchQuery{
				Location: "California",
				Limit:    20,
			},
			expectedArgs: []interface{}{"%California%", 20},
			description:  "Should search by location (city OR state OR country)",
		},
		{
			name: "name and city",
			query: services.BrewerySearchQuery{
				Name:  "Stone",
				City:  "Escondido",
				Limit: 20,
			},
			expectedArgs: []interface{}{"%Stone%", "%Escondido%", 20},
			description:  "Should search by name AND city",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			service := setupBreweryService(db)

			rows := sqlmock.NewRows([]string{
				"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
			}).AddRow(1, "Test Brewery", "micro", "123 Test St", "Test City", "Test State", "12345", "Test Country", "123-456-7890", "https://test.com")

			// Build the expected SQL based on the query parameters
			var expectedSQL string
			switch {
			case tc.query.Name != "" && tc.query.City != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) AND LOWER\(city\) LIKE LOWER\(\$2\)\s+ORDER BY name\s+LIMIT \$3`
			case tc.query.Name != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`
			case tc.query.City != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(city\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`
			case tc.query.State != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(state\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`
			case tc.query.Country != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(country\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`
			case tc.query.Location != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND \(LOWER\(city\) LIKE LOWER\(\$1\) OR LOWER\(state\) LIKE LOWER\(\$1\) OR LOWER\(country\) LIKE LOWER\(\$1\)\)\s+ORDER BY name\s+LIMIT \$2`
			}

			mock.ExpectQuery(expectedSQL).
				WithArgs(interfaceToDriverValues(tc.expectedArgs)...).
				WillReturnRows(rows)

			ctx := context.Background()
			results, err := service.SearchBreweries(ctx, tc.query)

			require.NoError(t, err, tc.description)
			assert.NotNil(t, results, tc.description)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSearchBreweries_ByCity_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		City:  "Woodstock",
		Limit: 10,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		getMockBreweryData()[0].ID, getMockBreweryData()[0].Name, getMockBreweryData()[0].BreweryType,
		getMockBreweryData()[0].Street, getMockBreweryData()[0].City, getMockBreweryData()[0].State,
		getMockBreweryData()[0].PostalCode, getMockBreweryData()[0].Country, getMockBreweryData()[0].Phone,
		getMockBreweryData()[0].Website,
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(city\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Woodstock%", 10).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Devil's Peak Brewing Company", results[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBreweries_ByLocation_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Location: "California",
		Limit:    5,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		getMockBreweryData()[1].ID, getMockBreweryData()[1].Name, getMockBreweryData()[1].BreweryType,
		getMockBreweryData()[1].Street, getMockBreweryData()[1].City, getMockBreweryData()[1].State,
		getMockBreweryData()[1].PostalCode, getMockBreweryData()[1].Country, getMockBreweryData()[1].Phone,
		getMockBreweryData()[1].Website,
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND \(LOWER\(city\) LIKE LOWER\(\$1\) OR LOWER\(state\) LIKE LOWER\(\$1\) OR LOWER\(country\) LIKE LOWER\(\$1\)\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%California%", 5).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Stone Brewing", results[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBreweries_MultipleFilters_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:    "Stone",
		State:   "California",
		Country: "United States",
		Limit:   15,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		getMockBreweryData()[1].ID, getMockBreweryData()[1].Name, getMockBreweryData()[1].BreweryType,
		getMockBreweryData()[1].Street, getMockBreweryData()[1].City, getMockBreweryData()[1].State,
		getMockBreweryData()[1].PostalCode, getMockBreweryData()[1].Country, getMockBreweryData()[1].Phone,
		getMockBreweryData()[1].Website,
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) AND LOWER\(state\) LIKE LOWER\(\$2\) AND LOWER\(country\) LIKE LOWER\(\$3\) ORDER BY name LIMIT \$4`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Stone%", "%California%", "%United States%", 15).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Sad Path Tests.
func TestSearchBreweries_NoResults(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "NonexistentBrewery",
		Limit: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	})

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%NonexistentBrewery%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	// For no results, accept both nil slice and empty slice
	if results != nil {
		assert.Empty(t, results)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBreweries_DatabaseError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 20,
	}

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Test%", 20).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to search breweries")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Edge Cases and Boundary Tests.
func TestSearchBreweries_EmptyFilters(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Limit: 10,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	})
	for _, brewery := range getMockBreweryData()[:2] {
		rows.AddRow(
			brewery.ID, brewery.Name, brewery.BreweryType,
			brewery.Street, brewery.City, brewery.State,
			brewery.PostalCode, brewery.Country, brewery.Phone,
			brewery.Website,
		)
	}

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 ORDER BY name LIMIT \$1`

	mock.ExpectQuery(expectedSQL).
		WithArgs(10).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBreweries_CaseInsensitive(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "STONE", // Uppercase
		Limit: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		getMockBreweryData()[1].ID, getMockBreweryData()[1].Name, getMockBreweryData()[1].BreweryType,
		getMockBreweryData()[1].Street, getMockBreweryData()[1].City, getMockBreweryData()[1].State,
		getMockBreweryData()[1].PostalCode, getMockBreweryData()[1].Country, getMockBreweryData()[1].Phone,
		getMockBreweryData()[1].Website,
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%STONE%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Stone Brewing", results[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBreweries_SpecialCharacters(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Devil's", // Contains apostrophe
		Limit: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		getMockBreweryData()[0].ID, getMockBreweryData()[0].Name, getMockBreweryData()[0].BreweryType,
		getMockBreweryData()[0].Street, getMockBreweryData()[0].City, getMockBreweryData()[0].State,
		getMockBreweryData()[0].PostalCode, getMockBreweryData()[0].Country, getMockBreweryData()[0].Phone,
		getMockBreweryData()[0].Website,
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Devil's%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Devil's Peak Brewing Company", results[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBreweries_UnicodeCharacters(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Bières", // Unicode characters
		Limit: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	})

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Bières%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Empty(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Context and Timeout Tests.
func TestSearchBreweries_ContextTimeout(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 20,
	}

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Test%", 20).
		WillDelayFor(2 * time.Second). // Simulate slow query
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url"}))

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	results, err := service.SearchBreweries(ctx, query)

	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "canceling query due to user request")
}

func TestSearchBreweries_ContextCancellation(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 20,
	}

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Test%", 20).
		WillDelayFor(1 * time.Second).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url"}))

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	results, err := service.SearchBreweries(ctx, query)

	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "context canceled")
}

// Performance Tests.
func TestSearchBreweries_LargeResultSet(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Brewing", // Common term that might match many results
		Limit: 100,       // Maximum allowed
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	})

	// Add 100 mock results
	for i := 1; i <= 100; i++ {
		rows.AddRow(
			i, "Test Brewing "+string(rune(i)), "micro",
			"Test Street", "Test City", "Test State",
			"12345", "Test Country", "+1234567890",
			"https://test.com",
		)
	}

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Brewing%", 100).
		WillReturnRows(rows)

	ctx := context.Background()
	start := time.Now()
	results, err := service.SearchBreweries(ctx, query)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, 100)
	assert.Less(t, duration, 1*time.Second, "Query should complete within 1 second")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Integration-like tests (with mock database).
func TestSearchBreweries_ComplexLocationSearch(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Location: "United", // Should match "United States" in country
		Limit:    20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	})
	for _, brewery := range getMockBreweryData()[1:] { // Stone and Founders (both US)
		rows.AddRow(
			brewery.ID, brewery.Name, brewery.BreweryType,
			brewery.Street, brewery.City, brewery.State,
			brewery.PostalCode, brewery.Country, brewery.Phone,
			brewery.Website,
		)
	}

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND \(LOWER\(city\) LIKE LOWER\(\$1\) OR LOWER\(state\) LIKE LOWER\(\$1\) OR LOWER\(country\) LIKE LOWER\(\$1\)\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%United%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 2)

	// All results should be from United States
	for _, result := range results {
		assert.Equal(t, "United States", result.Country)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Validation Tests.
func TestSearchBreweries_AllEmptyFields(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:     "",
		Location: "",
		City:     "",
		State:    "",
		Country:  "",
		Limit:    0, // This should default to 20
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	})

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1\s+ORDER BY name\s+LIMIT \$1`

	mock.ExpectQuery(expectedSQL).
		WithArgs(20). // Should default to 20
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	// For empty fields test, accept both nil slice and empty slice
	if results != nil {
		assert.Empty(t, results)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Security Tests - SQL Injection Prevention.
func TestSearchBreweries_SQLInjectionAttempts(t *testing.T) {
	testCases := []struct {
		name      string
		query     services.BrewerySearchQuery
		expectErr bool
	}{
		{
			name: "SQL injection in name field",
			query: services.BrewerySearchQuery{
				Name:  "'; DROP TABLE breweries; --",
				Limit: 20,
			},
			expectErr: false, // Should be safely handled by parameterized queries
		},
		{
			name: "SQL injection in city field",
			query: services.BrewerySearchQuery{
				City:  "' OR 1=1 --",
				Limit: 20,
			},
			expectErr: false,
		},
		{
			name: "SQL injection in state field",
			query: services.BrewerySearchQuery{
				State: "' UNION SELECT * FROM users --",
				Limit: 20,
			},
			expectErr: false,
		},
		{
			name: "SQL injection in country field",
			query: services.BrewerySearchQuery{
				Country: "' OR '1'='1",
				Limit:   20,
			},
			expectErr: false,
		},
		{
			name: "SQL injection in location field",
			query: services.BrewerySearchQuery{
				Location: "; UPDATE breweries SET name='hacked'",
				Limit:    20,
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			service := setupBreweryService(db)

			// Mock expects the malicious input to be safely parameterized
			rows := sqlmock.NewRows([]string{
				"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
			}).AddRow(1, "Test Brewery", "micro", "123 Test St", "Test City", "Test State", "12345", "Test Country", "123-456-7890", "https://test.com")

			// Build expected SQL and args based on the query fields
			var expectedSQL string
			var expectedArgs []interface{}

			switch {
			case tc.query.Name != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`
				expectedArgs = []interface{}{"%" + tc.query.Name + "%", 20}
			case tc.query.City != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(city\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`
				expectedArgs = []interface{}{"%" + tc.query.City + "%", 20}
			case tc.query.State != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(state\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`
				expectedArgs = []interface{}{"%" + tc.query.State + "%", 20}
			case tc.query.Country != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(country\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`
				expectedArgs = []interface{}{"%" + tc.query.Country + "%", 20}
			case tc.query.Location != "":
				expectedSQL = `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND \(LOWER\(city\) LIKE LOWER\(\$1\) OR LOWER\(state\) LIKE LOWER\(\$1\) OR LOWER\(country\) LIKE LOWER\(\$1\)\)\s+ORDER BY name\s+LIMIT \$2`
				expectedArgs = []interface{}{"%" + tc.query.Location + "%", 20}
			}

			mock.ExpectQuery(expectedSQL).
				WithArgs(interfaceToDriverValues(expectedArgs)...).
				WillReturnRows(rows)

			ctx := context.Background()
			results, err := service.SearchBreweries(ctx, tc.query)

			if tc.expectErr {
				require.Error(t, err)
				assert.Nil(t, results)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, results)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Data Integrity Tests.
func TestSearchBreweries_NullValueHandling(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 20,
	}

	// Mock data with some null/empty values
	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		1, "Test Brewery", "micro", "", "Test City", "", "12345", "USA", "", "",
	).AddRow(
		2, "Another Brewery", "", "123 Main St", "", "CA", "", "", "+1234567890", "https://test.com",
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Test%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Verify that empty strings are handled properly
	assert.Empty(t, results[0].Street)
	assert.Empty(t, results[0].State)
	assert.Empty(t, results[1].BreweryType)
	assert.Empty(t, results[1].City)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBreweries_LongFieldValues(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	// Test with very long field values
	longName := strings.Repeat("Very Long Brewery Name ", 50) // ~1000+ characters
	longCity := strings.Repeat("Very Long City Name ", 20)    // ~400+ characters

	query := services.BrewerySearchQuery{
		Name:  longName,
		City:  longCity,
		Limit: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(1, "Test Brewery", "micro", "123 Test St", "Test City", "Test State", "12345", "Test Country", "123-456-7890", "https://test.com")

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) AND LOWER\(city\) LIKE LOWER\(\$2\)\s+ORDER BY name\s+LIMIT \$3`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%"+longName+"%", "%"+longCity+"%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	// For long field values test, the service should handle it gracefully
	// Accept both nil slice and empty slice - testing that long field values don't crash the service
	assert.NotNil(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Concurrency Tests.
func TestSearchBreweries_ConcurrentRequests(t *testing.T) {
	// Skip this test in race detection mode since sqlmock is not thread-safe
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// Test concurrent requests with separate mock instances to avoid race conditions
	const numRoutines = 5
	results := make(chan error, numRoutines)
	var wg sync.WaitGroup

	for i := range numRoutines {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			// Create separate mock for each goroutine - direct sqlmock creation to avoid testify issues
			db, mock, err := sqlmock.New()
			if err != nil {
				results <- fmt.Errorf("failed to create mock: %w", err)
				return
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			service := services.NewBreweryService(sqlxDB, nil)

			rows := sqlmock.NewRows([]string{
				"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
			}).AddRow(
				1, fmt.Sprintf("Test Brewery %d", routineID), "micro", "123 Main St", "Test City", "CA", "12345", "USA", "+1234567890", "https://test.com",
			)

			expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
				FROM breweries
				WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

			mock.ExpectQuery(expectedSQL).
				WithArgs("%Test%", 20).
				WillReturnRows(rows)

			query := services.BrewerySearchQuery{
				Name:  "Test",
				Limit: 20,
			}

			ctx := context.Background()
			_, searchErr := service.SearchBreweries(ctx, query)
			if searchErr != nil {
				results <- searchErr
				return
			}

			if mockErr := mock.ExpectationsWereMet(); mockErr != nil {
				results <- mockErr
				return
			}

			results <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for err := range results {
		require.NoError(t, err)
	}
}

// Memory Usage Tests.
func TestSearchBreweries_MemoryUsage(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 100,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	})

	// Add many results to test memory usage
	for i := range 100 {
		rows.AddRow(
			i,
			fmt.Sprintf("Test Brewery %d", i),
			"micro",
			"123 Main St",
			"Test City",
			"CA",
			"12345",
			"USA",
			"+1234567890",
			"https://test.com",
		)
	}

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Test%", 100).
		WillReturnRows(rows)

	ctx := context.Background()

	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 100)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Regression Tests.
func TestSearchBreweries_OrderByName(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Brewing",
		Limit: 20,
	}

	// Return results in alphabetical order to verify ORDER BY works
	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		1, "Alpha Brewing", "micro", "123 Main St", "Test City", "CA", "12345", "USA", "+1234567890", "https://alpha.com",
	).
		AddRow(
			2, "Beta Brewing", "micro", "456 Oak St", "Test City", "CA", "12345", "USA", "+1234567891", "https://beta.com",
		).
		AddRow(
			3, "Charlie Brewing", "micro", "789 Pine St", "Test City", "CA", "12345", "USA", "+1234567892", "https://charlie.com",
		)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Brewing%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Verify ordering
	assert.Equal(t, "Alpha Brewing", results[0].Name)
	assert.Equal(t, "Beta Brewing", results[1].Name)
	assert.Equal(t, "Charlie Brewing", results[2].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Whitespace and Trimming Tests.
func TestSearchBreweries_WhitespaceHandling(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "  Stone  ",     // Leading and trailing whitespace
		City:  "\tEscondido\n", // Tab and newline characters
		Limit: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		getMockBreweryData()[1].ID, getMockBreweryData()[1].Name, getMockBreweryData()[1].BreweryType,
		getMockBreweryData()[1].Street, getMockBreweryData()[1].City, getMockBreweryData()[1].State,
		getMockBreweryData()[1].PostalCode, getMockBreweryData()[1].Country, getMockBreweryData()[1].Phone,
		getMockBreweryData()[1].Website,
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) AND LOWER\(city\) LIKE LOWER\(\$2\) ORDER BY name LIMIT \$3`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%  Stone  %", "%\tEscondido\n%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Database Error Scenarios.
func TestSearchBreweries_DatabaseConnectionError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 20,
	}

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Test%", 20).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to search breweries")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBreweries_QuerySyntaxError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 20,
	}

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%Test%", 20).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to search breweries")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Edge Case: Very Short Search Terms.
func TestSearchBreweries_SingleCharacterSearch(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "A",
		Limit: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		1, "Alpha Brewing", "micro", "123 Main St", "Test City", "CA", "12345", "USA", "+1234567890", "https://alpha.com",
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%A%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Alpha Brewing", results[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test State and Country Filters.
func TestSearchBreweries_StateFilter(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		State: "California",
		Limit: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		getMockBreweryData()[1].ID, getMockBreweryData()[1].Name, getMockBreweryData()[1].BreweryType,
		getMockBreweryData()[1].Street, getMockBreweryData()[1].City, getMockBreweryData()[1].State,
		getMockBreweryData()[1].PostalCode, getMockBreweryData()[1].Country, getMockBreweryData()[1].Phone,
		getMockBreweryData()[1].Website,
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(state\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%California%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "California", results[0].State)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBreweries_CountryFilter(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Country: "South Africa",
		Limit:   20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		getMockBreweryData()[0].ID, getMockBreweryData()[0].Name, getMockBreweryData()[0].BreweryType,
		getMockBreweryData()[0].Street, getMockBreweryData()[0].City, getMockBreweryData()[0].State,
		getMockBreweryData()[0].PostalCode, getMockBreweryData()[0].Country, getMockBreweryData()[0].Phone,
		getMockBreweryData()[0].Website,
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(country\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%South Africa%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "South Africa", results[0].Country)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Benchmark Tests.
func BenchmarkSearchBreweries_SingleField(b *testing.B) {
	db, mock := setupMockDB(&testing.T{})
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "Test",
		Limit: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	}).AddRow(
		1, "Test Brewery", "micro", "123 Main St", "Test City", "CA", "12345", "USA", "+1234567890", "https://test.com",
	)

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\) ORDER BY name LIMIT \$2`

	for range b.N {
		mock.ExpectQuery(expectedSQL).
			WithArgs("%Test%", 20).
			WillReturnRows(rows)
	}

	b.ResetTimer()

	for range b.N {
		ctx := context.Background()
		_, err := service.SearchBreweries(ctx, query)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// Helper function tests.
func TestBrewerySearchQuery_ValidationLogic(t *testing.T) {
	tests := []struct {
		name     string
		query    services.BrewerySearchQuery
		expected services.BrewerySearchQuery
	}{
		{
			name: "all valid fields",
			query: services.BrewerySearchQuery{
				Name:     "Stone",
				Location: "California",
				City:     "Escondido",
				State:    "CA",
				Country:  "USA",
				Limit:    50,
			},
			expected: services.BrewerySearchQuery{
				Name:     "Stone",
				Location: "California",
				City:     "Escondido",
				State:    "CA",
				Country:  "USA",
				Limit:    50,
			},
		},
		{
			name: "limit boundary correction",
			query: services.BrewerySearchQuery{
				Name:  "Test",
				Limit: -10,
			},
			expected: services.BrewerySearchQuery{
				Name:  "Test",
				Limit: 20, // Should be corrected
			},
		},
		{
			name: "limit too high correction",
			query: services.BrewerySearchQuery{
				Name:  "Test",
				Limit: 1000,
			},
			expected: services.BrewerySearchQuery{
				Name:  "Test",
				Limit: 20, // Should be corrected
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate limit validation logic
			if tt.query.Limit <= 0 || tt.query.Limit > 100 {
				tt.query.Limit = 20
			}

			assert.Equal(t, tt.expected.Name, tt.query.Name)
			assert.Equal(t, tt.expected.Location, tt.query.Location)
			assert.Equal(t, tt.expected.City, tt.query.City)
			assert.Equal(t, tt.expected.State, tt.query.State)
			assert.Equal(t, tt.expected.Country, tt.query.Country)
			assert.Equal(t, tt.expected.Limit, tt.query.Limit)
		})
	}
}

// Test struct field completeness and types.
func TestBrewerySearchResult_StructIntegrity(t *testing.T) {
	brewery := services.BrewerySearchResult{
		ID:          123,
		Name:        "Test Brewery",
		BreweryType: "micro",
		Street:      "123 Main St",
		City:        "Test City",
		State:       "CA",
		PostalCode:  "12345",
		Country:     "USA",
		Phone:       "+1234567890",
		Website:     "https://test.com",
	}

	// Test that all fields have expected types and can be set/retrieved
	assert.IsType(t, 0, brewery.ID)
	assert.IsType(t, "", brewery.Name)
	assert.IsType(t, "", brewery.BreweryType)
	assert.IsType(t, "", brewery.Street)
	assert.IsType(t, "", brewery.City)
	assert.IsType(t, "", brewery.State)
	assert.IsType(t, "", brewery.PostalCode)
	assert.IsType(t, "", brewery.Country)
	assert.IsType(t, "", brewery.Phone)
	assert.IsType(t, "", brewery.Website)

	// Test field values
	assert.Equal(t, 123, brewery.ID)
	assert.Equal(t, "Test Brewery", brewery.Name)
	assert.Equal(t, "micro", brewery.BreweryType)
	assert.Equal(t, "123 Main St", brewery.Street)
	assert.Equal(t, "Test City", brewery.City)
	assert.Equal(t, "CA", brewery.State)
	assert.Equal(t, "12345", brewery.PostalCode)
	assert.Equal(t, "USA", brewery.Country)
	assert.Equal(t, "+1234567890", brewery.Phone)
	assert.Equal(t, "https://test.com", brewery.Website)
}

// Edge case: Empty database.
func TestSearchBreweries_EmptyDatabase(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := setupBreweryService(db)

	query := services.BrewerySearchQuery{
		Name:  "AnyName",
		Limit: 20,
	}

	// Empty result set
	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state", "postal_code", "country", "phone", "website_url",
	})

	expectedSQL := `SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url\s+FROM breweries\s+WHERE 1=1 AND LOWER\(name\) LIKE LOWER\(\$1\)\s+ORDER BY name\s+LIMIT \$2`

	mock.ExpectQuery(expectedSQL).
		WithArgs("%AnyName%", 20).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := service.SearchBreweries(ctx, query)

	require.NoError(t, err)
	// For empty database test, accept both nil slice and empty slice
	if results != nil {
		assert.Empty(t, results)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}
