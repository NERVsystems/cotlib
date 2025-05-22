![Cursor On Target](cotlogo.png)

'The sum of all wisdom is a cursor over the target.'  — Gen. John Jumper

# CoT Library

[![Go Report Card](https://goreportcard.com/badge/github.com/NERVsystems/cotlib)](https://goreportcard.com/report/github.com/NERVsystems/cotlib)



A comprehensive Go library for creating, validating, and working with Cursor-on-Target (CoT) events.

## Features

- Complete CoT event creation and manipulation
- XML serialization and deserialization with security protections
- Full CoT type catalog with metadata
- Coordinate and spatial data handling
- Event relationship management
- Type validation and registration
- Secure logging with slog
- Thread-safe operations
- Predicate-based event classification
- Security-first design
- Wildcard pattern support for types
- Type search by description or full name

## Installation

```bash
go get github.com/NERVsystems/cotlib
```

## Usage

### Creating and Managing CoT Events

```go
package main

import (
    "fmt"
    "log/slog"
    "os"
    "github.com/NERVsystems/cotlib"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Create a new CoT event
    event, err := cotlib.NewEvent("UNIT-123", "a-f-G", 37.422, -122.084, 0.0)
    if err != nil {
        logger.Error("Failed to create event", "error", err)
        return
    }

    // Add detail information
    event.Detail = &cotlib.Detail{
        Contact: &cotlib.Contact{
            Callsign: "ALPHA-7",
        },
        Group: &cotlib.Group{
            Name: "Team Blue",
            Role: "Infantry",
        },
    }

    // Add relationship link
    event.AddLink(&cotlib.Link{
        Uid:      "HQ-1",
        Type:     "a-f-G-U-C",
        Relation: "p-p",
    })

    // Convert to XML
    xmlData, err := event.ToXML()
    if err != nil {
        logger.Error("Failed to convert to XML", "error", err)
        return
    }

    fmt.Println(string(xmlData))
}
```

### Parsing CoT XML

```go
package main

import (
    "fmt"
    "github.com/NERVsystems/cotlib"
)

func main() {
    xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<event version="2.0" uid="UNIT-123" type="a-f-G" time="2023-05-15T18:30:22Z"
       start="2023-05-15T18:30:22Z" stale="2023-05-15T18:30:32Z">
  <point lat="37.422000" lon="-122.084000" hae="0.0" ce="9999999.0" le="9999999.0"/>
  <detail>
    <contact callsign="ALPHA-7"/>
    <group name="Team Blue" role="Infantry"/>
  </detail>
</event>`

    // Parse XML into CoT event
    event, err := cotlib.UnmarshalXMLEvent([]byte(xmlData))
    if err != nil {
        fmt.Printf("Error parsing XML: %v\n", err)
        return
    }

    // Access event data
    fmt.Printf("Event Type: %s\n", event.Type)
    fmt.Printf("Location: %.6f, %.6f\n", event.Point.Lat, event.Point.Lon)
    fmt.Printf("Callsign: %s\n", event.Detail.Contact.Callsign)

    // Check event predicates
    if event.Is("friend") {
        fmt.Println("This is a friendly unit")
    }

    if event.Is("ground") {
        fmt.Println("This is a ground-based entity")
    }
}
```

### Type Validation and Catalog

The library provides comprehensive type validation and catalog management:

```go
package main

import (
    "fmt"
    "log"
    "github.com/NERVsystems/cotlib"
)

func main() {
    // Register a custom CoT type
    if err := cotlib.RegisterCoTType("a-f-G-U-C-F"); err != nil {
        log.Fatal(err)
    }

    // Validate a CoT type
    if err := cotlib.ValidateType("a-f-G-U-C-F"); err != nil {
        log.Fatal(err)
    }

    // Look up type metadata
    fullName, err := cotlib.GetTypeFullName("a-f-G-E-X-N")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Full name: %s\n", fullName)
    // Output: Full name: Gnd/Equip/Nbc Equipment

    // Get type description
    desc, err := cotlib.GetTypeDescription("a-f-G-E-X-N")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Description: %s\n", desc)
    // Output: Description: NBC EQUIPMENT

    // Search for types by description
    types := cotlib.FindTypesByDescription("NBC")
    for _, t := range types {
        fmt.Printf("Found type: %s (%s)\n", t.Name, t.Description)
    }

    // Search for types by full name
    types = cotlib.FindTypesByFullName("Equipment")
    for _, t := range types {
        fmt.Printf("Found type: %s (%s)\n", t.Name, t.FullName)
    }
}
```

`catalog.Upsert` precomputes upper-case versions of each type's `FullName` and
`Description`. `FindByDescription` and `FindByFullName` reuse these cached
strings so searches are allocation-free.

### Type Validation

The library enforces strict validation of CoT types:
- Basic syntax checking
- Standard prefix validation
- Length limits
- Wildcard pattern validation
- Type catalog verification

```go
// Examples of different validation scenarios:
cotlib.ValidateType("a-f-G")             // Valid - Friendly Ground
cotlib.ValidateType("a-h-A")             // Valid - Hostile Air
cotlib.ValidateType("b-d")               // Valid - Bits, Detection
cotlib.ValidateType("t-x-takp-v")        // Valid - TAK presence
cotlib.ValidateType("a-f-G-*")           // Valid - Wildcard pattern
cotlib.ValidateType("a-.-G")             // Valid - Atomic wildcard
cotlib.ValidateType("invalid")           // Invalid - Fails validation
cotlib.ValidateType("a-f-G-U-*-C")       // Invalid - Wildcard in wrong position
```

### Custom Types

You can register custom type codes that extend the standard prefixes:

```go
// Register a custom type
cotlib.RegisterCoTType("a-f-G-E-V-custom")

// Validate the custom type
if err := cotlib.ValidateType("a-f-G-E-V-custom"); err != nil {
    log.Fatal(err)
}

// Register types from a file
if err := cotlib.RegisterCoTTypesFromFile("my-types.xml"); err != nil {
    log.Fatal(err)
}

// Register types from a string
xmlContent := `<types>
    <cot cot="a-f-G-custom"/>
    <cot cot="a-h-A-custom"/>
</types>`
if err := cotlib.RegisterCoTTypesFromXMLContent(xmlContent); err != nil {
    log.Fatal(err)
}
```

### Generating Type Metadata (`cotgen`)

The `cmd/cotgen` utility expands the CoT XML definitions and writes the
`cottypes/generated_types.go` file used by the library. Ensure the
`cot-types` directory (or your own XML definitions) is present, then run:

```bash
go run ./cmd/cotgen
# or simply
go generate ./cottypes
```

Add your custom type entries to `cot-types/CoTtypes.xml` before running the
generator to embed them into the resulting Go code.

### Event Predicates

The library provides convenient type classification with the `Is()` method:

```go
// Create a friendly ground unit event
event, _ := cotlib.NewEvent("test123", "a-f-G", 30.0, -85.0, 0.0)

// Check various predicates
fmt.Printf("Is friendly: %v\n", event.Is("friend"))  // true
fmt.Printf("Is hostile: %v\n", event.Is("hostile")) // false
fmt.Printf("Is ground: %v\n", event.Is("ground"))   // true
fmt.Printf("Is air: %v\n", event.Is("air"))         // false
```

### Thread Safety

All operations in the library are thread-safe. The type catalog uses internal synchronization to ensure safe concurrent access.

### Security Features

The library implements several security measures:

- XML parsing restrictions to prevent XXE attacks
- Input validation on all fields
- Coordinate range enforcement
- Time field validation to prevent time-based attacks
- Maximum value length controls
- Configurable parser limits
- Secure logging practices

```go
// Set maximum allowed length for XML attribute values
// This protects against memory exhaustion attacks
cotlib.SetMaxValueLen(500 * 1024) // 500KB limit
cotlib.SetMaxXMLSize(2 << 20)    // 2MB overall XML size
cotlib.SetMaxElementDepth(32)    // nesting depth limit
cotlib.SetMaxElementCount(10000) // total element limit
cotlib.SetMaxTokenLen(1024)      // single token size
```

### Logging

The library uses `slog` for structured logging:
- Debug level for detailed operations
- Info level for normal events
- Warn level for recoverable issues
- Error level for critical problems

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

ctx := cotlib.WithLogger(context.Background(), logger)
```

### Event Pooling

`UnmarshalXMLEvent` reuses `Event` objects from an internal pool to reduce
allocations. When you are done with an event, return it to the pool:

```go
evt, _ := cotlib.UnmarshalXMLEvent(data)
defer cotlib.ReleaseEvent(evt)
```

## Benchmarks

Run benchmarks with the standard Go tooling:

```bash
go test -bench=. ./...
```

This executes any `Benchmark...` functions across the module, allowing you to
profile serialization, validation, or other operations.

## Documentation

For detailed documentation and examples, see:
- [GoDoc](https://pkg.go.dev/github.com/NERVsystems/cotlib)
- [CoT Specification](https://www.mitre.org/sites/default/files/pdf/09_4937.pdf)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Project History

Originally created by [@pdfinn](https://github.com/pdfinn).
All core functionality and initial versions developed prior to organisational transfer.
