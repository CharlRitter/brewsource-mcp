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
	port := flag.String("port", "8080", "Port for HTTP server")
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

	// Run server
	RunHTTPServer(mcpServer, webHandlers, *port)
	cleanup()
}

// RunHTTPServer starts the HTTP server for MCP connections over HTTP POST.
func RunHTTPServer(mcpServer *mcp.Server, webHandlers *handlers.WebHandlers, port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", webHandlers.ServeHome)
	mux.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(webHandlers.ServeStatic)))
	mux.HandleFunc("/api", webHandlers.ServeAPI)
	mux.HandleFunc("/health", webHandlers.ServeHealth)
	mux.HandleFunc("/version", webHandlers.ServeVersion)
	mux.HandleFunc("/mcp", mcpServer.HandleHTTP)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logrus.Info("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logrus.Errorf("HTTP server shutdown error: %v", err)
		}
	}()

	logrus.Infof("Starting HTTP MCP server on port %s", port)
	logrus.Infof("MCP endpoint: http://localhost:%s/mcp", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed to start: %v", err)
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
