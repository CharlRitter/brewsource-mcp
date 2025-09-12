// Package models_test contains tests for the data models and database schema in Brewsource MCP.
package models_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/CharlRitter/brewsource-mcp/app/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test models.StringArray type implementation.
func TestStringArray_Scan(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		expected  models.StringArray
		expectErr bool
	}{
		{
			name:     "scan nil value",
			value:    nil,
			expected: nil,
		},
		{
			name:     "scan empty json array as bytes",
			value:    []byte("[]"),
			expected: models.StringArray{},
		},
		{
			name:     "scan json array as bytes",
			value:    []byte(`["item1", "item2", "item3"]`),
			expected: models.StringArray{"item1", "item2", "item3"},
		},
		{
			name:     "scan json array as string",
			value:    `["beer", "brewing", "hops"]`,
			expected: models.StringArray{"beer", "brewing", "hops"},
		},
		{
			name:     "scan empty string array",
			value:    `[]`,
			expected: models.StringArray{},
		},
		{
			name:     "scan single item array",
			value:    `["single"]`,
			expected: models.StringArray{"single"},
		},
		// Error cases
		{
			name:      "scan invalid JSON",
			value:     []byte(`["invalid json`),
			expectErr: true,
		},
		{
			name:      "scan non-string type",
			value:     123,
			expectErr: true,
		},
		{
			name:      "scan invalid JSON string",
			value:     `{"key": "value"}`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s models.StringArray
			err := s.Scan(tt.value)

			if (err != nil) != tt.expectErr {
				t.Errorf("models.StringArray.Scan() error = %v, expectErr = %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr && !reflect.DeepEqual(s, tt.expected) {
				t.Errorf("models.StringArray.Scan() = %v, want %v", s, tt.expected)
			}
		})
	}
}

func TestStringArray_Value(t *testing.T) {
	tests := []struct {
		name      string
		array     models.StringArray
		expected  driver.Value
		expectErr bool
	}{
		{
			name:     "nil array",
			array:    nil,
			expected: "[]",
		},
		{
			name:     "empty array",
			array:    models.StringArray{},
			expected: "[]",
		},
		{
			name:     "single item array",
			array:    models.StringArray{"item1"},
			expected: []byte(`["item1"]`),
		},
		{
			name:     "multiple items array",
			array:    models.StringArray{"item1", "item2", "item3"},
			expected: []byte(`["item1","item2","item3"]`),
		},
		{
			name:     "array with special characters",
			array:    models.StringArray{"item with spaces", "item\nwith\nnewlines", "item\"with\"quotes"},
			expected: []byte(`["item with spaces","item\nwith\nnewlines","item\"with\"quotes"]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.array.Value()

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assertValueEquals(t, tt.expected, value)
		})
	}
}

// Helper function to reduce complexity.
func assertValueEquals(t *testing.T, expected, actual driver.Value) {
	t.Helper()

	switch exp := expected.(type) {
	case string:
		assert.Equal(t, exp, actual)
	case []byte:
		actualBytes, ok := actual.([]byte)
		require.True(t, ok, "expected actual value to be []byte")
		assert.Equal(t, exp, actualBytes)
	default:
		assert.Equal(t, expected, actual)
	}
}

// Test models.Beer model structure and JSON marshaling.
func TestBeer_JSONMarshaling(t *testing.T) {
	now := time.Now().UTC()
	beer := models.Beer{
		ID:          1,
		Name:        "Test IPA",
		BreweryID:   10,
		Style:       "American IPA",
		ABV:         6.5,
		IBU:         65,
		SRM:         8.5,
		Description: "A hoppy IPA with citrus notes",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Test marshaling
	data, err := json.Marshal(beer)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Test unmarshaling
	var unmarshaled models.Beer
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify fields
	if unmarshaled.ID != beer.ID {
		t.Errorf("Unmarshaled ID = %d, want %d", unmarshaled.ID, beer.ID)
	}
	if unmarshaled.Name != beer.Name {
		t.Errorf("Unmarshaled Name = %q, want %q", unmarshaled.Name, beer.Name)
	}
	if unmarshaled.BreweryID != beer.BreweryID {
		t.Errorf("Unmarshaled BreweryID = %d, want %d", unmarshaled.BreweryID, beer.BreweryID)
	}
	if unmarshaled.Style != beer.Style {
		t.Errorf("Unmarshaled Style = %q, want %q", unmarshaled.Style, beer.Style)
	}
	if unmarshaled.ABV != beer.ABV {
		t.Errorf("Unmarshaled ABV = %f, want %f", unmarshaled.ABV, beer.ABV)
	}
	if unmarshaled.IBU != beer.IBU {
		t.Errorf("Unmarshaled IBU = %d, want %d", unmarshaled.IBU, beer.IBU)
	}
	if unmarshaled.SRM != beer.SRM {
		t.Errorf("Unmarshaled SRM = %f, want %f", unmarshaled.SRM, beer.SRM)
	}
	if unmarshaled.Description != beer.Description {
		t.Errorf("Unmarshaled Description = %q, want %q", unmarshaled.Description, beer.Description)
	}
}

// Test models.Brewery model structure and JSON marshaling.
func TestBrewery_JSONMarshaling(t *testing.T) {
	now := time.Now().UTC()
	brewery := models.Brewery{
		ID:          1,
		Name:        "Test models.Brewery",
		BreweryType: "micro",
		Street:      "123 Test Street",
		City:        "Test City",
		State:       "Test State",
		PostalCode:  "12345",
		Country:     "Test Country",
		Phone:       "+1-234-567-8900",
		WebsiteURL:  "https://testbrewery.com",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Test marshaling
	data, err := json.Marshal(brewery)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Test unmarshaling
	var unmarshaled models.Brewery
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify fields
	if unmarshaled.ID != brewery.ID {
		t.Errorf("Unmarshaled ID = %d, want %d", unmarshaled.ID, brewery.ID)
	}
	if unmarshaled.Name != brewery.Name {
		t.Errorf("Unmarshaled Name = %q, want %q", unmarshaled.Name, brewery.Name)
	}
	if unmarshaled.BreweryType != brewery.BreweryType {
		t.Errorf("Unmarshaled BreweryType = %q, want %q", unmarshaled.BreweryType, brewery.BreweryType)
	}
	if unmarshaled.City != brewery.City {
		t.Errorf("Unmarshaled City = %q, want %q", unmarshaled.City, brewery.City)
	}
	if unmarshaled.State != brewery.State {
		t.Errorf("Unmarshaled State = %q, want %q", unmarshaled.State, brewery.State)
	}
	if unmarshaled.Country != brewery.Country {
		t.Errorf("Unmarshaled Country = %q, want %q", unmarshaled.Country, brewery.Country)
	}
}

// Test MigrateDatabase function.
func TestMigrateDatabase(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		expectErr bool
	}{
		{
			name: "successful migration",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect all migration queries to succeed
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS breweries").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS beers").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE INDEX IF NOT EXISTS idx_breweries_name").
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE INDEX IF NOT EXISTS idx_breweries_location").
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE INDEX IF NOT EXISTS idx_beers_name").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE INDEX IF NOT EXISTS idx_beers_brewery").
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE INDEX IF NOT EXISTS idx_beers_style").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE OR REPLACE FUNCTION update_updated_at_column").
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("DROP TRIGGER IF EXISTS update_breweries_updated_at").
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE TRIGGER update_breweries_updated_at").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("DROP TRIGGER IF EXISTS update_beers_updated_at").
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE TRIGGER update_beers_updated_at").WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectErr: false,
		},
		{
			name: "failed table creation",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS breweries").WillReturnError(sql.ErrConnDone)
			},
			expectErr: true,
		},
		{
			name: "failed index creation",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS breweries").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS beers").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE INDEX IF NOT EXISTS idx_breweries_name").WillReturnError(sql.ErrConnDone)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock: %v", err)
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			tt.setupMock(mock)

			err = models.MigrateDatabase(sqlxDB)

			if (err != nil) != tt.expectErr {
				t.Errorf("MigrateDatabase() error = %v, expectErr = %v", err, tt.expectErr)
			}

			if mockErr := mock.ExpectationsWereMet(); mockErr != nil {
				t.Errorf("Unfulfilled expectations: %s", mockErr)
			}
		})
	}
}

// Test edge cases for models.Beer model.
func TestBeer_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		beer models.Beer
	}{
		{
			name: "beer with zero values",
			beer: models.Beer{
				ID:        0,
				Name:      "",
				BreweryID: 0,
				Style:     "",
				ABV:       0.0,
				IBU:       0,
				SRM:       0.0,
			},
		},
		{
			name: "beer with maximum values",
			beer: models.Beer{
				ID:        2147483647, // max int32
				Name:      "Very Long models.Beer Name That Exceeds Normal Length Limits",
				BreweryID: 2147483647,
				Style:     "Very Long Style Name That Might Exceed Database Limits",
				ABV:       99.99, // theoretical maximum ABV
				IBU:       1000,  // theoretical maximum IBU
				SRM:       500.0, // theoretical maximum SRM
			},
		},
		{
			name: "beer with special characters",
			beer: models.Beer{
				Name:        "models.Beer with Special Characters: !@#$%^&*()_+-={}[]|\\:;\"'<>?,./",
				Style:       "Style with Ümlåuts and Açcénts",
				Description: "Description with\nnewlines\tand\ttabs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling doesn't panic or error
			data, err := json.Marshal(tt.beer)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
			}

			// Test unmarshaling works
			var unmarshaled models.Beer
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
			}

			// Basic sanity check
			if unmarshaled.Name != tt.beer.Name {
				t.Errorf("Name not preserved: got %q, want %q", unmarshaled.Name, tt.beer.Name)
			}
		})
	}
}

// Test edge cases for models.Brewery model.
func TestBrewery_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		brewery models.Brewery
	}{
		{
			name: "brewery with empty values",
			brewery: models.Brewery{
				Name:        "",
				BreweryType: "",
				Street:      "",
				City:        "",
				State:       "",
				PostalCode:  "",
				Country:     "",
				Phone:       "",
				WebsiteURL:  "",
			},
		},
		{
			name: "brewery with very long values",
			brewery: models.Brewery{
				Name:        "Very Long models.Brewery Name That Might Exceed Normal Database Field Limits",
				BreweryType: "Very Long models.Brewery Type",
				Street:      "Very Long Street Address That Includes Multiple Lines And Apartment Numbers",
				City:        "Very Long City Name",
				State:       "Very Long State Or Province Name",
				PostalCode:  "VERYLONGPOSTALCODE123456",
				Country:     "Very Long Country Name",
				Phone:       "+1-234-567-8900-ext-12345-very-long-extension",
				WebsiteURL:  "https://very-long-subdomain.very-long-domain-name.com/very-long-path",
			},
		},
		{
			name: "brewery with international characters",
			brewery: models.Brewery{
				Name:    "Bräuerei München",
				City:    "São Paulo",
				State:   "Zürich",
				Country: "España",
			},
		},
		{
			name: "brewery with special URLs and phone formats",
			brewery: models.Brewery{
				Phone:      "+27-21-123-4567",
				WebsiteURL: "https://test.co.za/path?param=value&other=123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling doesn't panic or error
			data, err := json.Marshal(tt.brewery)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
			}

			// Test unmarshaling works
			var unmarshaled models.Brewery
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
			}

			// Basic sanity check
			if unmarshaled.Name != tt.brewery.Name {
				t.Errorf("Name not preserved: got %q, want %q", unmarshaled.Name, tt.brewery.Name)
			}
		})
	}
}

// Test models.StringArray with edge cases.
func TestStringArray_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected models.StringArray
		hasError bool
	}{
		{
			name:     "very large array",
			input:    `["item1", "item2", "item3", "item4", "item5"]`,
			expected: models.StringArray{"item1", "item2", "item3", "item4", "item5"},
		},
		{
			name:     "array with unicode characters",
			input:    `["🍺", "🍻", "bière", "çerveza", "пиво"]`,
			expected: models.StringArray{"🍺", "🍻", "bière", "çerveza", "пиво"},
		},
		{
			name:     "array with nested quotes",
			input:    `["He said \"hello\"", "It's a test", "Multiple \"quotes\" here"]`,
			expected: models.StringArray{`He said "hello"`, "It's a test", `Multiple "quotes" here`},
		},
		{
			name:     "array with control characters",
			input:    `["line1\nline2", "tab\there", "carriage\rreturn"]`,
			expected: models.StringArray{"line1\nline2", "tab\there", "carriage\rreturn"},
		},
		{
			name:     "malformed JSON with missing bracket",
			input:    `["item1", "item2"`,
			hasError: true,
		},
		{
			name:     "non-array JSON",
			input:    `"not an array"`,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s models.StringArray
			err := s.Scan(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(s, tt.expected) {
					t.Errorf("models.StringArray.Scan() = %v, want %v", s, tt.expected)
				}
			}
		})
	}
}

// Benchmark tests for performance.
func BenchmarkStringArray_Scan(b *testing.B) {
	jsonData := `["beer", "brewing", "hops", "malt", "yeast", "water", "fermentation"]`

	b.ResetTimer()
	for range b.N {
		var s models.StringArray
		err := s.Scan(jsonData)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkStringArray_Value(b *testing.B) {
	s := models.StringArray{"beer", "brewing", "hops", "malt", "yeast", "water", "fermentation"}

	b.ResetTimer()
	for range b.N {
		_, err := s.Value()
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkBeer_JSONMarshal(b *testing.B) {
	beer := models.Beer{
		ID:          1,
		Name:        "Test IPA",
		BreweryID:   10,
		Style:       "American IPA",
		ABV:         6.5,
		IBU:         65,
		SRM:         8.5,
		Description: "A hoppy IPA with citrus and pine notes",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	b.ResetTimer()
	for range b.N {
		_, err := json.Marshal(beer)
		if err != nil {
			b.Fatalf("json.Marshal() error: %v", err)
		}
	}
}

// Test database field tags.
func TestModel_DatabaseTags(t *testing.T) {
	beerType := reflect.TypeOf(models.Beer{})

	// Test models.Beer struct tags
	expectedBeerTags := map[string]string{
		"ID":          "id",
		"Name":        "name",
		"BreweryID":   "brewery_id",
		"Style":       "style",
		"ABV":         "abv",
		"IBU":         "ibu",
		"SRM":         "srm",
		"Description": "description",
		"CreatedAt":   "created_at",
		"UpdatedAt":   "updated_at",
	}

	for fieldName, expectedTag := range expectedBeerTags {
		field, found := beerType.FieldByName(fieldName)
		if !found {
			t.Errorf("Field %s not found in models.Beer struct", fieldName)
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag != expectedTag {
			t.Errorf("Field %s has db tag %q, want %q", fieldName, dbTag, expectedTag)
		}
	}

	// Test models.Brewery struct tags
	breweryType := reflect.TypeOf(models.Brewery{})
	expectedBreweryTags := map[string]string{
		"ID":          "id",
		"Name":        "name",
		"BreweryType": "brewery_type",
		"Street":      "street",
		"City":        "city",
		"State":       "state",
		"PostalCode":  "postal_code",
		"Country":     "country",
		"Phone":       "phone",
		"WebsiteURL":  "website_url",
		"CreatedAt":   "created_at",
		"UpdatedAt":   "updated_at",
	}

	for fieldName, expectedTag := range expectedBreweryTags {
		field, found := breweryType.FieldByName(fieldName)
		if !found {
			t.Errorf("Field %s not found in models.Brewery struct", fieldName)
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag != expectedTag {
			t.Errorf("Field %s has db tag %q, want %q", fieldName, dbTag, expectedTag)
		}
	}
}
