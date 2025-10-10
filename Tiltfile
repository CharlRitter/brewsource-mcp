# Tiltfile for BrewSource MCP Server development

# Load Kubernetes YAML using Kustomize
k8s_yaml(kustomize('k8s/dev'))

docker_build(
    'ghcr.io/charlritter/brewsource-mcp:latest',
    '.',
    dockerfile='Dockerfile',
    platform='linux/amd64',
    live_update=[
        # Rebuild binary when Go files change
        fall_back_on(['app/cmd/**/*.go', 'app/internal/**/*.go', 'app/pkg/**/*.go', 'go.mod', 'go.sum']),
        # Sync data files without rebuilding
        sync('app/data/', '/app/data/'),
    ]
)

# Set up port forwarding for local development
k8s_resource('brewsource-mcp', port_forwards='8080:8080')
k8s_resource('postgres', port_forwards='5432:5432')
k8s_resource('redis', port_forwards='6379:6379')

# Set up file watching for live reload
watch_file('app/cmd')
watch_file('app/internal')
watch_file('app/pkg')
watch_file('go.mod')
watch_file('go.sum')

# Wait for dependencies
k8s_resource('brewsource-mcp', resource_deps=['postgres', 'redis'])

print("""
🍺 BrewSource MCP Server Development Environment

Services:
- MCP Server:   http://localhost:8080
- PostgreSQL:   localhost:5432
- Redis:        localhost:6379

Commands:
- make up:      Start development
- make down:    Stop development
- make k9s:     Explore cluster
""")
