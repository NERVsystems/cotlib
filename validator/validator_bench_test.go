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

func BenchmarkValidateAgainstSchemaStatus(b *testing.B) {
	xml := []byte(`<status battery="80"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-status", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func BenchmarkValidateAgainstSchemaVideo(b *testing.B) {
	xml := []byte(`<__video url="http://x"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-__video", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func BenchmarkValidateAgainstSchemaMission(b *testing.B) {
	xml := []byte(`<mission name="op" tool="t" type="x"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-mission", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func BenchmarkValidateAgainstSchemaTakv(b *testing.B) {
	xml := []byte(`<takv platform="Android" version="1"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-takv", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func BenchmarkValidateAgainstSchemaBullseye(b *testing.B) {
	xml := []byte(`<bullseye mils="true" distance="10" bearingRef="T" bullseyeUID="b" distanceUnits="u-r-b-bullseye" edgeToCenter="false" rangeRingVisible="true" title="t" hasRangeRings="false"/>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-bullseye", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func BenchmarkValidateAgainstSchemaRouteInfo(b *testing.B) {
	xml := []byte(`<__routeinfo><__navcues/></__routeinfo>`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validator.ValidateAgainstSchema("tak-details-routeinfo", xml); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}
