package data

import (
	"reflect"
	"testing"
)

func mockBJCPData() *BJCPData {
	return &BJCPData{
		Styles: map[string]BJCPStyle{
			"21A": {
				Code:               "21A",
				Name:               "American IPA",
				Category:           "IPA",
				CommercialExamples: []string{"Stone IPA"},
				Vitals:             Vitals{ABVMin: 5.5, ABVMax: 7.5, IBUMin: 40, IBUMax: 70, SRMMin: 6, SRMMax: 14, OGMin: 1.056, OGMax: 1.070, FGMin: 1.008, FGMax: 1.014},
			},
			"1A": {
				Code:               "1A",
				Name:               "American Light Lager",
				Category:           "Standard American Beer",
				CommercialExamples: []string{"Bud Light"},
				Vitals:             Vitals{ABVMin: 2.8, ABVMax: 4.2, IBUMin: 8, IBUMax: 12, SRMMin: 2, SRMMax: 3, OGMin: 1.028, OGMax: 1.040, FGMin: 0.998, FGMax: 1.008},
			},
		},
		Categories: []string{"IPA", "Standard American Beer"},
		Metadata: Metadata{
			Version:     "2021",
			Source:      "test",
			LastUpdated: "2025-07-26",
			TotalStyles: 2,
		},
	}
}

func TestGetStyleByCode(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())
	style, err := svc.GetStyleByCode("21A")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if style == nil || style.Name != "American IPA" {
		t.Errorf("expected American IPA, got %+v", style)
	}
	// Test case insensitivity
	style, err = svc.GetStyleByCode("21a")
	if err != nil || style.Name != "American IPA" {
		t.Errorf("case-insensitive code failed: %v, %+v", err, style)
	}
	// Test not found
	_, err = svc.GetStyleByCode("99Z")
	if err == nil {
		t.Error("expected error for unknown code")
	}
}

func TestGetStyleByName(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())
	// Exact match
	style, err := svc.GetStyleByName("American IPA")
	if err != nil || style.Code != "21A" {
		t.Errorf("expected 21A, got %v, %v", style, err)
	}
	// Case-insensitive exact
	style, err = svc.GetStyleByName("american ipa")
	if err != nil || style.Code != "21A" {
		t.Errorf("case-insensitive name failed: %v, %+v", err, style)
	}
	// Partial match (starts with)
	style, err = svc.GetStyleByName("American L")
	if err != nil || style.Code != "1A" {
		t.Errorf("partial match failed: %v, %+v", err, style)
	}
	// Contains match
	style, err = svc.GetStyleByName("Light")
	if err != nil || style.Code != "1A" {
		t.Errorf("contains match failed: %v, %+v", err, style)
	}
	// Not found
	_, err = svc.GetStyleByName("Nonexistent")
	if err == nil {
		t.Error("expected error for unknown name")
	}
}

func TestGetAllStyles(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())
	styles := svc.GetAllStyles()
	if len(styles) != 2 {
		t.Errorf("expected 2 styles, got %d", len(styles))
	}
}

func TestGetCategories(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())
	cats := svc.GetCategories()
	expected := []string{"IPA", "Standard American Beer"}
	if !reflect.DeepEqual(cats, expected) {
		t.Errorf("expected %v, got %v", expected, cats)
	}
}

func TestGetStylesByCategory(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())
	styles := svc.GetStylesByCategory("IPA")
	if len(styles) != 1 || styles[0].Code != "21A" {
		t.Errorf("expected 1 style 21A, got %+v", styles)
	}
	// Case-insensitive
	styles = svc.GetStylesByCategory("standard american beer")
	if len(styles) != 1 || styles[0].Code != "1A" {
		t.Errorf("expected 1 style 1A, got %+v", styles)
	}
	// No match
	styles = svc.GetStylesByCategory("Nonexistent")
	if len(styles) != 0 {
		t.Errorf("expected 0 styles, got %+v", styles)
	}
}

func TestGetMetadata(t *testing.T) {
	svc := NewBJCPServiceFromData(mockBJCPData())
	meta := svc.GetMetadata()
	if meta.Version != "2021" || meta.TotalStyles != 2 {
		t.Errorf("unexpected metadata: %+v", meta)
	}
}
