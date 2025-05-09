# CoT Library

A Go library for working with Cursor-on-Target (CoT) messages.

## Features

- Type validation and registration
- Secure logging with slog
- Thread-safe catalog system
- Wildcard pattern support

## Installation

```bash
go get github.com/NERVsystems/cotlib
```

## Usage

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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 