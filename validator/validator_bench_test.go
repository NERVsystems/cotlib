package validator_test

import (
	"testing"

	"github.com/NERVsystems/cotlib/validator"
)

func BenchmarkValidateAgainstSchemaContact(b *testing.B) {
	xml := []byte(`<contact callsign="A"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-contact", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func BenchmarkValidateAgainstSchemaTrack(b *testing.B) {
	xml := []byte(`<track course="90" speed="10"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-track", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func BenchmarkValidateAgainstSchemaColor(b *testing.B) {
	xml := []byte(`<color argb="1"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-color", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}
