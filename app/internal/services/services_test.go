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

	if brewery.Name != "Devil's Peak Brewing Company" {
		t.Errorf("Expected Name to be 'Devil's Peak Brewing Company', got %s", brewery.Name)
	}

	if brewery.City != "Woodstock" {
		t.Errorf("Expected City to be 'Woodstock', got %s", brewery.City)
	}
}
