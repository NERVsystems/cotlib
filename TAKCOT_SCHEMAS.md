# TAKCoT Schema Integration

This project now includes embedded XSD schemas from the [AndroidTacticalAssaultKit-CIV](https://github.com/deptofdefense/AndroidTacticalAssaultKit-CIV) repository's `takcot` directory.

## Available TAKCoT Schemas

The following TAKCoT detail schemas are available for validation:

- `tak-details-contact` - Contact information schema
- `tak-details-emergency` - Emergency event schema  
- `tak-details-status` - Status information schema
- `tak-details-track` - Track/movement schema
- `tak-details-remarks` - Remarks/comments schema

## Usage

### Validating XML against TAKCoT schemas

```go
import "github.com/NERVsystems/cotlib/validator"

// Validate XML against a TAKCoT schema
err := validator.ValidateAgainstSchema("tak-details-contact", xmlData)
if err != nil {
    // Handle validation error
}
```

### Listing available schemas

```go
schemas := validator.ListAvailableSchemas()
fmt.Printf("Available schemas: %v\n", schemas)
```

## Schema Sources

The schemas are sourced from the TAKCoT directory via a symbolic link to:
`../AndroidTacticalAssaultKit-CIV/takcot/xsd/`

The schemas are embedded at build time using Go's `go:embed` directive, so no external files are needed at runtime.

## Limitations

Currently, only self-contained detail schemas are included. Complex schemas with dependencies (like the main Chat.xsd and Route.xsd) require additional work to resolve their includes and dependencies.

## Future Enhancements

- Add support for complex schemas with dependencies
- Include more TAKCoT schemas as needed
- Add schema validation examples and test cases 