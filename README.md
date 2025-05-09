# CoT Library

A Go library for working with Cursor-on-Target (CoT) messages.

## Features

- Full CoT type catalog with metadata
- Type validation and registration
- Secure logging with slog
- Thread-safe catalog system
- Wildcard pattern support
- Search by description or full name

## Installation

```bash
go get github.com/NERVsystems/cotlib
```

## Usage

### Basic Event Creation

```go
package main

import (
    "log/slog"
    "os"
    "github.com/NERVsystems/cotlib"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    
    // Register a custom CoT type
    if err := cotlib.RegisterCoTType("a-f-G-U-C-F"); err != nil {
        logger.Error("Failed to register type", "error", err)
        return
    }
    
    // Validate a CoT type
    if valid := cotlib.ValidateType("a-f-G-U-C-F"); valid {
        logger.Info("Type is valid")
    }
}
```

### Type Catalog Operations

```go
package main

import (
    "fmt"
    "log"
    "github.com/NERVsystems/cotlib"
)

func main() {
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

### Thread Safety

All operations in the library are thread-safe. The type catalog uses internal synchronization to ensure safe concurrent access.

### Type Validation

The library enforces strict validation of CoT types:
- Basic syntax checking
- Standard prefix validation
- Length limits
- Wildcard pattern validation

### Custom Types

You can register custom type codes that extend the standard prefixes:

```go
// Register a custom type
cotlib.RegisterCoTType("a-f-G-E-V-custom")

// Validate the custom type
if err := cotlib.ValidateType("a-f-G-E-V-custom"); err != nil {
    log.Fatal(err)
}
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

## Documentation

For detailed documentation and examples, see:
- [GoDoc](https://pkg.go.dev/github.com/NERVsystems/cotlib)
- [CoT Specification](https://www.mitre.org/sites/default/files/pdf/09_4937.pdf)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 