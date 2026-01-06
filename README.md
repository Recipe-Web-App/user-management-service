# User Management Service

A Go microservice for managing user data, profiles, social connections, and preferences. Part of the Recipe Web App ecosystem.

## Features

- User profile management
- Social features (follow/unfollow, followers, following)
- User preferences management
- Admin endpoints for user statistics and cache management
- Health and readiness endpoints with dependency checks
- Prometheus metrics

## Prerequisites

- **Go 1.24**
- **PostgreSQL** - Primary database
- **Redis** - Caching layer
- **Make** - Build automation
- **Docker** and **Minikube** (for containerized deployment)

## Quick Start

### Local Development

1. Clone the repository:

    ```bash
    git clone https://github.com/Recipe-Web-App/user-management-service.git
    cd user-management-service
    ```

2. Install dependencies:

    ```bash
    go mod download
    ```

3. Set up environment variables by copying the example file:

    ```bash
    cp .env.example .env.local
    ```

4. Configure your `.env.local` with database and Redis connection details.

5. Run the service:

    ```bash
    make run
    ```

    The server starts on port 8080 by default.

## Build Commands

```bash
make build           # Build binary to bin/server
make run             # Run server directly (port 8080)
make clean           # Remove build artifacts
make lint            # Run pre-commit hooks (golangci-lint)
make check           # Lint + test + build (full validation)
```

## Testing

```bash
make test            # Run all tests (excludes performance)
make test-unit       # Unit tests only
make test-component  # Component tests
make test-dependency # Integration tests
make test-performance # Benchmark tests
make test-all        # Run all test suites including performance
make test-coverage   # Generate coverage.html report
```

Run a single test:

```bash
go test -v -run TestName ./path/to/package/...
```

## API Documentation

See the [OpenAPI specification](docs/openapi.yaml) for detailed API documentation.

## Configuration

Configuration is managed via YAML files in the `config/` directory:

- `server.yaml` - HTTP server settings
- `database.yaml` - PostgreSQL and Redis connection settings
- `logging.yaml` - Logging configuration
- `cors.yaml` - CORS settings
- `oauth2.yaml` - OAuth2/JWT settings

Environment variables override YAML config using the `USERMGMT_` prefix (e.g., `USERMGMT_SERVER_PORT`).

## Deployment (Minikube)

Requires Docker, Minikube, and Kubectl.

### Deploy

We provide shell scripts to manage the lifecycle of the application in Minikube.

1. **Start/Deploy**:

    ```bash
    ./scripts/containerManagement/deploy-container.sh
    ```

    This script will:
    - Check for required tools.
    - Start Minikube if not running.
    - Build the Docker image inside Minikube's Docker daemon.
    - Apply ConfigMaps, Secrets, and Deployments.
    - Wait for the pod to be ready.

2. **Monitor**:

    ```bash
    ./scripts/containerManagement/get-container-status.sh
    ```

    Displays a dashboard with pod status, logs, resource usage, and health checks.

3. **Update**:

    ```bash
    ./scripts/containerManagement/update-container.sh
    ```

    updates the image and performs a rollout restart.

4. **Stop/Start**:

    ```bash
    ./scripts/containerManagement/stop-container.sh
    ./scripts/containerManagement/start-container.sh
    ```

    Easily pause and resume the service without full redeployment.

5. **Cleanup**:

    ```bash
    ./scripts/containerManagement/cleanup-container.sh
    ```

    Removes the namespace and all associated resources.

### Kubernetes Resources

Manifests are located in the `k8s/` directory:

- `namespace.yaml`: Defines `user-management-dev-poc`.
- `deployment.yaml`: Main application deployment with probes.
- `service.yaml`: ClusterIP service.
- `configmap-template.yaml`: Configuration template.
- `secret-template.yaml`: Secret template.

## Contributing

See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for guidelines on contributing to this project.

## License

See [LICENSE](LICENSE) for details.
