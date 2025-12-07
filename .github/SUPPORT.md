# Support

Thank you for using the User Management Service! This document provides resources to help you get support.

## Documentation

Before asking for help, please check our documentation:

### Primary Documentation

- **[README.md](../README.md)** - Complete feature overview, setup instructions, and API documentation
- **[CLAUDE.md](../CLAUDE.md)** - Development commands, architecture overview, and developer guide
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Contribution guidelines and development workflow
- **[SECURITY.md](SECURITY.md)** - Security features, best practices, and vulnerability reporting

### Code Examples

- **[`.env.example`](../.env.example)** - Configuration examples
- **[Docker Compose](../docker-compose.yml)** - Deployment examples
- **[Kubernetes Manifests](../k8s/)** - K8s deployment configurations

## Getting Help

### 1. Search Existing Resources

Before creating a new issue, please search:

- [Existing Issues](https://github.com/Recipe-Web-App/user-management-service/issues) - Someone may have already asked
- [Closed Issues](https://github.com/Recipe-Web-App/user-management-service/issues?q=is%3Aissue+is%3Aclosed) - Your question
  may already be answered
- [Discussions](https://github.com/Recipe-Web-App/user-management-service/discussions) - Community Q&A

### 2. GitHub Discussions (Recommended for Questions)

For general questions, use [GitHub Discussions](https://github.com/Recipe-Web-App/user-management-service/discussions):

**When to use Discussions:**

- "How do I...?" questions
- Configuration help
- Best practice advice
- Integration questions
- Authentication flow clarifications
- Architecture discussions
- Troubleshooting (non-bug)

**Categories:**

- **Q&A** - Ask questions and get answers
- **Ideas** - Share feature ideas and proposals
- **Show and Tell** - Share your implementations
- **General** - General discussion about the project

### 3. GitHub Issues (For Bugs and Feature Requests)

Create a GitHub issue when you have:

- **Bug** - Something is not working as expected
- **Feature Request** - A new feature or enhancement
- **Performance Issue** - Service is slow or has performance problems
- **Security Issue** - Low-severity security concerns (use Security Advisories for critical issues)
- **Documentation** - Documentation is missing, incorrect, or unclear

**When creating an issue:**

1. Use the appropriate issue template
2. Search for duplicates first
3. Provide all requested information
4. Include logs, error messages, and environment details
5. Redact any sensitive information

### 4. Security Issues

**DO NOT** create public issues for security vulnerabilities.

Use [GitHub Security Advisories][security-advisories] to report security issues privately.

[security-advisories]: https://github.com/Recipe-Web-App/user-management-service/security/advisories/new

See [SECURITY.md](SECURITY.md) for detailed reporting guidelines.

## Common Questions

### Setup and Configuration

**Q: How do I set up the development environment?**

A: See [CONTRIBUTING.md](CONTRIBUTING.md) for complete setup instructions. Basic steps:

```bash
# Install Poetry
poetry install

# Copy environment template
cp .env.example .env

# Start services
make docker-up

# Run the application
poetry run app-local
```

**Q: What environment variables are required?**

A: See `.env.example` for all required configuration. Key variables include:

- Database: `POSTGRES_HOST`, `POSTGRES_DB`, `USER_MANAGEMENT_DB_USER`, `USER_MANAGEMENT_DB_PASSWORD`
- JWT: `JWT_SECRET_KEY`, `JWT_SIGNING_ALGORITHM`
- Redis: `REDIS_HOST`, `REDIS_PORT`

**Q: How do I run tests?**

A: Use Poetry:

```bash
poetry run test                 # All tests
pytest tests/unit/              # Unit tests only
pytest tests/component/         # Component tests only
pytest --cov                    # With coverage
```

### Common Errors

#### Q: Getting "connection refused" errors

A: Ensure PostgreSQL and Redis services are running:

```bash
make docker-up
make health
```

#### Q: Tests are failing with database errors

A: Reset your test database:

```bash
make db-reset
poetry run test
```

#### Q: Import errors or module not found

A: Reinstall dependencies:

```bash
poetry install --no-cache
```

### Integration

**Q: How do I integrate this service with my application?**

A: The service provides a RESTful API. See the API documentation:

- Interactive docs: `http://localhost:8000/docs` (when running locally)
- API endpoints: `/api/v1/auth/`, `/api/v1/users/`, `/api/v1/admin/`

**Q: What authentication methods are supported?**

A: JWT-based authentication with:

- Access tokens (short-lived, 30 minutes)
- Refresh tokens (long-lived, 7 days)
- Password reset tokens (15 minutes)

**Q: How do I handle CORS in my frontend?**

A: Configure `ALLOWED_ORIGIN_HOSTS` in your `.env`:

```bash
ALLOWED_ORIGIN_HOSTS=http://localhost:3000,https://myapp.com
ALLOWED_CREDENTIALS=true
```

### Performance

**Q: How can I improve performance?**

A: Optimization tips:

- Ensure Redis is properly configured for caching
- Use connection pooling for database
- Monitor slow queries with logging
- Run performance tests: `pytest tests/performance/`

**Q: What are the performance benchmarks?**

A: See `tests/performance/` for benchmark tests. Expected performance:

- Login: < 200ms
- User retrieval: < 50ms
- Search: < 100ms

## Response Times

We aim for:

- **Discussions**: Response within 2-3 business days
- **Issues**: Triage within 1 week
- **Security**: Within 48 hours (see SECURITY.md)

## Bug Report Best Practices

To get the fastest resolution:

1. **Clear title** - Summarize the issue in one line
2. **Detailed description** - What's wrong and what you expected
3. **Steps to reproduce** - Exact steps to trigger the issue
4. **Environment** - Python version, OS, deployment type
5. **Logs** - Include relevant logs (redact sensitive data!)
6. **Screenshots** - Visual aids if applicable

## Additional Resources

### Python/FastAPI Resources

- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Pydantic Documentation](https://docs.pydantic.dev/)
- [SQLAlchemy Documentation](https://docs.sqlalchemy.org/)
- [Poetry Documentation](https://python-poetry.org/docs/)

### Authentication & Security

- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)

### Docker & Kubernetes

- [Docker Documentation](https://docs.docker.com/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

## Contact

For general questions and community support, use GitHub Discussions.

For security issues, use GitHub Security Advisories.

For all other inquiries, create a GitHub issue with the appropriate template.
