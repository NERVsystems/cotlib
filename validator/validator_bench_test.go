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

func BenchmarkValidateAgainstSchemaEnvironment(b *testing.B) {
	xml := []byte(`<environment temperature="20" windDirection="10" windSpeed="5"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-environment", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func BenchmarkValidateAgainstSchemaPrecisionLocation(b *testing.B) {
	xml := []byte(`<precisionlocation altsrc="GPS"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-precisionlocation", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func BenchmarkValidateAgainstSchemaShape(b *testing.B) {
	xml := []byte(`<shape><polyline closed="true"><vertex lat="1" lon="1" hae="0"/></polyline></shape>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-shape", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func BenchmarkValidateAgainstSchemaEventPoint(b *testing.B) {
	xml := []byte(`<point lat="1" lon="2" hae="0" ce="0" le="0"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("event-point", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}
