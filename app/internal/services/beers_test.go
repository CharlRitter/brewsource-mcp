package services

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func TestNewBeerService(t *testing.T) {
	db := &sqlx.DB{}               // Use a mock or test DB connection in real tests
	redisClient := &redis.Client{} // Use a mock or test Redis client
	svc := NewBeerService(db, redisClient)
	if svc == nil {
		t.Error("expected non-nil BeerService")
	}
}

// Integration test: expects at least one seeded beer if test DB is available
func TestSearchBeers_SeededData(t *testing.T) {
	// This test assumes a test DB with seeded data is available
	db, err := sqlx.Open("postgres", "user=brewsource_user dbname=brewsource sslmode=disable")
	if err != nil {
		t.Skip("skipping: could not connect to test database")
	}
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	svc := NewBeerService(db, redisClient)
	query := BeerSearchQuery{Limit: 5}
	results, err := svc.SearchBeers(context.Background(), query)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) == 0 {
		t.Errorf("expected at least one seeded beer, got 0")
	}
}

func TestBeerSearchQuery_Fields(t *testing.T) {
	q := BeerSearchQuery{
		Name:     "King's Blockhouse IPA",
		Style:    "American IPA",
		Brewery:  "Devil's Peak Brewing Company",
		Location: "South Africa",
		Limit:    5,
	}
	if q.Name != "King's Blockhouse IPA" || q.Style != "American IPA" || q.Brewery != "Devil's Peak Brewing Company" || q.Location != "South Africa" || q.Limit != 5 {
		t.Errorf("unexpected BeerSearchQuery fields: %+v", q)
	}
}

func TestBeerSearchResult_Fields(t *testing.T) {
	r := &BeerSearchResult{
		ID:      1,
		Name:    "King's Blockhouse IPA",
		Style:   "American IPA",
		Brewery: "Devil's Peak Brewing Company",
		Country: "South Africa",
	}
	if r.Name != "King's Blockhouse IPA" {
		t.Errorf("Expected Name to be 'King's Blockhouse IPA', got %s", r.Name)
	}
	if r.Style != "American IPA" {
		t.Errorf("Expected Style to be 'American IPA', got %s", r.Style)
	}
	if r.Brewery != "Devil's Peak Brewing Company" {
		t.Errorf("Expected Brewery to be 'Devil's Peak Brewing Company', got %s", r.Brewery)
	}
	if r.Country != "South Africa" {
		t.Errorf("Expected Country to be 'South Africa', got %s", r.Country)
	}
}
