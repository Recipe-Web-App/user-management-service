# Contributing to User Management Service

First off, thanks for taking the time to contribute!

## Code of Conduct

This project and everyone participating in it is governed by the [Code of Conduct](CODE_OF_CONDUCT.md). By participating,
you are expected to uphold this code.

## Getting Started

1. **Fork** the repository on GitHub
2. **Clone** your fork locally
3. **Setup** your development environment

## Development Setup

### Prerequisites

- **Go 1.24**: Ensure you have Go 1.24 installed.
- **Make**: We use `make` for running tasks.

### Installation

```bash
git clone https://github.com/Recipe-Web-App/user-management-service.git
cd user-management-service
go mod download
```

## Development Workflow

1. Create a new branch for your feature or fix: `git checkout -b feature/my-new-feature`
2. Make your changes.
3. Run tests and linters locally.

## Testing

We use standard Go testing.

```bash
make test
```

To run with coverage:

```bash
make test-coverage
```

## Code Style

We strictly follow standard Go idioms and use `golangci-lint` to enforce style.

```bash
make lint # Runs pre-commit hooks including golangci-lint
```

Ensure all checks pass before submitting your PR.

## Commit Guidelines

We use [Conventional Commits](https://www.conventionalcommits.org/).

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Formatting (no code changes)
- `refactor:` Refactoring (no functional changes)
- `test:` Adding missing tests
- `chore:` Maintenance tasks

## Pull Request Process

1. Ensure local tests pass.
2. Update documentation if necessary.
3. Push your branch and open a Pull Request.
4. Fill out the Pull Request Template completely.
5. Wait for review from `@jsamuelsen11`.

## Questions?

Feel free to open an issue or start a Discussion if you have questions!
