//go:build !novalidator

package cotlib

// ValidateAgainstSchema validates the given XML against the CoT schema.
// This default implementation currently performs no validation and
// returns nil.
func ValidateAgainstSchema(data []byte) error {
	return nil
}
