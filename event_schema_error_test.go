package cotlib

import (
	"errors"
	"testing"
)

// TestValidateAgainstSchemaInitError ensures ValidateAgainstSchema returns the
// initialization error instead of panicking when the schema failed to compile.
func TestValidateAgainstSchemaInitError(t *testing.T) {
	// Save original values
	origSchema := eventPointSchema
	origErr := initErr
	defer func() {
		eventPointSchema = origSchema
		initErr = origErr
	}()

	// Simulate compilation failure
	initErr = errors.New("compile fail")
	eventPointSchema = nil

	if err := ValidateAgainstSchema(nil); err == nil || err.Error() != initErr.Error() {
		t.Fatalf("expected %v, got %v", initErr, err)
	}

	// Call again to ensure same error returned and no panic
	if err := ValidateAgainstSchema(nil); err == nil || err.Error() != initErr.Error() {
		t.Fatalf("expected %v on second call, got %v", initErr, err)
	}
}
