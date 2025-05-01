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

## Usage Examples

Basic usage example:

```go
// Create a new friendly ground unit
evt := cotlib.NewEvent("UNIT1", cotlib.TypePredFriend+"-G", 45.0, -120.0)
if evt == nil {
    log.Fatal("failed to create event")
}

// Add detail extensions
evt.DetailContent.Contact.Callsign = "ALPHA1"
evt.DetailContent.Shape = struct {
    Type   string  `xml:"type,attr,omitempty"`
    Points string  `xml:"points,attr,omitempty"`
    Radius float64 `xml:"radius,attr,omitempty"`
}{
    Type:   "circle",
    Radius: 1000, // meters
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

## Transport Considerations

CoT typically uses one of these transport patterns:

1. **Open-Squirt-Close (Recommended)**
   - One event per TCP connection
   - Open connection, send event, close connection
   - Simple and reliable
   - Example:
     ```go
     evt := cotlib.NewEvent(...)
     if evt == nil {
         return fmt.Errorf("failed to create event")
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
     evt := cotlib.NewEvent(...)
     if evt == nil {
         return fmt.Errorf("failed to create event")
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

This library uses Height Above Ellipsoid (HAE) exclusively, as recommended by the CoT specification. The `Point` struct validates HAE values to be within reasonable bounds (-12,000m to +9,000m, accommodating Mariana Trench to Mount Everest).

## Testing

The library includes comprehensive tests:

1. **Unit Tests** (`cotlib_test.go`)
   - Tests all public API functions
   - Validates XML marshaling/unmarshaling
   - Verifies time validation rules
   - Tests type predicates and event linking
   - Ensures logging functionality

2. **Integration Tests**
   - Tests real-world usage scenarios
   - Verifies XML round-trip functionality
   - Tests all detail extensions

Run tests with:
```bash
go test -v ./...
```

## Best Practices

1. **Event Creation**
   - Always use `NewEvent()` to ensure proper initialization
   - Check for nil return value
   - Use predefined type predicates (e.g., `TypePredFriend`)
   - Set reasonable stale times (> 5 seconds from event time)

2. **Validation**
   - Always check `Validate()` before transmission
   - Handle validation errors appropriately
   - Use context-aware logging for debugging

3. **Detail Extensions**
   - Use provided structs for known extensions
   - Validate detail content size
   - Keep extensions minimal and focused

4. **Security**
   - Use structured logging with appropriate levels
   - Handle all errors explicitly
   - Validate all input data
   - Use context for logger injection

## License

MIT License - See LICENSE file

## Contributing

Contributions welcome! Please read CONTRIBUTING.md first. 