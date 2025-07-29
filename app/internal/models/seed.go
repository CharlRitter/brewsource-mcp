package models

import (
	"context"
	"fmt"

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
	breweries := getSeedBreweries()
	if berr := insertBreweries(ctx, db, breweries); berr != nil {
		return berr
	}
	logrus.Infof("Seeded %d breweries", len(breweries))
	return nil
}

func getSeedBreweries() []Brewery {
	return []Brewery{
		{
			Name:        "Devil's Peak Brewing Company",
			BreweryType: "micro",
			Street:      "1st Floor, The Old Warehouse, 6 Beach Road",
			City:        "Woodstock",
			State:       "Western Cape",
			PostalCode:  "7925",
			Country:     "South Africa",
			Phone:       "+27 21 200 5818",
			WebsiteURL:  "https://www.devilspeak.beer",
		},
		{
			Name:        "Jack Black's Brewing Company",
			BreweryType: "micro",
			Street:      "10 Brigid Road",
			City:        "Diep River",
			State:       "Western Cape",
			PostalCode:  "7800",
			Country:     "South Africa",
			Phone:       "+27 21 447 4151",
			WebsiteURL:  "https://www.jackblackbeer.com",
		},
		{
			Name:        "Drifter Brewing Company",
			BreweryType: "micro",
			Street:      "156 Victoria Road",
			City:        "Woodstock",
			State:       "Western Cape",
			PostalCode:  "7925",
			Country:     "South Africa",
			Phone:       "+27 21 447 0835",
			WebsiteURL:  "https://www.drifterbrewing.co.za",
		},
		{
			Name:        "Stellenbosch Brewing Company",
			BreweryType: "micro",
			Street:      "Klein Joostenberg, R304",
			City:        "Stellenbosch",
			State:       "Western Cape",
			PostalCode:  "7600",
			Country:     "South Africa",
			Phone:       "+27 21 884 4014",
			WebsiteURL:  "https://www.stellenboschbrewing.co.za",
		},
		{
			Name:        "Woodstock Brewery",
			BreweryType: "micro",
			Street:      "252 Albert Road",
			City:        "Woodstock",
			State:       "Western Cape",
			PostalCode:  "7925",
			Country:     "South Africa",
			Phone:       "+27 21 447 0953",
			WebsiteURL:  "https://www.woodstockbrewery.co.za",
		},
		{
			Name:        "Darling Brew",
			BreweryType: "micro",
			Street:      "48 Caledon Street",
			City:        "Darling",
			State:       "Western Cape",
			PostalCode:  "7345",
			Country:     "South Africa",
			Phone:       "+27 21 286 1099",
			WebsiteURL:  "https://www.darlingbrew.co.za",
		},
		{
			Name:        "Cape Brewing Company (CBC)",
			BreweryType: "micro",
			Street:      "R44, Klapmuts",
			City:        "Paarl",
			State:       "Western Cape",
			PostalCode:  "7625",
			Country:     "South Africa",
			Phone:       "+27 21 863 2270",
			WebsiteURL:  "https://www.capebrewing.co.za",
		},
		{
			Name:        "Signal Gun Wines & Brewery",
			BreweryType: "micro",
			Street:      "Meerendal Wine Estate, Vissershok Road",
			City:        "Durbanville",
			State:       "Western Cape",
			PostalCode:  "7550",
			Country:     "South Africa",
			Phone:       "+27 21 558 6972",
			WebsiteURL:  "https://www.signalgun.com",
		},
	}
}

func insertBreweries(ctx context.Context, db *sqlx.DB, breweries []Brewery) error {
	for _, brewery := range breweries {
		query := `
			INSERT INTO breweries (
				name, brewery_type, street, city, state, postal_code, country, phone, website_url
			) VALUES (
				:name, :brewery_type, :street, :city, :state, :postal_code, :country, :phone, :website_url
			)
		`
		if _, berr := db.NamedExecContext(ctx, query, brewery); berr != nil {
			return fmt.Errorf("failed to insert brewery %s: %w", brewery.Name, berr)
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
	breweries, berr := getBreweryIDs(ctx, db)
	if berr != nil {
		return berr
	}
	beers := getSeedBeers()
	if berr := insertBeers(ctx, db, breweries, beers); berr != nil {
		return berr
	}
	logrus.Infof("Seeded %d beers", len(beers))
	return nil
}

func getBreweryIDs(ctx context.Context, db *sqlx.DB) (map[string]int, error) {
	breweries := map[string]int{}
	rows, err := db.QueryxContext(ctx, "SELECT id, name FROM breweries")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

type seedBeer struct {
	Name        string
	BreweryName string
	Style       string
	ABV         float64
	IBU         int
	SRM         float64
	Description string
}

func getSeedBeers() []seedBeer {
	return []seedBeer{
		{
			Name:        "King's Blockhouse IPA",
			BreweryName: "Devil's Peak Brewing Company",
			Style:       "American IPA",
			ABV:         6.0,
			IBU:         52,
			SRM:         10.0,
			Description: "A bold, hop-forward IPA with citrus and pine notes, regarded as one of South Africa's best IPAs.",
		},
		{
			Name:        "Four Lager",
			BreweryName: "Jack Black's Brewing Company",
			Style:       "Lager",
			ABV:         4.0,
			IBU:         18,
			SRM:         4.0,
			Description: "A crisp, easy-drinking lager with a light malt backbone and subtle hop aroma.",
		},
		{
			Name:        "The Stranded Coconut",
			BreweryName: "Drifter Brewing Company",
			Style:       "Coconut Ale",
			ABV:         4.5,
			IBU:         20,
			SRM:         6.0,
			Description: "A unique ale brewed with real coconut, offering a tropical aroma and smooth finish.",
		},
		{
			Name:        "Hoenderhok Bock",
			BreweryName: "Stellenbosch Brewing Company",
			Style:       "Bock",
			ABV:         6.5,
			IBU:         24,
			SRM:         20.0,
			Description: "A malty, rich bock with caramel and toffee notes, inspired by German brewing traditions.",
		},
		{
			Name:        "Rhythm Stick English Pale Ale",
			BreweryName: "Woodstock Brewery",
			Style:       "English Pale Ale",
			ABV:         4.5,
			IBU:         30,
			SRM:         8.0,
			Description: "A balanced pale ale with biscuit malt character and earthy English hops.",
		},
		{
			Name:        "Bone Crusher Witbier",
			BreweryName: "Darling Brew",
			Style:       "Witbier",
			ABV:         6.0,
			IBU:         15,
			SRM:         4.0,
			Description: "A refreshing Belgian-style witbier brewed with coriander and orange peel.",
		},
		{
			Name:        "Amber Weiss",
			BreweryName: "Cape Brewing Company (CBC)",
			Style:       "Weissbier",
			ABV:         5.0,
			IBU:         14,
			SRM:         7.0,
			Description: "A classic German-style wheat beer with banana and clove aromas and a smooth mouthfeel.",
		},
		{
			Name:        "Gun Powder IPA",
			BreweryName: "Signal Gun Wines & Brewery",
			Style:       "IPA",
			ABV:         5.5,
			IBU:         45,
			SRM:         9.0,
			Description: "A hop-forward IPA with citrus and tropical fruit notes, brewed on a historic Durbanville farm.",
		},
	}
}

func insertBeers(ctx context.Context, db *sqlx.DB, breweries map[string]int, beers []seedBeer) error {
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
		if _, berr := db.ExecContext(ctx, query, breweryID, beer.Name, beer.Style, beer.ABV, beer.IBU, beer.SRM, beer.Description); berr != nil {
			return fmt.Errorf("failed to insert beer %s: %w", beer.Name, berr)
		}
	}
	return nil
}
