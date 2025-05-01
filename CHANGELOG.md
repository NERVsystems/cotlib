# Changelog

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