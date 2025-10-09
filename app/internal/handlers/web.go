// Package handlers provides web handlers for serving HTML pages.
package handlers

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"regexp"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/redis/go-redis/v9"
	"github.com/russross/blackfriday/v2"
)

//go:embed templates/*
var templateFS embed.FS

// WebHandlers provides HTTP handlers for web pages.
type WebHandlers struct {
	templates   *template.Template
	db          interface{}
	redisClient interface{}
}

// NewWebHandlers creates a new instance of WebHandlers.
func NewWebHandlers(db interface{}, redisClient interface{}) *WebHandlers {
	templates := template.Must(template.ParseFS(templateFS, "templates/*.html"))
	return &WebHandlers{
		templates:   templates,
		db:          db,
		redisClient: redisClient,
	}
}

// LandingPageData represents the data passed to the landing page template.
type LandingPageData struct {
	ProjectName string
	Version     string
	Description string
	Host        string
	ReadmeHTML  template.HTML
	LastUpdated string
	Healthy     bool
}

// ServeHome handles the root path and serves the landing page with README content.
func (w *WebHandlers) ServeHome(writer http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(writer, r)
		return
	}

	// Read README.md file
	readmeContent, err := w.readReadmeFile()
	if err != nil {
		http.Error(writer, "Could not load README", http.StatusInternalServerError)
		return
	}

	// Convert markdown to HTML
	htmlContent := blackfriday.Run(readmeContent)

	// Post-process links: make relative links absolute to GitHub and set target="_blank"
	processedHTML := updateReadmeLinks(string(htmlContent))

	// Sanitize HTML to prevent XSS
	sanitizedHTML := bluemonday.UGCPolicy().Sanitize(processedHTML)

	health := isDBHealthy(w.db) && isRedisHealthy(w.redisClient)

	data := LandingPageData{
		ProjectName: "BrewSource MCP Server",
		Version:     getVersion(),
		Description: "A comprehensive Model Context Protocol (MCP) server for brewing resources, built with Go.",
		Host:        r.Host,
		ReadmeHTML:  template.HTML(sanitizedHTML), // #nosec G203
		LastUpdated: time.Now().Format("January 2, 2006"),
		Healthy:     health,
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if execErr := w.templates.ExecuteTemplate(writer, "landing.html", data); execErr != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// isDBHealthy checks if the database connection is healthy.
func isDBHealthy(db interface{}) bool {
	if db == nil {
		return true
	}
	type pinger interface {
		Ping() error
	}
	if p, ok := db.(pinger); ok {
		err := p.Ping()
		return err == nil
	}
	return false
}

const redisHealthTimeout = 2 * time.Second

// isRedisHealthy checks if the Redis client is healthy.
func isRedisHealthy(redisClient interface{}) bool {
	if redisClient == nil {
		return true
	}
	if client, ok := redisClient.(*redis.Client); ok {
		ctx, cancel := context.WithTimeout(context.Background(), redisHealthTimeout)
		defer cancel()
		err := client.Ping(ctx).Err()
		return err == nil
	}
	return false
}

// ServeStatic serves static assets (favicon, SVGs, etc.) from the embedded templates directory at /static/.
func (w *WebHandlers) ServeStatic(writer http.ResponseWriter, r *http.Request) {
	staticFS, err := fs.Sub(templateFS, "templates")
	if err != nil {
		http.Error(writer, "Static assets not found", http.StatusInternalServerError)
		return
	}
	http.FileServer(http.FS(staticFS)).ServeHTTP(writer, r)
}

// updateReadmeLinks updates relative links to absolute GitHub URLs and sets target="_blank" for all links.
func updateReadmeLinks(html string) string {
	repoURL := "https://github.com/CharlRitter/brewsource-mcp/tree/main"
	// Replace hrefs that start with ./ or ../
	reDot := regexp.MustCompile(`<a ([^>]*?)href=["'](\.?\.?/[^"'>]+)["']([^>]*)>`)
	html = reDot.ReplaceAllString(html, `<a $1href="`+repoURL+`/$2"$3 target="_blank">`)

	// Replace hrefs that are relative (do not start with http, https, mailto, #, /, ./, ../)
	reRel := regexp.MustCompile(`<a ([^>]*?)href=["']([^"'>]+)["']([^>]*)>`)
	html = reRel.ReplaceAllStringFunc(html, func(m string) string {
		// Extract the href value
		reHref := regexp.MustCompile(`href=["']([^"'>]+)["']`)
		match := reHref.FindStringSubmatch(m)
		if len(match) > 1 {
			href := match[1]
			if !(regexp.MustCompile(`^(https?://|mailto:|#|/|\./|\.\./)`).MatchString(href)) {
				// Replace with absolute GitHub URL
				m = reHref.ReplaceAllString(m, `href="`+repoURL+`/`+href+`"`)
			}
		}
		// Always add target="_blank" if not present
		if !regexp.MustCompile(`target=`).MatchString(m) {
			m = m[:len(m)-1] + ` target="_blank">`
		}
		return m
	})

	// For all other <a> tags, add target="_blank" if not present
	reA := regexp.MustCompile(`<a ([^>]*?)href=["']([^"'>]+)["']([^>]*)>`)
	html = reA.ReplaceAllStringFunc(html, func(m string) string {
		if !regexp.MustCompile(`target=`).MatchString(m) {
			m = m[:len(m)-1] + ` target="_blank">`
		}
		return m
	})

	return html
}

// ServeAPI handles the /api path and provides API information.
func (w *WebHandlers) ServeAPI(writer http.ResponseWriter, r *http.Request) {
	// Check if this is an HTMX request
	if r.Header.Get("Hx-Request") == "true" {
		// Return HTML fragment for HTMX
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusOK)

		data := struct {
			Version string
			Host    string
		}{
			Version: getVersion(),
			Host:    r.Host,
		}

		if err := w.templates.ExecuteTemplate(writer, "api-fragment.html", data); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Original JSON response for non-HTMX requests
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"name":        "BrewSource MCP Server",
		"version":     getVersion(),
		"description": "Model Context Protocol server for brewing resources",
		"endpoints": map[string]string{
			"mcp":    "/mcp",
			"health": "/health",
			"api":    "/api",
		},
		"phase": "Phase 1 MVP",
		"tools": []string{
			"bjcp_lookup",
			"search_beers",
			"find_breweries",
		},
		"resources": []string{
			"bjcp://styles",
			"bjcp://categories",
			"beers://catalog",
			"breweries://directory",
		},
		"connection": map[string]interface{}{
			"websocket":           "ws://" + r.Host + "/mcp",
			"supported_protocols": []string{"websocket", "stdio"},
		},
	}
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	_, _ = writer.Write(jsonBytes)
}

// ServeHealth handles the /health endpoint for both JSON and HTMX requests.
func (w *WebHandlers) ServeHealth(writer http.ResponseWriter, r *http.Request) {
	// Check if this is an HTMX request
	if r.Header.Get("Hx-Request") == "true" {
		// Return HTML fragment for HTMX
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusOK)

		// Check if this is for the status indicator specifically
		if r.URL.Query().Get("target") == "status" {
			if err := w.templates.ExecuteTemplate(writer, "status-indicator.html", nil); err != nil {
				http.Error(writer, "Internal server error", http.StatusInternalServerError)
				return
			}
			return
		}

		// Full health details
		data := struct {
			Version string
		}{
			Version: getVersion(),
		}

		if err := w.templates.ExecuteTemplate(writer, "health-fragment.html", data); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Original JSON response for non-HTMX requests
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "brewsource-mcp",
		"version": getVersion(),
	}
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	_, _ = writer.Write(jsonBytes)
}

// readReadmeFile reads the README.md file from the project root.
func (w *WebHandlers) readReadmeFile() ([]byte, error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://raw.githubusercontent.com/CharlRitter/brewsource-mcp/refs/heads/main/README.md",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request for README.md file: %w", err)
	}
	client := &http.Client{}
	resp, reqErr := client.Do(req)
	if reqErr == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			content, readmeReadErr := io.ReadAll(resp.Body)
			if readmeReadErr == nil {
				return content, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to read README.md file: %w", err)
}
