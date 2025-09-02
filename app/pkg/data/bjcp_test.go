package data

import (
	"reflect"
	"strings"
	"testing"
)

func mockBJCPData() *BJCPData {
	return &BJCPData{
		Styles: map[string]BJCPStyle{
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
				Vitals: Vitals{
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
				Vitals: Vitals{
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
				Vitals: Vitals{
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
				Vitals: Vitals{
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
		Metadata: Metadata{
			Version:     "2021",
			Source:      "BJCP Style Guidelines",
			LastUpdated: "2025-07-26",
			TotalStyles: 4,
		},
	}
}

func mockEmptyBJCPData() *BJCPData {
	return &BJCPData{
		Styles:     make(map[string]BJCPStyle),
		Categories: []string{},
		Metadata: Metadata{
			Version:     "2021",
			Source:      "test-empty",
			LastUpdated: "2025-07-26",
			TotalStyles: 0,
		},
	}
}

// Test GetStyleByCode - Happy Path Cases.
func TestGetStyleByCode_HappyPath(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"Valid uppercase code", "21A", "American IPA"},
		{"Valid lowercase code", "21a", "American IPA"},
		{"Valid mixed case code", "21a", "American IPA"},
		{"Different valid code", "1A", "American Light Lager"},
		{"Numeric code", "9A", "Doppelbock"},
		{"Higher numbered code", "34A", "Clone Beer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style, err := svc.GetStyleByCode(tt.code)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if style == nil {
				t.Errorf("expected style, got nil")
				return
			}
			if style.Name != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, style.Name)
			}
			// Verify all fields are populated for complete styles
			if tt.code == "21A" || tt.code == "21a" {
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
		})
	}
}

// Test GetStyleByCode - Sad Path Cases.
func TestGetStyleByCode_SadPath(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())

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
		svc  *BJCPService
		code string
	}{
		{"Empty database", NewBJCPServiceFromData(mockEmptyBJCPData()), "21A"},
		{"Empty string code", NewBJCPServiceFromData(mockBJCPData()), ""},
		{"Whitespace only code", NewBJCPServiceFromData(mockBJCPData()), "   "},
		{"Code with leading/trailing spaces", NewBJCPServiceFromData(mockBJCPData()), " 21A "},
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
	svc := NewBJCPServiceFromData(mockBJCPData())

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
	svc := NewBJCPServiceFromData(mockBJCPData())

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
	svc := NewBJCPServiceFromData(mockBJCPData())

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
	svc := NewBJCPServiceFromData(mockEmptyBJCPData())

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
	svc := NewBJCPServiceFromData(mockBJCPData())
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
	svc := NewBJCPServiceFromData(mockEmptyBJCPData())
	styles := svc.GetAllStyles()

	if len(styles) != 0 {
		t.Errorf("expected 0 styles for empty database, got %d", len(styles))
	}
}

// Test GetCategories - Happy Path.
func TestGetCategories_HappyPath(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())
	cats := svc.GetCategories()
	expected := []string{"IPA", "Standard American Beer", "Strong European Beer", "Specialty Beer"}

	if !reflect.DeepEqual(cats, expected) {
		t.Errorf("expected %v, got %v", expected, cats)
	}
}

// Test GetCategories - Empty Database.
func TestGetCategories_EmptyDatabase(t *testing.T) {
	svc := NewBJCPServiceFromData(mockEmptyBJCPData())
	cats := svc.GetCategories()

	if len(cats) != 0 {
		t.Errorf("expected 0 categories for empty database, got %d", len(cats))
	}
}

// Test GetStylesByCategory - Happy Path.
func TestGetStylesByCategory_HappyPath(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())

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
	svc := NewBJCPServiceFromData(mockBJCPData())

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
	svc := NewBJCPServiceFromData(mockEmptyBJCPData())
	styles := svc.GetStylesByCategory("IPA")

	if len(styles) != 0 {
		t.Errorf("expected 0 styles for empty database, got %d", len(styles))
	}
}

// Test GetMetadata - Happy Path.
func TestGetMetadata_HappyPath(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())
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
	svc := NewBJCPServiceFromData(mockEmptyBJCPData())
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
	svc := NewBJCPServiceFromData(mockBJCPData())
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
		data *BJCPData
	}{
		{"Nil data", nil},
		{"Empty data", mockEmptyBJCPData()},
		{"Valid data", mockBJCPData()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewBJCPServiceFromData(tt.data)
			if svc == nil {
				t.Error("expected non-nil service")
			}
			if tt.data != nil && svc.data != tt.data {
				t.Error("expected service data to match input data")
			}
		})
	}
}

// Benchmark tests for performance requirements.
func BenchmarkGetStyleByCode(b *testing.B) {
	svc := NewBJCPServiceFromData(mockBJCPData())

	b.ResetTimer()
	for range b.N {
		_, err := svc.GetStyleByCode("21A")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkGetStyleByName(b *testing.B) {
	svc := NewBJCPServiceFromData(mockBJCPData())

	b.ResetTimer()
	for range b.N {
		_, err := svc.GetStyleByName("American IPA")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkGetStylesByCategory(b *testing.B) {
	svc := NewBJCPServiceFromData(mockBJCPData())

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
	svc := NewBJCPServiceFromData(mockBJCPData())

	// Run multiple goroutines accessing the service simultaneously
	done := make(chan bool, 10)

	for i := range 10 {
		go func(id int) {
			defer func() { done <- true }()

			// Test different operations concurrently
			switch id % 4 {
			case 0:
				_, err := svc.GetStyleByCode("21A")
				if err != nil {
					t.Errorf("goroutine %d: GetStyleByCode failed: %v", id, err)
				}
			case 1:
				_, err := svc.GetStyleByName("American IPA")
				if err != nil {
					t.Errorf("goroutine %d: GetStyleByName failed: %v", id, err)
				}
			case 2:
				styles := svc.GetStylesByCategory("IPA")
				if len(styles) == 0 {
					t.Errorf("goroutine %d: GetStylesByCategory returned no results", id)
				}
			case 3:
				styles := svc.GetAllStyles()
				if len(styles) == 0 {
					t.Errorf("goroutine %d: GetAllStyles returned no results", id)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}
}
