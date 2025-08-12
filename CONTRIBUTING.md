# Contributing to User Management Service

Thank you for your interest in contributing to the User Management Service! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Security Vulnerabilities](#security-vulnerabilities)

## Code of Conduct

### Our Pledge

We are committed to making participation in this project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

Examples of behavior that contributes to creating a positive environment include:

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

Examples of unacceptable behavior include:

- The use of sexualized language or imagery and unwelcome sexual attention or advances
- Trolling, insulting/derogatory comments, and personal or political attacks
- Public or private harassment
- Publishing others' private information without explicit permission
- Other conduct which could reasonably be considered inappropriate in a professional setting

## Getting Started

### Ways to Contribute

- **Bug Reports**: Report bugs and issues
- **Feature Requests**: Suggest new features or improvements
- **Code Contributions**: Fix bugs, implement features, improve performance
- **Documentation**: Improve documentation, add examples, fix typos
- **Testing**: Add tests, improve test coverage
- **Reviews**: Review pull requests and provide feedback

### Before You Start

1. **Check existing issues** to see if your bug report or feature request already exists
2. **Search pull requests** to see if someone is already working on it
3. **Read the documentation** to understand the project structure and conventions
4. **Set up the development environment** following the [Development Guide](docs/DEVELOPMENT.md)

## Development Setup

### Prerequisites

- Python 3.11+
- Poetry 1.8+
- Docker & Docker Compose
- Git

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/jsamuelsen/user-management-service.git
cd user-management-service

# Install dependencies
poetry install

# Set up environment
cp .env.example .env
# Edit .env with your configuration

# Start development services
docker-compose up -d postgres redis

# Run the application
poetry run app-local

# Run tests
poetry run pytest
```

For detailed setup instructions, see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

## Contributing Guidelines

### Branching Strategy

We use **Git Flow** branching strategy:

- `main` - Production-ready code
- `develop` - Integration branch for features
- `feature/*` - Feature development branches
- `bugfix/*` - Bug fix branches
- `hotfix/*` - Critical production fixes

### Creating Feature Branches

```bash
# Start from develop branch
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/your-feature-name

# Work on your feature
# ... make changes ...

# Push your branch
git push origin feature/your-feature-name
```

### Commit Message Convention

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

**Examples:**
```
feat(auth): add password reset functionality

fix(users): resolve profile update validation error

docs(api): update authentication endpoint documentation

test(social): add integration tests for follow functionality
```

## Code Standards

### Python Code Style

We use several tools to maintain code quality:

- **Black** - Code formatting (line length: 88)
- **Ruff** - Fast linting with comprehensive rules
- **MyPy** - Static type checking
- **isort** - Import sorting
- **Bandit** - Security vulnerability scanning

### Pre-commit Hooks

Install and use pre-commit hooks to automatically check code quality:

```bash
poetry run pre-commit install
poetry run pre-commit run --all-files
```

### Code Quality Requirements

- **Type Hints**: All functions must have type hints
- **Docstrings**: All public functions and classes must have docstrings
- **Error Handling**: Proper exception handling with custom exceptions
- **Logging**: Use structured logging with appropriate levels
- **Security**: Follow security best practices, no hardcoded secrets

### Example Code Style

```python
"""Module docstring describing the module's purpose."""

import logging
from typing import Optional

from fastapi import HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.models.user import User
from app.exceptions.custom_exceptions import UserNotFoundError

logger = logging.getLogger(__name__)


async def get_user_by_id(
    db_session: AsyncSession,
    user_id: str
) -> Optional[User]:
    """Get a user by their ID.

    Args:
        db_session: Database session
        user_id: User's unique identifier

    Returns:
        User object if found, None otherwise

    Raises:
        UserNotFoundError: If user does not exist
    """
    try:
        user = await db_session.get(User, user_id)
        if not user:
            raise UserNotFoundError(f"User with ID {user_id} not found")

        logger.info("Retrieved user", extra={"user_id": user_id})
        return user

    except Exception as e:
        logger.error(
            "Failed to retrieve user",
            extra={"user_id": user_id, "error": str(e)}
        )
        raise
```

## Testing Requirements

### Test Coverage

- **Minimum coverage**: 90%
- **Unit tests**: Test individual functions and methods
- **Component tests**: Test API endpoints end-to-end
- **Performance tests**: Test under load conditions

### Test Structure

```
tests/
├── unit/           # Unit tests
├── component/      # API integration tests
├── performance/    # Load and performance tests
├── conftest.py     # Test configuration and fixtures
└── __init__.py
```

### Writing Tests

```python
"""Example test file."""

import pytest
from httpx import AsyncClient


@pytest.mark.unit
@pytest.mark.auth
async def test_password_validation():
    """Test password validation logic."""
    # Test implementation
    pass


@pytest.mark.component
@pytest.mark.auth
async def test_user_registration_endpoint(async_client: AsyncClient):
    """Test user registration API endpoint."""
    response = await async_client.post(
        "/api/v1/auth/register",
        json={
            "username": "testuser",
            "email": "test@example.com",
            "password": "TestPass123!",
            "first_name": "Test",
            "last_name": "User"
        }
    )

    assert response.status_code == 201
    assert "user_id" in response.json()
```

### Running Tests

```bash
# All tests
poetry run pytest

# Specific test types
poetry run pytest tests/unit/       # Unit tests
poetry run pytest tests/component/  # Component tests
poetry run pytest tests/performance/ # Performance tests

# With coverage
poetry run pytest --cov=app --cov-report=html

# Run tests for specific functionality
poetry run pytest -m auth          # Authentication tests
poetry run pytest -k "test_user"   # Tests with "test_user" in name
```

## Pull Request Process

### Before Submitting

1. **Update your branch** with the latest changes from develop:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout feature/your-feature
   git rebase develop
   ```

2. **Run all quality checks**:
   ```bash
   poetry run pre-commit run --all-files
   poetry run pytest --cov=app
   ```

3. **Update documentation** if needed
4. **Add changelog entry** if it's a significant change

### Pull Request Template

When creating a pull request, include:

```markdown
## Description
Brief description of the changes and their purpose.

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?
Describe the tests you ran and provide instructions to reproduce.

## Checklist
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Coverage remains above 90%

## Screenshots (if applicable)
Add screenshots to help explain your changes.

## Related Issues
Closes #123
```

### Review Process

1. **Automated checks** must pass (CI pipeline)
2. **Code review** by at least one maintainer
3. **Testing** in development environment
4. **Documentation** review if applicable

### Merge Requirements

- All CI checks pass
- At least one approving review from a maintainer
- No unresolved review comments
- Branch is up-to-date with target branch
- Test coverage remains above 90%

## Issue Reporting

### Bug Reports

When reporting bugs, please include:

1. **Bug description**: Clear and concise description
2. **Steps to reproduce**: Detailed steps to reproduce the behavior
3. **Expected behavior**: What you expected to happen
4. **Actual behavior**: What actually happened
5. **Environment information**:
   - OS and version
   - Python version
   - Service version
   - Browser (if applicable)
6. **Screenshots**: If applicable
7. **Additional context**: Logs, error messages, etc.

### Feature Requests

When requesting features, please include:

1. **Feature description**: Clear and concise description
2. **Use case**: Explain why this feature would be useful
3. **Proposed solution**: Describe how you envision the feature working
4. **Alternatives considered**: Other solutions you've considered
5. **Additional context**: Any other relevant information

### Issue Labels

We use labels to categorize issues:

- **Type**: `bug`, `feature`, `documentation`, `question`
- **Priority**: `low`, `medium`, `high`, `critical`
- **Status**: `needs-triage`, `in-progress`, `blocked`
- **Component**: `auth`, `users`, `social`, `admin`, `api`

## Security Vulnerabilities

### Reporting Security Issues

**Do not report security vulnerabilities through public GitHub issues.**

Instead, please email security@recipe-app.com with:

1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact
4. Suggested fix (if any)

We will acknowledge receipt of your report within 48 hours and provide a detailed response within 5 business days.

### Security Guidelines

- Never commit secrets, API keys, or passwords
- Use environment variables for sensitive configuration
- Follow the principle of least privilege
- Validate all inputs and sanitize outputs
- Use parameterized queries to prevent SQL injection
- Implement proper authentication and authorization
- Keep dependencies up to date

## Development Resources

### Documentation

- [API Documentation](docs/API.md)
- [Development Guide](docs/DEVELOPMENT.md)
- [Architecture Overview](README.md#architecture-overview)

### Tools and Services

- **Code Quality**: GitHub Actions CI/CD
- **Testing**: pytest with coverage reporting
- **Security**: Bandit, safety, gitleaks
- **Documentation**: Swagger/OpenAPI, Markdown
- **Monitoring**: Health checks, logging

### Communication

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: General questions, ideas
- **Pull Request Reviews**: Code feedback and discussions

## Recognition

Contributors are recognized in several ways:

- **Contributors list** in README.md
- **Changelog acknowledgments** for significant contributions
- **GitHub contributor statistics** automatically tracked

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project (GNU General Public License v3.0).

---

Thank you for contributing to the User Management Service! Your efforts help make this project better for everyone.
