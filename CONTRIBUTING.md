# Contributing to CoT Library

Thank you for your interest in contributing to the CoT Library for Go! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a standard code of conduct. By participating, you are expected to:
- Be respectful and inclusive
- Focus on constructive feedback
- Maintain professional discourse
- Support a welcoming environment

## Getting Started

1. **Fork the Repository**
   - Create your own fork of the code
   - Clone your fork locally

2. **Set Up Development Environment**
   ```bash
   git clone https://github.com/YOUR-USERNAME/cotlib.git
   cd cotlib
   go mod download
   ```

3. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Guidelines

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format code
- Follow the project's existing patterns
- Document all exported symbols
- Use meaningful variable and function names

### Testing

1. **Test Organization**
   - Integration tests go in `integration_test.go`
   - Internal tests go in `internal/internal_test.go`
   - Examples go in `examples_test.go`

2. **Test Requirements**
   - Write tests for all new features
   - Maintain existing test coverage
   - Include both positive and negative test cases
   - Test edge cases thoroughly

3. **Running Tests**
   ```bash
   go test -v ./...
   ```

### Logging

- Use `slog` for all logging
- Follow the established logging patterns:
  ```go
  logger.Debug("detailed state", "key1", value1, "key2", value2)
  logger.Info("operational message", "key", value)
  logger.Warn("unexpected condition", "error", err)
  logger.Error("critical problem", "error", err)
  ```

### Security

- Never store sensitive data in logs
- Always validate input
- Use secure XML parsing
- Consider security implications of changes
- Document security considerations

## Pull Request Process

1. **Before Submitting**
   - Update documentation
   - Add/update tests
   - Run all tests
   - Format code with `gofmt`
   - Update README.md if needed

2. **Pull Request Content**
   - Provide a clear description
   - Reference any related issues
   - List notable changes
   - Include testing steps
   - Note any breaking changes

3. **Review Process**
   - Address review feedback
   - Keep discussions focused
   - Be patient and responsive

## Commit Messages

Follow these guidelines for commit messages:

```
type(scope): Brief description

Detailed description of changes and rationale.

Fixes #123
```

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation changes
- style: Formatting changes
- refactor: Code restructuring
- test: Test changes
- chore: Maintenance tasks

## Documentation

- Update README.md for significant changes
- Document all new features
- Include examples for new functionality
- Keep documentation current
- Use clear, concise language

## Release Process

1. **Version Numbers**
   - Follow semantic versioning (MAJOR.MINOR.PATCH)
   - Document breaking changes

2. **Release Notes**
   - List significant changes
   - Note any deprecations
   - Include upgrade instructions

## Questions?

- Open an issue for general questions
- Tag security issues appropriately
- Use discussions for design proposals

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License. 