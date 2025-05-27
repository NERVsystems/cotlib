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

All TAKCoT XSD schemas are included in the repository at `validator/schemas/` and embedded at build time using Go's `go:embed` directive. This ensures:

- ✅ **Airgapped builds work** - No external dependencies required
- ✅ **Container builds work** - All files are self-contained
- ✅ **No network access needed** - Everything is embedded at compile time
- ✅ **Reproducible builds** - Schema versions are locked to the committed files

The schemas were sourced from: `../AndroidTacticalAssaultKit-CIV/takcot/xsd/`

## Available Schema Files

The repository includes the complete set of TAKCoT schemas:

-### Main Schemas
- Chat.xsd (integrated as tak-chat.xsd)
- Route.xsd
- Various marker and drawing shape schemas

### Detail Schemas  
- All schemas in `details/` directory for validating CoT detail elements
- Includes contact, emergency, status, track, remarks, and many others

### Event Schemas
- Point schema for event positioning

## Limitations

All TAKCoT schemas from the upstream repository are embedded in the validator. This includes larger schemas such as `Route.xsd` that reference other files. These cross-file dependencies are resolved automatically during validation.

## Future Enhancements

- Provide additional schema validation examples and test cases.
- Monitor the upstream ATAK repository for new TAKCoT schemas and
  integrate them as they appear.
