package validator_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/NERVsystems/cotlib/validator"
)

func TestValidateAgainstSchemaInitErrors(t *testing.T) {
	t.Run("mktemp", func(t *testing.T) {
		validator.ResetForTest()
		validator.SetMkTempForTest(func(string, string) (string, error) {
			return "", errors.New("mktemp fail")
		})
		err := validator.ValidateAgainstSchema("chat", nil)
		if err == nil || !strings.Contains(err.Error(), "mktemp fail") {
			t.Fatalf("expected mktemp error, got %v", err)
		}
		if err2 := validator.ValidateAgainstSchema("chat", nil); err2 == nil || err2.Error() != err.Error() {
			t.Fatalf("expected same error on subsequent call, got %v", err2)
		}
	})

	t.Run("write", func(t *testing.T) {
		validator.ResetForTest()
		validator.SetWriteSchemasForTest(func(string) error {
			return errors.New("write fail")
		})
		err := validator.ValidateAgainstSchema("chat", nil)
		if err == nil || !strings.Contains(err.Error(), "write schemas") {
			t.Fatalf("expected write error, got %v", err)
		}
	})
}

func TestInitSchemasRemovesTempDir(t *testing.T) {
	validator.ResetForTest()
	var dir string
	validator.SetMkTempForTest(func(string, string) (string, error) {
		var err error
		dir, err = os.MkdirTemp("", "cotlib-test")
		return dir, err
	})

	if err := validator.ValidateAgainstSchema("chat", []byte(`<__chat sender="A" message="hi"/>`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("temp dir not removed: %v", err)
	}
}
