# GEMINI.md

This file provides guidance to Gemini (Antigravity) when working with code in this repository.

## Build & Development Commands

```bash
make build           # Build binary to bin/server
make run             # Run server directly (port 8080)
make clean           # Remove build artifacts
make test            # Run all tests (excludes performance)
make test-unit       # Unit tests only (./internal/handler/...)
make test-component  # Component tests (./tests/component/...)
make test-dependency # Integration tests (./tests/dependency/...)
make test-performance # Benchmark tests
make test-all        # Run all test suites including performance
make test-coverage   # Generate coverage.html report
make lint            # Run pre-commit hooks (golangci-lint)
make check           # Lint + test + build (full validation)
```

Run a single test: `go test -v -run TestName ./path/to/package/...`

## Architecture

Go 1.24 microservice using Chi router for HTTP handling, following a **layered architecture**:

```text
handler → service → repository → database
```

**Entry point**: `cmd/api/main.go` - Initializes config, database, redis, logger, then starts HTTP server with graceful shutdown.

**Package structure** (`internal/`):

- `app/` - Dependency injection container
- `config/` - Viper-based config loading from `config/*.yaml` files
- `server/` - HTTP server setup and route registration
- `handler/` - HTTP request handlers (health, user, social, admin, metrics)
- `service/` - Business logic layer
- `repository/` - Database access layer (users, social, preferences)
- `dto/` - Data Transfer Objects for requests/responses
- `validation/` - Request validation using go-playground/validator
- `middleware/` - Custom middleware (logging, auth, metrics, context)
- `oauth2/` - JWT token management and OAuth2 authentication
- `database/` - PostgreSQL connection (pgx/v5) with pooling
- `redis/` - Redis client (go-redis/v9)
- `metrics/` - Prometheus metrics collection
- `notification/` - Notification service client
- `logger/` - Custom slog fanout handler for multi-output logging

**Configuration**: YAML files in `config/` directory, overridable via environment variables with `USERMGMT_` prefix.

**Key patterns**:

- Global singletons for Config, Database, Redis, Logger
- Non-fatal initialization for external dependencies (service starts even if DB/Redis unavailable)
- Health endpoint returns "UP"; Ready endpoint aggregates dependency health with graceful degradation

## Code Style

- golangci-lint with all linters enabled (120 char line limit, cyclomatic complexity max 20)
- Conventional Commits required (feat:, fix:, docs:, etc.)
- Import ordering: goimports with local prefix `github.com/jsamuelsen/recipe-web-app/user-management-service`
- Tests use testify for assertions, sqlmock for DB mocking, miniredis for Redis mocking

## Agent Behavior

- **Cleanup**: Any temporary files created during a session (e.g., test output logs, temp code files) MUST be deleted
  immediately after their purpose is served. Do not leave clutter in the workspace.
