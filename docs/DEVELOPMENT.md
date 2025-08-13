# Development Guide

This guide provides comprehensive instructions for setting up and developing the User Management Service.

## Prerequisites

### Required Software

- **Python 3.11+** - Required for modern syntax and performance improvements
- **Poetry 1.8+** - For dependency management and virtual environments
- **Docker & Docker Compose** - For local development services
- **Git** - For version control
- **Make** (optional) - For convenient command shortcuts

### Recommended Tools

- **VS Code** with Python extension
- **Postman** or **Insomnia** for API testing
- **DBeaver** or **pgAdmin** for database management
- **Redis Commander** for Redis management

## Initial Setup

### 1. Clone and Enter Repository

```bash
git clone https://github.com/jsamuelsen/user-management-service.git
cd user-management-service
```

### 2. Install Python Dependencies

```bash
# Install Poetry if not already installed
curl -sSL https://install.python-poetry.org | python3 -

# Install project dependencies
poetry install

# Activate virtual environment
poetry shell
```

### 3. Set Up Environment Variables

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your preferred editor
nano .env  # or vim .env, code .env, etc.
```

**Important environment variables to configure:**

```bash
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=recipe_manager
USER_MANAGEMENT_DB_USER=user_management
USER_MANAGEMENT_DB_PASSWORD=your_secure_password

# JWT
JWT_SECRET_KEY=your_super_secret_jwt_key_change_this
ACCESS_TOKEN_EXPIRE_MINUTES=30

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_secure_redis_password

# CORS
ALLOWED_ORIGIN_HOSTS=http://localhost:3000,http://localhost:8080
```

### 4. Start Development Services

```bash
# Start PostgreSQL and Redis with Docker Compose
docker-compose up -d postgres redis

# Wait for services to be ready (about 30 seconds)
docker-compose logs postgres redis

# Optional: Start admin tools
docker-compose --profile admin up -d pgadmin redis-commander
```

### 5. Initialize Database

```bash
# Run database migrations (if using Alembic)
# poetry run alembic upgrade head

# Or let the application create tables on startup
poetry run app-local
```

### 6. Verify Setup

```bash
# Check service readiness (includes dependency status)
curl http://localhost:8000/api/v1/user-management/health

# Check service liveness (simple alive check)
curl http://localhost:8000/api/v1/user-management/live

# View API documentation
open http://localhost:8000/docs
```

## Development Workflow

### Running the Application

#### Local Development

```bash
# Run with hot reload
poetry run app-local

# Or run directly with uvicorn
poetry run uvicorn app.main:app --host 127.0.0.1 --port 8000 --reload
```

#### Docker Development

```bash
# Build and run with Docker Compose
docker-compose up --build user-management-service

# View logs
docker-compose logs -f user-management-service
```

### Code Quality and Testing

#### Pre-commit Hooks

```bash
# Install pre-commit hooks
poetry run pre-commit install

# Run all hooks manually
poetry run pre-commit run --all-files

# Run specific hooks
poetry run pre-commit run black
poetry run pre-commit run ruff-check
poetry run pre-commit run mypy
```

#### Running Tests

```bash
# All tests
poetry run pytest

# Unit tests only
poetry run pytest tests/unit/ -v

# Component tests only
poetry run pytest tests/component/ -v

# Performance tests (excluding slow ones)
poetry run pytest tests/performance/ -v -m "not slow"

# With coverage report
poetry run pytest --cov=app --cov-report=html --cov-report=term

# Open coverage report
open htmlcov/index.html
```

#### Code Quality Checks

```bash
# Format code
poetry run black .

# Sort imports
poetry run isort .

# Lint code
poetry run ruff check . --fix

# Type checking
poetry run mypy app/

# Security scanning
poetry run bandit -r app/

# Dead code detection
poetry run vulture app/
```

## Database Development

### Using PostgreSQL

```bash
# Connect to database
docker exec -it user-management-postgres psql -U user_management -d recipe_manager

# Run SQL commands
\dt  # List tables
\d users  # Describe users table
SELECT * FROM users LIMIT 5;
```

### Using pgAdmin

- URL: <http://localhost:5050>
- Email: <admin@recipe-app.local>
- Password: (from your .env file)

### Database Migrations (if using Alembic)

```bash
# Create new migration
poetry run alembic revision --autogenerate -m "Add new field to users table"

# Apply migrations
poetry run alembic upgrade head

# Rollback migration
poetry run alembic downgrade -1

# View migration history
poetry run alembic history
```

## Redis Development

### Using Redis CLI

```bash
# Connect to Redis
docker exec -it user-management-redis redis-cli -a your_redis_password

# Common Redis commands
KEYS *  # List all keys
GET session:user:123  # Get session data
FLUSHALL  # Clear all data (development only!)
```

### Using Redis Commander

- URL: <http://localhost:8081>
- Username: admin
- Password: (from your .env file)

## API Development

### Interactive API Documentation

- **Swagger UI**: <http://localhost:8000/docs>
- **ReDoc**: <http://localhost:8000/redoc>

### Testing API Endpoints

#### Using curl

```bash
# Register a new user
curl -X POST "http://localhost:8000/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "TestPass123!",
    "first_name": "Test",
    "last_name": "User"
  }'

# Login
curl -X POST "http://localhost:8000/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123!"
  }'

# Use access token for authenticated requests
export ACCESS_TOKEN="your_access_token_here"
curl -X GET "http://localhost:8000/api/v1/users/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

#### Using HTTP Files

The project includes HTTP test files in `tests/http/`:

- `auth.http` - Authentication endpoints
- `users.http` - User management endpoints
- `social.http` - Social features
- `admin.http` - Admin endpoints

Use with VS Code REST Client extension or similar tools.

## Debugging

### Application Debugging

```bash
# Run with debug logging
export LOG_LEVEL=DEBUG
poetry run app-local

# Use Python debugger
# Add breakpoint in code: import pdb; pdb.set_trace()
poetry run python -m pdb app/main.py
```

### Database Debugging

```bash
# Enable SQL query logging in config
export SQLALCHEMY_ECHO=true

# Or add to .env file
echo "SQLALCHEMY_ECHO=true" >> .env
```

### Performance Profiling

```bash
# Run performance tests
poetry run pytest tests/performance/ -v -s

# Profile specific functions
poetry run python -m cProfile -o profile.out app/main.py
poetry run python -c "import pstats; pstats.Stats('profile.out').sort_stats('time').print_stats(20)"
```

## Development Best Practices

### Code Organization

- **Routes**: Keep route handlers thin, delegate to services
- **Services**: Implement business logic, interact with database
- **Models**: Define data structures and relationships
- **Schemas**: Handle request/response validation
- **Dependencies**: Centralize dependency injection

### Error Handling

```python
# Use custom exceptions
from app.exceptions.custom_exceptions import UserNotFoundError

# Raise with proper HTTP status codes
raise UserNotFoundError("User with ID 123 not found")

# Log errors appropriately
import logging
logger = logging.getLogger(__name__)
logger.error("Database connection failed", exc_info=True)
```

### Testing Guidelines

- **Unit tests**: Test individual functions and methods
- **Component tests**: Test API endpoints end-to-end
- **Performance tests**: Test under load and stress
- **Use fixtures**: Create reusable test data
- **Mock external dependencies**: Database, Redis, external APIs

### Security Considerations

- **Never log sensitive data**: Passwords, tokens, personal information
- **Validate all inputs**: Use Pydantic models for validation
- **Use parameterized queries**: Prevent SQL injection
- **Rate limiting**: Implement for public endpoints
- **HTTPS only**: In production environments

## Troubleshooting

### Common Issues

#### Poetry Issues

```bash
# Clear Poetry cache
poetry cache clear --all pypi

# Reinstall dependencies
rm poetry.lock
poetry install
```

#### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Reset database
docker-compose down postgres
docker volume rm user-management-service_postgres_data
docker-compose up -d postgres
```

#### Redis Connection Issues

```bash
# Check if Redis is running
docker-compose ps redis

# Test Redis connection
docker exec -it user-management-redis redis-cli ping

# Clear Redis data
docker exec -it user-management-redis redis-cli FLUSHALL
```

#### Application Startup Issues

```bash
# Check environment variables
poetry run python -c "from app.core.config import settings; print(settings.model_dump())"

# Check application logs
tail -f logs/app.log

# Run with debug mode
export DEBUG=true
poetry run app-local
```

### Performance Issues

```bash
# Check database query performance
# Enable query logging and review slow queries

# Monitor memory usage
docker stats user-management-service

# Profile application
poetry run python -m cProfile app/main.py
```

### Getting Help

1. **Check logs**: Application and service logs often contain helpful error messages
2. **Review documentation**: API documentation and code comments
3. **Search issues**: Check GitHub issues for similar problems
4. **Debug step by step**: Use debugger to trace execution
5. **Ask for help**: Create detailed GitHub issues with logs and steps to reproduce

## Contributing Workflow

### Before Making Changes

1. Create feature branch: `git checkout -b feature/your-feature-name`
2. Run tests: `poetry run pytest`
3. Check code quality: `poetry run pre-commit run --all-files`

### Making Changes

1. Write code following project conventions
2. Add/update tests for new functionality
3. Update documentation if needed
4. Run tests and quality checks frequently

### Before Committing

1. Run full test suite: `poetry run pytest`
2. Check coverage: `poetry run pytest --cov=app`
3. Run pre-commit hooks: `poetry run pre-commit run --all-files`
4. Update CHANGELOG.md if needed

### Submitting Changes

1. Push branch: `git push origin feature/your-feature-name`
2. Create pull request with clear description
3. Ensure CI pipeline passes
4. Address review feedback

## IDE Setup

### VS Code Configuration

Create `.vscode/settings.json`:

```json
{
  "python.defaultInterpreterPath": ".venv/bin/python",
  "python.formatting.provider": "black",
  "python.linting.enabled": true,
  "python.linting.ruffEnabled": true,
  "python.linting.mypyEnabled": true,
  "python.testing.pytestEnabled": true,
  "python.testing.pytestArgs": ["tests/"],
  "files.exclude": {
    "**/__pycache__": true,
    "**/.pytest_cache": true,
    "**/.mypy_cache": true
  }
}
```

### Recommended VS Code Extensions

- Python
- Python Docstring Generator
- REST Client (for HTTP files)
- Docker
- GitLens
- Error Lens

## Environment-Specific Notes

### Development Environment

- Debug mode enabled
- Hot reload enabled
- Detailed logging
- Test data seeding available

### Testing Environment

- Separate test database
- Mock external services
- Coverage reporting
- Performance testing

### Production Environment

- Debug mode disabled
- Optimized logging
- Security headers enabled
- Performance monitoring
- Error tracking

This guide should get you up and running with the User Management Service development environment. For additional
help, refer to the API documentation or create an issue on GitHub.
