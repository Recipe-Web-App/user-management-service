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

- **API Layer:** `/app/api/v1/routes/` - FastAPI routers (auth, users, admin, social, notifications)
- **Service Layer:** `/app/services/` - Business logic services
- **Data Layer:** `/app/db/` - SQLAlchemy models and session management
- **Schema Layer:** `/app/api/v1/schemas/` - Pydantic request/response models

**Dual Database Setup:**

- **PostgreSQL** - Primary data storage with async SQLAlchemy
- **Redis** - JWT session management and caching

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

- **auth.py** - Registration, login, logout, password reset
- **users.py** - Profile management, search, CRUD operations
- **admin.py** - Administrative functions and system monitoring
- **social.py** - Following/followers functionality
- **notifications.py** - User notification management

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

- JWT authentication with refresh tokens
- bcrypt password hashing
- Role-based access control (USER/ADMIN)
- Sensitive data protection utilities
- Request ID middleware for tracing
