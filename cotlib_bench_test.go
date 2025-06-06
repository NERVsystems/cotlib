package cotlib

import (
	"bytes"
	"context"
	"encoding/xml"
	"testing"
)

func BenchmarkNewEvent(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewEvent("bench", "a-f-G", 30.0, -85.0, 0.0); err != nil {
			b.Fatalf("NewEvent returned error: %v", err)
		}
	}
}

func BenchmarkToXML(b *testing.B) {
	evt, err := NewEvent("bench", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		b.Fatalf("NewEvent returned error: %v", err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := evt.ToXML(); err != nil {
			b.Fatalf("ToXML error: %v", err)
		}
	}
}

func BenchmarkUnmarshalXMLEvent(b *testing.B) {
	evt, err := NewEvent("bench", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		b.Fatalf("NewEvent returned error: %v", err)
	}
	xmlData, err := evt.ToXML()
	if err != nil {
		b.Fatalf("ToXML error: %v", err)
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := UnmarshalXMLEvent(ctx, xmlData); err != nil {
			b.Fatalf("UnmarshalXMLEvent error: %v", err)
		}
	}
}

func BenchmarkValidateType(b *testing.B) {
	types := []string{
		"a-f-G",       // Valid basic type
		"a-f-G-E-X-N", // Valid catalog type
		"a-f-G-*",     // Valid wildcard
		"a-.-X",       // Valid atomic wildcard
		"invalid",     // Invalid
		"a-f-INVALID", // Invalid format
		"a-f-G-Z-Z-Z", // Unknown but valid format
	}
	b.ReportAllocs()
	for _, typ := range types {
		b.Run(typ, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ValidateType(typ)
			}
		})
	}
}

func BenchmarkDecodeWithLimits(b *testing.B) {
	xmlStr := `<event version="2.0" uid="bench" type="a-f-G"><point lat="1" lon="2" hae="3"/></event>`
	data := []byte(xmlStr)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dec := xml.NewDecoder(bytes.NewReader(data))
		evt := Event{}
		if err := decodeWithLimits(dec, &evt); err != nil {
			b.Fatalf("decodeWithLimits error: %v", err)
		}
	}
}
