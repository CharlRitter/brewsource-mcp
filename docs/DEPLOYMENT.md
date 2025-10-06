# BrewSource MCP Server Deployment Guide

This guide provides comprehensive instructions for deploying the BrewSource MCP Server using Docker and Kubernetes workflows.
It also includes production deployment on Oracle Cloud with K3s, Traefik, and Cert-Manager. The guide covers local development,
production deployment, CI/CD integration, security, and troubleshooting.

---

## Table of Contents

1. [Overview](#overview)
2. [Deployment Options](#deployment-options)
   - [Docker Workflow](#docker-workflow)
   - [Kubernetes Workflow](#kubernetes-workflow)
   - [Production on Oracle Cloud (K3s, Traefik, Cert-Manager)](#production-on-oracle-cloud-k3s-traefik-cert-manager)
3. [CI/CD Integration](#cicd-integration)
4. [Security & Optimization](#security--optimization)
5. [Testing & Troubleshooting](#testing--troubleshooting)
6. [Environment Variables & Configuration](#environment-variables--configuration)
7. [Migration & Advanced Topics](#migration--advanced-topics)
8. [Benefits & Next Steps](#benefits--next-steps)

---

## Overview

BrewSource MCP Server is an open-source Model Context Protocol (MCP) server providing AI assistants with comprehensive brewing
knowledge and tools. This guide helps you set up, develop, and deploy the server using modern containerization and orchestration
tools.

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

- **BrewSource MCP Server**: <http://localhost:8080>
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

Located in `k8s/base/`:

- `namespace.yaml`    # Namespace for the application (in overlays, not base)
- `brewsource-mcp.yaml` # BrewSource MCP Server deployment, config, and service
- `postgres.yaml`     # PostgreSQL deployment, config, PVC, and service
- `redis.yaml`        # Redis deployment, PVC, and service

#### Ingress and TLS

The production Ingress (`k8s/prod/ingress.yaml`) uses Traefik and includes:

- HTTP Ingress with middleware for automatic HTTP-to-HTTPS redirect
- HTTPS Ingress with cert-manager TLS annotation and reference to the generated TLS secret
- Host: `brewsource.charlritter.com` (update as needed)

#### Network Policies

Production includes detailed network policies (`k8s/prod/network-policies.yaml`) to restrict traffic between app, database,
redis, and cert-manager pods, following security best practices.

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

### Production on Oracle Cloud (K3s, Traefik, Cert-Manager)

This section provides step-by-step instructions to set up a production-ready Kubernetes environment on Oracle Cloud Infrastructure
(OCI) using K3s, Traefik, and Cert-Manager. It covers networking, cluster installation, secrets management,
and automated HTTPS.
**All steps should be run on the remote server via SSH unless specified otherwise.**

#### Step 1: OCI Instance and Firewall Setup

1. **Create an OCI Compute Instance:**
   - Use an "Always Free Eligible" shape (e.g., `VM.Standard.E2.1.Micro`).
   - Select the Canonical Ubuntu image.
   - Ensure it is assigned a Public IPv4 address.
   - Add your public SSH key for access.

2. **Configure VCN Security Lists:**
   - Go to your instance's Virtual Cloud Network (VCN) → Security Lists → Default Security List.
   - Add the following Ingress Rules:
     - **Port 22 (SSH):** Source `0.0.0.0/0`, Protocol TCP, Destination Port 22
     - **Port 80 (HTTP):** Source `0.0.0.0/0`, Protocol TCP, Destination Port 80 _(required for Let's Encrypt validation)_
     - **Port 443 (HTTPS):** Source `0.0.0.0/0`, Protocol TCP, Destination Port 443
     - **Port 6443 (K8s API):** Source `0.0.0.0/0`, Protocol TCP, Destination Port 6443

---

#### Step 2: Install and Configure K3s and Local Tooling

1. **SSH into the OCI Instance.**

2. **Install K3s (Kubernetes):**
   - Disable the default Traefik Ingress controller (we will install our own):

     ```sh
     curl -sfL https://get.k3s.io | sh -s - --disable=traefik
     ```

3. **Configure kubectl for Server-Side Use:**
   - K3s includes `kubectl`, but you need to configure your shell to find it:

     ```sh
     mkdir -p ~/.kube
     sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
     sudo chown $(id -u):$(id -g) ~/.kube/config
     ```

   - Your `kubectl` commands will now work correctly on the server.

4. **Install Helm (Kubernetes Package Manager):**

   ```sh
   curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
   chmod 700 get_helm.sh
   ./get_helm.sh
   ```

5. **Clone Your Application Repository:**

   ```sh
   sudo apt update && sudo apt install git -y
   git clone https://github.com/CharlRitter/brewsource-mcp.git
   cd brewsource-mcp
   ```

---

#### Step 3: Install Cluster Controllers

1. **Install Traefik (Ingress Controller):**
   - Uninstall any old versions first:

     ```sh
     helm uninstall traefik
     ```

   - Install Traefik with explicit arguments:

     ```sh
     helm repo add traefik https://helm.traefik.io/traefik
     helm repo update
     helm install traefik traefik/traefik \
       --set="additionalArguments={--providers.kubernetesingress.ingressclass=traefik}" \
       --set="ports.websecure.tls.enabled=true"
     ```

2. **Install Cert-Manager (for SSL/TLS):**

   ```sh
   helm repo add jetstack https://charts.jetstack.io
   helm repo update
   helm install cert-manager jetstack/cert-manager \
     --namespace cert-manager \
     --create-namespace \
     --set installCRDs=true
   ```

3. **Install Sealed Secrets (for Encrypted Secrets):**

   ```sh
   helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
   helm install sealed-secrets sealed-secrets/sealed-secrets -n kube-system
   ```

_Wait about a minute for all controller pods to be in a Running state before proceeding._

---

#### Step 4: Nuke, Pave, and Deploy the Application

This "nuke and pave" approach ensures both the application and the database start with the same, correct credentials,
preventing authentication failures.

1. **Delete the Namespace (if it exists):**

   ```sh
   kubectl delete namespace brewsource-prod
   ```

2. **Create the Namespace:**

   ```sh
   kubectl create namespace brewsource-prod
   ```

3. **Create and Seal the Database Secret:**
   - **IMPORTANT:** The password must be URL-safe. Use only letters (a-z, A-Z) and numbers (0-9). Special characters will
     cause the application to crash.

   - Create a temporary plain-text secret file:

     ```sh
     cat <<EOF > temp-postgres-secret.yaml
     apiVersion: v1
     kind: Secret
     metadata:
       name: postgres-secret
       namespace: brewsource-prod
     stringData:
       POSTGRES_USER: "brewsource_user"
       POSTGRES_PASSWORD: "averylongsafepassword123"
     EOF
     ```

   - Find the exact name of the Sealed Secrets service:

     ```sh
     kubectl get svc -n kube-system
     # The name will be 'sealed-secrets'
     ```

   - Use `kubeseal` to encrypt the file, specifying the correct controller name and namespace. This overwrites the old
     `postgres-sealedsecret.yaml` in your repo:

     ```sh
     kubeseal --controller-namespace kube-system --controller-name sealed-secrets < temp-postgres-secret.yaml > k8s/prod/postgres-sealedsecret.yaml
     ```

   - Clean up the temporary plain-text file:

     ```sh
     rm temp-postgres-secret.yaml
     ```

4. **Deploy the Application Stack:**
   - Apply your entire Kustomize configuration. This will create all resources, including the ClusterIssuer for cert-manager
     and the final, corrected Ingress rules:

     ```sh
     kubectl apply -k brewsource-mcp/k8s/prod
     ```

---

#### Step 5: Final DNS Configuration

1. **Find your server's public IP address.**

2. **Go to your DNS provider** (the service where you manage your domain).

3. **Create an A Record:**
   - **Type:** A
   - **Host/Name:** `brewsource` (or your chosen subdomain)
   - **Value:** Your server's public IP address

_After DNS propagates (a few minutes to an hour), cert-manager will automatically complete the SSL/TLS challenge,_
_and your site will be live and secure._

---

#### Step 6: Run Codacy CLI Analysis

1. **Install Codacy CLI:**

- Follow the instructions on the [Codacy CLI documentation](https://docs.codacy.com/cli/installation/) to install the CLI
  tool.

1. **Run the analysis:**

- Navigate to your project directory and run:

    ```sh
    codacy-cli analyze
    ```

1. **Review the results:**

- After the analysis completes, review the results and make any necessary changes to improve code quality.

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

In development, environment variables are set via a ConfigMap and a plain Secret. In production, sensitive values are
provided via a SealedSecret, and `DATABASE_URL` is constructed from secret values using a patch:

```yaml
env:
  - name: DATABASE_URL
    value: "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:5432/brewsource_prod?sslmode=disable"
  - name: REDIS_URL
    value: "redis://redis:6379"
  - name: LOG_LEVEL
    value: "info"
  - name: PORT
    value: "8080"
  - name: POSTGRES_USER
    valueFrom:
      secretKeyRef:
        name: postgres-secret
        key: POSTGRES_USER
  - name: POSTGRES_PASSWORD
    valueFrom:
      secretKeyRef:
        name: postgres-secret
        key: POSTGRES_PASSWORD
```

> **Note:** In production, do not hardcode credentials. Use the provided SealedSecret and patches for secure configuration.

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
