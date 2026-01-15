# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

cotlib is a Go library for creating, validating, and working with Cursor-on-Target (CoT) events. CoT is an XML-based protocol used for situational awareness, originally developed by MITRE and extended by TAK (Team Awareness Kit) applications.

## Build and Test Commands

```bash
# Build the library (requires CGO for libxml2-based schema validation)
go build ./...

# Build without schema validation (no CGO required)
go build -tags novalidator ./...

# Run all tests
go test -v ./...

# Run tests for a specific package
go test -v ./cottypes
go test -v ./validator

# Run a single test
go test -v -run TestEventCreation ./...

# Run benchmarks
go test -bench=. ./...

# Regenerate type catalog (after modifying XML type definitions)
go generate ./cottypes

# Format code before committing
gofmt -s -w $(git ls-files '*.go')

# Run security scanner
go install github.com/securego/gosec/v2/cmd/gosec@latest && gosec -no-fail ./...
```

## Architecture

### Core Package (`cotlib`)

The main package in `cotlib.go` provides:
- `Event` struct: The primary data structure representing a CoT message with XML marshaling/unmarshaling
- `NewEvent()`, `NewPresenceEvent()`: Event constructors
- `EventBuilder`: Fluent API for constructing events with `NewEventBuilder().WithContact().WithGroup().Build()`
- `UnmarshalXMLEvent()`: Parses XML into Event with security limits and validation
- `ValidateType()`, `ValidateHow()`, `ValidateRelation()`: Validation functions for CoT type codes
- Event pooling via `getEvent()`/`ReleaseEvent()` for reduced allocations

### Subpackages

- **`cottypes/`**: Type catalog system with ~10k CoT type definitions
  - `Catalog`: Thread-safe registry with zero-allocation lookups
  - `generated_types.go`: Auto-generated from `CoTtypes.xml` (MITRE) and `TAKtypes.xml` (TAK extensions)
  - `GetCatalog()`: Returns the singleton catalog instance
  - How/relation value mappings for position source and relationship types

- **`validator/`**: XML Schema (XSD) validation using libxml2 via CGO
  - Embedded schemas in `schemas/` directory for TAK detail extensions
  - `ValidateAgainstSchema(name, xml)`: Validates against named schema
  - Build with `-tags novalidator` to disable (returns nil for all validations)

- **`ctxlog/`**: Context-based slog logger propagation

- **`cmd/cotgen/`**: Generator for `cottypes/generated_types.go` from XML definitions

### Key Data Structures

The `Detail` struct contains pointers to ~30 TAK extension types (Chat, Track, Shape, Marti, etc.) that are validated against embedded XSD schemas during `Event.Validate()`.

### Type System

- MITRE types use `a-` prefix (atoms) with affiliation codes: `f` (friendly), `h` (hostile), `n` (neutral), `u` (unknown)
- TAK types use `TAK/` namespace prefix in FullName (e.g., `b-t-f` for file transfer)
- Wildcard validation: `a-.-G` matches any affiliation, `a-f-G-*` matches subtypes

## Code Conventions

- Use `slog` for logging via context: `LoggerFromContext(ctx).Info("msg")`
- Commit messages: `type(scope): Brief description` (feat, fix, docs, refactor, test, chore)
- Integration tests go in `integration_test.go`, examples in `examples_test.go`
- XML security: Always check for DOCTYPE with `doctypePattern.Match(data)` before parsing

## CGO Dependency

Schema validation requires libxml2 and CGO. The `novalidator` build tag provides a stub implementation that skips all schema validation - use only for trusted input or when CGO is unavailable.
