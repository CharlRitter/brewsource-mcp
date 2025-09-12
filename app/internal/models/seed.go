// Package models defines the data models and database schema for Brewsource MCP.
package models

import (
	"context"
	"fmt"

	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// SeedDatabase populates the database with initial data for Phase 1.
// It seeds both breweries and beers tables with sample data if they are empty.
// Returns an error if any seeding step fails.
func SeedDatabase(db *sqlx.DB) error {
	ctx := context.Background()

	logrus.Info("Starting database seeding...")

	// Seed breweries
	if err := seedBreweries(ctx, db); err != nil {
		return fmt.Errorf("failed to seed breweries: %w", err)
	}

	// Seed beers
	if err := seedBeers(ctx, db); err != nil {
		return fmt.Errorf("failed to seed beers: %w", err)
	}

	logrus.Info("Database seeding completed successfully")
	return nil
}

// seedBreweries inserts a set of sample breweries into the database if none exist.
// It checks for existing data to ensure idempotency.
// Returns an error if the operation fails.
func seedBreweries(ctx context.Context, db *sqlx.DB) error {
	// Check if breweries already exist
	var count int
	err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM breweries")
	if err != nil {
		return err
	}
	if count > 0 {
		logrus.Info("Breweries already exist, skipping seeding")
		return nil
	}
	logrus.Info("Seeding breweries...")
	breweries := services.GetSeedBreweries()
	if insertErr := insertBreweries(ctx, db, breweries); insertErr != nil {
		return insertErr
	}
	logrus.Infof("Seeded %d breweries", len(breweries))
	return nil
}

func insertBreweries(ctx context.Context, db *sqlx.DB, breweries []services.Brewery) error {
	for _, brewery := range breweries {
		query := `
			INSERT INTO breweries (
				name, brewery_type, street, city, state, postal_code, country, phone, website_url
			) VALUES (
				:name, :brewery_type, :street, :city, :state, :postal_code, :country, :phone, :website_url
			)
		`
		if _, insertErr := db.NamedExecContext(ctx, query, brewery); insertErr != nil {
			return fmt.Errorf("failed to insert brewery %s: %w", brewery.Name, insertErr)
		}
	}
	return nil
}

// seedBeers inserts a set of sample beers into the database if none exist.
// It looks up brewery IDs to associate beers with breweries.
// Returns an error if the operation fails.
func seedBeers(ctx context.Context, db *sqlx.DB) error {
	// Check if beers already exist
	var count int
	err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM beers")
	if err != nil {
		return err
	}
	if count > 0 {
		logrus.Info("Beers already exist, skipping seeding")
		return nil
	}
	logrus.Info("Seeding beers...")
	breweries, breweryErr := GetBreweryIDs(ctx, db)
	if breweryErr != nil {
		return breweryErr
	}
	beers := services.GetSeedBeers()
	if insertErr := insertBeers(ctx, db, breweries, beers); insertErr != nil {
		return insertErr
	}
	logrus.Infof("Seeded %d beers", len(beers))
	return nil
}

func GetBreweryIDs(ctx context.Context, db *sqlx.DB) (map[string]int, error) {
	breweries := map[string]int{}
	rows, err := db.QueryxContext(ctx, "SELECT id, name FROM breweries")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logrus.Warnf("Failed to close rows: %v", closeErr)
		}
	}()

	for rows.Next() {
		var id int
		var name string
		if scanErr := rows.Scan(&id, &name); scanErr != nil {
			return nil, scanErr
		}
		breweries[name] = id
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return breweries, nil
}

// SeedBreweries exports the internal seedBreweries function for testing.
func SeedBreweries(ctx context.Context, db *sqlx.DB) error {
	return seedBreweries(ctx, db)
}

// SeedBeers exports the internal seedBeers function for testing.
func SeedBeers(ctx context.Context, db *sqlx.DB) error {
	return seedBeers(ctx, db)
}

func insertBeers(ctx context.Context, db *sqlx.DB, breweries map[string]int, beers []services.SeedBeer) error {
	for _, beer := range beers {
		breweryID, exists := breweries[beer.BreweryName]
		if !exists {
			logrus.Warnf("Brewery not found: %s, skipping beer: %s", beer.BreweryName, beer.Name)
			continue
		}
		query := `
			INSERT INTO beers (
				brewery_id, name, style, abv, ibu, srm, description
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7
			)
		`
		if _, insertErr := db.ExecContext(ctx, query, breweryID, beer.Name, beer.Style, beer.ABV, beer.IBU, beer.SRM, beer.Description); insertErr != nil {
			return fmt.Errorf("failed to insert beer %s: %w", beer.Name, insertErr)
		}
	}
	return nil
}
