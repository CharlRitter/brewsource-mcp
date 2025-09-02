package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

// Test StringArray type implementation.
func TestStringArray_Scan(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		expected  StringArray
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
			expected: StringArray{},
		},
		{
			name:     "scan json array as bytes",
			value:    []byte(`["item1", "item2", "item3"]`),
			expected: StringArray{"item1", "item2", "item3"},
		},
		{
			name:     "scan json array as string",
			value:    `["beer", "brewing", "hops"]`,
			expected: StringArray{"beer", "brewing", "hops"},
		},
		{
			name:     "scan empty string array",
			value:    `[]`,
			expected: StringArray{},
		},
		{
			name:     "scan single item array",
			value:    `["single"]`,
			expected: StringArray{"single"},
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
			var s StringArray
			err := s.Scan(tt.value)

			if (err != nil) != tt.expectErr {
				t.Errorf("StringArray.Scan() error = %v, expectErr = %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr && !reflect.DeepEqual(s, tt.expected) {
				t.Errorf("StringArray.Scan() = %v, want %v", s, tt.expected)
			}
		})
	}
}

func TestStringArray_Value(t *testing.T) {
	tests := []struct {
		name      string
		array     StringArray
		expected  driver.Value
		expectErr bool
	}{
		{
			name:     "nil array",
			array:    nil,
			expected: nil,
		},
		{
			name:     "empty array",
			array:    StringArray{},
			expected: []byte("[]"),
		},
		{
			name:     "single item array",
			array:    StringArray{"item1"},
			expected: []byte(`["item1"]`),
		},
		{
			name:     "multiple items array",
			array:    StringArray{"item1", "item2", "item3"},
			expected: []byte(`["item1","item2","item3"]`),
		},
		{
			name:     "array with special characters",
			array:    StringArray{"item with spaces", "item\nwith\nnewlines", "item\"with\"quotes"},
			expected: []byte(`["item with spaces","item\nwith\nnewlines","item\"with\"quotes"]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.array.Value()

			if (err != nil) != tt.expectErr {
				t.Errorf("StringArray.Value() error = %v, expectErr = %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr {
				if tt.expected == nil && value != nil {
					t.Errorf("StringArray.Value() = %v, want nil", value)
					return
				}
				if tt.expected != nil && value == nil {
					t.Errorf("StringArray.Value() = nil, want %v", tt.expected)
					return
				}
				if tt.expected != nil && value != nil {
					expectedBytes, _ := tt.expected.([]byte)
					actualBytes, _ := value.([]byte)
					if !reflect.DeepEqual(actualBytes, expectedBytes) {
						t.Errorf("StringArray.Value() = %s, want %s", actualBytes, expectedBytes)
					}
				}
			}
		})
	}
}

// Test Beer model structure and JSON marshaling.
func TestBeer_JSONMarshaling(t *testing.T) {
	now := time.Now().UTC()
	beer := Beer{
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
	var unmarshaled Beer
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

// Test Brewery model structure and JSON marshaling.
func TestBrewery_JSONMarshaling(t *testing.T) {
	now := time.Now().UTC()
	brewery := Brewery{
		ID:          1,
		Name:        "Test Brewery",
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
	var unmarshaled Brewery
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

			err = MigrateDatabase(sqlxDB)

			if (err != nil) != tt.expectErr {
				t.Errorf("MigrateDatabase() error = %v, expectErr = %v", err, tt.expectErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}
		})
	}
}

// Test edge cases for Beer model.
func TestBeer_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		beer Beer
	}{
		{
			name: "beer with zero values",
			beer: Beer{
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
			beer: Beer{
				ID:        2147483647, // max int32
				Name:      "Very Long Beer Name That Exceeds Normal Length Limits",
				BreweryID: 2147483647,
				Style:     "Very Long Style Name That Might Exceed Database Limits",
				ABV:       99.99, // theoretical maximum ABV
				IBU:       1000,  // theoretical maximum IBU
				SRM:       500.0, // theoretical maximum SRM
			},
		},
		{
			name: "beer with special characters",
			beer: Beer{
				Name:        "Beer with Special Characters: !@#$%^&*()_+-={}[]|\\:;\"'<>?,./",
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
			var unmarshaled Beer
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

// Test edge cases for Brewery model.
func TestBrewery_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		brewery Brewery
	}{
		{
			name: "brewery with empty values",
			brewery: Brewery{
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
			brewery: Brewery{
				Name:        "Very Long Brewery Name That Might Exceed Normal Database Field Limits",
				BreweryType: "Very Long Brewery Type",
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
			brewery: Brewery{
				Name:    "Bräuerei München",
				City:    "São Paulo",
				State:   "Zürich",
				Country: "España",
			},
		},
		{
			name: "brewery with special URLs and phone formats",
			brewery: Brewery{
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
			var unmarshaled Brewery
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

// Test StringArray with edge cases.
func TestStringArray_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected StringArray
		hasError bool
	}{
		{
			name:     "very large array",
			input:    `["item1", "item2", "item3", "item4", "item5"]`,
			expected: StringArray{"item1", "item2", "item3", "item4", "item5"},
		},
		{
			name:     "array with unicode characters",
			input:    `["🍺", "🍻", "bière", "çerveza", "пиво"]`,
			expected: StringArray{"🍺", "🍻", "bière", "çerveza", "пиво"},
		},
		{
			name:     "array with nested quotes",
			input:    `["He said \"hello\"", "It's a test", "Multiple \"quotes\" here"]`,
			expected: StringArray{`He said "hello"`, "It's a test", `Multiple "quotes" here`},
		},
		{
			name:     "array with control characters",
			input:    `["line1\nline2", "tab\there", "carriage\rreturn"]`,
			expected: StringArray{"line1\nline2", "tab\there", "carriage\rreturn"},
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
			var s StringArray
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
					t.Errorf("StringArray.Scan() = %v, want %v", s, tt.expected)
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
		var s StringArray
		err := s.Scan(jsonData)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkStringArray_Value(b *testing.B) {
	s := StringArray{"beer", "brewing", "hops", "malt", "yeast", "water", "fermentation"}

	b.ResetTimer()
	for range b.N {
		_, err := s.Value()
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkBeer_JSONMarshal(b *testing.B) {
	beer := Beer{
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
	beerType := reflect.TypeOf(Beer{})

	// Test Beer struct tags
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
			t.Errorf("Field %s not found in Beer struct", fieldName)
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag != expectedTag {
			t.Errorf("Field %s has db tag %q, want %q", fieldName, dbTag, expectedTag)
		}
	}

	// Test Brewery struct tags
	breweryType := reflect.TypeOf(Brewery{})
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
			t.Errorf("Field %s not found in Brewery struct", fieldName)
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag != expectedTag {
			t.Errorf("Field %s has db tag %q, want %q", fieldName, dbTag, expectedTag)
		}
	}
}
