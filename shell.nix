{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    # Kubernetes development tools
    kind
    kubectl
    tilt
    k9s

    # Container tools
    docker
    docker-compose

    # Development tools
    direnv

    # Application development
    go_1_21
    postgresql
    redis

    # Utilities
    curl
    wget
    jq
  ];

  shellHook = ''
    echo "🍺 BrewSource MCP Development Environment"
    echo "========================================"
    echo ""
    echo "Available commands:"
    echo "  make up    - Start development environment"
    echo "  make down  - Stop development environment"
    echo "  make clean - Clean up everything"
    echo ""
    echo "Tools available:"
    echo "  kind, tilt, k9s, kubectl, direnv"
    echo ""

    # Set up environment
    export KUBECONFIG="$HOME/.kube/config"
    export TILT_HOST="0.0.0.0"
    export DATABASE_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

    # Allow direnv if .envrc exists
    if [ -f .envrc ]; then
      direnv allow
    fi
  '';
}
