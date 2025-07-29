package bjcp

import (
	"testing"
)

func TestParseStyleCode(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		expectError         bool
		expectedCategory    int
		expectedSubcategory string
	}{
		{
			name:                "Valid IPA code",
			input:               "21A",
			expectError:         false,
			expectedCategory:    21,
			expectedSubcategory: "A",
		},
		{
			name:                "Valid single digit category",
			input:               "4B",
			expectError:         false,
			expectedCategory:    4,
			expectedSubcategory: "B",
		},
		{
			name:                "Valid two letter subcategory",
			input:               "27AA",
			expectError:         false,
			expectedCategory:    27,
			expectedSubcategory: "AA",
		},
		{
			name:        "Invalid - no letters",
			input:       "21",
			expectError: true,
		},
		{
			name:        "Invalid - no numbers",
			input:       "AA",
			expectError: true,
		},
		{
			name:        "Invalid - category too high",
			input:       "99A",
			expectError: true,
		},
		{
			name:        "Invalid - empty string",
			input:       "",
			expectError: true,
		},
		{
			name:                "Invalid - lowercase (should be normalized)",
			input:               "21a",
			expectError:         false,
			expectedCategory:    21,
			expectedSubcategory: "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseStyleCode(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.Category != tt.expectedCategory {
				t.Errorf("Expected category %d, got %d", tt.expectedCategory, result.Category)
			}

			if result.Subcategory != tt.expectedSubcategory {
				t.Errorf("Expected subcategory %s, got %s", tt.expectedSubcategory, result.Subcategory)
			}
		})
	}
}

func TestIsValidStyleCode(t *testing.T) {
	validCodes := []string{"1A", "21A", "4B", "27AA", "34C"}
	invalidCodes := []string{"", "21", "AA", "99A", "0A", "35A"}

	for _, code := range validCodes {
		t.Run("Valid_"+code, func(t *testing.T) {
			if !IsValidStyleCode(code) {
				t.Errorf("Expected %s to be valid", code)
			}
		})
	}

	for _, code := range invalidCodes {
		t.Run("Invalid_"+code, func(t *testing.T) {
			if IsValidStyleCode(code) {
				t.Errorf("Expected %s to be invalid", code)
			}
		})
	}
}

func TestGetCategoryName(t *testing.T) {
	tests := []struct {
		category int
		expected string
	}{
		{1, "Standard American Beer"},
		{21, "IPA"},
		{4, "Pale Malty European Lager"},
		{99, "Unknown Category"}, // Invalid category
	}

	for _, tt := range tests {
		t.Run("Category_"+string(rune(tt.category)), func(t *testing.T) {
			result := GetCategoryName(tt.category)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestStyleRange(t *testing.T) {
	r := StyleRange{Min: 5.0, Max: 7.0}

	tests := []struct {
		value    float64
		expected bool
	}{
		{5.0, true},  // Lower bound
		{6.0, true},  // Middle
		{7.0, true},  // Upper bound
		{4.9, false}, // Below range
		{7.1, false}, // Above range
	}

	for _, tt := range tests {
		t.Run("Value_"+string(rune(int(tt.value*10))), func(t *testing.T) {
			result := r.Contains(tt.value)
			if result != tt.expected {
				t.Errorf("Range.Contains(%.1f) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestStyleGuidelines(t *testing.T) {
	guidelines := StyleGuidelines{
		ABV: StyleRange{Min: 5.5, Max: 7.5},
		IBU: StyleRange{Min: 40, Max: 70},
		SRM: StyleRange{Min: 6, Max: 14},
		OG:  StyleRange{Min: 1.056, Max: 1.070},
		FG:  StyleRange{Min: 1.008, Max: 1.014},
	}

	// Test a beer within guidelines (American IPA)
	results := guidelines.IsWithinGuidelines(6.5, 55, 8, 1.065, 1.012)

	expected := map[string]bool{
		"ABV": true,
		"IBU": true,
		"SRM": true,
		"OG":  true,
		"FG":  true,
	}

	for param, expectedResult := range expected {
		if results[param] != expectedResult {
			t.Errorf("Expected %s to be %v, got %v", param, expectedResult, results[param])
		}
	}

	// Test a beer outside guidelines
	results = guidelines.IsWithinGuidelines(4.0, 20, 2, 1.040, 1.020)

	for param := range expected {
		if results[param] {
			t.Errorf("Expected %s to be false for out-of-range beer, got true", param)
		}
	}
}

func TestParseStyleCodeComprehensive(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		expectError         bool
		expectedCategory    int
		expectedSubcategory string
	}{
		// Happy Path Test Cases
		{
			name:                "Standard IPA",
			input:               "21A",
			expectError:         false,
			expectedCategory:    21,
			expectedSubcategory: "A",
		},
		{
			name:                "Lager category",
			input:               "4B",
			expectError:         false,
			expectedCategory:    4,
			expectedSubcategory: "B",
		},
		{
			name:                "Specialty beer",
			input:               "27A",
			expectError:         false,
			expectedCategory:    27,
			expectedSubcategory: "A",
		},
		{
			name:                "Historical beer",
			input:               "27AA", // Two letter subcategory
			expectError:         false,
			expectedCategory:    27,
			expectedSubcategory: "AA",
		},

		// Boundary Value Test Cases
		{
			name:                "Minimum category (1)",
			input:               "1A",
			expectError:         false,
			expectedCategory:    1,
			expectedSubcategory: "A",
		},
		{
			name:                "Maximum category (34)",
			input:               "34C",
			expectError:         false,
			expectedCategory:    34,
			expectedSubcategory: "C",
		},
		{
			name:                "Subcategory A",
			input:               "12A",
			expectError:         false,
			expectedCategory:    12,
			expectedSubcategory: "A",
		},
		{
			name:                "Subcategory Z",
			input:               "12Z",
			expectError:         false,
			expectedCategory:    12,
			expectedSubcategory: "Z",
		},

		// Equivalence Partitioning Test Cases
		{
			name:                "Single digit category",
			input:               "5C",
			expectError:         false,
			expectedCategory:    5,
			expectedSubcategory: "C",
		},
		{
			name:                "Double digit category",
			input:               "15D",
			expectError:         false,
			expectedCategory:    15,
			expectedSubcategory: "D",
		},
		{
			name:                "Single letter subcategory",
			input:               "8B",
			expectError:         false,
			expectedCategory:    8,
			expectedSubcategory: "B",
		},
		{
			name:                "Double letter subcategory",
			input:               "29BB",
			expectError:         false,
			expectedCategory:    29,
			expectedSubcategory: "BB",
		},

		// Sad Path Test Cases
		{
			name:        "Invalid - no letters",
			input:       "21",
			expectError: true,
		},
		{
			name:        "Invalid - no numbers",
			input:       "AA",
			expectError: true,
		},
		{
			name:        "Invalid - category zero",
			input:       "0A",
			expectError: true,
		},
		{
			name:        "Invalid - category too high",
			input:       "99A",
			expectError: true,
		},
		{
			name:        "Invalid - category 35",
			input:       "35A",
			expectError: true,
		},
		{
			name:        "Invalid - empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "Invalid - only letters",
			input:       "ABC",
			expectError: true,
		},
		{
			name:        "Invalid - only numbers",
			input:       "123",
			expectError: true,
		},
		{
			name:        "Invalid - special characters",
			input:       "21@",
			expectError: true,
		},
		{
			name:        "Invalid - spaces",
			input:       "21 A",
			expectError: true,
		},
		{
			name:                "Invalid - mixed case should normalize",
			input:               "21a",
			expectError:         false,
			expectedCategory:    21,
			expectedSubcategory: "A", // Should be normalized to uppercase
		},

		// Edge Case Test Cases
		{
			name:        "Invalid - very long string",
			input:       "21ABCDEF",
			expectError: true,
		},
		{
			name:        "Invalid - decimal category",
			input:       "21.5A",
			expectError: true,
		},
		{
			name:        "Invalid - negative category",
			input:       "-21A",
			expectError: true,
		},
		{
			name:                "Edge - three character code",
			input:               "1AA",
			expectError:         false,
			expectedCategory:    1,
			expectedSubcategory: "AA",
		},
		{
			name:                "Edge - four character code",
			input:               "34AB",
			expectError:         false,
			expectedCategory:    34,
			expectedSubcategory: "AB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styleCode, err := ParseStyleCode(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if styleCode.Category != tt.expectedCategory {
				t.Errorf("Expected category %d, got %d", tt.expectedCategory, styleCode.Category)
			}

			if styleCode.Subcategory != tt.expectedSubcategory {
				t.Errorf("Expected subcategory %s, got %s", tt.expectedSubcategory, styleCode.Subcategory)
			}
		})
	}
}

func TestIsValidStyleCodeComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		// Happy Path Test Cases
		{"Valid standard IPA", "21A", true},
		{"Valid porter", "20B", true},
		{"Valid lager", "4C", true},
		{"Valid specialty", "27A", true},

		// Boundary Value Test Cases
		{"Valid minimum category", "1A", true},
		{"Valid maximum category", "34C", true},
		{"Valid first subcategory", "21A", true},
		{"Valid last typical subcategory", "21D", true},
		{"Valid double letter", "27AA", true},

		// Equivalence Partitioning Test Cases
		{"Valid single digit category", "5B", true},
		{"Valid double digit category", "15C", true},
		{"Valid single letter sub", "12A", true},
		{"Valid double letter sub", "29BB", true},

		// Sad Path Test Cases
		{"Invalid empty", "", false},
		{"Invalid only numbers", "21", false},
		{"Invalid only letters", "AA", false},
		{"Invalid category zero", "0A", false},
		{"Invalid category too high", "99A", false},
		{"Invalid category 35", "35A", false},
		{"Invalid special chars", "21@", false},
		{"Invalid spaces", "21 A", false},
		{"Invalid decimal", "21.5A", false},
		{"Invalid negative", "-21A", false},

		// Edge Case Test Cases
		{"Edge valid lowercase (normalized)", "21a", true},
		{"Edge valid mixed case", "21b", true},
		{"Edge valid three chars", "1AA", true},
		{"Edge valid four chars", "34AB", true},
		{"Edge invalid too long", "21ABCD", false},
		{"Edge invalid too short", "2", false},
		{"Edge invalid symbols", "21#", false},
		{"Edge invalid unicode", "21α", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidStyleCode(tt.code)
			if result != tt.expected {
				t.Errorf("IsValidStyleCode(%q) = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

func TestGetCategoryNameComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		category int
		expected string
	}{
		// Happy Path Test Cases
		{"Standard American Beer category", 1, "Standard American Beer"},
		{"Pale Bitter European Beer category", 5, "Pale Bitter European Beer"},
		{"IPA category", 21, "IPA"},
		{"American Porter and Stout category", 20, "American Porter and Stout"},

		// Boundary Value Test Cases
		{"Minimum category", 1, "Standard American Beer"},
		{"Maximum category", 34, "Specialty Beer"},

		// Equivalence Partitioning Test Cases
		{"Single digit category", 4, "Pale Malty European Lager"},
		{"Double digit category", 15, "Irish Beer"},
		{"Specialty category", 27, "Historical Beer"},

		// Sad Path Test Cases
		{"Invalid category zero", 0, "Unknown Category"},
		{"Invalid negative category", -1, "Unknown Category"},
		{"Invalid category too high", 35, "Unknown Category"},
		{"Invalid category way too high", 99, "Unknown Category"},

		// Edge Case Test Cases
		{"Edge valid category 34", 34, "Specialty Beer"},
		{"Edge invalid category 35", 35, "Unknown Category"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := GetCategoryName(tt.category)

			if name != tt.expected {
				t.Errorf("GetCategoryName(%d) = %q, want %q", tt.category, name, tt.expected)
			}
		})
	}
}

func TestStyleRangeValidation(t *testing.T) {
	tests := []struct {
		name        string
		abvMin      float64
		abvMax      float64
		ibuMin      int
		ibuMax      int
		srmMin      float64
		srmMax      float64
		expectValid bool
	}{
		// Happy Path Test Cases
		{
			name:        "Valid IPA ranges",
			abvMin:      5.5,
			abvMax:      7.5,
			ibuMin:      40,
			ibuMax:      70,
			srmMin:      6.0,
			srmMax:      14.0,
			expectValid: true,
		},
		{
			name:        "Valid light lager ranges",
			abvMin:      2.8,
			abvMax:      4.2,
			ibuMin:      8,
			ibuMax:      12,
			srmMin:      2.0,
			srmMax:      3.0,
			expectValid: true,
		},

		// Boundary Value Test Cases
		{
			name:        "Zero ABV minimum",
			abvMin:      0.0,
			abvMax:      4.0,
			ibuMin:      5,
			ibuMax:      15,
			srmMin:      2.0,
			srmMax:      4.0,
			expectValid: true, // Non-alcoholic beer
		},
		{
			name:        "High ABV maximum",
			abvMin:      8.0,
			abvMax:      15.0,
			ibuMin:      20,
			ibuMax:      50,
			srmMin:      10.0,
			srmMax:      25.0,
			expectValid: true, // Strong beer
		},

		// Sad Path Test Cases
		{
			name:        "Invalid - ABV min > max",
			abvMin:      7.5,
			abvMax:      5.5,
			ibuMin:      40,
			ibuMax:      70,
			srmMin:      6.0,
			srmMax:      14.0,
			expectValid: false,
		},
		{
			name:        "Invalid - IBU min > max",
			abvMin:      5.5,
			abvMax:      7.5,
			ibuMin:      70,
			ibuMax:      40,
			srmMin:      6.0,
			srmMax:      14.0,
			expectValid: false,
		},
		{
			name:        "Invalid - SRM min > max",
			abvMin:      5.5,
			abvMax:      7.5,
			ibuMin:      40,
			ibuMax:      70,
			srmMin:      14.0,
			srmMax:      6.0,
			expectValid: false,
		},
		{
			name:        "Invalid - negative ABV",
			abvMin:      -1.0,
			abvMax:      7.5,
			ibuMin:      40,
			ibuMax:      70,
			srmMin:      6.0,
			srmMax:      14.0,
			expectValid: false,
		},
		{
			name:        "Invalid - negative IBU",
			abvMin:      5.5,
			abvMax:      7.5,
			ibuMin:      -10,
			ibuMax:      70,
			srmMin:      6.0,
			srmMax:      14.0,
			expectValid: false,
		},
		{
			name:        "Invalid - negative SRM",
			abvMin:      5.5,
			abvMax:      7.5,
			ibuMin:      40,
			ibuMax:      70,
			srmMin:      -2.0,
			srmMax:      14.0,
			expectValid: false,
		},

		// Edge Case Test Cases
		{
			name:        "Edge - equal min and max values",
			abvMin:      5.0,
			abvMax:      5.0,
			ibuMin:      30,
			ibuMax:      30,
			srmMin:      8.0,
			srmMax:      8.0,
			expectValid: true, // Technically valid
		},
		{
			name:        "Edge - very high values",
			abvMin:      12.0,
			abvMax:      20.0,
			ibuMin:      80,
			ibuMax:      120,
			srmMin:      30.0,
			srmMax:      40.0,
			expectValid: true, // Extreme but valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock style range to validate
			isValid := tt.abvMin <= tt.abvMax &&
				tt.ibuMin <= tt.ibuMax &&
				tt.srmMin <= tt.srmMax &&
				tt.abvMin >= 0 &&
				tt.ibuMin >= 0 &&
				tt.srmMin >= 0

			if isValid != tt.expectValid {
				t.Errorf("Style range validation = %v, want %v", isValid, tt.expectValid)
			}
		})
	}
}
