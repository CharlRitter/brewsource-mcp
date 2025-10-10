// Package models defines the data models and database schema for Brewsource MCP.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

// ErrStringArrayEmpty represents an error when StringArray is empty.
var ErrStringArrayEmpty = errors.New("string array is empty")

// StringArray handles PostgreSQL string arrays in Go.
type StringArray []string

// Scan implements the Scanner interface for StringArray.
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("cannot scan non-string into StringArray")
	}
}

// Value implements the Valuer interface for StringArray.
func (s *StringArray) Value() (driver.Value, error) {
	if s == nil || len(*s) == 0 {
		return "[]", nil
	}
	return json.Marshal(*s)
}

// Beer represents a commercial beer.
type Beer struct {
	ID          int       `json:"id"          db:"id"`
	Name        string    `json:"name"        db:"name"`
	BreweryID   int       `json:"brewery_id"  db:"brewery_id"`
	Style       string    `json:"style"       db:"style"`
	ABV         float64   `json:"abv"         db:"abv"`
	IBU         int       `json:"ibu"         db:"ibu"`
	SRM         float64   `json:"srm"         db:"srm"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"  db:"updated_at"`
}

// Brewery represents a brewery.
type Brewery struct {
	ID          int       `json:"id"           db:"id"`
	Name        string    `json:"name"         db:"name"`
	BreweryType string    `json:"brewery_type" db:"brewery_type"`
	Street      string    `json:"street"       db:"street"`
	City        string    `json:"city"         db:"city"`
	State       string    `json:"state"        db:"state"`
	PostalCode  string    `json:"postal_code"  db:"postal_code"`
	Country     string    `json:"country"      db:"country"`
	Phone       string    `json:"phone"        db:"phone"`
	WebsiteURL  string    `json:"website_url"  db:"website_url"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"   db:"updated_at"`
}

// MigrateDatabase creates the necessary database tables and indexes for the brewsource application.
func MigrateDatabase(db *sqlx.DB) error {
	queries := []string{
		// Breweries table
		`CREATE TABLE IF NOT EXISTS breweries (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			brewery_type VARCHAR(50),
			street VARCHAR(255),
			city VARCHAR(255),
			state VARCHAR(255),
			postal_code VARCHAR(20),
			country VARCHAR(255) DEFAULT 'United States',
			phone VARCHAR(50),
			website_url VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Beers table
		`CREATE TABLE IF NOT EXISTS beers (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			brewery_id INTEGER REFERENCES breweries(id),
			style VARCHAR(255),
			abv DECIMAL(4,2),
			ibu INTEGER,
			srm DECIMAL(4,1),
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE INDEX IF NOT EXISTS idx_breweries_name ON breweries USING gin(to_tsvector('english', name))`,
		`CREATE INDEX IF NOT EXISTS idx_breweries_location ON breweries(city, state, country)`,

		`CREATE INDEX IF NOT EXISTS idx_beers_name ON beers USING gin(to_tsvector('english', name))`,
		`CREATE INDEX IF NOT EXISTS idx_beers_brewery ON beers(brewery_id)`,
		`CREATE INDEX IF NOT EXISTS idx_beers_style ON beers(style)`,

		// Update triggers for updated_at timestamps
		`CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql'`,

		`DROP TRIGGER IF EXISTS update_breweries_updated_at ON breweries`,
		`CREATE TRIGGER update_breweries_updated_at
			BEFORE UPDATE ON breweries
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,

		`DROP TRIGGER IF EXISTS update_beers_updated_at ON beers`,
		`CREATE TRIGGER update_beers_updated_at
			BEFORE UPDATE ON beers
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
