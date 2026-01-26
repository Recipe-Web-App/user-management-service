# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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

```
handler → service → repository → database
```

**Entry point**: `cmd/api/main.go` - Initializes config, database, redis, logger, then starts HTTP server with graceful shutdown.

**Package structure** (`internal/`):

- `app/` - Dependency injection container (`Container` struct wires all dependencies)
- `config/` - Viper-based config loading from `config/*.yaml` files
- `server/` - HTTP server setup and route registration (`routes.go` defines all endpoints)
- `handler/` - HTTP request handlers (health, user, social, admin, metrics, preference)
- `service/` - Business logic layer (interfaces defined in `interfaces.go`)
- `repository/` - Database access layer (interfaces in `interfaces.go`)
- `dto/` - Data Transfer Objects for requests/responses
- `validation/` - Request validation using go-playground/validator
- `middleware/` - Custom middleware (logging, auth, metrics, context)
- `oauth2/` - JWT token management and OAuth2 authentication
- `database/` - PostgreSQL connection (pgx/v5) with pooling
- `redis/` - Redis client (go-redis/v9)
- `metrics/` - Prometheus metrics collection
- `notification/` - Notification service client
- `logger/` - Custom slog fanout handler for multi-output logging

**API Routes** (base path: `/api/v1/user-management`):
- `/health`, `/ready` - Health checks (public)
- `/users/*` - User profile, search, account deletion (authenticated)
- `/users/{user_id}/following|followers` - Social features (authenticated)
- `/users/{user_id}/preferences` - User preferences (authenticated)
- `/admin/*` - User stats, cache management (authenticated)
- `/metrics/*` - Performance, cache, system metrics (authenticated)
- `/metrics` (root) - Prometheus endpoint (public)

**Configuration**: YAML files in `config/` directory, overridable via environment variables with `USERMGMT_` prefix.

**Key patterns**:

- Dependency injection via `app.Container` - supports mock injection for testing via `ContainerConfig`
- Non-fatal initialization for external dependencies (service starts even if DB/Redis unavailable)
- Health endpoint returns "UP"; Ready endpoint aggregates dependency health with graceful degradation
- Authenticated user context set via `middleware.SetAuthenticatedUser` and retrieved via `middleware.GetAuthenticatedUser`

**Deployment**: Minikube scripts in `scripts/containerManagement/` (deploy, update, stop, start, cleanup).

## Testing

**Test organization**:
- `internal/handler/*_test.go` - Unit tests with mocked services (run via `make test-unit`)
- `tests/component/` - Component tests using full router with mock dependencies via `TestMain` setup
- `tests/dependency/` - Integration tests requiring external dependencies
- `tests/performance/` - Benchmark tests (`go test -bench=.`)

**Mocking patterns**:
- DB mocking: `github.com/DATA-DOG/go-sqlmock`
- Redis mocking: `github.com/alicebob/miniredis/v2`
- Service interfaces enable easy mocking in handler tests
- Use `app.ContainerConfig` to inject mock repositories in component tests

**Auth in tests**: Use `middleware.SetAuthenticatedUser(ctx, &middleware.AuthenticatedUser{...})` to set authenticated user context. Handler tests have `setAuthenticatedUser` helper in `test_helpers_test.go`.

## Code Style

- golangci-lint with all linters enabled (120 char line limit, cyclomatic complexity max 20)
- Conventional Commits required (feat:, fix:, docs:, etc.)
- Import ordering: goimports with local prefix `github.com/jsamuelsen/recipe-web-app/user-management-service`
- Tests use testify for assertions
