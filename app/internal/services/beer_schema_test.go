package services_test

import (
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
)

func TestGetSeedBeers(t *testing.T) {
	beers := services.GetSeedBeers()
	checkNonEmptySlice(t, beers)
	checkMinBeers(t, beers, 50)
	checkRequiredFields(t, beers)
	checkExpectedBeersExist(t, beers, []string{"Castle Lager", "King's Blockhouse IPA", "Brewers Lager"})
	checkABVRange(t, beers, 3.0, 15.0)
	checkIBURange(t, beers, 0, 120)
	checkSRMRange(t, beers, 1.0, 50.0)
}

func checkNonEmptySlice(t *testing.T, beers []services.SeedBeer) {
	if len(beers) == 0 {
		t.Fatal("GetSeedBeers() returned empty slice")
	}
}

func checkMinBeers(t *testing.T, beers []services.SeedBeer, minCount int) {
	if len(beers) < minCount {
		t.Errorf("GetSeedBeers() returned %d beers, expected at least %d", len(beers), minCount)
	}
}

func checkRequiredFields(t *testing.T, beers []services.SeedBeer) {
	for i, beer := range beers {
		if beer.Name == "" {
			t.Errorf("Beer at index %d has empty Name", i)
		}
		if beer.BreweryName == "" {
			t.Errorf("Beer at index %d has empty BreweryName", i)
		}
		if beer.Style == "" {
			t.Errorf("Beer at index %d has empty Style", i)
		}
		if beer.ABV <= 0 {
			t.Errorf("Beer at index %d has invalid ABV: %f", i, beer.ABV)
		}
		if beer.IBU < 0 {
			t.Errorf("Beer at index %d has negative IBU: %d", i, beer.IBU)
		}
		if beer.SRM <= 0 {
			t.Errorf("Beer at index %d has invalid SRM: %f", i, beer.SRM)
		}
		if beer.Description == "" {
			t.Errorf("Beer at index %d has empty Description", i)
		}
	}
}

func checkExpectedBeersExist(t *testing.T, beers []services.SeedBeer, expected []string) {
	found := make(map[string]bool)
	for _, name := range expected {
		found[name] = false
	}
	for _, beer := range beers {
		if _, exists := found[beer.Name]; exists {
			found[beer.Name] = true
		}
	}
	for beerName, ok := range found {
		if !ok {
			t.Errorf("Expected beer %s not found in seed data", beerName)
		}
	}
}

func checkABVRange(t *testing.T, beers []services.SeedBeer, minABV, maxABV float64) {
	for i, beer := range beers {
		if beer.ABV < minABV || beer.ABV > maxABV {
			t.Errorf("Beer at index %d has unreasonable ABV: %f", i, beer.ABV)
		}
	}
}

func checkIBURange(t *testing.T, beers []services.SeedBeer, minIBU, maxIBU int) {
	for i, beer := range beers {
		if beer.IBU < minIBU || beer.IBU > maxIBU {
			t.Errorf("Beer at index %d has unreasonable IBU: %d", i, beer.IBU)
		}
	}
}

func checkSRMRange(t *testing.T, beers []services.SeedBeer, minSRM, maxSRM float64) {
	for i, beer := range beers {
		if beer.SRM < minSRM || beer.SRM > maxSRM {
			t.Errorf("Beer at index %d has unreasonable SRM: %f", i, beer.SRM)
		}
	}
}

func TestSeedBeerStructure(t *testing.T) {
	// Test that SeedBeer struct has the expected fields
	beer := services.SeedBeer{
		Name:        "Test Beer",
		BreweryName: "Test Brewery",
		Style:       "Test Style",
		ABV:         5.0,
		IBU:         30,
		SRM:         10.0,
		Description: "Test Description",
	}

	if beer.Name != "Test Beer" {
		t.Errorf("Expected Name to be 'Test Beer', got %s", beer.Name)
	}
	if beer.BreweryName != "Test Brewery" {
		t.Errorf("Expected BreweryName to be 'Test Brewery', got %s", beer.BreweryName)
	}
	if beer.Style != "Test Style" {
		t.Errorf("Expected Style to be 'Test Style', got %s", beer.Style)
	}
	if beer.ABV != 5.0 {
		t.Errorf("Expected ABV to be 5.0, got %f", beer.ABV)
	}
	if beer.IBU != 30 {
		t.Errorf("Expected IBU to be 30, got %d", beer.IBU)
	}
	if beer.SRM != 10.0 {
		t.Errorf("Expected SRM to be 10.0, got %f", beer.SRM)
	}
	if beer.Description != "Test Description" {
		t.Errorf("Expected Description to be 'Test Description', got %s", beer.Description)
	}
}
