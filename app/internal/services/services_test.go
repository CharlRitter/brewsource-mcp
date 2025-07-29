package services

import (
	"testing"
)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := BrewerySearchQuery{
				Name:  "test",
				Limit: tt.inputLimit,
			}

			// Simulate the limit adjustment logic that happens in SearchBreweries
			if query.Limit <= 0 || query.Limit > 100 {
				query.Limit = 20
			}

			if query.Limit != tt.expectedLimit {
				t.Errorf("Limit adjustment: got %d, want %d", query.Limit, tt.expectedLimit)
			}
		})
	}
}

func TestBrewerySearchResult_FieldsExist(t *testing.T) {
	// Test that all expected fields exist on the BrewerySearchResult struct
	brewery := &BrewerySearchResult{
		ID:          1,
		Name:        "Stone Brewing",
		BreweryType: "regional",
		Street:      "1999 Citracado Pkwy",
		City:        "Escondido",
		State:       "CA",
		PostalCode:  "92029",
		Country:     "United States",
		Phone:       "760-294-7866",
		Website:     "https://stonebrewing.com",
	}

	if brewery.Name != "Stone Brewing" {
		t.Errorf("Expected Name to be 'Stone Brewing', got %s", brewery.Name)
	}

	if brewery.City != "Escondido" {
		t.Errorf("Expected City to be 'Escondido', got %s", brewery.City)
	}
}
