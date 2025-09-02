package services

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// BeerServiceInterface abstracts beer search for handler injection and testing
type BeerServiceInterface interface {
	SearchBeers(ctx context.Context, query BeerSearchQuery) ([]*BeerSearchResult, error)
}

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
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Style   string  `json:"style"`
	Brewery string  `json:"brewery"`
	Country string  `json:"country"`
	ABV     float64 `json:"abv"`
	IBU     int     `json:"ibu"`
}

// BeerService handles beer-related operations.
type BeerService struct {
	db          *sqlx.DB
	redisClient *redis.Client // Optional caching
}

// NewBeerService creates a new BeerService instance.
func NewBeerService(db *sqlx.DB, redisClient *redis.Client) *BeerService {
	return &BeerService{
		db:          db,
		redisClient: redisClient,
	}
}

// SearchBeers performs a search for beers based on the provided criteria.
// Requires a *sqlx.DB to be available (add as a field to BeerService if needed)
func (s *BeerService) SearchBeers(ctx context.Context, query BeerSearchQuery) ([]*BeerSearchResult, error) {
	q := `
		  SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu
		  FROM beers b
		  JOIN breweries br ON b.brewery_id = br.id
		  WHERE 1=1
	  `
	args := []interface{}{}
	argIdx := 1

	if query.Name != "" {
		q += " AND b.name ILIKE $" + fmt.Sprint(argIdx)
		args = append(args, "%"+query.Name+"%")
		argIdx++
	}
	if query.Style != "" {
		q += " AND b.style ILIKE $" + fmt.Sprint(argIdx)
		args = append(args, "%"+query.Style+"%")
		argIdx++
	}
	if query.Brewery != "" {
		q += " AND br.name ILIKE $" + fmt.Sprint(argIdx)
		args = append(args, "%"+query.Brewery+"%")
		argIdx++
	}
	if query.Location != "" {
		q += " AND br.city ILIKE $" + fmt.Sprint(argIdx)
		args = append(args, "%"+query.Location+"%")
		argIdx++
	}
	if query.Limit > 0 {
		q += " LIMIT $" + fmt.Sprint(argIdx)
		args = append(args, query.Limit)
		argIdx++
	}

	rows, err := s.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*BeerSearchResult{}
	for rows.Next() {
		var r BeerSearchResult
		if err := rows.Scan(&r.ID, &r.Name, &r.Style, &r.Brewery, &r.Country, &r.ABV, &r.IBU); err != nil {
			return nil, err
		}
		results = append(results, &r)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return results, nil
}
