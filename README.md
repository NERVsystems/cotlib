![Cursor On Target](cotlogo.png)

'…we want the target dead or saved…we gotta get away from platform centric thinking…and we gotta focus on this thing where the sum of the wisdom is a cursor over the target…and we're indifferent [to the source]'  — Gen. John Jumper

# CoT Library

[![Go Report Card](https://goreportcard.com/badge/github.com/NERVsystems/cotlib)](https://goreportcard.com/report/github.com/NERVsystems/cotlib)



A comprehensive Go library for creating, validating, and working with Cursor-on-Target (CoT) events.

## Features

- **High-performance processing**: Sub-microsecond event creation, millions of validations/sec
- Complete CoT event creation and manipulation
- XML serialization and deserialization with security protections
- Full CoT type catalog with metadata
- **Zero-allocation type lookups** and optimized memory usage
- **How and relation value support** with comprehensive validation
- Coordinate and spatial data handling
- Event relationship management
- Type validation and registration
- Secure logging with slog
- Thread-safe operations
- Detail extensions with round-trip preservation
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
    "errors"
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

#### Handling Detail Extensions

CoT events often include TAK-specific extensions inside the `<detail>` element.
`cotlib` now preserves several common extensions and round-trips them
unchanged:

- `__chat`
- `__chatReceipt`
- `__geofence`
- `__serverdestination`
- `__video`
- `__group`
- `archive`
- `attachmentList`
- `environment`
- `fileshare`
- `precisionlocation`
- `takv`
- `track`
- `mission`
- `status`
- `shape`

The `__chat` and `__chatReceipt` extensions are validated against embedded
XML schemas when decoding and during event validation. Invalid chat XML will
cause an error.

Example: adding a `shape` extension with a `strokeColor` attribute:

```go
event.Detail = &cotlib.Detail{
    Shape: &cotlib.Shape{Raw: []byte(`<shape strokeColor="#00FF00"/>`)},
}
```

Any unknown elements are stored in `Detail.Unknown` and serialized back
verbatim.

```go
xmlData := `<?xml version="1.0"?>
<event version="2.0" uid="EXT-1" type="t-x-c" time="2023-05-15T18:30:22Z" start="2023-05-15T18:30:22Z" stale="2023-05-15T18:30:32Z">
  <point lat="0" lon="0" ce="9999999.0" le="9999999.0"/>
  <detail>
    <__chat message="Hello world!" sender="Alpha"/>
    <__video url="http://example/video"/>
  </detail>
</event>`

evt, _ := cotlib.UnmarshalXMLEvent([]byte(xmlData))
out, _ := evt.ToXML()
fmt.Println(string(out)) // prints the same XML
```

### Type Validation and Catalog

The library provides comprehensive type validation and catalog management:

```go
package main

import (
    "errors"
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
        if errors.Is(err, cotlib.ErrInvalidType) {
            log.Fatal(err)
        }
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
- Automatic resolution of `f`, `h`, `n`, or `u` segments to catalog
  entries containing `.`

```go
// Examples of different validation scenarios:
cotlib.ValidateType("a-f-G")             // Valid - Friendly Ground
cotlib.ValidateType("b-m-r")             // Valid - Route
cotlib.ValidateType("invalid")           // Error - Unknown type
```

### How and Relation Values

The library provides full support for CoT how values (indicating position source) and relation values (for event relationships):

#### How Values

How values indicate the source or method of position determination:

```go
package main

import (
    "errors"
    "fmt"
    "log"
    "github.com/NERVsystems/cotlib"
)

func main() {
    // Create an event
    event, _ := cotlib.NewEvent("UNIT-123", "a-f-G", 37.422, -122.084, 0.0)
    
    // Set how value using descriptor (recommended)
    err := cotlib.SetEventHowFromDescriptor(event, "gps")
    if err != nil {
        log.Fatal(err)
    }
    // This sets event.How to "h-g-i-g-o"
    
    // Or set directly if you know the code
    event.How = "h-e" // manually entered
    
    // Validate how value
    if err := cotlib.ValidateHow(event.How); err != nil {
        if errors.Is(err, cotlib.ErrInvalidHow) {
            log.Fatal(err)
        }
    }
    
    // Get human-readable description
    desc, _ := cotlib.GetHowDescriptor("h-g-i-g-o")
    fmt.Printf("How: %s\n", desc) // Output: How: gps
}
```

#### Relation Values

Relation values specify the relationship type in link elements:

```go
// Add a validated link with parent-point relation
err := event.AddValidatedLink("HQ-1", "a-f-G-U-C", "p-p")
if err != nil {
    if errors.Is(err, cotlib.ErrInvalidRelation) {
        log.Fatal(err)
    }
}

// Or add manually (validation happens during event.Validate())
event.AddLink(&cotlib.Link{
    Uid:      "CHILD-1",
    Type:     "a-f-G",
    Relation: "p-c", // parent-child
})

// Validate relation value
if err := cotlib.ValidateRelation("p-c"); err != nil {
    if errors.Is(err, cotlib.ErrInvalidRelation) {
        log.Fatal(err)
    }
}

// Get relation description
desc, _ := cotlib.GetRelationDescription("p-p")
fmt.Printf("Relation: %s\n", desc) // Output: Relation: parent-point
```

#### Available Values

**How values include:**
- `h-e` (manual entry)
- `h-g-i-g-o` (GPS)
- `m-g` (GPS - MITRE)
- And many others from both MITRE and TAK specifications

**Relation values include:**
- `c` (connected)
- `p-p` (parent-point)
- `p-c` (parent-child)  
- `p` (parent - MITRE)
- And many others from both MITRE and TAK specifications

#### Validation

Event validation automatically checks how and relation values:

```go
event.How = "invalid-how"
err := event.Validate() // Will fail

event.AddLink(&cotlib.Link{
    Uid:      "test",
    Type:     "a-f-G", 
    Relation: "invalid-relation",
})
err = event.Validate() // Will fail
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

ctx := cotlib.WithLogger(context.Background(), logger)

// Register types from a file
if err := cotlib.RegisterCoTTypesFromFile(ctx, "my-types.xml"); err != nil {
    log.Fatal(err)
}

// Register types from a string
xmlContent := `<types>
    <cot cot="a-f-G-custom"/>
    <cot cot="a-h-A-custom"/>
</types>`
if err := cotlib.RegisterCoTTypesFromXMLContent(ctx, xmlContent); err != nil {
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

## TAK Types and Extensions

The library supports both canonical MITRE CoT types and TAK-specific extensions. TAK types are maintained separately to ensure clear namespace separation and avoid conflicts with official MITRE specifications.

### Adding New CoT Types

**For MITRE/canonical types:** Add entries to `cot-types/CoTtypes.xml`
**For TAK-specific types:** Add entries to `cot-types/TAKtypes.xml`

The generator automatically discovers and processes all `*.xml` files in the `cot-types/` directory.

### TAK Namespace

All TAK-specific types use the `TAK/` namespace prefix in their `full` attribute to distinguish them from MITRE types:

```xml
<!-- TAK-specific types in cot-types/TAKtypes.xml -->
<cot cot="b-t-f" full="TAK/Bits/File" desc="File Transfer" />
<cot cot="u-d-f" full="TAK/Drawing/FreeForm" desc="Free Form Drawing" />
<cot cot="t-x-c" full="TAK/Chat/Message" desc="Chat Message" />
```

### Working with TAK Types

```go
// Check if a type is TAK-specific
typ, err := cottypes.GetCatalog().GetType("b-t-f")
if err != nil {
    log.Fatal(err)
}

if cottypes.IsTAK(typ) {
    fmt.Printf("%s is a TAK type: %s\n", typ.Name, typ.FullName)
    // Output: b-t-f is a TAK type: TAK/Bits/File
}

// Search for TAK types specifically
takTypes := cottypes.GetCatalog().FindByFullName("TAK/")
fmt.Printf("Found %d TAK types\n", len(takTypes))

// Validate TAK types
if err := cotlib.ValidateType("b-t-f"); err != nil {
    log.Fatal(err) // TAK types are fully validated
}
```

### Generator Workflow

1. The generator scans `cot-types/*.xml` for type definitions
2. Parses each XML file into the standard `<types><cot>` structure  
3. Validates TAK namespace integrity (no `a-` prefixes with `TAK/` full names)
4. Expands MITRE wildcards (`a-.-`) but leaves TAK types unchanged
5. Generates `cottypes/generated_types.go` with all types

### Adding New Types

To add new CoT types to the catalog:

1. **For MITRE types:** Edit `cot-types/CoTtypes.xml`
2. **For TAK extensions:** Edit `cot-types/TAKtypes.xml` 
3. **For new categories:** Create a new XML file in `cot-types/`
4. Run `go generate ./cottypes` to regenerate the catalog
5. Verify with tests: `go test ./cottypes -v -run TestTAK`

Example TAK type entry:
```xml
<cot cot="b-m-p-c-z" full="TAK/Map/Zone" desc="Map Zone" />
```

**Important:** TAK types should never use the `a-` prefix (reserved for MITRE affiliation-based types) and must always use the `TAK/` namespace prefix.

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
- UID values are limited to 64 characters and may not contain whitespace
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
log := cotlib.LoggerFromContext(ctx)
log.Info("logger ready")
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

## Performance

This library is optimized for high-performance CoT processing with minimal memory allocations:

### Core Operations (Apple M4)

| Operation | Speed | Allocations | Memory | Throughput |
|-----------|--------|-------------|---------|------------|
| **Event Creation** | 156.7 ns/op | 1 alloc | 240 B | ~6.4M events/sec |
| **XML Generation** | 558.9 ns/op | 4 allocs | 360 B | ~1.8M events/sec |
| **XML Parsing** | 4.9 μs/op | 73 allocs | 3.4 KB | ~200K events/sec |

### Type Validation

| Type Pattern | Speed | Allocations | Throughput |
|--------------|--------|-------------|------------|
| **Simple Types** (`a-f-G`) | 22.6 ns/op | 0 allocs | ~44M validations/sec |
| **Complex Types** (`a-f-G-E-X-N`) | 26.7 ns/op | 0 allocs | ~37M validations/sec |  
| **Wildcards** (`a-f-G-*`) | 41.4 ns/op | 1 alloc | ~24M validations/sec |

### Catalog Operations

| Operation | Speed | Allocations | Throughput |
|-----------|--------|-------------|------------|
| **Type Lookup** | 17.3 ns/op | 0 allocs | ~57M lookups/sec |
| **Search by Description** | 72.4 μs/op | 4 allocs | ~13K searches/sec |
| **Search by Full Name** | 104.7 μs/op | 5 allocs | ~9K searches/sec |

### Key Performance Features

- **Zero-allocation lookups**: Type catalog operations don't allocate memory
- **Object pooling**: XML parsing reuses event objects to minimize GC pressure  
- **Optimized validation**: Fast-path validation for common type patterns
- **Efficient searching**: Pre-computed uppercase strings for case-insensitive search
- **Minimal serialization overhead**: Direct byte buffer manipulation for XML generation

### Real-World Scenarios

**High-frequency tracking**: Process 200,000+ position updates per second
**Bulk operations**: Validate millions of type codes with zero GC impact  
**Memory-constrained environments**: Minimal allocation footprint
**Low-latency systems**: Sub-microsecond event processing

*Benchmarks run on Apple M4 with Go 1.21. Your mileage may vary by platform.*

## Documentation

For detailed documentation and examples, see:
- [GoDoc](https://pkg.go.dev/github.com/NERVsystems/cotlib)
- [CoT Specification](https://www.mitre.org/sites/default/files/pdf/09_4937.pdf)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Project History

Originally created by [@pdfinn](https://github.com/pdfinn).
All core functionality and initial versions developed prior to organisational transfer.
