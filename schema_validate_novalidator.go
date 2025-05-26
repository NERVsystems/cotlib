//go:build novalidator

package cotlib

// ValidateAgainstSchema is a no-op when built with the 'novalidator' tag.
// It immediately returns nil.
func ValidateAgainstSchema(data []byte) error {
	return nil
}
