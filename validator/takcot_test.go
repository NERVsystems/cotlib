package validator

import (
	"testing"
)

func TestTAKCoTSchemas(t *testing.T) {
	// Test that all TAKCoT detail schemas are available
	expectedSchemas := []string{
		"tak-details-contact",
		"tak-details-emergency",
		"tak-details-status",
		"tak-details-track",
		"tak-details-remarks",
	}

	for _, schemaName := range expectedSchemas {
		t.Run(schemaName, func(t *testing.T) {
			// Try to validate an empty XML document to ensure schema loads
			err := ValidateAgainstSchema(schemaName, []byte("<root></root>"))
			// We expect validation to fail (since we're using invalid XML),
			// but we should not get an "unknown schema" error
			if err != nil && err.Error() == "unknown schema "+schemaName {
				t.Errorf("Schema %s not found", schemaName)
			}
		})
	}
}

func TestTAKCoTSchemaAvailability(t *testing.T) {
	// Initialize schemas
	initSchemas()

	// Check that TAKCoT detail schemas are loaded
	takSchemas := []string{
		"tak-details-contact",
		"tak-details-emergency",
		"tak-details-status",
		"tak-details-track",
		"tak-details-remarks",
	}

	for _, schemaName := range takSchemas {
		if _, exists := schemas[schemaName]; !exists {
			t.Errorf("TAKCoT schema %s not loaded", schemaName)
		}
	}
}

func TestListAvailableSchemas(t *testing.T) {
	schemas := ListAvailableSchemas()

	// Should have at least the original schemas plus TAKCoT schemas
	expectedMinimum := []string{
		"chat",
		"chatReceipt",
		"tak-details-contact",
		"tak-details-emergency",
		"tak-details-status",
		"tak-details-track",
		"tak-details-remarks",
	}

	schemaMap := make(map[string]bool)
	for _, schema := range schemas {
		schemaMap[schema] = true
	}

	for _, expected := range expectedMinimum {
		if !schemaMap[expected] {
			t.Errorf("Expected schema %s not found in available schemas", expected)
		}
	}

	t.Logf("Available schemas: %v", schemas)
}
