# Contributing to opencligen

Thank you for your interest in contributing to opencligen! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.22 or later
- Make (optional, but recommended)
- golangci-lint (for linting)

### Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/crunchloop/opencligen.git
   cd opencligen
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run tests:
   ```bash
   make test
   # or
   go test ./...
   ```

4. Run linting:
   ```bash
   make lint
   # or
   golangci-lint run
   ```

5. Build the binary:
   ```bash
   make build
   # or
   go build -o bin/opencligen ./cmd/opencligen
   ```

## Making Changes

### Code Style

- Follow standard Go conventions and idioms
- Run `make fmt` before committing to format code
- Ensure all linting checks pass with `make lint`
- Add tests for new functionality

### Testing

- Write tests for any new functionality
- Ensure all existing tests pass before submitting a PR
- Aim for at least 80% test coverage for new code
- Run `make coverage` to check coverage

### Commit Messages

- Use clear, concise commit messages
- Start with a verb in imperative mood (e.g., "Add", "Fix", "Update")
- Reference related issues when applicable

## Pull Request Process

1. Fork the repository and create a feature branch
2. Make your changes and ensure all tests pass
3. Update documentation if needed
4. Submit a pull request with a clear description of the changes

## Project Structure

```
opencligen/
├── cmd/opencligen/     # CLI entry point
├── internal/
│   ├── spec/          # OpenAPI spec parsing
│   ├── plan/          # CLI command planning
│   ├── gen/           # Code generation
│   └── runtime_skel/  # Runtime library skeleton
├── .github/workflows/ # CI/CD configuration
└── Makefile          # Build automation
```

## Questions?

If you have questions or need help, please open an issue on GitHub.
