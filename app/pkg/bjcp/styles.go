package bjcp

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	// ErrInvalidStyleCode is returned when a style code doesn't match BJCP format.
	ErrInvalidStyleCode = errors.New("invalid BJCP style code format")

	// styleCodeRegex matches valid BJCP style codes (e.g., 21A, 1B, 23C).
	styleCodeRegex = regexp.MustCompile(`^(\d{1,2})([A-Z]{1,2})$`)
)

// StyleCode represents a BJCP style code with category and subcategory.
type StyleCode struct {
	Category    int
	Subcategory string
	Full        string
}

// ParseStyleCode parses a BJCP style code string into its components.
func ParseStyleCode(code string) (*StyleCode, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))

	matches := styleCodeRegex.FindStringSubmatch(normalized)
	if len(matches) != 3 {
		return nil, ErrInvalidStyleCode
	}

	category, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, ErrInvalidStyleCode
	}

	// Validate category range (1-34 for standard BJCP categories)
	if category < 1 || category > 34 {
		return nil, ErrInvalidStyleCode
	}

	return &StyleCode{
		Category:    category,
		Subcategory: matches[2],
		Full:        normalized,
	}, nil
}

// IsValidStyleCode checks if a string is a valid BJCP style code.
func IsValidStyleCode(code string) bool {
	_, err := ParseStyleCode(code)
	return err == nil
}

// FormatStyleCode ensures consistent formatting of style codes.
func FormatStyleCode(code string) (string, error) {
	parsed, err := ParseStyleCode(code)
	if err != nil {
		return "", err
	}
	return parsed.Full, nil
}

// StyleRange represents a range of values for beer characteristics.
type StyleRange struct {
	Min float64
	Max float64
}

// Contains checks if a value falls within the style range.
func (r StyleRange) Contains(value float64) bool {
	return value >= r.Min && value <= r.Max
}

// String returns a formatted string representation of the range.
func (r StyleRange) String() string {
	if r.Min == r.Max {
		return strconv.FormatFloat(r.Min, 'f', 1, 64)
	}
	return strconv.FormatFloat(r.Min, 'f', 1, 64) + " - " + strconv.FormatFloat(r.Max, 'f', 1, 64)
}

// StyleGuidelines represents the numerical guidelines for a BJCP style.
type StyleGuidelines struct {
	ABV StyleRange // Alcohol By Volume percentage
	IBU StyleRange // International Bitterness Units
	SRM StyleRange // Standard Reference Method (color)
	OG  StyleRange // Original Gravity
	FG  StyleRange // Final Gravity
}

// IsWithinGuidelines checks if beer characteristics fall within style guidelines.
func (sg StyleGuidelines) IsWithinGuidelines(abv, ibu, srm, og, fg float64) map[string]bool {
	return map[string]bool{
		"ABV": sg.ABV.Contains(abv),
		"IBU": sg.IBU.Contains(ibu),
		"SRM": sg.SRM.Contains(srm),
		"OG":  sg.OG.Contains(og),
		"FG":  sg.FG.Contains(fg),
	}
}

// GetCategoryName returns the name of a BJCP category by number.
func GetCategoryName(categoryNum int) string {
	categories := map[int]string{
		1:  "Standard American Beer",
		2:  "International Lager",
		3:  "Czech Lager",
		4:  "Pale Malty European Lager",
		5:  "Pale Bitter European Beer",
		6:  "Amber Malty European Lager",
		7:  "Amber Bitter European Beer",
		8:  "Dark European Lager",
		9:  "Strong European Beer",
		10: "German Wheat Beer",
		11: "British Bitter",
		12: "Pale Commonwealth Beer",
		13: "Brown British Beer",
		14: "Scottish Ale",
		15: "Irish Beer",
		16: "Dark British Beer",
		17: "Strong British Ale",
		18: "Pale American Ale",
		19: "Amber and Brown American Beer",
		20: "American Porter and Stout",
		21: "IPA",
		22: "Strong American Ale",
		23: "European Sour Ale",
		24: "Belgian Ale",
		25: "Strong Belgian Ale",
		26: "Trappist Ale",
		27: "Historical Beer",
		28: "American Wild Ale",
		29: "Fruit Beer",
		30: "Spiced Beer",
		31: "Alternative Fermentables Beer",
		32: "Smoked Beer",
		33: "Wood Beer",
		34: "Specialty Beer",
	}

	if name, exists := categories[categoryNum]; exists {
		return name
	}
	return "Unknown Category"
}

// GetSubcategoryStyles returns common subcategory letters and their typical meanings.
func GetSubcategoryStyles() map[string]string {
	return map[string]string{
		"A": "Primary or most common variant",
		"B": "Secondary variant or different strength",
		"C": "Third variant or specialty version",
		"D": "Additional variant (less common)",
	}
}

// ValidateStyleFormat performs comprehensive validation of a style code.
func ValidateStyleFormat(code string) error {
	if code == "" {
		return errors.New("style code cannot be empty")
	}

	if len(code) < 2 || len(code) > 4 {
		return errors.New("style code must be 2-4 characters long")
	}

	_, err := ParseStyleCode(code)
	return err
}

// SuggestSimilarCodes suggests similar style codes based on partial input.
func SuggestSimilarCodes(partial string) []string {
	// This would typically query a database, but for now we'll provide
	// some common examples based on the partial input
	commonCodes := []string{
		"1A", "1B", "1C", "1D",
		"2A", "2B", "2C",
		"3A", "3B",
		"4A", "4B", "4C",
		"5A", "5B", "5C", "5D",
		"6A", "6B", "6C",
		"7A", "7B", "7C",
		"8A", "8B",
		"9A", "9B", "9C",
		"10A", "10B", "10C",
		"11A", "11B", "11C",
		"12A", "12B", "12C",
		"13A", "13B", "13C",
		"14A", "14B", "14C",
		"15A", "15B",
		"16A", "16B", "16C", "16D",
		"17A", "17B", "17C", "17D",
		"18A", "18B",
		"19A", "19B", "19C",
		"20A", "20B", "20C",
		"21A", "21B", "21C",
		"22A", "22B", "22C", "22D",
		"23A", "23B", "23C", "23D", "23E", "23F", "23G",
		"24A", "24B", "24C",
		"25A", "25B", "25C",
		"26A", "26B", "26C", "26D",
		"27A", "27B", "27C", "27D", "27E", "27F", "27G",
	}

	var suggestions []string
	upperPartial := strings.ToUpper(partial)

	for _, code := range commonCodes {
		if strings.HasPrefix(code, upperPartial) {
			suggestions = append(suggestions, code)
		}
	}

	return suggestions
}
