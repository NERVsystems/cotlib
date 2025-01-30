# CoT Library for Go

A Go implementation of the Cursor on Target (CoT) protocol, focusing on security, extensibility, and standards compliance.

## Features

- Full support for CoT base schema (Event.xsd)
- Secure XML parsing with XXE prevention
- Structured logging via slog
- Type predicate support (CoTtypes.xml patterns)
- Event linking for complex relationships
- Common sub-schema support (shape, request, flow-tags, uid)
- Height Above Ellipsoid (HAE) for all elevations

## Security Features

- Input validation on all fields
- Size limits on detail content
- Secure XML parsing (no XXE)
- Time-based attack prevention
- Structured logging with sensitive data protection

## Usage Examples

The library includes comprehensive examples in `examples_test.go` demonstrating common use cases:

```go
// Create a new friendly ground unit
evt := cotlib.NewEvent("UNIT1", cotlib.TypePredFriend+"-G", 45.0, -120.0)
evt.How = "m-g" // GPS measurement

// Add detail extensions
evt.DetailContent.UidAliases = &cotlib.UidAliases{
    Callsign: "ALPHA1",
    Platform: "HMMWV",
}

// Add a shape for area of operations
evt.DetailContent.Shape = &cotlib.Shape{
    Type:   "circle",
    Radius: 1000, // meters
}

// Marshal to XML
xmlData, err := evt.ToXML()
```

See `examples_test.go` for more examples including:
- Creating and validating events
- Working with type predicates
- Adding links between events
- Using detail extensions

## Transport Considerations

CoT typically uses one of these transport patterns:

1. **Open-Squirt-Close (Recommended)**
   - One event per TCP connection
   - Open connection, send event, close connection
   - Simple and reliable
   - Example:
     ```go
     evt := cotlib.NewEvent(...)
     conn, _ := net.Dial("tcp", "target:port")
     defer conn.Close()
     data, _ := evt.ToXML()
     conn.Write(data)
     ```

2. **UDP Multicast**
   - One event per UDP packet
   - Efficient for "SA spam" (frequent position updates)
   - Best for situational awareness data
   - Example:
     ```go
     evt := cotlib.NewEvent(...)
     addr, _ := net.ResolveUDPAddr("udp", "239.2.3.1:6969")
     conn, _ := net.DialUDP("udp", nil, addr)
     data, _ := evt.ToXML()
     conn.Write(data)
     ```

3. **Persistent Connections (Not Recommended)**
   - Requires additional framing protocol
   - More complex error handling
   - Not typical in CoT deployments

## Height Representation

This library uses Height Above Ellipsoid (HAE) exclusively, as recommended by the CoT specification. If you need Mean Sea Level (MSL) or Above Ground Level (AGL), implement these in a custom detail sub-schema rather than modifying the base `point` element.

## Testing

The library uses a well-organized testing structure:

1. **Integration Tests** (`integration_test.go`)
   - Tests the public API from a consumer's perspective
   - Verifies high-level functionality like XML marshaling
   - Tests real-world usage scenarios

2. **Internal Tests** (`internal/internal_test.go`)
   - Tests internal implementation details
   - Focuses on validation and edge cases
   - Verifies internal invariants

3. **Examples** (`examples_test.go`)
   - Provides documented examples of common use cases
   - Serves as both tests and documentation
   - Shows idiomatic usage patterns

Run tests with:
```bash
go test -v ./...
```

## Best Practices

1. **Event Creation**
   - Always use `NewEvent()` to ensure proper initialization
   - Validate events before transmission
   - Use type predicates instead of hard-coded strings

2. **Linking**
   - Use `AddLink()` to create relationships
   - Keep relationship graphs acyclic
   - Document link types and relations

3. **Detail Extensions**
   - Create typed structs for known sub-schemas
   - Use `RawXML` only for experimental fields
   - Validate detail content size

4. **Transport**
   - Prefer Open-Squirt-Close for reliability
   - Use UDP for high-frequency updates
   - Implement appropriate timeouts

5. **Security**
   - Always validate input
   - Use structured logging
   - Handle errors appropriately
   - Protect sensitive data

## License

MIT License - See LICENSE file

## Contributing

Contributions welcome! Please read CONTRIBUTING.md first. 