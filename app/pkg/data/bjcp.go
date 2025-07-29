package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BJCPStyle represents a beer style from the BJCP guidelines.
type BJCPStyle struct {
	Code                      string   `json:"code"`
	Name                      string   `json:"name"`
	Category                  string   `json:"category"`
	OverallImpression         string   `json:"overall_impression"`
	Appearance                string   `json:"appearance"`
	Aroma                     string   `json:"aroma"`
	Flavor                    string   `json:"flavor"`
	Mouthfeel                 string   `json:"mouthfeel"`
	Comments                  string   `json:"comments"`
	History                   string   `json:"history"`
	CharacteristicIngredients string   `json:"characteristic_ingredients"`
	StyleComparison           string   `json:"style_comparison"`
	CommercialExamples        []string `json:"commercial_examples"`
	Vitals                    Vitals   `json:"vitals"`
}

// Vitals represents the technical specifications of a beer style.
type Vitals struct {
	ABVMin float64 `json:"abv_min"`
	ABVMax float64 `json:"abv_max"`
	IBUMin int     `json:"ibu_min"`
	IBUMax int     `json:"ibu_max"`
	SRMMin float64 `json:"srm_min"`
	SRMMax float64 `json:"srm_max"`
	OGMin  float64 `json:"og_min"`
	OGMax  float64 `json:"og_max"`
	FGMin  float64 `json:"fg_min"`
	FGMax  float64 `json:"fg_max"`
}

// BJCPData represents the complete BJCP style guide data.
type BJCPData struct {
	Styles     map[string]BJCPStyle `json:"styles"`
	Categories []string             `json:"categories"`
	Metadata   Metadata             `json:"metadata"`
}

// Metadata contains information about the BJCP data version and source.
type Metadata struct {
	Version     string `json:"version"`
	Source      string `json:"source"`
	LastUpdated string `json:"last_updated"`
	TotalStyles int    `json:"total_styles"`
}

// LoadBJCPData loads and parses the BJCP style data from file.
func LoadBJCPData() (*BJCPData, error) {
	dataPath := filepath.Join("data", "bjcp_2021_beer.json")
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read BJCP data file: %w", err)
	}

	var bjcpData BJCPData
	if err := json.Unmarshal(data, &bjcpData); err != nil {
		return nil, fmt.Errorf("failed to parse BJCP data: %w", err)
	}

	return &bjcpData, nil
}

// BJCPService provides access to BJCP style information from JSON data.
type BJCPService struct {
	data *BJCPData
}

// NewBJCPServiceFromData creates a new BJCPService instance from BJCPData.
func NewBJCPServiceFromData(data *BJCPData) *BJCPService {
	return &BJCPService{data: data}
}

// NewBJCPService creates a new BJCPService instance with JSON data.
func NewBJCPService() (*BJCPService, error) {
	data, err := LoadBJCPData()
	if err != nil {
		return nil, err
	}

	return &BJCPService{data: data}, nil
}

// GetStyleByCode retrieves a BJCP style by its code (e.g., "21A").
func (s *BJCPService) GetStyleByCode(code string) (*BJCPStyle, error) {
	style, exists := s.data.Styles[strings.ToUpper(code)]
	if !exists {
		return nil, fmt.Errorf("BJCP style not found: %s", code)
	}

	return &style, nil
}

// GetStyleByName retrieves a BJCP style by searching for its name.
func (s *BJCPService) GetStyleByName(name string) (*BJCPStyle, error) {
	nameLower := strings.ToLower(name)

	// First, try exact match
	for _, style := range s.data.Styles {
		if strings.ToLower(style.Name) == nameLower {
			return &style, nil
		}
	}

	// Then, try partial match (starts with)
	for _, style := range s.data.Styles {
		if strings.HasPrefix(strings.ToLower(style.Name), nameLower) {
			return &style, nil
		}
	}

	// Finally, try contains match
	for _, style := range s.data.Styles {
		if strings.Contains(strings.ToLower(style.Name), nameLower) {
			return &style, nil
		}
	}

	return nil, fmt.Errorf("BJCP style not found: %s", name)
}

// GetAllStyles returns all BJCP styles.
func (s *BJCPService) GetAllStyles() map[string]BJCPStyle {
	return s.data.Styles
}

// GetCategories returns all BJCP style categories.
func (s *BJCPService) GetCategories() []string {
	return s.data.Categories
}

// GetStylesByCategory returns all styles in a given category.
func (s *BJCPService) GetStylesByCategory(category string) []BJCPStyle {
	var styles []BJCPStyle
	categoryLower := strings.ToLower(category)

	for _, style := range s.data.Styles {
		if strings.ToLower(style.Category) == categoryLower {
			styles = append(styles, style)
		}
	}

	return styles
}

// GetMetadata returns metadata about the BJCP data.
func (s *BJCPService) GetMetadata() Metadata {
	return s.data.Metadata
}
