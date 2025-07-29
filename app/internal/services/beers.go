package services

import (
	"context"
)

// BeerSearchQuery represents search parameters for beer lookup.
type BeerSearchQuery struct {
	Name     string
	Style    string
	Brewery  string
	Location string
	Limit    int
}

// BeerSearchResult represents a beer search result.
type BeerSearchResult struct {
	ID      int
	Name    string
	Style   string
	Brewery string
	Country string
}

// BeerService handles beer-related operations.
type BeerService struct{}

// NewBeerService creates a new BeerService instance.
func NewBeerService() *BeerService {
	return &BeerService{}
}

// SearchBeers performs a search for beers based on the provided criteria.
func (s *BeerService) SearchBeers(ctx context.Context, query BeerSearchQuery) ([]*BeerSearchResult, error) {
	// Minimal stub: returns a static sample
	return []*BeerSearchResult{
		{ID: 1, Name: "Sample IPA", Style: "IPA", Brewery: "Sample Brewery", Country: "USA"},
		{ID: 2, Name: "Sample Stout", Style: "Stout", Brewery: "Sample Brewery", Country: "USA"},
	}, nil
}
