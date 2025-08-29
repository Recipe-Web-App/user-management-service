# User Management Service

[![Build Status](https://github.com/jsamuelsen/user-management-service/actions/workflows/security.yml/badge.svg)](https://github.com/jsamuelsen/user-management-service/actions/workflows/security.yml)
[![Coverage Status](https://img.shields.io/badge/coverage-90%25-brightgreen)](https://github.com/jsamuelsen/user-management-service)
[![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-blue.svg)](LICENSE)
[![Python Version](https://img.shields.io/badge/python-3.11-blue.svg)](https://www.python.org/downloads/release/python-3110/)
[![FastAPI](https://img.shields.io/badge/FastAPI-0.115.14-009688.svg)](https://fastapi.tiangolo.com)
[![Code style: black](https://img.shields.io/badge/code%20style-black-000000.svg)](https://github.com/psf/black)

A modern, high-performance microservice for user management in the Recipe Web Application ecosystem.
Built with FastAPI, featuring comprehensive authentication, profile management, social features, and administrative
capabilities.

## 🏗️ Architecture Overview

This service follows **Clean Architecture** principles with clear separation of concerns:

```text
├── app/
│   ├── api/v1/           # API Layer - FastAPI routers & schemas
│   ├── services/         # Business Logic Layer
│   ├── db/              # Data Layer - SQLAlchemy & Redis models
│   ├── core/            # Configuration & utilities
│   ├── middleware/      # Request processing middleware
│   └── exceptions/      # Custom exception handling
├── tests/               # Comprehensive test suite
├── k8s/                 # Kubernetes deployment manifests
└── scripts/             # Container management scripts
```

### Key Technologies

- **FastAPI 0.115.14** - Modern async web framework
- **SQLAlchemy 2.0.41** - Async ORM with PostgreSQL
- **Redis 6.2.0** - Session management & caching
- **Pydantic 2.0.0** - Data validation & serialization
- **Poetry** - Dependency management
- **Kubernetes** - Container orchestration

## 🚀 Quick Start

### Prerequisites

- Python 3.11+
- Poetry
- Docker & Docker Compose (for local development)
- Kubernetes cluster (for deployment)

### Local Development Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/jsamuelsen/user-management-service.git
   cd user-management-service
   ```

2. **Install dependencies**

   ```bash
   poetry install
   ```

3. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start local services** (PostgreSQL & Redis)

   ```bash
   docker-compose up -d
   ```

5. **Run the application**

   ```bash
   poetry run app-local
   ```

The API will be available at `http://localhost:8000` with interactive documentation at `http://localhost:8000/docs`.

## 📋 Core Features

### 🔐 Authentication & Security

- **JWT-based authentication** with access & refresh tokens
- **Secure password hashing** using bcrypt
- **Password reset** functionality with time-limited tokens
- **Session management** via Redis
- **Role-based access control** (USER/ADMIN)

### 👤 User Management

- **User registration & profile management**
- **Privacy preferences** with granular controls
- **Account preferences** (theme, language, notifications)
- **User search** with privacy-aware filtering
- **Account deletion** with data cleanup

### 🤝 Social Features

- **Follow/unfollow** other users
- **Social preferences** management
- **Privacy-first design** with comprehensive preference checking

### 🔔 Notifications System

- **Real-time notifications** delivery
- **Notification preferences** per user
- **Read/unread status** tracking
- **Bulk operations** (mark all as read, delete)

### ⚙️ Administrative Functions

- **User statistics** and analytics
- **System health monitoring**
- **Session management** (force logout, clear sessions)
- **Redis metrics** and monitoring

## 🛠️ API Reference

### Base URL

```text
Production: https://api.recipe-app.com/user-management
Development: http://localhost:8000/api/v1
```

### Authentication Endpoints

```bash
# User Registration
POST /auth/register
Content-Type: application/json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe"
}

# Login
POST /auth/login
Content-Type: application/json
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}

# Token Refresh
POST /auth/refresh
Authorization: Bearer <refresh_token>
```

### User Management Endpoints

```bash
# Get User Profile
GET /users/{user_id}/profile
Authorization: Bearer <access_token>

# Update Profile
PUT /users/profile
Authorization: Bearer <access_token>
Content-Type: application/json
{
  "first_name": "John",
  "last_name": "Smith",
  "bio": "Updated bio"
}

# Search Users
GET /users/search?query=john&limit=10
Authorization: Bearer <access_token>
```

### Social Features

```bash
# Follow User
POST /social/follow
Authorization: Bearer <access_token>
Content-Type: application/json
{
  "user_id": "uuid-here"
}

# Get Followers
GET /social/followers
Authorization: Bearer <access_token>
```

### Notifications

```bash
# Get Notifications
GET /notifications?limit=20&offset=0
Authorization: Bearer <access_token>

# Mark as Read
PUT /notifications/{notification_id}/read
Authorization: Bearer <access_token>
```

For complete API documentation, visit `/docs` when running the service.

## 🧪 Testing

### Run Tests

```bash
# All tests
poetry run test

# Unit tests only
pytest tests/unit/

# Component tests only
pytest tests/component/

# With coverage
pytest --cov=app --cov-report=html
```

### Test Structure

- **Unit Tests** - Service layer and utilities
- **Component Tests** - API endpoint integration
- **Performance Tests** - Load testing and benchmarks
- **90% Coverage Requirement** - Enforced in CI/CD

## 🚀 Deployment

### Local Kubernetes

```bash
# Start deployment
poetry run container-start

# Check status
poetry run container-status

# Stop deployment
poetry run container-stop
```

### Production Deployment

```bash
# Deploy to production
poetry run container-deploy

# Monitor deployment
kubectl get pods -n user-management

# Check deployment status and logs
kubectl describe pods -n user-management
kubectl logs -n user-management -l app=user-management --tail=100

# Verify health checks
kubectl exec -n user-management deployment/user-management -- curl localhost:8000/api/v1/user-management/live
```

### Kubernetes Features

- **High Availability**: 2 replicas with pod anti-affinity
- **Health Monitoring**: Separate liveness and readiness probes
- **Network Security**: NetworkPolicy for controlled access
- **Resource Management**: CPU/memory requests and limits with ephemeral storage
- **Security**: Non-root user, read-only filesystem, dropped capabilities
- **Disruption Budget**: Ensures minimum availability during updates

### Environment Variables

| Variable                      | Description          | Required |
| ----------------------------- | -------------------- | -------- |
| `POSTGRES_HOST`               | PostgreSQL host      | ✅       |
| `POSTGRES_PORT`               | PostgreSQL port      | ✅       |
| `POSTGRES_DB`                 | Database name        | ✅       |
| `USER_MANAGEMENT_DB_USER`     | DB username          | ✅       |
| `USER_MANAGEMENT_DB_PASSWORD` | DB password          | ✅       |
| `JWT_SECRET_KEY`              | JWT signing secret   | ✅       |
| `REDIS_HOST`                  | Redis host           | ✅       |
| `REDIS_PASSWORD`              | Redis password       | ✅       |
| `ALLOWED_ORIGIN_HOSTS`        | CORS allowed origins | ✅       |

See `.env.example` for complete configuration.

## 🔧 Development

### Code Quality

```bash
# Format code
black .
ruff check . --fix

# Type checking
mypy .

# Security scan
bandit -r app/

# Pre-commit hooks
pre-commit run --all-files
```

### Database Migrations

```bash
# Create migration
alembic revision --autogenerate -m "description"

# Apply migrations
alembic upgrade head
```

### Container Management

```bash
# Build container
docker build -t user-management-service .

# Run with Docker Compose
docker-compose up --build
```

## 📊 Monitoring & Observability

### Health Checks

- `GET /user-management/health` - **Readiness check** with dependency status
  - Returns 200 OK when ready to serve requests
  - Returns 503 Service Unavailable when dependencies are unhealthy
  - Includes detailed database and Redis connectivity status
  - Response time metrics for each dependency
- `GET /user-management/live` - **Liveness check** for Kubernetes
  - Simple alive status (always returns 200 OK if service is running)
  - Used by Kubernetes liveness probes

### Logging

- **Structured JSON logging** with correlation IDs
- **Multiple log levels** (DEBUG, INFO, WARNING, ERROR, CRITICAL)
- **File rotation** with compression
- **Request tracing** via middleware

### Metrics

- Request/response metrics
- Database query performance
- Redis operation metrics
- Custom business metrics

## 📦 Releases & Versioning

This project uses [Semantic Versioning](https://semver.org/) and
[Conventional Commits](https://conventionalcommits.org/) for automated releases.

### Version Format

```text
MAJOR.MINOR.PATCH (e.g., 1.2.3)
```

- **MAJOR**: Breaking changes (`feat!:` or `BREAKING CHANGE:`)
- **MINOR**: New features (`feat:`)
- **PATCH**: Bug fixes (`fix:`)

### Release Process

Releases are **fully automated** via GitHub Actions:

1. **Commits** follow conventional commit format
2. **Release Please** analyzes commits and creates release PRs
3. **Merging** release PR triggers:
   - Version bump in `pyproject.toml`
   - CHANGELOG.md update
   - GitHub release creation
   - Package build and artifact upload

### Commit Types & Versioning

| Commit Type | Example                                | Version Impact        |
| ----------- | -------------------------------------- | --------------------- |
| `feat:`     | `feat(auth): add OAuth support`        | Minor (0.1.0 → 0.2.0) |
| `fix:`      | `fix(users): resolve validation error` | Patch (0.1.0 → 0.1.1) |
| `feat!:`    | `feat!: remove legacy API`             | Major (0.1.0 → 1.0.0) |
| `docs:`     | `docs: update API examples`            | None                  |
| `chore:`    | `chore: update dependencies`           | None                  |

### Latest Release

[![GitHub Release](https://img.shields.io/github/v/release/jsamuelsen/user-management-service?include_prereleases&sort=semver)](https://github.com/jsamuelsen/user-management-service/releases)
[![GitHub Release Date](https://img.shields.io/github/release-date/jsamuelsen/user-management-service)](https://github.com/jsamuelsen/user-management-service/releases)

## 🤝 Contributing

We use **Conventional Commits** for consistent messaging and automated releases.

### Quick Contribution Steps

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following our code standards
4. Use conventional commit messages:

   ```bash
   git commit -m "feat(users): add advanced search filters"
   git commit -m "fix(auth): resolve token refresh timing issue"
   git commit -m "docs(api): add OAuth endpoint examples"
   ```

5. Run tests and quality checks
6. Push to your branch and open a Pull Request

### Commit Message Format

```text
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Examples:**

```bash
feat(auth): add password reset functionality
fix(users): resolve profile update validation error
docs(api): update authentication endpoint documentation
test(social): add integration tests for follow functionality
chore(deps): update FastAPI to 0.115.14
ci: add automated release workflow
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

### Code Standards

- **Conventional Commits** enforced via pre-commit hooks
- **Black** formatting (line length: 88)
- **Ruff** linting with comprehensive rules
- **MyPy** strict type checking
- **90% test coverage** minimum
- **Comprehensive docstrings** for all public methods

## 📄 License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## 🛡️ Security

For security concerns, please email <security@recipe-app.com> instead of using the issue tracker.

## 📞 Support

- **Documentation**: [API Docs](docs/API.md) | [Development Guide](docs/DEVELOPMENT.md)
- **Issues**: [GitHub Issues](https://github.com/jsamuelsen/user-management-service/issues)
- **Discussions**: [GitHub Discussions](https://github.com/jsamuelsen/user-management-service/discussions)

---

## 🚧 Remaining Work & Recommended Enhancements

Based on comprehensive codebase analysis, this service is **90%+ production-ready** with the following remaining items:

### 🔥 HIGH PRIORITY (Production Gaps)

#### Email Service Integration

- Password reset emails currently only log tokens (`app/services/auth_service.py:437-439`)
- Need integration with SMTP service or email provider (SendGrid, AWS SES)

#### Missing Service Implementations

- Complete `SocialService` implementation for follow/unfollow functionality
- Complete `NotificationService` implementation for notification management
- Finalize `AdminService` functionality for administrative operations

#### Database Migration System

- No automated migration system (currently uses manual `init.sql`)
- Recommend implementing Alembic for schema version control

#### Incomplete API Endpoint

- `GET /users/{user_id}` returns placeholder (`app/api/v1/routes/users.py:358`)

### 🔶 MEDIUM PRIORITY (Operational Enhancements)

#### Monitoring & Observability

- Add Prometheus metrics exposure (`/metrics` endpoint)
- Integrate Application Performance Monitoring (APM)
- Set up distributed tracing (OpenTelemetry)

#### Performance & Caching

- Implement application-level caching for user profiles
- Add database query optimization and connection pooling configuration
- Redis currently only used for sessions, expand for general caching

#### Security Enhancements

- API rate limiting and DDoS protection
- Account lockout after failed login attempts
- Consider MFA/2FA implementation

#### File Storage Support

- User avatar/profile image upload functionality
- File storage backend integration (S3, GCS, etc.)

### 🔷 LOW PRIORITY (Nice-to-Have)

#### Advanced Features

- API versioning strategy for future v2 development
- Automated backup/restore procedures
- Data retention and cleanup policies
- Kubernetes HPA policies for auto-scaling

#### Development Experience

- Database seed data for development environments
- API client SDK generation
- Enhanced development containerization

### 📋 Implementation Priority

| Priority    | Item             | Effort | Impact |
| ----------- | ---------------- | ------ | ------ |
| 🔥 Critical | Email Service    | Medium | High   |
| 🔥 Critical | Missing Services | High   | High   |
| 🔶 High     | Monitoring Stack | Medium | Medium |
| 🔶 High     | Rate Limiting    | Low    | Medium |
| 🔷 Medium   | File Storage     | Medium | Low    |

---

## Built with ❤️ using FastAPI and modern Python practices
