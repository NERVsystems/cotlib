package cotlib

import (
	"testing"
)

func BenchmarkNewEvent(b *testing.B) {
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := UnmarshalXMLEvent(xmlData); err != nil {
			b.Fatalf("UnmarshalXMLEvent error: %v", err)
		}
	}
}
