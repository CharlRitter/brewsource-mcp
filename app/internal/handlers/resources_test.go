package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/CharlRitter/brewsource-mcp/app/internal/handlers"
	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
	"github.com/CharlRitter/brewsource-mcp/app/pkg/data"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// Test RegisterResourceHandlers function
func TestRegisterResourceHandlers(t *testing.T) {
	bjcpData := &data.BJCPData{
		Styles: map[string]data.BJCPStyle{
			"21A": {Code: "21A", Name: "American IPA", Category: "IPA"},
		},
		Categories: []string{"IPA"},
		Metadata:   data.Metadata{Version: "2021", Source: "test"},
	}

	resourceHandlers := handlers.NewResourceHandlers(bjcpData, nil, nil)
	server := mcp.NewServer(nil, resourceHandlers)

	// Call RegisterResourceHandlers
	resourceHandlers.RegisterResourceHandlers(server)

	// Test that the handlers were registered by attempting to use them
	ctx := context.Background()

	// Test BJCP resource
	bjcpRequest := mcp.ReadResourceRequest{URI: "bjcp://styles/21A"}
	bjcpMsg := mcp.NewMessage("resources/read", bjcpRequest)
	bjcpMsgData, err := json.Marshal(bjcpMsg)
	if err != nil {
		t.Fatalf("Failed to marshal BJCP request: %v", err)
	}

	bjcpResponse := server.ProcessMessage(ctx, bjcpMsgData)
	if bjcpResponse.Error != nil {
		t.Errorf("Expected successful BJCP resource read, got error: %v", bjcpResponse.Error)
	}
}

func newTestHandlersWithDB(db *sqlx.DB) *handlers.ResourceHandlers {
	bjcpData := &data.BJCPData{
		Styles: map[string]data.BJCPStyle{
			"21A": {
				Code:     "21A",
				Name:     "American IPA",
				Category: "IPA",
				Vitals: data.Vitals{
					ABVMin: 5.5,
					ABVMax: 7.5,
					IBUMin: 40,
					IBUMax: 70,
					SRMMin: 6.0,
					SRMMax: 14.0,
					OGMin:  1.056,
					OGMax:  1.070,
					FGMin:  1.010,
					FGMax:  1.015,
				},
				OverallImpression:  "A decidedly hoppy and bitter, moderately strong American pale ale.",
				CommercialExamples: []string{"Bell's Two Hearted Ale", "Stone IPA"},
			},
		},
		Categories: []string{"IPA", "Lager"},
		Metadata:   data.Metadata{Version: "2021"},
	}
	redisClient := &redis.Client{}
	beerService := services.NewBeerService(db, redisClient)
	breweryService := services.NewBreweryService(db, redisClient)
	return handlers.NewResourceHandlers(bjcpData, beerService, breweryService)
} // For legacy tests that do not require DB
func newTestHandlers() *handlers.ResourceHandlers {
	bjcpData := &data.BJCPData{
		Styles: map[string]data.BJCPStyle{
			"21A": {
				Code:     "21A",
				Name:     "American IPA",
				Category: "IPA",
				Vitals: data.Vitals{
					ABVMin: 5.5,
					ABVMax: 7.5,
					IBUMin: 40,
					IBUMax: 70,
					SRMMin: 6.0,
					SRMMax: 14.0,
					OGMin:  1.056,
					OGMax:  1.070,
					FGMin:  1.010,
					FGMax:  1.015,
				},
				OverallImpression:  "A decidedly hoppy and bitter, moderately strong American pale ale.",
				CommercialExamples: []string{"Bell's Two Hearted Ale", "Stone IPA"},
			},
		},
		Categories: []string{"IPA", "Lager"},
		Metadata:   data.Metadata{Version: "2021"},
	}
	beerService := &services.BeerService{}       // Not used in these tests
	breweryService := &services.BreweryService{} // Use real type for compatibility
	return handlers.NewResourceHandlers(bjcpData, beerService, breweryService)
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
	if parseErr := json.Unmarshal([]byte(res.Text), &parsed); parseErr != nil {
		t.Errorf("invalid JSON in resource text: %v", parseErr)
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

// Helper functions for reducing cognitive complexity

func setupNormalCategoriesData() *data.BJCPData {
	return &data.BJCPData{
		Categories: []string{"IPA", "Lager", "Stout"},
		Metadata:   data.Metadata{Version: "2021"},
	}
}

func setupEmptyCategoriesData() *data.BJCPData {
	return &data.BJCPData{
		Categories: []string{},
		Metadata:   data.Metadata{Version: "2021"},
	}
}

func checkNormalCategoriesResult(t *testing.T, res *mcp.ResourceContent) {
	var parsed map[string]interface{}
	if parseErr := json.Unmarshal([]byte(res.Text), &parsed); parseErr != nil {
		t.Errorf("invalid JSON response: %v", parseErr)
	}
	categories, ok := parsed["categories"].([]interface{})
	if !ok {
		t.Error("expected categories to be an array")
		return
	}
	if len(categories) != 3 {
		t.Errorf("expected 3 categories, got %d", len(categories))
	}
	count, ok := parsed["count"].(float64)
	if !ok || count != 3 {
		t.Errorf("expected count of 3, got %v", parsed["count"])
	}
}

func checkEmptyCategoriesResult(t *testing.T, res *mcp.ResourceContent) {
	var parsed map[string]interface{}
	if parseErr := json.Unmarshal([]byte(res.Text), &parsed); parseErr != nil {
		t.Errorf("invalid JSON response: %v", parseErr)
	}
	categories, ok := parsed["categories"].([]interface{})
	if !ok {
		t.Error("expected categories to be an array")
		return
	}
	if len(categories) != 0 {
		t.Error("expected empty categories array")
	}
	count, ok := parsed["count"].(float64)
	if !ok || count != 0 {
		t.Errorf("expected count of 0, got %v", parsed["count"])
	}
}

func checkNoDuplicatesResult(t *testing.T, res *mcp.ResourceContent) {
	var parsed map[string]interface{}
	if parseErr := json.Unmarshal([]byte(res.Text), &parsed); parseErr != nil {
		t.Errorf("invalid JSON response: %v", parseErr)
	}
	categories, ok := parsed["categories"].([]interface{})
	if !ok {
		t.Error("expected categories to be an array")
		return
	}
	// Check for duplicates
	seen := make(map[string]bool)
	for _, cat := range categories {
		catStr, stringOk := cat.(string)
		if !stringOk {
			t.Error("category is not a string")
			continue
		}
		if seen[catStr] {
			t.Errorf("found duplicate category: %s", catStr)
		}
		seen[catStr] = true
	}
}

func TestHandleBJCPResource_Categories(t *testing.T) {
	tests := []struct {
		name        string
		setupData   func() *data.BJCPData
		checkResult func(t *testing.T, res *mcp.ResourceContent)
	}{
		{
			name:        "normal categories",
			setupData:   setupNormalCategoriesData,
			checkResult: checkNormalCategoriesResult,
		},
		{
			name:        "empty categories",
			setupData:   setupEmptyCategoriesData,
			checkResult: checkEmptyCategoriesResult,
		},
		{
			name:        "no duplicate categories",
			setupData:   setupNormalCategoriesData,
			checkResult: checkNoDuplicatesResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bjcpData := tt.setupData()
			redisClient := &redis.Client{}
			beerService := services.NewBeerService(nil, redisClient)
			breweryService := services.NewBreweryService(nil, redisClient)
			h := handlers.NewResourceHandlers(bjcpData, beerService, breweryService)

			res, err := h.HandleBJCPResource(context.Background(), "bjcp://categories")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if res.URI != "bjcp://categories" {
				t.Errorf("unexpected URI: %s", res.URI)
			}

			if res.MimeType != "application/json" {
				t.Errorf("unexpected MIME type: %s", res.MimeType)
			}

			tt.checkResult(t, res)
		})
	}
}

// Helper functions for TestHandleBJCPResource_StyleDetail.
func checkValidStyleResponse(t *testing.T, res *mcp.ResourceContent) {
	if res.URI != "bjcp://styles/21A" {
		t.Errorf("unexpected URI: %s", res.URI)
	}
	var style data.BJCPStyle
	if err := json.Unmarshal([]byte(res.Text), &style); err != nil {
		t.Errorf("invalid JSON response: %v", err)
	}
	if style.Code != "21A" || style.Name != "American IPA" {
		t.Errorf("expected American IPA (21A), got %s (%s)", style.Name, style.Code)
	}
	if style.Category != "IPA" {
		t.Errorf("expected category IPA, got %s", style.Category)
	}
	if style.OverallImpression == "" {
		t.Error("expected non-empty overall impression")
	}
	if len(style.CommercialExamples) == 0 {
		t.Error("expected commercial examples")
	}
}

func checkCaseSensitivityResponse(t *testing.T, res *mcp.ResourceContent) {
	var style data.BJCPStyle
	if err := json.Unmarshal([]byte(res.Text), &style); err != nil {
		t.Errorf("invalid JSON response: %v", err)
	}
	if style.Code != "21A" {
		t.Errorf("expected normalized style code 21A, got %s", style.Code)
	}
}

func getStyleDetailTestCases() []struct {
	name          string
	styleCode     string
	expectedError bool
	checkResponse func(t *testing.T, res *mcp.ResourceContent)
} {
	return []struct {
		name          string
		styleCode     string
		expectedError bool
		checkResponse func(t *testing.T, res *mcp.ResourceContent)
	}{
		{
			name:          "valid style code",
			styleCode:     "21A",
			expectedError: false,
			checkResponse: checkValidStyleResponse,
		},
		{
			name:          "unknown style code",
			styleCode:     "99Z",
			expectedError: true,
		},
		{
			name:          "empty style code",
			styleCode:     "",
			expectedError: true,
		},
		{
			name:          "invalid format - too long",
			styleCode:     "21AAA",
			expectedError: true,
		},
		{
			name:          "invalid format - no letter",
			styleCode:     "21",
			expectedError: true,
		},
		{
			name:          "invalid format - special characters",
			styleCode:     "2!A",
			expectedError: true,
		},
		{
			name:          "case sensitivity test",
			styleCode:     "21a",
			expectedError: false,
			checkResponse: checkCaseSensitivityResponse,
		},
		{
			name:          "whitespace in style code",
			styleCode:     "21 A",
			expectedError: true,
		},
		{
			name:          "leading/trailing whitespace",
			styleCode:     " 21A ",
			expectedError: true,
		},
		{
			name:          "unicode characters",
			styleCode:     "21Ａ",
			expectedError: true,
		},
	}
}

func TestHandleBJCPResource_StyleDetail(t *testing.T) {
	tests := getStyleDetailTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandlers()
			res, err := h.HandleBJCPResource(context.Background(), fmt.Sprintf("bjcp://styles/%s", tt.styleCode))

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, res)
			}
		})
	}
}

func TestInvalidResourceURIs(t *testing.T) {
	h := newTestHandlers()
	ctx := context.Background()

	tests := []struct {
		name     string
		uri      string
		testFunc func() error
	}{
		{
			name: "invalid BJCP resource path",
			uri:  "bjcp://unknown",
			testFunc: func() error {
				_, err := h.HandleBJCPResource(ctx, "bjcp://unknown")
				return err
			},
		},
		{
			name: "malformed BJCP style code",
			uri:  "bjcp://styles/invalid!code",
			testFunc: func() error {
				_, err := h.HandleBJCPResource(ctx, "bjcp://styles/invalid!code")
				return err
			},
		},
		{
			name: "invalid beer resource path",
			uri:  "beers://unknown",
			testFunc: func() error {
				_, err := h.HandleBeerResource(ctx, "beers://unknown")
				return err
			},
		},
		{
			name: "invalid brewery resource path",
			uri:  "breweries://unknown",
			testFunc: func() error {
				_, err := h.HandleBreweryResource(ctx, "breweries://unknown")
				return err
			},
		},
		{
			name: "completely invalid URI scheme",
			uri:  "invalid://something",
			testFunc: func() error {
				_, err := h.HandleBJCPResource(ctx, "invalid://something")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc()
			if err == nil {
				t.Errorf("expected error for invalid URI: %s", tt.uri)
			}
		})
	}
}

// Helper functions for TestHandleBeerResource_Catalog.
func setupSuccessfulBeerQuery(mock sqlmock.Sqlmock) {
	rows := sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"}).
		AddRow(1, "Stone IPA", "American IPA", "Stone Brewing", "USA", 6.9, 71).
		AddRow(2, "Pliny the Elder", "Double IPA", "Russian River", "USA", 8.0, 100).
		AddRow(3, "Sierra Nevada Pale Ale", "American Pale Ale", "Sierra Nevada", "USA", 5.6, 38)
	mock.ExpectQuery("SELECT (.+) FROM beers b").
		WillReturnRows(rows)
}

func setupBeerDatabaseError(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu FROM beers b").
		WithArgs("%IPA%", 10).
		WillReturnError(errors.New("database error"))
}

func setupEmptyBeerResults(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT b.id, b.name, b.style, br.name as brewery, br.country, b.abv, b.ibu FROM beers b").
		WithArgs("%IPA%", 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "style", "brewery", "country", "abv", "ibu"}))
}

func checkSuccessfulBeerResult(t *testing.T, res *mcp.ResourceContent) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Errorf("invalid JSON in beer catalog: %v", err)
	}
	beers, ok := parsed["sample_beers"].([]interface{})
	if !ok {
		t.Error("expected sample_beers array in response")
		return
	}
	if len(beers) != 3 {
		t.Errorf("expected 3 beers, got %d", len(beers))
	}

	// Check first beer details
	beer := beers[0].(map[string]interface{})
	if beer["name"] != "Stone IPA" {
		t.Errorf("expected Stone IPA, got %v", beer["name"])
	}
	if beer["abv"].(float64) != 6.9 {
		t.Errorf("expected ABV 6.9, got %v", beer["abv"])
	}
}

func checkEmptyBeerResult(t *testing.T, res *mcp.ResourceContent) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Errorf("invalid JSON in beer catalog: %v", err)
	}
	beers, ok := parsed["sample_beers"].([]interface{})
	if !ok {
		t.Error("expected sample_beers to be an array")
	}
	if len(beers) != 0 {
		t.Error("expected empty sample_beers array")
	}
}

func getBeerCatalogTestCases() []struct {
	name          string
	setupMock     func(mock sqlmock.Sqlmock)
	expectedError bool
	checkResult   func(t *testing.T, res *mcp.ResourceContent)
} {
	return []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		expectedError bool
		checkResult   func(t *testing.T, res *mcp.ResourceContent)
	}{
		{
			name:          "successful query with multiple results",
			setupMock:     setupSuccessfulBeerQuery,
			expectedError: false,
			checkResult:   checkSuccessfulBeerResult,
		},
		{
			name:          "database error",
			setupMock:     setupBeerDatabaseError,
			expectedError: true,
		},
		{
			name:          "empty result set",
			setupMock:     setupEmptyBeerResults,
			expectedError: false,
			checkResult:   checkEmptyBeerResult,
		},
	}
}

func runBeerCatalogTestCase(t *testing.T, tt struct {
	name          string
	setupMock     func(mock sqlmock.Sqlmock)
	expectedError bool
	checkResult   func(t *testing.T, res *mcp.ResourceContent)
},
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	sqlxdb := sqlx.NewDb(db, "sqlmock")
	tt.setupMock(mock)

	h := newTestHandlersWithDB(sqlxdb)
	res, err := h.HandleBeerResource(context.Background(), "beers://catalog")

	if tt.expectedError {
		if err == nil {
			t.Error("expected error but got none")
		}
		return
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validateBeerCatalogResponse(t, res)

	if tt.checkResult != nil {
		tt.checkResult(t, res)
	}

	if mockErr := mock.ExpectationsWereMet(); mockErr != nil {
		t.Errorf("unmet sqlmock expectations: %v", mockErr)
	}
}

func validateBeerCatalogResponse(t *testing.T, res *mcp.ResourceContent) {
	if res.URI != "beers://catalog" {
		t.Errorf("unexpected URI: %s", res.URI)
	}
	if res.MimeType != "application/json" {
		t.Errorf("unexpected MIME type: %s", res.MimeType)
	}
}

func TestHandleBeerResource_Catalog(t *testing.T) {
	tests := getBeerCatalogTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBeerCatalogTestCase(t, tt)
		})
	}

	// Test invalid resource URI
	t.Run("invalid resource URI", func(t *testing.T) {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer db.Close()

		h := newTestHandlersWithDB(sqlx.NewDb(db, "sqlmock"))
		_, err = h.HandleBeerResource(context.Background(), "beers://unknown")
		if err == nil {
			t.Error("expected error for invalid beer resource URI")
		}
	})
}

// Helper functions for TestHandleBreweryResource_Directory.
func setupSuccessfulBreweryQuery(mock sqlmock.Sqlmock) {
	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state",
		"postal_code", "country", "phone", "website_url",
	}).
		AddRow(
			1, "Stone Brewing", "regional", "1999 Citracado Pkwy",
			"Escondido", "CA", "92029", "USA", "760-294-7866",
			"http://www.stonebrewing.com",
		).
		AddRow(
			2, "Russian River", "brewpub", "725 4th Street",
			"Santa Rosa", "CA", "95404", "USA", "707-545-2337",
			"http://www.russianriverbrewing.com",
		)
	mock.ExpectQuery(`SELECT (.+) FROM breweries WHERE 1=1`).
		WillReturnRows(rows)
}

func setupSpecialCharactersBreweryQuery(mock sqlmock.Sqlmock) {
	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state",
		"postal_code", "country", "phone", "website_url",
	}).AddRow(
		1, "Bräu & Co.", "brewpub", "123 Main St",
		"Milwaukee", "WI", "53202", "USA", "414-555-1234",
		"http://www.brauandco.com",
	)
	mock.ExpectQuery(`SELECT (.+) FROM breweries WHERE 1=1`).
		WillReturnRows(rows)
}

func setupMinimalBreweryQuery(mock sqlmock.Sqlmock) {
	rows := sqlmock.NewRows([]string{
		"id", "name", "brewery_type", "street", "city", "state",
		"postal_code", "country", "phone", "website_url",
	}).AddRow(
		1, "Minimalist Brewing", "micro", "",
		"Portland", "OR", "", "USA", "",
		"",
	)
	mock.ExpectQuery(`SELECT (.+) FROM breweries WHERE 1=1`).
		WillReturnRows(rows)
}

func setupBreweryDatabaseError(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT (.+) FROM breweries WHERE 1=1`).
		WillReturnError(errors.New("database error"))
}

func setupEmptyBreweryResults(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT (.+) FROM breweries WHERE 1=1`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "brewery_type", "street", "city", "state",
			"postal_code", "country", "phone", "website_url",
		}))
}

func checkSuccessfulBreweryResult(t *testing.T, res *mcp.ResourceContent) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Errorf("invalid JSON in brewery directory: %v", err)
	}
	breweries, ok := parsed["sample_breweries"].([]interface{})
	if !ok {
		t.Error("expected sample_breweries to be an array")
		return
	}
	if len(breweries) != 2 {
		t.Errorf("expected 2 breweries, got %d", len(breweries))
	}

	// Check first brewery details
	brewery := breweries[0].(map[string]interface{})
	if brewery["name"] != "Stone Brewing" {
		t.Errorf("expected Stone Brewing, got %v", brewery["name"])
	}
	if brewery["brewery_type"] != "regional" {
		t.Errorf("expected regional type, got %v", brewery["brewery_type"])
	}
	if brewery["city"] != "Escondido" || brewery["state"] != "CA" {
		t.Errorf("unexpected location: %v, %v", brewery["city"], brewery["state"])
	}

	// Check second brewery details
	brewery = breweries[1].(map[string]interface{})
	if brewery["name"] != "Russian River" {
		t.Errorf("expected Russian River, got %v", brewery["name"])
	}
}

func checkSpecialCharactersBreweryResult(t *testing.T, res *mcp.ResourceContent) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Errorf("invalid JSON in brewery directory: %v", err)
	}
	breweries := parsed["sample_breweries"].([]interface{})
	brewery := breweries[0].(map[string]interface{})
	if brewery["name"] != "Bräu & Co." {
		t.Errorf("expected Bräu & Co., got %v", brewery["name"])
	}
}

func checkMinimalBreweryResult(t *testing.T, res *mcp.ResourceContent) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Errorf("invalid JSON in brewery directory: %v", err)
	}
	breweries := parsed["sample_breweries"].([]interface{})
	brewery := breweries[0].(map[string]interface{})
	if brewery["name"] != "Minimalist Brewing" {
		t.Errorf("expected Minimalist Brewing, got %v", brewery["name"])
	}
	// Check that optional fields are empty but present
	if brewery["street"] != "" {
		t.Error("expected empty street")
	}
	if brewery["phone"] != "" {
		t.Error("expected empty phone")
	}
	if brewery["website_url"] != "" {
		t.Error("expected empty website_url")
	}
}

func checkEmptyBreweryResult(t *testing.T, res *mcp.ResourceContent) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.Text), &parsed); err != nil {
		t.Errorf("invalid JSON in brewery directory: %v", err)
	}
	breweries, ok := parsed["sample_breweries"].([]interface{})
	if !ok {
		t.Error("expected sample_breweries to be an array")
	}
	if len(breweries) != 0 {
		t.Error("expected empty sample_breweries array")
	}
}

func getBreweryDirectoryTestCases() []struct {
	name          string
	setupMock     func(mock sqlmock.Sqlmock)
	uri           string
	expectedError bool
	checkResult   func(t *testing.T, res *mcp.ResourceContent)
} {
	return []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		uri           string
		expectedError bool
		checkResult   func(t *testing.T, res *mcp.ResourceContent)
	}{
		{
			name:          "successful query with multiple results",
			setupMock:     setupSuccessfulBreweryQuery,
			uri:           "breweries://directory",
			expectedError: false,
			checkResult:   checkSuccessfulBreweryResult,
		},
		{
			name:          "special characters in brewery name",
			setupMock:     setupSpecialCharactersBreweryQuery,
			uri:           "breweries://directory",
			expectedError: false,
			checkResult:   checkSpecialCharactersBreweryResult,
		},
		{
			name:          "missing optional fields",
			setupMock:     setupMinimalBreweryQuery,
			uri:           "breweries://directory",
			expectedError: false,
			checkResult:   checkMinimalBreweryResult,
		},
		{
			name:          "database error",
			setupMock:     setupBreweryDatabaseError,
			uri:           "breweries://directory",
			expectedError: true,
		},
		{
			name:          "empty result set",
			setupMock:     setupEmptyBreweryResults,
			uri:           "breweries://directory",
			expectedError: false,
			checkResult:   checkEmptyBreweryResult,
		},
		{
			name:          "invalid resource URI",
			uri:           "breweries://unknown",
			expectedError: true,
			setupMock:     func(_ sqlmock.Sqlmock) {},
		},
	}
}

func createBreweryTestHandlers(sqlxdb *sqlx.DB) *handlers.ResourceHandlers {
	bjcpData := &data.BJCPData{
		Styles:     map[string]data.BJCPStyle{"21A": {Code: "21A", Name: "IPA", Category: "IPA"}},
		Categories: []string{"IPA", "Lager"},
		Metadata:   data.Metadata{Version: "2021"},
	}
	redisClient := &redis.Client{}
	beerService := services.NewBeerService(sqlxdb, redisClient)
	breweryService := services.NewBreweryService(sqlxdb, redisClient)
	return handlers.NewResourceHandlers(bjcpData, beerService, breweryService)
}

func runBreweryDirectoryTestCase(t *testing.T, tt struct {
	name          string
	setupMock     func(mock sqlmock.Sqlmock)
	uri           string
	expectedError bool
	checkResult   func(t *testing.T, res *mcp.ResourceContent)
},
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	sqlxdb := sqlx.NewDb(db, "sqlmock")
	tt.setupMock(mock)

	h := createBreweryTestHandlers(sqlxdb)
	res, err := h.HandleBreweryResource(context.Background(), tt.uri)

	if tt.expectedError {
		if err == nil {
			t.Error("expected error but got none")
		}
		return
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validateBreweryDirectoryResponse(t, res, tt.uri)

	if tt.checkResult != nil {
		tt.checkResult(t, res)
	}

	if mockErr := mock.ExpectationsWereMet(); mockErr != nil {
		t.Errorf("unmet sqlmock expectations: %v", mockErr)
	}
}

func validateBreweryDirectoryResponse(t *testing.T, res *mcp.ResourceContent, expectedURI string) {
	if res.URI != expectedURI {
		t.Errorf("unexpected URI: %s", res.URI)
	}
	if res.MimeType != "application/json" {
		t.Errorf("unexpected MIME type: %s", res.MimeType)
	}
}

func TestHandleBreweryResource_Directory(t *testing.T) {
	tests := getBreweryDirectoryTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBreweryDirectoryTestCase(t, tt)
		})
	}
}

func TestResourceHandlersRegistry(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	h := newTestHandlersWithDB(sqlx.NewDb(db, "sqlmock"))

	// Test GetResourceDefinitions
	defs := h.GetResourceDefinitions()
	requiredURIs := []string{
		"bjcp://styles",
		"bjcp://styles/{code}",
		"bjcp://categories",
		"beers://catalog",
		"breweries://directory",
	}

	for _, uri := range requiredURIs {
		found := false
		for _, def := range defs {
			if def.URI == uri {
				found = true
				if def.MimeType == "" {
					t.Errorf("resource %s missing MIME type", uri)
				}
				if def.Description == "" {
					t.Errorf("resource %s missing description", uri)
				}
				break
			}
		}
		if !found {
			t.Errorf("required resource URI %s not found in definitions", uri)
		}
	}
}

func TestGetResourceDefinitions(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxdb := sqlx.NewDb(db, "sqlmock")
	h := newTestHandlersWithDB(sqlxdb)
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
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxdb := sqlx.NewDb(db, "sqlmock")
	bjcpData := &data.BJCPData{
		Styles:     map[string]data.BJCPStyle{},
		Categories: []string{},
		Metadata:   data.Metadata{Version: "2021"},
	}
	redisClient := &redis.Client{} // Use a mock or test Redis client
	beerService := services.NewBeerService(sqlxdb, redisClient)
	breweryService := &services.BreweryService{} // Use real type for compatibility
	h := handlers.NewResourceHandlers(bjcpData, beerService, breweryService)
	res, err := h.HandleBJCPResource(context.Background(), "bjcp://styles")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if parseErr := json.Unmarshal([]byte(res.Text), &parsed); parseErr != nil {
		t.Errorf("invalid JSON in resource text: %v", parseErr)
	}
	if cats, ok := parsed["categories"].([]interface{}); ok && len(cats) != 0 {
		t.Errorf("expected empty categories, got %v", cats)
	}
	if total, ok := parsed["total_styles"].(float64); !ok || total != 0 {
		t.Errorf("expected total_styles 0, got %v", parsed["total_styles"])
	}
}
