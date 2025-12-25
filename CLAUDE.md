# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Core Development:**

```bash
poetry run app-local              # Run app locally (FastAPI with Uvicorn)
poetry run test                   # Run pytest test suite
make dev                         # Alternative: Run development server
make dev-reload                  # Run with hot reload for development
```

**Setup & Installation:**

```bash
make install                     # Install dependencies and pre-commit hooks
make setup                       # Complete development setup (install + docker)
cp .env.example .env             # Copy environment template
```

**Container Management:**

```bash
poetry run container-deploy       # Deploy to Kubernetes
poetry run container-start        # Start K8s deployment
poetry run container-stop         # Stop K8s deployment
poetry run container-status       # Check deployment status
poetry run container-cleanup      # Clean up K8s resources
```

**Docker Development:**

```bash
make docker-up                   # Start PostgreSQL and Redis services
make docker-up-all               # Start all services including admin tools
make docker-down                 # Stop development services
make docker-logs                 # View service logs
make docker-clean                # Clean up Docker resources
```

**Code Quality:**

```bash
black .                          # Format code
ruff check .                     # Lint code
mypy .                          # Type checking
bandit -r app/                  # Security analysis
make lint                        # Run all quality checks via pre-commit
make format                      # Format code (black + ruff + isort)
make security                    # Run security scans (bandit + safety)
```

**Conventional Commits:**

This project uses [Conventional Commits](https://conventionalcommits.org/) for consistent commit messages and automated releases:

```bash
git commit -m "feat: add user search functionality"     # New feature
git commit -m "fix: resolve authentication timeout"     # Bug fix
git commit -m "docs: update API documentation"          # Documentation
git commit -m "refactor: improve database queries"      # Code refactoring
git commit -m "test: add unit tests for user service"   # Tests
git commit -m "chore: update dependencies"              # Maintenance
```

**Supported commit types:**

- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes
- `build`: Build system changes
- `revert`: Reverting previous commits

**Release Management:**

Releases are automated based on conventional commits:

- `feat:` triggers minor version bump (0.1.0 → 0.2.0)
- `fix:` triggers patch version bump (0.1.0 → 0.1.1)
- `feat!:` or `BREAKING CHANGE:` triggers major version bump (0.1.0 → 1.0.0)

**Testing:**

```bash
pytest tests/unit/              # Run unit tests only
pytest tests/component/         # Run component tests only
pytest --cov=app --cov-report=html  # Run with coverage report
make test                        # Run all tests
make test-unit                   # Run unit tests with verbose output
make test-comp                   # Run component tests with verbose output
make test-perf                   # Run performance tests (excluding slow)
make coverage                    # Run tests with coverage (report in htmlcov/)
make load-test                   # Run load tests (including slow tests)
```

## Architecture Overview

**Clean Architecture Pattern** with FastAPI microservice for user management:

- **API Layer:** `/app/api/v1/routes/` - FastAPI routers (users, admin, social, notifications, health, metrics)
- **Service Layer:** `/app/services/` - Business logic services
- **Data Layer:** `/app/db/` - SQLAlchemy models and session management
- **Schema Layer:** `/app/api/v1/schemas/` - Pydantic request/response models

**Dual Database Setup:**

- **PostgreSQL** - Primary data storage with async SQLAlchemy
- **Redis** - Session management and caching

**Authentication Architecture:**

This service supports **dual authentication modes**:

1. **OAuth2 Integration** (recommended for production):
   - External OAuth2 service handles authentication
   - Authorization Code Flow with PKCE support
   - Token introspection OR JWT validation (configurable)
   - Service-to-service authentication via client credentials
   - Scope-based authorization: `user:read`, `user:write`, `admin`, `openid`, `profile`
   - Configure via: `OAUTH2_SERVICE_ENABLED=true`, `OAUTH2_INTROSPECTION_ENABLED=false` (for JWT)

2. **Legacy JWT Authentication** (backward compatibility):
   - Internal JWT token generation and validation
   - Access & refresh tokens with Redis session management
   - Role-based access control (USER/ADMIN)
   - Enable when: `OAUTH2_SERVICE_ENABLED=false`

**Note:** This service has NO `/auth/*` endpoints - authentication is delegated to external OAuth2 service.
All endpoints expect valid OAuth2 tokens or legacy JWT tokens in Authorization header.

**Development Dependencies:**

- Local: Docker Compose for PostgreSQL + Redis
- Production: Kubernetes with full manifests
- Testing: Comprehensive test suite with performance benchmarks

**Key Technologies:**

- FastAPI 0.115.14 with async/await patterns
- SQLAlchemy 2.0.41 ORM with async support
- Pydantic 2.0.0 for data validation
- Poetry for dependency management
- Kubernetes deployment with complete manifests in `/k8s/`

**Code Standards:**

- Python 3.11+ with strict type hints everywhere
- Google or NumPy docstring style required
- 90% test coverage enforced
- Pre-commit hooks mandatory before merging
- Security and dead code checks enforced

## API Structure

All endpoints use `/api/v1/` prefix. Main route modules:

- **users.py** - Profile management, search, CRUD operations
- **admin.py** - Administrative functions and system monitoring
- **social.py** - Following/followers functionality
- **notifications.py** - User notification management
- **health.py** - Health checks (`/health` for readiness, `/live` for liveness)
- **metrics.py** - Prometheus-compatible metrics endpoint

**Request/Response Pattern:**

- Structured schemas in `/schemas/request/` and `/schemas/response/`
- Custom exception handling with HTTP status codes
- Privacy-first design with built-in privacy preference checking

## Testing & Quality

**Test Structure:**

- Unit tests: `/tests/unit/`
- Component tests: `/tests/component/`
- Performance tests: `/tests/performance/`
- HTTP test files: `/tests/http/` (manual API testing)

**Test Markers & Categories:**

```bash
pytest -m unit                   # Run only unit tests
pytest -m component              # Run only component tests
pytest -m performance            # Run performance tests (excluding slow)
pytest -m slow                   # Run slow/load tests
pytest -m auth                   # Run authentication tests
pytest -m admin                  # Run admin functionality tests
pytest -m social                 # Run social features tests
pytest -m notifications          # Run notification system tests
```

**CI/CD Testing:**

```bash
make ci-test                     # Run CI test suite with XML coverage
make ci-quality                  # Run CI quality checks
make quick-test                  # Run lint + unit tests
make full-test                   # Run lint + all tests + coverage
```

**Quality Requirements:**

- 90% minimum test coverage (enforced)
- Black code formatting (line length: 88)
- Ruff linting with comprehensive rules
- MyPy strict type checking
- Bandit security analysis
- Type hints required everywhere
- Google or NumPy docstring style for all functions
- Pre-commit hooks must pass before merging

## Development Workflow

**Health Checks:**

```bash
make health                      # Check basic service health
make health-detailed             # Check detailed service health
curl http://localhost:8000/api/v1/health  # Direct health check
```

**Database Management:**

```bash
make db-reset                    # Reset development database
make db-shell                    # Connect to PostgreSQL shell
make redis-shell                 # Connect to Redis CLI
```

**Development Utilities:**

```bash
make clean                       # Clean cache and temp files
make shell                       # Python shell with app context
make docs                        # Open API documentation (/docs)
```

## Configuration

**Environment Variables Required:**

- Database connections (PostgreSQL + Redis)
- JWT authentication settings
- CORS configuration
- Logging configuration via `/config/logging.json`

**Security Features:**

- OAuth2 integration with scope-based authorization OR legacy JWT authentication
- bcrypt password hashing
- Role-based access control (USER/ADMIN) with OAuth2 scope fallback
- Sensitive data protection utilities
- Request ID middleware for tracing

## Known Limitations & Development Notes

**Missing Implementations (from README):**

- **Email service integration** - Password reset currently only logs tokens
  (see `app/services/auth_service.py:437-439`). Needs SMTP/email provider integration.
- **Incomplete services** - `SocialService`, `NotificationService`, and some `AdminService` methods are partial implementations
- **Placeholder endpoint** - `GET /users/{user_id}` returns placeholder response (`app/api/v1/routes/users.py:358`)
- **Database migrations** - No automated migration system. Currently uses manual `init.sql`.
  Consider implementing Alembic for production.

**OAuth2 Configuration:**

When enabling OAuth2 integration, ensure these environment variables are set:

- `JWT_SECRET` - Must match the shared secret with external OAuth2 service
- `OAUTH2_CLIENT_ID` and `OAUTH2_CLIENT_SECRET` - For service-to-service auth
- `OAUTH2_INTROSPECTION_ENABLED=false` - Recommended for better performance (uses JWT validation instead)

**Health Check Endpoints:**

- `/api/v1/user-management/health` - Readiness check with dependency status (PostgreSQL + Redis)
- `/api/v1/user-management/live` - Liveness check for Kubernetes probes

## Notification Service Integration

**Overview:**

The user-management-service integrates with the notification-service to send email notifications for social
interactions. This integration uses OAuth2 client credentials flow for service-to-service authentication.

**Integration Points:**

- **New Follower Notifications** - When a user follows another user, a notification is sent to the followed user
  - Endpoint: `POST /notifications/new-follower` on notification-service
  - Checks user notification preferences before sending
  - Graceful degradation: follow operations succeed even if notifications fail

**Configuration:**

Notification service URLs and scopes are hard-coded in `app/clients/constants.py`:

```python
# Local development
NOTIFICATION_SERVICE_URL_LOCAL = "http://sous-chef-proxy.local/api/v1/notification"

# Kubernetes/production
NOTIFICATION_SERVICE_URL_K8S = "http://notification-service.notification.svc.cluster.local:8000/api/v1/notification"

# Required OAuth2 scopes
NOTIFICATION_SERVICE_SCOPES = ["notification:admin"]
```

**Architecture:**

1. **OAuth2 Token Manager** (`app/clients/oauth2_token_manager.py`)
   - Manages OAuth2 access tokens for service-to-service calls
   - Automatically refreshes tokens when expired (no arbitrary caching buffer)
   - Thread-safe for concurrent requests

2. **Base OAuth2 Service Client** (`app/clients/base_oauth2_service_client.py`)
   - Reusable HTTP client with persistent connection pooling (httpx)
   - Automatic OAuth2 token injection in Authorization header
   - Configurable timeouts and error handling

3. **Notification Client** (`app/clients/notification_client.py`)
   - Extends BaseOAuth2ServiceClient
   - Implements `send_new_follower_notification()` method
   - Handles graceful degradation on failures

4. **Integration in Social Service** (`app/services/social_service.py`)
   - Checks user notification preferences before sending
   - Required preferences: `email_notifications=True` AND `social_interactions=True`
   - Non-blocking: notification failures don't affect core functionality

**Preference Checking:**

Before sending notifications, the service checks the target user's preferences:

```python
# Required database fields (UserNotificationPreferences model)
email_notifications: bool       # Must be True
social_interactions: bool       # Must be True
```

If either preference is `False` or preferences don't exist, notifications are skipped.

**Environment Detection:**

The service automatically selects the correct notification service URL based on the `settings.environment` value:

- `production` → Uses Kubernetes URL
- Any other value → Uses local URL

**Graceful Degradation:**

All notification operations implement graceful degradation:

- HTTP errors (503, timeout, etc.) are logged but don't raise exceptions
- Invalid responses are logged and return `None`
- User operations (like following) always succeed regardless of notification status

**Testing:**

- Unit tests: `tests/unit/clients/test_oauth2_token_manager.py`
- Unit tests: `tests/unit/clients/test_notification_client.py`
- Component tests: `tests/component/test_notification_integration.py`

**Downstream Schemas:**

Request/response schemas for notification service are in `app/api/v1/schemas/downstream/notification/`:

- `NewFollowerNotificationRequest` - Request schema
- `BatchNotificationResponse` - Response schema (202 Accepted)
- `NotificationItem` - Individual notification in batch response
