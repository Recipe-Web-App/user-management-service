# Contributing to User Management Service

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing
to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Style](#code-style)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Security](#security)
- [Questions](#questions)

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please report
unacceptable behavior through the project's issue tracker.

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for details.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:

   ```bash
   git clone https://github.com/YOUR_USERNAME/user-management-service.git
   cd user-management-service
   ```

3. **Add upstream remote**:

   ```bash
   git remote add upstream https://github.com/Recipe-Web-App/user-management-service.git
   ```

## Development Setup

### Prerequisites

- Python 3.11 or higher
- Poetry 1.8.3 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (for local development)
- Redis 7+ (for local development)
- Make

### Initial Setup

1. **Install Poetry** (if not already installed):

   ```bash
   curl -sSL https://install.python-poetry.org | python3 -
   ```

2. **Install dependencies**:

   ```bash
   poetry install
   ```

3. **Set up pre-commit hooks**:

   ```bash
   poetry run pre-commit install
   ```

4. **Set up environment**:

   ```bash
   cp .env.example .env
   # Edit .env with your local configuration
   ```

5. **Start development services**:

   ```bash
   make docker-up  # Starts PostgreSQL and Redis
   ```

6. **Run the service**:

   ```bash
   poetry run app-local  # Run with Uvicorn
   # OR
   make dev              # Alternative command
   ```

7. **Verify setup**:

   ```bash
   make health          # Check service health
   poetry run test      # Run tests
   ```

## Development Workflow

1. **Create a feature branch**:

   ```bash
   git checkout -b feature/your-feature-name
   # OR for bug fixes
   git checkout -b fix/bug-description
   ```

2. **Make your changes** following the code style guidelines

3. **Run tests frequently**:

   ```bash
   poetry run test
   # OR
   make test
   ```

4. **Run code quality checks**:

   ```bash
   poetry run pre-commit run --all-files
   # This runs Black, Ruff, isort, mypy
   ```

5. **Commit your changes** following commit guidelines (see below)

6. **Keep your branch updated**:

   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

7. **Push to your fork**:

   ```bash
   git push origin feature/your-feature-name
   ```

8. **Open a Pull Request** on GitHub

## Testing

This project maintains 90% minimum test coverage.

### Running Tests

```bash
# All tests
poetry run test
# OR
make test

# Unit tests only
pytest tests/unit/ -v

# Component tests only
pytest tests/component/ -v

# Performance tests
pytest tests/performance/ -v -m "not slow"

# With coverage
pytest --cov=app --cov-report=html
make coverage
```

### Test Organization

- **Unit Tests** (`tests/unit/`) - Fast, isolated tests for individual functions/classes
- **Component Tests** (`tests/component/`) - Test API endpoints and integration
- **Performance Tests** (`tests/performance/`) - Benchmark and load testing

### Test Requirements

- All new features must include tests
- Bug fixes should include regression tests
- Maintain or improve coverage (90% minimum)
- Tests must pass before PR approval

### Writing Tests

```python
import pytest
from fastapi.testclient import TestClient

def test_user_registration(client: TestClient):
    """Test user registration endpoint."""
    response = client.post(
        "/api/v1/auth/register",
        json={"username": "testuser", "email": "test@example.com", "password": "SecurePass123!"}
    )
    assert response.status_code == 201
    assert "id" in response.json()
```

## Code Style

### Code Quality Tools

This project uses strict code quality enforcement:

- **Black** - Code formatting (line length: 88)
- **Ruff** - Fast Python linter
- **isort** - Import sorting
- **mypy** - Static type checking
- **Bandit** - Security linting

### Running Code Quality Checks

```bash
# Format code
black .
poetry run black .

# Check linting
ruff check .
poetry run ruff check .

# Type checking
mypy .
poetry run mypy app/

# Security scan
bandit -r app/
poetry run bandit -r app/

# All checks (via pre-commit)
poetry run pre-commit run --all-files
make lint
```

### Code Standards

1. **Type Hints Required**

   ```python
   def get_user(user_id: int) -> User:
       """Retrieve user by ID."""
       return db.query(User).filter(User.id == user_id).first()
   ```

2. **Docstrings Required** (Google or NumPy style)

   ```python
   def create_user(username: str, email: str) -> User:
       """Create a new user account.

       Args:
           username: Unique username for the account
           email: User's email address

       Returns:
           Created User object

       Raises:
           ValueError: If username or email already exists
       """
       # Implementation
   ```

3. **Follow PEP 8** - Enforced by Black and Ruff

4. **Use Async/Await** - For database and I/O operations

   ```python
   async def get_user_async(user_id: int) -> User:
       """Async user retrieval."""
       async with get_session() as session:
           result = await session.execute(select(User).where(User.id == user_id))
           return result.scalar_one_or_none()
   ```

## Commit Guidelines

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for automated releases and changelog generation.

### Commit Message Format

```text
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat` - New features
- `fix` - Bug fixes
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `perf` - Performance improvements
- `test` - Adding or updating tests
- `chore` - Maintenance tasks
- `ci` - CI/CD changes
- `build` - Build system changes
- `revert` - Reverting previous commits

### Examples

```bash
# Feature
git commit -m "feat(auth): add OAuth2 password flow"

# Bug fix
git commit -m "fix(users): resolve duplicate email validation"

# Breaking change
git commit -m "feat(api)!: change user endpoint response format

BREAKING CHANGE: User endpoint now returns nested profile object"

# Multiple changes
git commit -m "chore: update dependencies and improve error handling

- Update FastAPI to 0.115.14
- Add better error messages for validation failures
- Improve logging for database errors"
```

### Release Impact

- `feat:` → Minor version bump (0.1.0 → 0.2.0)
- `fix:` → Patch version bump (0.1.0 → 0.1.1)
- `feat!:` or `BREAKING CHANGE:` → Major version bump (0.1.0 → 1.0.0)

## Pull Request Process

### Before Creating a PR

1. ✅ All tests pass (`make test`)
2. ✅ Code quality checks pass (`make lint`)
3. ✅ Coverage meets minimum (90%)
4. ✅ Pre-commit hooks pass
5. ✅ Branch is rebased on latest `main`
6. ✅ Commit messages follow conventional commits

### PR Requirements

1. **Use the PR template** - Fill out all sections
2. **Link related issues** - Use `Fixes #123` or `Closes #456`
3. **Provide context** - Explain why and what
4. **Include tests** - New code must be tested
5. **Update documentation** - README, CLAUDE.md, etc.
6. **Add screenshots** - For UI changes
7. **Describe breaking changes** - If any
8. **Security review** - Note any security implications

### PR Title Format

Follow conventional commit format:

```text
feat(auth): add social login support
fix(users): resolve avatar upload issue
docs: update API documentation
```

### Review Process

1. **Automated checks** - CI must pass
2. **Code review** - At least one approval required
3. **Security scan** - Automated security checks
4. **Coverage check** - Must maintain 90%
5. **Maintainer approval** - Final approval from maintainer

### After Approval

- PRs are merged via **squash merge** for clean history
- Branch is automatically deleted after merge
- Release automation triggers if appropriate

## Security

### Reporting Vulnerabilities

**DO NOT** create public issues for security vulnerabilities.

Use [GitHub Security Advisories][security-advisories] to report privately.

[security-advisories]: https://github.com/Recipe-Web-App/user-management-service/security/advisories/new

See [SECURITY.md](SECURITY.md) for detailed reporting guidelines.

### Security Guidelines

1. **No secrets in code** - Use environment variables
2. **Redact sensitive data** - In logs and errors
3. **SQL injection prevention** - Use parameterized queries (SQLAlchemy handles this)
4. **Input validation** - Use Pydantic models
5. **Password security** - Use bcrypt for hashing
6. **JWT best practices** - Short-lived access tokens, longer refresh tokens

## Questions

### Where to Ask

- **Usage questions** - [GitHub Discussions](https://github.com/Recipe-Web-App/user-management-service/discussions)
- **Bug reports** - [GitHub Issues](https://github.com/Recipe-Web-App/user-management-service/issues) (use template)
- **Feature requests** - [GitHub Issues](https://github.com/Recipe-Web-App/user-management-service/issues) (use template)
- **Security issues** - [GitHub Security Advisories](https://github.com/Recipe-Web-App/user-management-service/security/advisories/new)

### Additional Resources

- [SUPPORT.md](SUPPORT.md) - Support resources and FAQs
- [CLAUDE.md](../CLAUDE.md) - Development commands and architecture
- [README.md](../README.md) - Project overview and API documentation

---

Thank you for contributing! Your efforts help make this project better for everyone. 🎉
