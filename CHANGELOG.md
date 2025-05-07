# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

## [Unreleased]

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

### Bug Fixes
- Fixed UnknownElements capture during Detail unmarshaling
- Corrected attribute preservation in Detail struct
- Removed redundant typePredicateMap
- Fixed regex patterns for better type string validation 