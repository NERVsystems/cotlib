package validator_test

import (
	"strings"
	"testing"

	"github.com/NERVsystems/cotlib/validator"
)

func TestValidateAgainstSchemaCompileError(t *testing.T) {
	validator.ResetForTest()
	orig := validator.EventPointXSD()
	validator.SetEventPointXSDForTest(nil)
	t.Cleanup(func() {
		validator.SetEventPointXSDForTest(orig)
		validator.ResetForTest()
	})
	err := validator.ValidateAgainstSchema("event-point", []byte(`<point lat="1" lon="1" hae="0" ce="0" le="0"/>`))
	if err == nil || !strings.Contains(err.Error(), "compile event point schema") {
		t.Fatalf("expected compile error, got %v", err)
	}
	if err2 := validator.ValidateAgainstSchema("event-point", nil); err2 == nil || err2.Error() != err.Error() {
		t.Fatalf("expected same error on subsequent call, got %v", err2)
	}
}
