# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.3.6] - 2025-06-02

### Bug Fixes
- Fixed Go vet issues with unexported struct fields that had XML tags but were not used for marshaling/unmarshaling
- Resolved struct field validation errors in `detail_extensions.go` for better code quality and compliance

### Security
- Conducted comprehensive security scan with gosec showing clean codebase with no critical vulnerabilities
- All detected issues were in build cache files rather than application code, confirming secure implementation
- Enhanced CI/CD pipeline security practices with proper gosec integration

### Improvements
- Improved code quality by addressing static analysis warnings
- Enhanced development workflow with better linting and security scanning integration

## [v0.3.5] - 2025-01-27

### Added
- Added remarks element schema validation and comprehensive tests
- Enhanced TAK chat fallback test coverage for improved validation reliability

### Bug Fixes
- Fixed schema initialization error handling to properly return initialization errors
- Improved error propagation in validator initialization

### Improvements
- Strengthened validation test suite with additional TAK chat fallback scenarios
- Enhanced schema validation coverage for remarks elements

## [v0.3.4] - 2025-05-29

### Improvements
- Enhanced chat validation with comprehensive schema support for both simple and TAK-specific chat messages
- Improved event pool tests to be more resilient to parallel execution and race conditions
- Added better test coverage for event pool reuse scenarios after XML parsing failures
- Strengthened validation for chat extensions with dual schema support (simple chat and TAK chat)
- Added `remarks` element schema derived from the MITRE CoT release, allowing validation via `tak-details-remarks`
- Added regression test ensuring chat messages with `<chatgrp>` fall back to the TAK-specific schema

### Bug Fixes
- Fixed race condition in event pool tests that caused failures when running with race detection
- Resolved test interference issues when running multiple tests in parallel
- Improved event pool test reliability by using multiple baseline events instead of single object identity checks

### Security
- Maintained secure event pool management with proper cleanup on validation failures
- Ensured no resource leaks in concurrent event processing scenarios

## [v0.2.3] - 2025-05-11

### Breaking Changes
- Security constants (`MaxDetailSize`, `MinStaleOffset`, `MaxStaleOffset`, etc.) are now unexported to prevent API breakage in future versions. External callers should use the validation methods instead.
- Detail struct now preserves XML attributes for known elements (shape, remarks, contact, etc.) instead of storing them as plain strings.
- Unknown elements in Detail are now properly captured during unmarshaling, improving round-trip fidelity.

### Improvements
- Enhanced XML security with proper handling of unknown elements and attributes
- Improved size validation using sync.Pool for buffer reuse
- Unified regex patterns for type and UID validation
- Simplified predicate logic with consistent anchoring
- Fixed duplicate tag issues in Detail marshaling
- Added comprehensive docstrings that pass go vet
- Refactored `ValidateType` to use the type catalog for validation, improving consistency and reliability
- Added improved test coverage for CoT type validation

### Bug Fixes
- Fixed UnknownElements capture during Detail unmarshaling
- Corrected attribute preservation in Detail struct
- Removed redundant typePredicateMap
- Fixed regex patterns for better type string validation
- Fixed test cases to handle empty FullName and Description fields in certain CoT types

## [v0.1.1] - 2024-03-14

### Changed
- Improved error handling in NewEvent function to return errors instead of nil
- Added init function to ensure a safe no-op logger is always available
- Fixed test cases for TAK system message validation

### Security
- Improved error handling to prevent potential nil pointer dereferences
- Enhanced logging safety with guaranteed no-op logger

## [v0.1.0] - 2024-03-13

### Added
- Initial release of the CoT library
- Basic CoT event creation and validation
- XML marshaling and unmarshaling
- Secure logging practices
- Input validation and sanitization
- Support for detail extensions
- Type predicate checking
- Event linking capabilities

## [v0.3.2] - 2025-05-27

### Added
- Comprehensive TAKCoT schema validation, including complex schemas such as `Route.xsd`
- Benchmarks and tests covering the expanded schema set

### Changed
- Secure XML decoding enforced across the library
- Type catalog lookup now supports wildcard entries

### Fixed
- XML escape handling for newline and tab characters
