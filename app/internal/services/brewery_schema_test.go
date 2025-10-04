package services_test

import (
	"testing"
	"time"

	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
)

func TestGetSeedBreweries(t *testing.T) {
	breweries := services.GetSeedBreweries()
	checkNonEmptyBreweries(t, breweries)
	checkMinBreweries(t, breweries, 20)
	checkRequiredBreweryFields(t, breweries)
	checkExpectedBreweriesExist(t, breweries, []string{
		"SAB - Newlands Brewery",
		"Jack Black's Brewing Company",
		"Devil's Peak Brewing Company",
		"Cape Brewing Company (CBC)",
		"Darling Brew",
	})
	checkValidBreweryTypes(t, breweries)
	checkProvinceCoverage(t, breweries, []string{"Western Cape", "Gauteng", "KwaZulu-Natal"})
}

func checkNonEmptyBreweries(t *testing.T, breweries []services.Brewery) {
	if len(breweries) == 0 {
		t.Fatal("GetSeedBreweries() returned empty slice")
	}
}

func checkMinBreweries(t *testing.T, breweries []services.Brewery, minCount int) {
	if len(breweries) < minCount {
		t.Errorf("GetSeedBreweries() returned %d breweries, expected at least %d", len(breweries), minCount)
	}
}

func checkRequiredBreweryFields(t *testing.T, breweries []services.Brewery) {
	for i, brewery := range breweries {
		if brewery.Name == "" {
			t.Errorf("Brewery at index %d has empty Name", i)
		}
		if brewery.BreweryType == "" {
			t.Errorf("Brewery at index %d has empty BreweryType", i)
		}
		if brewery.City == "" {
			t.Errorf("Brewery at index %d has empty City", i)
		}
		if brewery.State == "" {
			t.Errorf("Brewery at index %d has empty State", i)
		}
		if brewery.Country == "" {
			t.Errorf("Brewery at index %d has empty Country", i)
		}
		if brewery.Country != "South Africa" {
			t.Errorf("Brewery at index %d has unexpected Country: %s", i, brewery.Country)
		}
	}
}

func checkExpectedBreweriesExist(t *testing.T, breweries []services.Brewery, expected []string) {
	found := make(map[string]bool)
	for _, name := range expected {
		found[name] = false
	}
	for _, brewery := range breweries {
		if _, exists := found[brewery.Name]; exists {
			found[brewery.Name] = true
		}
	}
	for breweryName, ok := range found {
		if !ok {
			t.Errorf("Expected brewery %s not found in seed data", breweryName)
		}
	}
}

func checkValidBreweryTypes(t *testing.T, breweries []services.Brewery) {
	validTypes := map[string]bool{
		"micro":    false,
		"macro":    false,
		"brewpub":  false,
		"regional": false,
		"contract": false,
		"planning": false,
		"closed":   false,
		"large":    false,
		"nano":     false,
	}
	for _, brewery := range breweries {
		if _, valid := validTypes[brewery.BreweryType]; !valid {
			t.Errorf("Brewery %s has invalid BreweryType: %s", brewery.Name, brewery.BreweryType)
		} else {
			validTypes[brewery.BreweryType] = true
		}
	}
	typesFound := 0
	for _, found := range validTypes {
		if found {
			typesFound++
		}
	}
	if typesFound < 2 {
		t.Errorf("Expected at least 2 different brewery types, found %d", typesFound)
	}
}

func checkProvinceCoverage(t *testing.T, breweries []services.Brewery, expectedProvinces []string) {
	found := make(map[string]bool)
	for _, province := range expectedProvinces {
		found[province] = false
	}
	for _, brewery := range breweries {
		if _, exists := found[brewery.State]; exists {
			found[brewery.State] = true
		}
	}
	for province, ok := range found {
		if !ok {
			t.Errorf("Expected province %s not found in seed data", province)
		}
	}
}

func TestBreweryStructure(t *testing.T) {
	// Test that Brewery struct has the expected fields
	now := time.Now()
	brewery := services.Brewery{
		ID:          1,
		Name:        "Test Brewery",
		BreweryType: "micro",
		Street:      "123 Test Street",
		City:        "Test City",
		State:       "Test State",
		PostalCode:  "12345",
		Country:     "Test Country",
		Phone:       "+1234567890",
		WebsiteURL:  "https://test.com",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if brewery.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", brewery.ID)
	}
	if brewery.Name != "Test Brewery" {
		t.Errorf("Expected Name to be 'Test Brewery', got %s", brewery.Name)
	}
	if brewery.BreweryType != "micro" {
		t.Errorf("Expected BreweryType to be 'micro', got %s", brewery.BreweryType)
	}
	if brewery.Street != "123 Test Street" {
		t.Errorf("Expected Street to be '123 Test Street', got %s", brewery.Street)
	}
	if brewery.City != "Test City" {
		t.Errorf("Expected City to be 'Test City', got %s", brewery.City)
	}
	if brewery.State != "Test State" {
		t.Errorf("Expected State to be 'Test State', got %s", brewery.State)
	}
	if brewery.PostalCode != "12345" {
		t.Errorf("Expected PostalCode to be '12345', got %s", brewery.PostalCode)
	}
	if brewery.Country != "Test Country" {
		t.Errorf("Expected Country to be 'Test Country', got %s", brewery.Country)
	}
	if brewery.Phone != "+1234567890" {
		t.Errorf("Expected Phone to be '+1234567890', got %s", brewery.Phone)
	}
	if brewery.WebsiteURL != "https://test.com" {
		t.Errorf("Expected WebsiteURL to be 'https://test.com', got %s", brewery.WebsiteURL)
	}
	if brewery.CreatedAt != now {
		t.Errorf("Expected CreatedAt to be %v, got %v", now, brewery.CreatedAt)
	}
	if brewery.UpdatedAt != now {
		t.Errorf("Expected UpdatedAt to be %v, got %v", now, brewery.UpdatedAt)
	}
}

func TestBreweryTypesConsistency(t *testing.T) {
	breweries := services.GetSeedBreweries()

	// Count brewery types
	typeCounts := make(map[string]int)
	for _, brewery := range breweries {
		typeCounts[brewery.BreweryType]++
	}

	// Should have macro breweries (SAB)
	if typeCounts["macro"] == 0 {
		t.Error("Expected at least one macro brewery in seed data")
	}

	// Should have micro breweries
	if typeCounts["micro"] == 0 {
		t.Error("Expected at least one micro brewery in seed data")
	}

	// Should have brewpubs
	if typeCounts["brewpub"] == 0 {
		t.Error("Expected at least one brewpub in seed data")
	}
}

func TestBreweryGeographicDistribution(t *testing.T) {
	breweries := services.GetSeedBreweries()

	// Count by province
	provinceCounts := make(map[string]int)
	for _, brewery := range breweries {
		provinceCounts[brewery.State]++
	}

	// Western Cape should have the most craft breweries
	if provinceCounts["Western Cape"] == 0 {
		t.Error("Expected Western Cape breweries in seed data")
	}

	// Should have representation from other provinces
	if provinceCounts["Gauteng"] == 0 {
		t.Error("Expected Gauteng breweries in seed data")
	}

	// Check for reasonable geographic distribution
	totalBreweries := len(breweries)
	wcPercentage := float64(provinceCounts["Western Cape"]) / float64(totalBreweries)
	if wcPercentage < 0.3 { // Western Cape should have at least 30% of breweries
		t.Errorf("Expected Western Cape to have at least 30%% of breweries, got %.1f%%", wcPercentage*100)
	}
}
