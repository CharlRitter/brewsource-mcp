// Package services provides business logic and service layer functions for Brewsource MCP, including beer and brewery operations.
package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// BreweryServiceInterface abstracts brewery search for handler injection and testing.
type BreweryServiceInterface interface {
	SearchBreweries(ctx context.Context, query BrewerySearchQuery) ([]*BrewerySearchResult, error)
}

// BrewerySearchQuery represents search parameters for brewery lookup.
type BrewerySearchQuery struct {
	Name     string
	Location string
	City     string
	State    string
	Country  string
	Limit    int
}

// BrewerySearchResult represents a brewery search result.
type BrewerySearchResult struct {
	ID          int    `db:"id"           json:"id"`
	Name        string `db:"name"         json:"name"`
	BreweryType string `db:"brewery_type" json:"brewery_type"`
	Street      string `db:"street"       json:"street"`
	City        string `db:"city"         json:"city"`
	State       string `db:"state"        json:"state"`
	PostalCode  string `db:"postal_code"  json:"postal_code"`
	Country     string `db:"country"      json:"country"`
	Phone       string `db:"phone"        json:"phone"`
	Website     string `db:"website_url"  json:"website_url"`
}

// BreweryService handles brewery-related operations.
type BreweryService struct {
	db          *sqlx.DB
	redisClient *redis.Client // Optional caching
}

// NewBreweryService creates a new BreweryService instance.
func NewBreweryService(db *sqlx.DB, redisClient *redis.Client) *BreweryService {
	return &BreweryService{
		db:          db,
		redisClient: redisClient,
	}
}

// SearchBreweries performs a search for breweries based on the provided criteria.
func (s *BreweryService) SearchBreweries(
	ctx context.Context,
	query BrewerySearchQuery,
) ([]*BrewerySearchResult, error) {
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 20
	}

	var conditions []string
	var args []interface{}
	argCount := 0

	baseQuery := `
		SELECT id, name, brewery_type, street, city, state, postal_code, country, phone, website_url
		FROM breweries
		WHERE 1=1`

	if query.Name != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("LOWER(name) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+query.Name+"%")
	}

	if query.City != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("LOWER(city) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+query.City+"%")
	}

	if query.State != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("LOWER(state) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+query.State+"%")
	}

	if query.Country != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("LOWER(country) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+query.Country+"%")
	}

	if query.Location != "" {
		argCount++
		conditions = append(
			conditions,
			fmt.Sprintf(
				"(LOWER(city) LIKE LOWER($%d) OR LOWER(state) LIKE LOWER($%d) OR LOWER(country) LIKE LOWER($%d))",
				argCount,
				argCount,
				argCount,
			),
		)
		args = append(args, "%"+query.Location+"%")
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY name"
	argCount++
	baseQuery += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, query.Limit)

	var results []*BrewerySearchResult
	err := s.db.SelectContext(ctx, &results, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search breweries: %w", err)
	}

	return results, nil
}
