package services

import (
	"context"
	"reflect"
	"testing"
)

func TestNewBeerService(t *testing.T) {
	svc := NewBeerService()
	if svc == nil {
		t.Error("expected non-nil BeerService")
	}
}

func TestSearchBeers_ReturnsSampleData(t *testing.T) {
	svc := NewBeerService()
	query := BeerSearchQuery{
		Name:     "Sample",
		Style:    "IPA",
		Brewery:  "Sample Brewery",
		Location: "USA",
		Limit:    10,
	}
	results, err := svc.SearchBeers(context.Background(), query)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	want := &BeerSearchResult{ID: 1, Name: "Sample IPA", Style: "IPA", Brewery: "Sample Brewery", Country: "USA"}
	if !reflect.DeepEqual(results[0], want) {
		t.Errorf("unexpected first result: got %+v, want %+v", results[0], want)
	}
}

func TestBeerSearchQuery_Fields(t *testing.T) {
	q := BeerSearchQuery{
		Name:     "Test Beer",
		Style:    "Lager",
		Brewery:  "Test Brewery",
		Location: "Germany",
		Limit:    5,
	}
	if q.Name != "Test Beer" || q.Style != "Lager" || q.Brewery != "Test Brewery" || q.Location != "Germany" || q.Limit != 5 {
		t.Errorf("unexpected BeerSearchQuery fields: %+v", q)
	}
}

func TestBeerSearchResult_Fields(t *testing.T) {
	r := &BeerSearchResult{
		ID:      42,
		Name:    "Test Stout",
		Style:   "Stout",
		Brewery: "Test Brewery",
		Country: "Ireland",
	}
	if r.ID != 42 || r.Name != "Test Stout" || r.Style != "Stout" || r.Brewery != "Test Brewery" || r.Country != "Ireland" {
		t.Errorf("unexpected BeerSearchResult fields: %+v", r)
	}
}
