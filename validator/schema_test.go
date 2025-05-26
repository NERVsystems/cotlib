package validator_test

import (
	"testing"

	"github.com/NERVsystems/cotlib/validator"
)

func TestValidateAgainstSchemaNonet(t *testing.T) {
	good := []byte(`<__chat sender="Alice" message="hi"/>`)
	if err := validator.ValidateAgainstSchema("chat", good); err != nil {
		t.Fatalf("valid chat rejected: %v", err)
	}

	bad := []byte(`<!DOCTYPE foo SYSTEM "http://example.com/foo.dtd"><__chat sender="Alice" message="hi">&ext;</__chat>`)
	if err := validator.ValidateAgainstSchema("chat", bad); err == nil {
		t.Fatal("expected error for external entity")
	}
}
