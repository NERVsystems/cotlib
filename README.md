# CoT Library for Go

A Go implementation of the Cursor on Target (CoT) protocol, focusing on security, extensibility, and standards compliance.

## Features

- Full support for CoT base schema (Event.xsd)
- Secure XML parsing with XXE prevention and size limits
- Structured logging via slog with context support
- Type predicate support (CoTtypes.xml patterns)
- Event linking for complex relationships
- Detail extensions with validation (shape, remarks, contact, status, flow-tags, uid-aliases)
- Height Above Ellipsoid (HAE) for all elevations
- Comprehensive time validation with replay attack prevention
- XML round-trip support with secure marshaling/unmarshaling
- Strict validation of UID, Type, and coordinates
- Semantic validation of shape parameters
- Consistent error wrapping with sentinel errors

## Security Features

- Input validation on all fields with detailed error messages
- Size limits on XML content (2 MiB max) and detail content (1 MiB max)
- Secure XML parsing with:
  - XXE prevention
  - Entity expansion limits
  - Maximum element depth (32)
  - Maximum element count (10,000)
  - Token length limits (1024 bytes)
- Time-based attack prevention:
  - Minimum stale time offset (5 seconds)
  - Maximum stale time offset (7 days)
  - Past/future time bounds (24 hours)
- String sanitization for all XML content
- Structured logging with sensitive data protection
- Context-aware logging support
- Strict validation of presence events and system messages
- Semantic validation of shape parameters (radius, points)

## Installation

```bash
go get github.com/NERVsystems/cotlib
```

## CoT Types

This library implements a comprehensive set of CoT types directly embedded in the code for fast and secure validation. The types are derived from the official CoT specification and cover all common use cases.

### Type Validation

The library provides fast, secure type validation through embedded types:

```go
// Validate a CoT type
err := cotlib.ValidateType("a-f-G") // Friendly ground unit
if err != nil {
    log.Fatal(err)
}

// Register a custom type if needed
cotlib.RegisterCoTType("a-c-my-custom-type")
```

The embedded types include:
- All standard tactical symbols (a-*)
- Common bits types (b-*)
- Message types (t-*)
- Tasking types (t-k, t-s, etc.)
- Reply types (y-*)
- Capability types (c-*)

### Type Predicates

The library provides type predicates for common patterns:

```go
// Check if an event is a friendly unit
if evt.Is("friend") {
    // Handle friendly unit
}

// Check if an event is a ground track
if evt.Is("ground") {
    // Handle ground track
}
```

Available predicates:
- `atom`: Any type starting with "a"
- `friend`: Friendly force (a-f)
- `hostile`: Hostile force (a-h)
- `unknown`: Unknown force (a-u)
- `neutral`: Neutral force (a-n)
- `ground`: Ground track (-G)
- `air`: Air track (-A)
- `pending`: Pending/planned track (-P)

## Usage Examples

Basic usage example:

```go
// Create a new friendly ground unit
evt, err := cotlib.NewEvent("UNIT1", cotlib.TypePredFriend+"-G", 45.0, -120.0, 0.0)
if err != nil {
    log.Fatal(err)
}

// Add detail extensions
evt.Detail = &cotlib.Detail{
    Contact: &cotlib.Contact{
        Callsign: "ALPHA1",
    },
    Shape: &cotlib.Shape{
        Type:   "circle",
        Radius: 1000, // meters
    },
}

// Validate the event
if err := evt.Validate(); err != nil {
    log.Fatal(err)
}

// Marshal to XML
xmlData, err := evt.ToXML()
if err != nil {
    log.Fatal(err)
}
```

See `examples_test.go` for more examples including:
- Creating and validating events
- Working with type predicates
- Adding links between events
- Using detail extensions
- Context-based logging
- Custom stale time configuration
- Presence event handling

## Time Validation

The library enforces strict time validation rules:

1. **Event Time**
   - Must be within 24 hours of current time
   - Used as reference for start/stale validation

2. **Start Time**
   - Must be before or equal to event time
   - Used to indicate when the event becomes valid

3. **Stale Time**
   - Must be more than 5 seconds after event time
   - Must be within 7 days of event time
   - Used to prevent replay attacks
   - Customizable via `WithStale()` method

## Transport Considerations

CoT typically uses one of these transport patterns:

1. **Open-Squirt-Close (Recommended)**
   - One event per TCP connection
   - Open connection, send event, close connection
   - Simple and reliable
   - Example:
     ```go
     evt, err := cotlib.NewEvent(...)
     if err != nil {
         return fmt.Errorf("failed to create event: %w", err)
     }
     conn, err := net.Dial("tcp", "target:port")
     if err != nil {
         return err
     }
     defer conn.Close()
     data, err := evt.ToXML()
     if err != nil {
         return err
     }
     _, err = conn.Write(data)
     return err
     ```

2. **UDP Multicast**
   - One event per UDP packet
   - Efficient for "SA spam" (frequent position updates)
   - Best for situational awareness data
   - Example:
     ```go
     evt, err := cotlib.NewEvent(...)
     if err != nil {
         return fmt.Errorf("failed to create event: %w", err)
     }
     addr, err := net.ResolveUDPAddr("udp", "239.2.3.1:6969")
     if err != nil {
         return err
     }
     conn, err := net.DialUDP("udp", nil, addr)
     if err != nil {
         return err
     }
     data, err := evt.ToXML()
     if err != nil {
         return err
     }
     _, err = conn.Write(data)
     return err
     ```

## Height Representation

This library uses Height Above Ellipsoid (HAE) exclusively, as recommended by the CoT specification. The `Point` struct validates HAE values to be within reasonable bounds (-12,000m to +999,999m, accommodating Mariana Trench to space systems).

## Testing

The library includes comprehensive tests:

1. **Unit Tests** (`cotlib_test.go`)
   - Tests all public API functions
   - Validates XML marshaling/unmarshaling
   - Verifies time validation rules
   - Tests type predicates and event linking
   - Ensures logging functionality
   - Validates shape parameters
   - Tests presence event handling

2. **Integration Tests**
   - Tests real-world usage scenarios
   - Verifies XML round-trip functionality
   - Tests all detail extensions
   - Validates error handling

Run tests with:
```bash
go test -v ./...
```

## Best Practices

1. **Event Creation**
   - Always use `NewEvent()` to ensure proper initialization
   - Check for errors returned by `NewEvent()`
   - Use predefined type predicates (e.g., `TypePredFriend`)
   - Set reasonable stale times using `WithStale()`

2. **Validation**
   - Always check `Validate()` before transmission
   - Handle validation errors appropriately
   - Use context-aware logging for debugging
   - Check shape parameters for semantic validity

3. **Detail Extensions**
   - Use provided structs for known extensions
   - Validate detail content size
   - Keep extensions minimal and focused
   - Ensure shape parameters are valid

4. **Security**
   - Use structured logging with appropriate levels
   - Handle all errors explicitly
   - Validate all input data
   - Use context for logger injection
   - Check for presence event validity

## License

MIT License - See LICENSE file

## Contributing

Contributions welcome! Please read CONTRIBUTING.md first. 