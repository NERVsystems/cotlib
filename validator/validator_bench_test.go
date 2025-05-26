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

func BenchmarkValidateAgainstSchemaEnvironment(b *testing.B) {
	xml := []byte(`<environment temperature="20" windDirection="10" windSpeed="5"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-environment", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}
