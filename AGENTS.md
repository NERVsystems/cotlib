# Contribution and Coding Guidelines

These instructions apply to the entire repository.

## Development
- Follow standard Go conventions and existing patterns.
- Format all Go code using `gofmt -s -w` before committing.
- Document all exported identifiers.
- Use meaningful variable and function names.
- Use the `slog` package for logging following the examples in `CONTRIBUTING.md`.
- Validate all input and avoid logging sensitive data.

## Testing
- Write tests for all new features and maintain coverage.
- Integration tests belong in `integration_test.go`.
- Internal tests belong in `internal/internal_test.go`.
- Examples go in `examples_test.go`.
- Run `go test -v ./...` and `go test -bench .` (when benchmarking) before submitting a PR.

## Documentation
- Update `README.md` and other docs for significant changes.
- Include examples for new functionality.

## Commit Messages
- Use the format `type(scope): Brief description`.
- Provide a detailed explanation if necessary.
- Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`.

## Pull Requests
- Ensure code is formatted and all tests pass.
- Provide a clear description of changes and reference related issues.
- List notable changes and any breaking changes.
