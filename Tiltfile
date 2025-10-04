# Tiltfile for BrewSource MCP Server development

# Load Kubernetes YAML using Kustomize
k8s_yaml(kustomize('k8s/dev'))

# Build the Go application
local_resource(
    'build-go-binary',
    cmd='cd app && CGO_ENABLED=0 GOOS=linux go build -o bin/brewsource-mcp cmd/server/main.go',
    deps=['app/cmd', 'app/internal', 'app/pkg', 'app/go.mod', 'app/go.sum']
)

docker_build(
    'ghcr.io/charlritter/brewsource-mcp:latest',
    '.',
    dockerfile='Dockerfile',
    platform='linux/amd64'
)

# Set up port forwarding for local development
k8s_resource('brewsource-mcp', port_forwards='8080:8080')
k8s_resource('postgres', port_forwards='5432:5432')
k8s_resource('redis', port_forwards='6379:6379')

# Set up file watching for live reload
watch_file('app/cmd')
watch_file('app/internal')
watch_file('app/pkg')
watch_file('app/go.mod')
watch_file('app/go.sum')

# Wait for dependencies
k8s_resource('brewsource-mcp', resource_deps=['postgres', 'redis', 'build-go-binary'])

print("""
🍺 BrewSource MCP Server Development Environment

Services:
- MCP Server: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379

Commands:
- tilt up     - Start development
- tilt down   - Stop development
- k9s         - Explore cluster
""")
