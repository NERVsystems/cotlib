//go:build !cgo || novalidator

package validator

// ValidateAgainstSchema is a no-op when built without CGO/with novalidator.
// Returns nil to allow events to pass through without XSD validation.
func ValidateAgainstSchema(name string, xml []byte) error {
	return nil
}

// ListAvailableSchemas returns an empty slice when built without CGO/with novalidator.
func ListAvailableSchemas() []string {
	return nil
}
