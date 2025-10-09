// Package main starts the Brewsource MCP server.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CharlRitter/brewsource-mcp/app/internal/handlers"
	"github.com/CharlRitter/brewsource-mcp/app/internal/mcp"
	"github.com/CharlRitter/brewsource-mcp/app/internal/models"
	"github.com/CharlRitter/brewsource-mcp/app/internal/services"
	"github.com/CharlRitter/brewsource-mcp/app/pkg/data"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// Database connection pool settings.
	maxOpenConns    = 25
	maxIdleConns    = 5
	connMaxLifetime = 5 * time.Minute

	// Network timeout settings.
	redisTimeout    = 5 * time.Second
	readTimeout     = 30 * time.Second
	writeTimeout    = 30 * time.Second
	idleTimeout     = 120 * time.Second
	shutdownTimeout = 30 * time.Second
)

func main() {
	// Command line flags
	var (
		mode = flag.String("mode", "websocket", "Server mode: websocket or stdio")
		port = flag.String("port", "8080", "Port for WebSocket server")
	)
	flag.Parse()

	// Initialize logger
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if os.Getenv("LOG_LEVEL") == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Initialize database
	db, err := InitDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis (optional)
	var redisClient *redis.Client
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		redisClient = InitRedis(redisURL)
	}

	// Cleanup function for early exits and normal execution
	cleanup := func() {
		if redisClient != nil {
			if closeErr := redisClient.Close(); closeErr != nil {
				logrus.Warnf("Failed to close Redis client: %v", closeErr)
			}
		}
		if closeErr := db.Close(); closeErr != nil {
			logrus.Warnf("Failed to close database: %v", closeErr)
		}
	}

	// Load BJCP data
	bjcpData, err := data.LoadBJCPData()
	if err != nil {
		cleanup()
		log.Fatalf("Failed to load BJCP data: %v", err)
	}

	// Initialize services
	beerService := services.NewBeerService(db, redisClient)
	breweryService := services.NewBreweryService(db, redisClient)

	// Initialize handlers
	toolHandlers := handlers.NewToolHandlers(bjcpData, beerService, breweryService)
	resourceHandlers := handlers.NewResourceHandlers(bjcpData, beerService, breweryService)
	webHandlers := handlers.NewWebHandlers(db, redisClient)

	// Initialize MCP server
	mcpServer := mcp.NewServer(toolHandlers, resourceHandlers)

	// Register /version resource handler
	mcpServer.RegisterResourceHandler("/version", handlers.VersionResourceHandler())

	// Register /health resource handler
	mcpServer.RegisterResourceHandler("/health", handlers.HealthResourceHandler())

	// Run server based on mode
	var shouldDefer bool
	switch *mode {
	case "websocket":
		shouldDefer = true
		RunWebSocketServer(mcpServer, webHandlers, *port)
	case "stdio":
		shouldDefer = true
		RunStdioServer(mcpServer)
	default:
		cleanup()
		log.Fatalf("Unknown mode: %s. Use 'websocket' or 'stdio'", *mode)
	}

	// Cleanup for normal execution paths
	if shouldDefer {
		cleanup()
	}
}

// InitDatabase initializes and configures the PostgreSQL database connection.
func InitDatabase() (*sqlx.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL environment variable is required")
	}

	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Auto-migrate database schema
	if migrationErr := models.MigrateDatabase(db); migrationErr != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", migrationErr)
	}

	// Seed database with initial data
	if seedErr := models.SeedDatabase(db); seedErr != nil {
		logrus.Warnf("Failed to seed database: %v", seedErr)
		// Don't fail startup if seeding fails
	}

	logrus.Info("Database initialized successfully")
	return db, nil
}

// InitRedis initializes and configures the Redis client connection.
func InitRedis(redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		logrus.Warnf("Failed to parse Redis URL: %v", err)
		return nil
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	if pingErr := client.Ping(ctx).Err(); pingErr != nil {
		logrus.Warnf("Failed to connect to Redis: %v", pingErr)
		return nil
	}

	logrus.Info("Redis initialized successfully")
	return client
}

// RunWebSocketServer starts the HTTP server with WebSocket support for MCP connections.
func RunWebSocketServer(mcpServer *mcp.Server, webHandlers *handlers.WebHandlers, port string) {
	// Create HTTP server
	mux := http.NewServeMux()

	// Landing page (serves HTML with README content)
	mux.HandleFunc("/", webHandlers.ServeHome)

	// Static assets handler (favicon, SVGs, etc.)
	mux.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(webHandlers.ServeStatic)))

	// API information endpoint (serves JSON)
	mux.HandleFunc("/api", webHandlers.ServeAPI)

	// Health check endpoint (supports both JSON and HTMX)
	mux.HandleFunc("/health", webHandlers.ServeHealth)

	// MCP WebSocket endpoint
	mux.HandleFunc("/mcp", mcpServer.HandleWebSocket)

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logrus.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logrus.Errorf("Server shutdown error: %v", err)
		}
	}()

	logrus.Infof("Starting WebSocket server on port %s", port)
	logrus.Infof("MCP endpoint: ws://localhost:%s/mcp", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// RunStdioServer starts the stdio-based MCP server for direct communication.
func RunStdioServer(mcpServer *mcp.Server) {
	logrus.Info("Starting stdio server")

	if err := mcpServer.HandleStdio(); err != nil {
		log.Fatalf("Stdio server error: %v", err)
	}
}
