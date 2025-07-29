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
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Redis (optional)
	var redisClient *redis.Client
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		redisClient = initRedis(redisURL)
		defer redisClient.Close()
	}

	// Load BJCP data
	bjcpData, err := data.LoadBJCPData()
	if err != nil {
		log.Fatalf("Failed to load BJCP data: %v", err)
	}

	// Initialize services
	beerService := services.NewBeerService()
	breweryService := services.NewBreweryService(db, redisClient)

	// Initialize handlers
	toolHandlers := handlers.NewToolHandlers(bjcpData, beerService, breweryService)
	resourceHandlers := handlers.NewResourceHandlers(bjcpData, beerService, breweryService)

	// Initialize MCP server
	mcpServer := mcp.NewServer(toolHandlers, resourceHandlers)

	// Run server based on mode
	switch *mode {
	case "websocket":
		runWebSocketServer(mcpServer, *port)
	case "stdio":
		runStdioServer(mcpServer)
	default:
		log.Fatalf("Unknown mode: %s. Use 'websocket' or 'stdio'", *mode)
	}
}

func initDatabase() (*sqlx.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL environment variable is required")
	}

	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Auto-migrate database schema
	if err := models.MigrateDatabase(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Seed database with initial data
	if err := models.SeedDatabase(db); err != nil {
		logrus.Warnf("Failed to seed database: %v", err)
		// Don't fail startup if seeding fails
	}

	logrus.Info("Database initialized successfully")
	return db, nil
}

func initRedis(redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		logrus.Warnf("Failed to parse Redis URL: %v", err)
		return nil
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logrus.Warnf("Failed to connect to Redis: %v", err)
		return nil
	}

	logrus.Info("Redis initialized successfully")
	return client
}

func runWebSocketServer(mcpServer *mcp.Server, port string) {
	// Create HTTP server
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "service": "brewsource-mcp", "version": "1.0.0"}`)
	})

	// Server info endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"name": "BrewSource MCP Server",
			"version": "1.0.0",
			"description": "Model Context Protocol server for brewing resources",
			"endpoints": {
				"mcp": "/mcp",
				"health": "/health"
			},
			"phase": "Phase 1 MVP"
		}`)
	})

	// MCP WebSocket endpoint
	mux.HandleFunc("/mcp", mcpServer.HandleWebSocket)

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logrus.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

func runStdioServer(mcpServer *mcp.Server) {
	logrus.Info("Starting stdio server")

	if err := mcpServer.HandleStdio(); err != nil {
		log.Fatalf("Stdio server error: %v", err)
	}
}
