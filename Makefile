# BrewSource MCP Server Makefile

.PHONY: help setup up down clean k9s build test format lint lint-fix security

# Default target
help:
	@echo "🍺 BrewSource MCP Server - Development Commands"
	@echo "=============================================="
	@echo ""
	@echo "Main commands:"
	@echo "  setup             Install required tools (Kind, Tilt, k9s)"
	@echo "  up                Start development environment (Kind + Tilt)"
	@echo "  down              Stop development environment"
	@echo "  clean             Clean up everything (delete cluster)"
	@echo "  k9s               Explore cluster with k9s (in brewsource-dev namespace)"
	@echo ""
	@echo "Development:"
	@echo "  build             Build the application"
	@echo "  test              Run tests"
	@echo "  format            Format code"
	@echo "  lint              Run linter"
	@echo "  lint-fix          Run linter with auto-fix"
	@echo "  security          Run security scans"
	@echo ""

# Kubernetes Development Environment

# Install required tools (Kind, Tilt, k9s) for development
setup:
	@echo "🛠️  Installing development tools..."
	@echo ""
	@if command -v brew >/dev/null 2>&1; then \
		echo "📦 Installing via Homebrew..."; \
		brew install kind tilt k9s; \
	elif command -v curl >/dev/null 2>&1; then \
		echo "📦 Installing via direct download..."; \
		echo "Installing Kind..."; \
		curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64 && \
		chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind; \
		echo "Installing Tilt..."; \
		curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash; \
		echo "Installing k9s..."; \
		curl -sS https://webinstall.dev/k9s | bash; \
	else \
		echo "❌ Please install curl or Homebrew to continue"; \
		exit 1; \
	fi
	@echo ""
	@echo "✅ Development tools installed!"
	@echo "💡 You can now run 'make up' to start the development environment"
	@echo ""

# Main development command - the "git clone && make up" experience
up:
	@echo "🚀 Starting BrewSource MCP development environment..."
	@echo "Building the binary"
	@make build
	@echo "Checking Kind cluster..."
	@if kind get clusters | grep -q "brewsource-dev"; then \
		echo "✅ Cluster 'brewsource-dev' already exists"; \
	else \
		echo "Creating Kind cluster..."; \
		if kind create cluster --name brewsource-dev; then \
			echo "✅ Cluster created successfully without config"; \
		elif kind create cluster --config kind-config.yaml --name brewsource-dev; then \
			echo "✅ Cluster created with custom config"; \
		else \
			echo "❌ Failed to create cluster. Trying with minimal config..."; \
			echo "kind: Cluster" > /tmp/minimal-kind-config.yaml; \
			echo "apiVersion: kind.x-k8s.io/v1alpha4" >> /tmp/minimal-kind-config.yaml; \
			echo "nodes:" >> /tmp/minimal-kind-config.yaml; \
			echo "- role: control-plane" >> /tmp/minimal-kind-config.yaml; \
			echo "  extraPortMappings:" >> /tmp/minimal-kind-config.yaml; \
			echo "  - containerPort: 8080" >> /tmp/minimal-kind-config.yaml; \
			echo "    hostPort: 8080" >> /tmp/minimal-kind-config.yaml; \
			kind create cluster --config /tmp/minimal-kind-config.yaml --name brewsource-dev; \
		fi; \
	fi
	@echo "Starting Tilt..."
	@tilt up --context kind-brewsource-dev
	@echo ""
	@echo "🍺 Development environment ready!"
	@echo "  📡 MCP Server:    http://localhost:8080"
	@echo "  🗄️  PostgreSQL:   localhost:5432"
	@echo "  🔴 Redis:         localhost:6379"
	@echo ""
	@echo "🛠️  Next steps:"
	@echo "  make k9s       - Explore cluster with k9s (opens in brewsource-dev namespace)"
	@echo "  make down      - Stop everything"
	@echo "  tilt ui        - Open Tilt dashboard"
	@echo ""

# Stop development environment
down:
	@echo "🛑 Stopping Tilt..."
	@tilt down || true
	@echo "🛑 Development environment stopped"
	@echo "Cluster still running. Use 'make clean' to delete the cluster."

# Clean up everything - delete Kind cluster and all resources
clean: down
	@echo "🗑️  Deleting Kind cluster..."
	@kind delete cluster --name brewsource-dev || true
	@echo "✅ Everything cleaned up"

# Open k9s for interactive cluster exploration and debugging
k9s:
	@echo "🔍 Opening k9s cluster explorer..."
	@echo "🎯 Starting in 'brewsource-dev' namespace where your services are running"
	@k9s --context kind-brewsource-dev --namespace brewsource-dev

# Application development

# Build the Go application binary
build:
	@echo "🔨 Building application..."
	@cd app && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/brewsource-mcp cmd/server/main.go
	@echo "✅ Build complete: app/bin/brewsource-mcp"

# Run all unit tests for the application
test:
	@echo "🧪 Running tests..."
	@go test -coverprofile=coverage.out ./app/...
	@echo ""
	@echo "📊 Coverage summary:"
	@go tool cover -func=coverage.out
	@echo "✅ Tests complete"

# Format Go code using gofmt
format:
	@echo "🎨 Formatting code..."
	@cd app && go fmt ./...
	@echo "✅ Code formatted"

# Run golangci-lint for code quality checks
lint:
	@echo "🔍 Running linter..."
	@cd app && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run
	@echo "✅ Linting complete"

# Run linter with automatic fixes applied
lint-fix:
	@echo "🔧 Running linter with auto-fix..."
	@cd app && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --fix
	@echo "✅ Linting with auto-fix complete"

# Run security scans using gosec (Go Security Checker) and govulncheck (dependency scanner)
security:
	@echo "🔒 Running security scans..."
	@echo "Running gosec (Go Security Checker)..."
	@cd app && go run github.com/securego/gosec/v2/cmd/gosec@latest -fmt json -out ../gosec-report.json ./...
	@echo "✅ gosec scan complete (see gosec-report.json)"
	@echo "Running govulncheck (dependency vulnerability scanner)..."
	@cd app && go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	@echo "✅ Dependency vulnerability check complete"
