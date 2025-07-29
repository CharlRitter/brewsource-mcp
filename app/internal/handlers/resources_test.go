package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
	"github.com/CharlRitter/brewsource-mcp/app/pkg/data"
)

type stubBreweryService struct {
	services.BreweryService
}

func (s *stubBreweryService) SearchBreweries(ctx context.Context, query services.BrewerySearchQuery) ([]*services.BrewerySearchResult, error) {
	return []*services.BrewerySearchResult{{ID: 1, Name: "Stone Brewing", City: "Escondido", State: "CA", Country: "USA"}}, nil
}

func newTestHandlers() *ResourceHandlers {
	bjcpData := &data.BJCPData{
		Styles: map[string]data.BJCPStyle{
			"21A": {Code: "21A", Name: "IPA", Category: "IPA"},
		},
		Categories: []string{"IPA", "Lager"},
		Metadata:   data.Metadata{Version: "2021"},
	}
	beerService := services.NewBeerService()
	breweryService := &services.BreweryService{}
	return NewResourceHandlers(bjcpData, beerService, breweryService)
}

func TestHandleBJCPResource_Styles(t *testing.T) {
	h := newTestHandlers()
	res, err := h.HandleBJCPResource(context.Background(), "bjcp://styles")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.URI != "bjcp://styles" || res.MimeType != "application/json" {
		t.Errorf("unexpected resource content: %+v", res)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Errorf("invalid JSON in resource text: %v", err)
	}
	// Check that version/metadata is present
	if _, ok := parsed["version"]; !ok {
		t.Error("expected version in BJCP styles response")
	}
	// Check that categories is a slice
	if cats, ok := parsed["categories"].([]interface{}); !ok || len(cats) == 0 {
		t.Error("expected non-empty categories in BJCP styles response")
	}
}

func TestHandleBJCPResource_Categories(t *testing.T) {
	h := newTestHandlers()
	res, err := h.HandleBJCPResource(context.Background(), "bjcp://categories")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.URI != "bjcp://categories" {
		t.Errorf("unexpected URI: %s", res.URI)
	}
}

func TestHandleBJCPResource_StyleDetail(t *testing.T) {
	h := newTestHandlers()
	res, err := h.HandleBJCPResource(context.Background(), "bjcp://styles/21A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.URI != "bjcp://styles/21A" {
		t.Errorf("unexpected URI: %s", res.URI)
	}
	// Not found
	_, err = h.HandleBJCPResource(context.Background(), "bjcp://styles/99Z")
	if err == nil {
		t.Error("expected error for unknown style code")
	}
}

func TestHandleBJCPResource_InvalidURI(t *testing.T) {
	h := newTestHandlers()
	_, err := h.HandleBJCPResource(context.Background(), "bjcp://unknown")
	if err == nil {
		t.Error("expected error for invalid URI")
	}
}

func TestHandleBeerResource_Catalog(t *testing.T) {
	h := newTestHandlers()
	res, err := h.HandleBeerResource(context.Background(), "beers://catalog")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.URI != "beers://catalog" {
		t.Errorf("unexpected URI: %s", res.URI)
	}
	if res.MimeType != "application/json" {
		t.Errorf("unexpected MIME type: %s", res.MimeType)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Errorf("invalid JSON in beer catalog: %v", err)
	}
	if _, ok := parsed["sample_beers"]; !ok {
		t.Error("expected sample_beers in beer catalog response")
	}
	// Not found
	_, err = h.HandleBeerResource(context.Background(), "beers://unknown")
	if err == nil {
		t.Error("expected error for invalid beer resource URI")
	}
}

func TestHandleBreweryResource_Directory(t *testing.T) {
	h := newTestHandlers()
	res, err := h.HandleBreweryResource(context.Background(), "breweries://directory")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.URI != "breweries://directory" {
		t.Errorf("unexpected URI: %s", res.URI)
	}
	if res.MimeType != "application/json" {
		t.Errorf("unexpected MIME type: %s", res.MimeType)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Errorf("invalid JSON in brewery directory: %v", err)
	}
	if _, ok := parsed["sample_breweries"]; !ok {
		t.Error("expected sample_breweries in brewery directory response")
	}
	// Not found
	_, err = h.HandleBreweryResource(context.Background(), "breweries://unknown")
	if err == nil {
		t.Error("expected error for invalid brewery resource URI")
	}
}

func TestGetResourceDefinitions(t *testing.T) {
	h := newTestHandlers()
	defs := h.GetResourceDefinitions()
	if len(defs) == 0 {
		t.Fatal("expected resource definitions, got none")
	}
	uris := map[string]bool{}
	for _, def := range defs {
		uris[def.URI] = true
		if def.MimeType == "" {
			t.Errorf("resource %s missing MIME type", def.URI)
		}
	}
	// Check for required URIs
	for _, want := range []string{"bjcp://styles", "bjcp://styles/{code}", "bjcp://categories", "beers://catalog", "breweries://directory"} {
		if !uris[want] {
			t.Errorf("expected resource definition for %s", want)
		}
	}
}

func TestHandleBJCPResource_EmptyCategories(t *testing.T) {
	bjcpData := &data.BJCPData{
		Styles:     map[string]data.BJCPStyle{},
		Categories: []string{},
		Metadata:   data.Metadata{Version: "2021"},
	}
	beerService := services.NewBeerService()
	breweryService := &services.BreweryService{}
	h := NewResourceHandlers(bjcpData, beerService, breweryService)
	res, err := h.HandleBJCPResource(context.Background(), "bjcp://styles")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Errorf("invalid JSON in resource text: %v", err)
	}
	if cats, ok := parsed["categories"].([]interface{}); ok && len(cats) != 0 {
		t.Errorf("expected empty categories, got %v", cats)
	}
	if total, ok := parsed["total_styles"].(float64); !ok || total != 0 {
		t.Errorf("expected total_styles 0, got %v", parsed["total_styles"])
	}
}
