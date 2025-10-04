// Package data provides utilities for loading and handling BJCP data for Brewsource MCP.

// Package data_test contains tests for the BJCP data utilities in Brewsource MCP.
package data_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/pkg/data"
)

func mockBJCPData() *data.BJCPData {
	return &data.BJCPData{
		Styles: map[string]data.BJCPStyle{
			"21A": {
				Code:                      "21A",
				Name:                      "American IPA",
				Category:                  "IPA",
				OverallImpression:         "A decidedly hoppy and bitter, moderately strong American pale ale.",
				Appearance:                "Color ranges from medium gold to light reddish-amber.",
				Aroma:                     "A prominent to intense hop aroma featuring one or more characteristics of American hops.",
				Flavor:                    "Hop flavor is medium to very high, and should reflect an American hop character.",
				Mouthfeel:                 "Medium-light to medium body, with a smooth texture.",
				Comments:                  "A modern American craft beer interpretation of the historical English style.",
				History:                   "An American craft beer innovation first developed in the mid-1970s.",
				CharacteristicIngredients: "Pale ale malt (well-modified and suitable for single-temperature infusion mashing).",
				StyleComparison:           "Stronger and more hoppy than an American Pale Ale.",
				CommercialExamples: []string{
					"Stone IPA",
					"Bell's Two Hearted IPA",
					"Russian River Blind Pig IPA",
				},
				Vitals: data.Vitals{
					ABVMin: 5.5,
					ABVMax: 7.5,
					IBUMin: 40,
					IBUMax: 70,
					SRMMin: 6,
					SRMMax: 14,
					OGMin:  1.056,
					OGMax:  1.070,
					FGMin:  1.008,
					FGMax:  1.014,
				},
			},
			"1A": {
				Code:                      "1A",
				Name:                      "American Light Lager",
				Category:                  "Standard American Beer",
				OverallImpression:         "Highly carbonated, very light-bodied, nearly flavorless lager.",
				Appearance:                "Very pale straw to pale yellow color.",
				Aroma:                     "Low to no malt aroma, although it can be perceived as grainy, sweet, or corn-like if present.",
				Flavor:                    "Relatively neutral palate with a crisp and dry finish.",
				Mouthfeel:                 "Very light (sometimes watery) body.",
				Comments:                  "Designed to appeal to as broad a range of the general public as possible.",
				History:                   "Coors Light was introduced in 1978, Bud Light in 1982.",
				CharacteristicIngredients: "Two or six-row barley with high percentage (up to 40%) of rice or corn as adjuncts.",
				StyleComparison:           "Lower in alcohol content and lighter in body than an American Lager.",
				CommercialExamples:        []string{"Bud Light", "Coors Light", "Keystone Light"},
				Vitals: data.Vitals{
					ABVMin: 2.8,
					ABVMax: 4.2,
					IBUMin: 8,
					IBUMax: 12,
					SRMMin: 2,
					SRMMax: 3,
					OGMin:  1.028,
					OGMax:  1.040,
					FGMin:  0.998,
					FGMax:  1.008,
				},
			},
			"9A": {
				Code:                      "9A",
				Name:                      "Doppelbock",
				Category:                  "Strong European Beer",
				OverallImpression:         "A strong, rich, and very malty German lager.",
				Appearance:                "Deep gold to dark brown in color.",
				Aroma:                     "Very strong maltiness.",
				Flavor:                    "Very rich and malty.",
				Mouthfeel:                 "Medium-full to full body.",
				Comments:                  "Most versions are dark colored and may display the caramelizing and Maillard products of decoction mashing.",
				History:                   "A Bavarian specialty first brewed in Munich by the monks of St. Francis of Paula.",
				CharacteristicIngredients: "Pils and/or Vienna malt for pale versions.",
				StyleComparison:           "A stronger, richer, more full-bodied version of either a Dunkles Bock or a Helles Bock.",
				CommercialExamples: []string{
					"Ayinger Celebrator",
					"Weihenstephaner Korbinian",
					"Spaten Optimator",
				},
				Vitals: data.Vitals{
					ABVMin: 7.0,
					ABVMax: 10.0,
					IBUMin: 16,
					IBUMax: 26,
					SRMMin: 6,
					SRMMax: 25,
					OGMin:  1.072,
					OGMax:  1.112,
					FGMin:  1.016,
					FGMax:  1.024,
				},
			},
			"34A": {
				Code:     "34A",
				Name:     "Clone Beer",
				Category: "Specialty Beer",
				Vitals: data.Vitals{
					ABVMin: 0.0,
					ABVMax: 15.0,
					IBUMin: 0,
					IBUMax: 200,
					SRMMin: 1,
					SRMMax: 40,
					OGMin:  1.000,
					OGMax:  1.200,
					FGMin:  0.990,
					FGMax:  1.030,
				},
			},
		},
		Categories: []string{"IPA", "Standard American Beer", "Strong European Beer", "Specialty Beer"},
		Metadata: data.Metadata{
			Version:     "2021",
			Source:      "BJCP Style Guidelines",
			LastUpdated: "2025-07-26",
			TotalStyles: 4,
		},
	}
}

func mockEmptyBJCPData() *data.BJCPData {
	return &data.BJCPData{
		Styles:     make(map[string]data.BJCPStyle),
		Categories: []string{},
		Metadata: data.Metadata{
			Version:     "2021",
			Source:      "test-empty",
			LastUpdated: "2025-07-26",
			TotalStyles: 0,
		},
	}
}

// Test GetStyleByCode - Basic Happy Path Cases.
func TestGetStyleByCode_BasicCases(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	// Test valid uppercase code
	style, err := svc.GetStyleByCode("21A")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if style == nil {
		t.Fatal("expected style, got nil")
	}
	if style.Name != "American IPA" {
		t.Errorf("expected American IPA, got %s", style.Name)
	}
}

// Test GetStyleByCode - Case Sensitivity.
func TestGetStyleByCode_CaseInsensitive(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	testCodes := []string{"21a", "21A", "1a", "1A"}
	expectedNames := []string{"American IPA", "American IPA", "American Light Lager", "American Light Lager"}

	for i, code := range testCodes {
		style, err := svc.GetStyleByCode(code)
		if err != nil {
			t.Errorf("code %s: expected no error, got %v", code, err)
		}
		if style == nil {
			t.Errorf("code %s: expected style, got nil", code)
			continue
		}
		if style.Name != expectedNames[i] {
			t.Errorf("code %s: expected %s, got %s", code, expectedNames[i], style.Name)
		}
	}
}

// Test GetStyleByCode - Field Validation.
func TestGetStyleByCode_FieldValidation(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	style, err := svc.GetStyleByCode("21A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if style == nil {
		t.Fatal("expected style, got nil")
	}

	// Verify all fields are populated for complete styles
	if style.OverallImpression == "" {
		t.Error("expected OverallImpression to be populated")
	}
	if style.Appearance == "" {
		t.Error("expected Appearance to be populated")
	}
	if len(style.CommercialExamples) == 0 {
		t.Error("expected CommercialExamples to be populated")
	}
}

// Test GetStyleByCode - Sad Path Cases.
func TestGetStyleByCode_SadPath(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	tests := []struct {
		name string
		code string
	}{
		{"Invalid code format", "99Z"},
		{"Non-existent numeric", "100A"},
		{"Invalid characters", "XX1"},
		{"Too long code", "21AA"},
		{"Single character", "A"},
		{"Numbers only", "123"},
		{"Special characters", "21@"},
		{"Unicode characters", "21Ą"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style, err := svc.GetStyleByCode(tt.code)
			if err == nil {
				t.Errorf("expected error for code %s, got nil", tt.code)
			}
			if style != nil {
				t.Errorf("expected nil style for invalid code %s, got %+v", tt.code, style)
			}
			if !strings.Contains(err.Error(), "BJCP style not found") {
				t.Errorf("expected 'BJCP style not found' error message, got: %s", err.Error())
			}
		})
	}
}

// Test GetStyleByCode - Edge Cases.
func TestGetStyleByCode_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		svc  *data.BJCPService
		code string
	}{
		{"Empty database", data.NewBJCPServiceFromData(mockEmptyBJCPData()), "21A"},
		{"Empty string code", data.NewBJCPServiceFromData(mockBJCPData()), ""},
		{"Whitespace only code", data.NewBJCPServiceFromData(mockBJCPData()), "   "},
		{"Code with leading/trailing spaces", data.NewBJCPServiceFromData(mockBJCPData()), " 21A "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style, err := tt.svc.GetStyleByCode(tt.code)
			if err == nil {
				t.Errorf("expected error for edge case '%s', got nil", tt.code)
			}
			if style != nil {
				t.Errorf("expected nil style for edge case '%s', got %+v", tt.code, style)
			}
		})
	}
}

// Test GetStyleByName - Happy Path Cases.
func TestGetStyleByName_HappyPath(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	tests := []struct {
		name         string
		searchName   string
		expectedCode string
		matchType    string
	}{
		{"Exact match", "American IPA", "21A", "exact"},
		{"Case insensitive exact", "american ipa", "21A", "exact"},
		{"Mixed case exact", "AmErIcAn IpA", "21A", "exact"},
		{"Partial match - starts with", "American L", "1A", "partial"},
		{"Partial match - starts with case insensitive", "american l", "1A", "partial"},
		{"Contains match", "Light", "1A", "contains"},
		{"Contains match case insensitive", "light", "1A", "contains"},
		{"Single word match", "Doppelbock", "9A", "exact"},
		{"Partial single word", "Doppel", "9A", "partial"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style, err := svc.GetStyleByName(tt.searchName)
			if err != nil {
				t.Errorf("expected no error for '%s', got %v", tt.searchName, err)
			}
			if style == nil {
				t.Errorf("expected style for '%s', got nil", tt.searchName)
				return
			}
			if style.Code != tt.expectedCode {
				t.Errorf("expected code %s for '%s', got %s", tt.expectedCode, tt.searchName, style.Code)
			}
		})
	}
}

// Test GetStyleByName - Sad Path Cases.
func TestGetStyleByName_SadPath(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	tests := []string{
		"Nonexistent Style",
		"XYZ Beer",
		"NotAStyle123",
		"Random Text",
		"   ",
		"",
	}

	for _, searchName := range tests {
		t.Run("Search for "+searchName, func(t *testing.T) {
			style, err := svc.GetStyleByName(searchName)
			if err == nil && searchName != "" && searchName != "   " {
				t.Errorf("expected error for nonexistent style '%s', got nil", searchName)
			}
			if style != nil {
				t.Errorf("expected nil style for nonexistent style '%s', got %+v", searchName, style)
			}
		})
	}
}

// Test GetStyleByName - Boundary Cases.
func TestGetStyleByName_BoundaryCase(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	tests := []struct {
		name       string
		searchName string
	}{
		{"Very long name", strings.Repeat("a", 1000)},
		{"Single character", "A"},
		{"Only numbers", "123"},
		{"Special characters", "@#$%"},
		{"Unicode characters", "Ąćęłńóśźż"},
		{"Mixed unicode and ascii", "Beer ąćę 123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style, err := svc.GetStyleByName(tt.searchName)
			if err == nil {
				t.Errorf("expected error for boundary case '%s', got nil", tt.searchName)
			}
			if style != nil {
				t.Errorf("expected nil style for boundary case '%s', got %+v", tt.searchName, style)
			}
		})
	}
}

// Test GetStyleByName - Empty Database.
func TestGetStyleByName_EmptyDatabase(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockEmptyBJCPData())

	style, err := svc.GetStyleByName("American IPA")
	if err == nil {
		t.Error("expected error for empty database, got nil")
	}
	if style != nil {
		t.Errorf("expected nil style for empty database, got %+v", style)
	}
}

// Test GetAllStyles - Happy Path.
func TestGetAllStyles_HappyPath(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())
	styles := svc.GetAllStyles()

	if len(styles) != 4 {
		t.Errorf("expected 4 styles, got %d", len(styles))
	}

	// Verify specific styles exist
	expectedCodes := []string{"21A", "1A", "9A", "34A"}
	for _, code := range expectedCodes {
		if _, exists := styles[code]; !exists {
			t.Errorf("expected style %s to exist", code)
		}
	}
}

// Test GetAllStyles - Empty Database.
func TestGetAllStyles_EmptyDatabase(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockEmptyBJCPData())
	styles := svc.GetAllStyles()

	if len(styles) != 0 {
		t.Errorf("expected 0 styles for empty database, got %d", len(styles))
	}
}

// Test GetCategories - Happy Path.
func TestGetCategories_HappyPath(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())
	cats := svc.GetCategories()
	expected := []string{"IPA", "Standard American Beer", "Strong European Beer", "Specialty Beer"}

	if !reflect.DeepEqual(cats, expected) {
		t.Errorf("expected %v, got %v", expected, cats)
	}
}

// Test GetCategories - Empty Database.
func TestGetCategories_EmptyDatabase(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockEmptyBJCPData())
	cats := svc.GetCategories()

	if len(cats) != 0 {
		t.Errorf("expected 0 categories for empty database, got %d", len(cats))
	}
}

// Test GetStylesByCategory - Happy Path.
func TestGetStylesByCategory_HappyPath(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	tests := []struct {
		name          string
		category      string
		expectedCount int
		expectedCodes []string
	}{
		{"IPA category", "IPA", 1, []string{"21A"}},
		{"Standard American Beer", "Standard American Beer", 1, []string{"1A"}},
		{"Strong European Beer", "Strong European Beer", 1, []string{"9A"}},
		{"Specialty Beer", "Specialty Beer", 1, []string{"34A"}},
		{"Case insensitive", "ipa", 1, []string{"21A"}},
		{"Mixed case", "StAnDaRd AmErIcAn BeEr", 1, []string{"1A"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styles := svc.GetStylesByCategory(tt.category)
			if len(styles) != tt.expectedCount {
				t.Errorf("expected %d styles for category '%s', got %d", tt.expectedCount, tt.category, len(styles))
			}

			for i, expectedCode := range tt.expectedCodes {
				if i < len(styles) && styles[i].Code != expectedCode {
					t.Errorf("expected code %s at position %d, got %s", expectedCode, i, styles[i].Code)
				}
			}
		})
	}
}

// Test GetStylesByCategory - Sad Path.
func TestGetStylesByCategory_SadPath(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	tests := []string{
		"Nonexistent Category",
		"",
		"   ",
		"XYZ Category",
		"123",
		"@#$%",
	}

	for _, category := range tests {
		t.Run("Category: "+category, func(t *testing.T) {
			styles := svc.GetStylesByCategory(category)
			if len(styles) != 0 {
				t.Errorf("expected 0 styles for nonexistent category '%s', got %d", category, len(styles))
			}
		})
	}
}

// Test GetStylesByCategory - Empty Database.
func TestGetStylesByCategory_EmptyDatabase(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockEmptyBJCPData())
	styles := svc.GetStylesByCategory("IPA")

	if len(styles) != 0 {
		t.Errorf("expected 0 styles for empty database, got %d", len(styles))
	}
}

// Test GetMetadata - Happy Path.
func TestGetMetadata_HappyPath(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())
	meta := svc.GetMetadata()

	if meta.Version != "2021" {
		t.Errorf("expected version '2021', got '%s'", meta.Version)
	}
	if meta.TotalStyles != 4 {
		t.Errorf("expected total styles 4, got %d", meta.TotalStyles)
	}
	if meta.Source != "BJCP Style Guidelines" {
		t.Errorf("expected source 'BJCP Style Guidelines', got '%s'", meta.Source)
	}
	if meta.LastUpdated != "2025-07-26" {
		t.Errorf("expected last updated '2025-07-26', got '%s'", meta.LastUpdated)
	}
}

// Test GetMetadata - Empty Database.
func TestGetMetadata_EmptyDatabase(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockEmptyBJCPData())
	meta := svc.GetMetadata()

	if meta.Version != "2021" {
		t.Errorf("expected version '2021', got '%s'", meta.Version)
	}
	if meta.TotalStyles != 0 {
		t.Errorf("expected total styles 0, got %d", meta.TotalStyles)
	}
	if meta.Source != "test-empty" {
		t.Errorf("expected source 'test-empty', got '%s'", meta.Source)
	}
}

// Test Vitals Struct Validation.
func TestVitals_Validation(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())
	style, err := svc.GetStyleByCode("21A")
	if err != nil {
		t.Fatalf("unexpected error getting style: %v", err)
	}

	vitals := style.Vitals

	// Test ABV ranges
	if vitals.ABVMin > vitals.ABVMax {
		t.Error("ABVMin should not be greater than ABVMax")
	}
	if vitals.ABVMin < 0 {
		t.Error("ABVMin should not be negative")
	}

	// Test IBU ranges
	if vitals.IBUMin > vitals.IBUMax {
		t.Error("IBUMin should not be greater than IBUMax")
	}
	if vitals.IBUMin < 0 {
		t.Error("IBUMin should not be negative")
	}

	// Test SRM ranges
	if vitals.SRMMin > vitals.SRMMax {
		t.Error("SRMMin should not be greater than SRMMax")
	}
	if vitals.SRMMin < 0 {
		t.Error("SRMMin should not be negative")
	}

	// Test OG ranges
	if vitals.OGMin > vitals.OGMax {
		t.Error("OGMin should not be greater than OGMax")
	}

	// Test FG ranges
	if vitals.FGMin > vitals.FGMax {
		t.Error("FGMin should not be greater than FGMax")
	}
}

// Test Service Constructor Edge Cases.
func TestNewBJCPServiceFromData_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		data *data.BJCPData
	}{
		{"Nil data", nil},
		{"Empty data", mockEmptyBJCPData()},
		{"Valid data", mockBJCPData()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := data.NewBJCPServiceFromData(tt.data)
			if svc == nil {
				t.Error("expected non-nil service")
			}
			// Test that the service is functional by checking it can return metadata
			if tt.data != nil {
				metadata := svc.GetMetadata()
				if metadata.Version != tt.data.Metadata.Version {
					t.Error("expected service metadata to match input data metadata")
				}
			}
		})
	}
}

// Benchmark tests for performance requirements.
func BenchmarkGetStyleByCode(b *testing.B) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	b.ResetTimer()
	for range b.N {
		_, err := svc.GetStyleByCode("21A")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkGetStyleByName(b *testing.B) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	b.ResetTimer()
	for range b.N {
		_, err := svc.GetStyleByName("American IPA")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkGetStylesByCategory(b *testing.B) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	b.ResetTimer()
	for range b.N {
		styles := svc.GetStylesByCategory("IPA")
		if len(styles) == 0 {
			b.Fatal("expected styles to be returned")
		}
	}
}

// Test concurrent access for thread safety.
func TestConcurrentAccess(t *testing.T) {
	svc := data.NewBJCPServiceFromData(mockBJCPData())

	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	// Launch concurrent goroutines
	for i := range numGoroutines {
		go testConcurrentOperation(t, svc, i, done)
	}

	// Wait for all goroutines to complete
	for range numGoroutines {
		<-done
	}
}

// Helper function to test different operations concurrently.
func testConcurrentOperation(t *testing.T, svc *data.BJCPService, id int, done chan bool) {
	defer func() { done <- true }()

	// Test different operations based on ID
	switch id % 4 {
	case 0:
		testConcurrentGetStyleByCode(t, svc, id)
	case 1:
		testConcurrentGetStyleByName(t, svc, id)
	case 2:
		testConcurrentGetStylesByCategory(t, svc, id)
	case 3:
		testConcurrentGetAllStyles(t, svc, id)
	}
}

func testConcurrentGetStyleByCode(t *testing.T, svc *data.BJCPService, id int) {
	_, err := svc.GetStyleByCode("21A")
	if err != nil {
		t.Errorf("goroutine %d: GetStyleByCode failed: %v", id, err)
	}
}

func testConcurrentGetStyleByName(t *testing.T, svc *data.BJCPService, id int) {
	_, err := svc.GetStyleByName("American IPA")
	if err != nil {
		t.Errorf("goroutine %d: GetStyleByName failed: %v", id, err)
	}
}

func testConcurrentGetStylesByCategory(t *testing.T, svc *data.BJCPService, id int) {
	styles := svc.GetStylesByCategory("IPA")
	if len(styles) == 0 {
		t.Errorf("goroutine %d: GetStylesByCategory returned no results", id)
	}
}

func testConcurrentGetAllStyles(t *testing.T, svc *data.BJCPService, id int) {
	styles := svc.GetAllStyles()
	if len(styles) == 0 {
		t.Errorf("goroutine %d: GetAllStyles returned no results", id)
	}
}

// Test LoadBJCPData function.
func TestLoadBJCPData(t *testing.T) {
	// Test that LoadBJCPData returns an error when file doesn't exist
	// We can't test successful loading without the actual data file in the test environment,
	// but we can test the error path
	_, err := data.LoadBJCPData()
	// In test environment, this will likely fail due to missing data file, which is expected
	if err == nil {
		// If it succeeds, that's actually fine too - means the data file exists
		t.Log("LoadBJCPData succeeded - data file is available")
	} else if !strings.Contains(err.Error(), "data file") && !strings.Contains(err.Error(), "no such file") {
		// Expected in test environment
		t.Errorf("Expected file-related error, got: %v", err)
	}
}

// Test LoadBJCPData with simulated data file.
func TestLoadBJCPData_WithMockData(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")
	err := os.MkdirAll(dataDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create temp data directory: %v", err)
	}

	// Create a valid JSON file
	validJSON := `{
		"styles": {
			"21A": {
				"code": "21A",
				"name": "American IPA",
				"category": "IPA",
				"overall_impression": "Test impression",
				"vitals": {
					"abv_min": 5.5,
					"abv_max": 7.5,
					"ibu_min": 40,
					"ibu_max": 70,
					"srm_min": 6,
					"srm_max": 14,
					"og_min": 1.056,
					"og_max": 1.070,
					"fg_min": 1.008,
					"fg_max": 1.014
				}
			}
		},
		"categories": ["IPA"],
		"metadata": {
			"version": "2021",
			"source": "BJCP Style Guidelines",
			"last_updated": "2025-01-01",
			"total_styles": 1
		}
	}`

	bjcpFile := filepath.Join(dataDir, "bjcp_2021_beer.json")
	err = os.WriteFile(bjcpFile, []byte(validJSON), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test data file: %v", err)
	}

	// Change to temp directory
	t.Chdir(tempDir)

	// Test successful loading
	bjcpData, err := data.LoadBJCPData()
	if err != nil {
		t.Fatalf("Expected successful loading, got error: %v", err)
	}

	if bjcpData == nil {
		t.Fatal("Expected non-nil BJCP data")
	}

	if len(bjcpData.Styles) != 1 {
		t.Errorf("Expected 1 style, got %d", len(bjcpData.Styles))
	}

	if bjcpData.Metadata.Version != "2021" {
		t.Errorf("Expected version '2021', got '%s'", bjcpData.Metadata.Version)
	}
}

// Test LoadBJCPData with invalid JSON.
func TestLoadBJCPData_InvalidJSON(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")
	err := os.MkdirAll(dataDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create temp data directory: %v", err)
	}

	// Create invalid JSON file
	invalidJSON := `{
		"styles": {
			"21A": {
				"code": "21A",
				"name": "American IPA"
				// Missing comma here makes it invalid JSON
			}
		}
	}`

	bjcpFile := filepath.Join(dataDir, "bjcp_2021_beer.json")
	err = os.WriteFile(bjcpFile, []byte(invalidJSON), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test data file: %v", err)
	}

	// Change to temp directory
	t.Chdir(tempDir)

	// Test that invalid JSON returns parse error
	_, err = data.LoadBJCPData()
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

// Test LoadBJCPData path traversal protection.
func TestLoadBJCPData_PathTraversalProtection(t *testing.T) {
	// The LoadBJCPData function has built-in path traversal protection
	// We can test this by temporarily changing the working directory
	// to a location where the path validation would detect traversal

	// Create a temporary directory structure that would trigger path validation
	tempDir := t.TempDir()

	// Change to a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err := os.MkdirAll(subDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	t.Chdir(subDir)

	// Now try to load BJCP data - this should fail with file not found
	// since there's no data directory in our temp subdirectory
	_, err = data.LoadBJCPData()
	if err == nil {
		t.Error("Expected error when data file doesn't exist")
	}

	// The error should be about file not being found, not path traversal
	if strings.Contains(err.Error(), "path traversal") {
		t.Error("Unexpected path traversal error in normal scenario")
	}
}

// Test NewBJCPService function.
func TestNewBJCPService(t *testing.T) {
	// Test that NewBJCPService returns an error when LoadBJCPData fails
	_, err := data.NewBJCPService()
	// In test environment, this will likely fail due to missing data file, which is expected
	if err == nil {
		// If it succeeds, that's actually fine too - means the data file exists
		t.Log("NewBJCPService succeeded - data file is available")
	} else if !strings.Contains(err.Error(), "data file") && !strings.Contains(err.Error(), "no such file") {
		// Expected in test environment
		t.Errorf("Expected file-related error, got: %v", err)
	}
}
