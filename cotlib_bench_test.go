package cotlib

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
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
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := UnmarshalXMLEvent(xmlData); err != nil {
			b.Fatalf("UnmarshalXMLEvent error: %v", err)
		}
	}
}

func BenchmarkUnmarshalXMLEventNoPool(b *testing.B) {
	evt, err := NewEvent("bench", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		b.Fatalf("NewEvent returned error: %v", err)
	}
	xmlData, err := evt.ToXML()
	if err != nil {
		b.Fatalf("ToXML error: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := unmarshalXMLEventNoPool(xmlData); err != nil {
			b.Fatalf("decode error: %v", err)
		}
	}
}

func unmarshalXMLEventNoPool(data []byte) (*Event, error) {
	if doctypePattern.Match(data) {
		return nil, ErrInvalidInput
	}
	if idx := bytes.Index(data, []byte(`xmlns="`)); idx >= 0 {
		end := bytes.Index(data[idx+7:], []byte(`"`))
		if end > 1024 {
			return nil, ErrInvalidInput
		}
	}
	if err := checkXMLLimits(data); err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(io.LimitReader(bytes.NewReader(data), int64(len(data))))
	dec.CharsetReader = nil
	dec.Entity = nil
	var evt Event
	if err := dec.Decode(&evt); err != nil {
		return nil, fmt.Errorf("failed to decode XML: %w", err)
	}
	if err := evt.Validate(); err != nil {
		return nil, err
	}
	return &evt, nil
}
