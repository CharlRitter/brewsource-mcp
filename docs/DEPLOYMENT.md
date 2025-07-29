# BrewSource MCP Server Deployment Guide

This guide provides comprehensive instructions for deploying the BrewSource MCP Server using both Docker and Kubernetes workflows. It covers local development, production deployment, CI/CD integration, security, and troubleshooting.

---

## Table of Contents

1. [Overview](#overview)
2. [Deployment Options](#deployment-options)
   - [Docker Workflow](#docker-workflow)
   - [Kubernetes Workflow](#kubernetes-workflow)
3. [CI/CD Integration](#cicd-integration)
4. [Security & Optimization](#security--optimization)
5. [Testing & Troubleshooting](#testing--troubleshooting)
6. [Environment Variables & Configuration](#environment-variables--configuration)
7. [Migration & Advanced Topics](#migration--advanced-topics)
8. [Benefits & Next Steps](#benefits--next-steps)

---

## Overview

BrewSource MCP Server is an open-source Model Context Protocol (MCP) server providing AI assistants with comprehensive brewing knowledge and tools. This guide helps you set up, develop, and deploy the server using modern containerization and orchestration tools.

---

## Deployment Options

### Docker Workflow

#### Quick Start

```bash
# Start complete environment
make docker-compose-up

# View logs
make docker-compose-logs

# Stop everything
make docker-compose-down
```

#### Development

```bash
# Development environment with debug logging
make docker-dev

# Run tests in Docker
make docker-test

# Integration testing
make docker-integration-test
```

#### Production Deployment

```bash
# Build production image
make docker-build

# Run with environment variables
docker run -d --name brewsource-mcp \
  -p 8080:8080 \
  -e DATABASE_URL=your_postgres_url \
  -e REDIS_URL=your_redis_url \
  brewsource-mcp:latest
```

#### Services Available

- **BrewSource MCP Server**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

#### Image Optimization

- Multi-stage build (final image ~27.2MB)
- Static compilation, no runtime dependencies
- Layer caching for CI/CD
- Security scanning ready

---

### Kubernetes Workflow

#### Quick Start

```bash
# Start all services (Kind cluster + Tilt + all services)
make up

# Access the Tilt dashboard
# Open http://localhost:10350 in your browser

# Use k9s for interactive cluster management
make k9s

# Stop everything
make down
```

#### Core Tools

- **Kind**: Local Kubernetes cluster in Docker
- **Tilt**: Live development orchestration and dashboard
- **k9s**: Interactive Kubernetes CLI management
- **Nix + direnv**: Reproducible development environment

#### Tool Installation

Tools are managed via Nix and direnv. For manual installation:

```bash
# Install Kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Install Tilt
curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash

# Install k9s
curl -sS https://webinstall.dev/k9s | bash
```

#### Kubernetes Manifests

Located in `k8s/`:

- `namespace.yaml`    # Namespace for the application
- `postgres.yaml`     # PostgreSQL deployment and service
- `redis.yaml`        # Redis deployment and service
- `app.yaml`          # BrewSource MCP Server deployment and service

#### Tilt Dashboard

- Access at [http://localhost:10350](http://localhost:10350)
- Real-time build status, logs, resource monitoring, port forwarding

#### k9s Interface

```bash
make k9s
# Use :pods, :services, :logs, :describe, :edit in k9s
```

#### Advanced Usage

```bash
# Use custom Kind cluster name
KIND_CLUSTER_NAME=my-cluster make up

# Override resource limits
TILT_ARGS="--web-port=8080" make up

# Debug mode with verbose logging
TILT_DEBUG=1 make up
```

#### Manual Debugging

```bash
# Check cluster status
kubectl get pods -n brewsource

# View logs for specific pod
kubectl logs -n brewsource -l app=brewsource-mcp

# Port forward manually
kubectl port-forward -n brewsource svc/brewsource-mcp 8080:8080

# Execute commands in pod
kubectl exec -n brewsource -it <pod-name> -- sh
```

---

## CI/CD Integration

- GitHub Actions workflow for Docker and Kubernetes
- Docker build step with cache optimization
- Container health check validation
- Integration testing with Docker Compose
- Multi-architecture build preparation
- Artifact upload for Docker images

---

## Security & Optimization

### Security Features

- Non-root user (UID 1001)
- Minimal Alpine base image
- Static binary compilation
- No secrets in image layers
- Health check monitoring
- Resource limits ready
- Network policies (Kubernetes)
- Secrets management (Kubernetes)

### Image Optimization

- Multi-stage builds
- Layer caching
- Security scanning

---

## Testing & Troubleshooting

### Testing

```bash
# Run all tests
make test

# Run security scans
make security

# Format code
make format

# Run linting
make lint
```

### Troubleshooting

- **Kind cluster not starting:**
  - Check Docker is running: `docker ps`
  - Restart Kind cluster: `make clean && make up`
- **Tilt dashboard not accessible:**
  - Check Tilt is running: `tilt status`
  - Restart Tilt: `make down && make up`
- **Services not starting:**
  - Check pod status: `make k9s`
  - View pod logs: `kubectl logs -n brewsource -l app=brewsource-mcp`
- **Database connection issues:**
  - Check PostgreSQL pod: `kubectl get pods -n brewsource -l app=postgres`
  - Check database logs: `kubectl logs -n brewsource -l app=postgres`

---

## Environment Variables & Configuration

### Required Variables

- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string (optional)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `PORT`: Server port (default: 8080)

#### Docker Example

```bash
docker run -d --name brewsource-mcp \
  -p 8080:8080 \
  -e DATABASE_URL=your_postgres_url \
  -e REDIS_URL=your_redis_url \
  brewsource-mcp:latest
```

#### Kubernetes Example

```yaml
env:
  - name: DATABASE_URL
    value: "postgres://brewsource:password@postgres:5432/brewsource?sslmode=disable"
  - name: REDIS_URL
    value: "redis://redis:6379"
  - name: LOG_LEVEL
    value: "info"
  - name: PORT
    value: "8080"
```

---

## Migration & Advanced Topics

### Migration from Docker Compose

1. Remove old Docker Compose files (already done):
   - `docker-compose.yml`
   - `docker-compose.dev.yml`
   - Root `Dockerfile`
2. Update workflow:
   - Replace `docker-compose up` with `make up`
   - Replace `docker-compose down` with `make down`
   - Use `make k9s` instead of `docker ps`
3. New features:
   - Access Tilt dashboard at `http://localhost:10350`
   - Use `make k9s` for interactive cluster management
   - Faster rebuilds with optimized Kubernetes workflow

### Advanced Configuration

- Custom Kind cluster name, resource limits, debug mode, and more (see above)
- Production deployment: Build and push Docker image, apply Kubernetes manifests
- Security: Use Kubernetes secrets, network policies, and regular updates

---

## Benefits & Next Steps

### Benefits

1. **Developer Onboarding**: One-command setup for Docker or Kubernetes
2. **Consistent Environments**: Parity between development, testing, and production
3. **Easy Deployment**: Single Docker image or Kubernetes manifests
4. **Isolation**: No conflicts with local tools
5. **Scalability**: Native scaling with Kubernetes or Docker orchestration
6. **CI/CD Ready**: Integrated with GitHub Actions
7. **Security**: Following best practices for both Docker and Kubernetes

### Next Steps

- CI/CD integration with GitHub Actions
- Monitoring (Prometheus, Grafana)
- Pod Security Standards (Kubernetes)
- Horizontal Pod Autoscaler (Kubernetes)
- Production deployment to cloud Kubernetes

---

Your BrewSource MCP Server is now fully containerized and ready for any deployment scenario! 🐳🍺
