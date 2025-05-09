package cotlib

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidationBaseline(t *testing.T) {
	t.Run("vague_error_strings", func(t *testing.T) {
		// Test for specific error messages in validateTimes
		evt := &Event{
			Time:  CoTTime(time.Now().Add(-25 * time.Hour)),
			Start: CoTTime(time.Now().Add(-25 * time.Hour)),
			Stale: CoTTime(time.Now()),
		}
		err := evt.validateTimes()
		if err == nil || !strings.Contains(err.Error(), "time must be in RFC3339 format") {
			t.Error("Expected specific error message for past time validation")
		}
	})

	t.Run("cot_type_registry_duplication", func(t *testing.T) {
		// Test for duplicate type registration
		typ := "a-f-G-E-V-test"
		RegisterCoTType(typ)
		RegisterCoTType(typ)
		if err := ValidateType(typ); err != nil {
			t.Error("Type should be registered")
		}
	})

	t.Run("wildcard_pattern_expansion", func(t *testing.T) {
		// Test wildcard pattern handling
		pattern := "a-.-G"
		if err := ValidateType(pattern); err != nil {
			t.Error("Wildcard pattern should be valid")
		}
	})

	t.Run("point_validation_mutation", func(t *testing.T) {
		// Test Point validation mutation
		p := Point{Lat: 45.0, Lon: -120.0, Hae: 100.0}
		original := p
		p.Validate()
		if p.Ce != original.Ce || p.Le != original.Le {
			t.Error("Point.Validate should not modify the receiver")
		}
	})

	t.Run("maxValueLen_race", func(t *testing.T) {
		// Test for race conditions in maxValueLen access
		go func() {
			SetMaxValueLen(2000)
		}()
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("doctype_rejection", func(t *testing.T) {
		// Test DOCTYPE rejection
		xmlData := []byte(`<?xml version="1.0"?>
<!DOCTYPE lolz [
  <!ENTITY lol "lol">
  <!ENTITY lol1 "&lol;&lol;&lol;&lol;">
  <!ENTITY lol2 "&lol1;&lol1;&lol1;&lol1;">
]>
<event></event>`)
		_, err := UnmarshalXMLEvent(xmlData)
		if !errors.Is(err, ErrInvalidInput) {
			t.Error("Expected DOCTYPE to be rejected")
		}
	})

	t.Run("logger_context", func(t *testing.T) {
		// Test logger context propagation
		ctx := context.Background()
		evt := &Event{}
		err := evt.Validate()
		if err == nil {
			t.Error("Expected validation error")
		}
		_ = ctx // Used for future context-based validation
	})

	t.Run("secure_decoder_limits", func(t *testing.T) {
		// Test decoder limits
		xmlData := []byte(`<?xml version="1.0"?>
<event>
  <detail>
    <nested1>
      <nested2>
        <nested3>
          <nested4>
            <nested5>
              <nested6>Too deep</nested6>
            </nested5>
          </nested4>
        </nested3>
      </nested2>
    </nested1>
  </detail>
</event>`)
		_, err := UnmarshalXMLEvent(xmlData)
		if err == nil {
			t.Error("Expected depth limit error")
		}
	})

	t.Run("exported_sentinels", func(t *testing.T) {
		// Test exported error sentinels
		err := ValidateLatLon(91, 0)
		if !errors.Is(err, ErrInvalidLatitude) {
			t.Error("Expected ErrInvalidLatitude")
		}
	})

	t.Run("uid_validation", func(t *testing.T) {
		// Test UID validation
		cases := []struct {
			uid     string
			wantErr bool
		}{
			{"7", false},   // Single char
			{"A_", false},  // Trailing underscore
			{"-a", true},   // Leading hyphen
			{"a..b", true}, // Double dot
			{"abc", false}, // Normal case
		}
		for _, tc := range cases {
			err := ValidateUID(tc.uid)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateUID(%q) error = %v, wantErr %v", tc.uid, err, tc.wantErr)
			}
		}
	})

	t.Run("string_sanitizer", func(t *testing.T) {
		// Test string sanitizer
		input := "<![CDATA[test]]>"
		evt := &Event{
			Detail: &Detail{
				Contact: &Contact{
					Callsign: input,
				},
			},
		}
		xmlData, err := evt.ToXML()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(xmlData), input) {
			t.Error("Sanitizer should preserve CDATA")
		}
	})

	t.Run("namespace_validation", func(t *testing.T) {
		// Test namespace validation
		xmlData := []byte(`<?xml version="1.0"?>
<event xmlns="` + strings.Repeat("x", 1025) + `">
</event>`)
		_, err := UnmarshalXMLEvent(xmlData)
		if err == nil {
			t.Error("Expected namespace length error")
		}
	})
}
