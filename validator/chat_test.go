package validator_test

import (
	"testing"

	"github.com/NERVsystems/cotlib/validator"
)

func TestValidateChat(t *testing.T) {
	valid := []byte(`<__chat sender="Alice" message="hello"/>`)
	if err := validator.ValidateChat(valid); err != nil {
		t.Fatalf("valid chat rejected: %v", err)
	}

	invalid := []byte(`<__chat sender="Alice"/>`)
	if err := validator.ValidateChat(invalid); err == nil {
		t.Fatal("expected error for missing message")
	}
}
