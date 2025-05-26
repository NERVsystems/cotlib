package validation_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/NERVsystems/cotlib"
)

func TestValidationBaseline(t *testing.T) {
	t.Run("vague_error_strings", func(t *testing.T) {
		// Test for specific error messages in validateTimes
		evt := &cotlib.Event{
			Version: "2.0",
			Uid:     "TEST-1",
			Type:    "a-f-G",
			Time:    cotlib.CoTTime(time.Now().Add(-25 * time.Hour)),
			Start:   cotlib.CoTTime(time.Now().Add(-25 * time.Hour)),
			Stale:   cotlib.CoTTime(time.Now()),
			Point: cotlib.Point{
				Lat: 37.7749,
				Lon: -122.4194,
				Hae: 100.0,
				Ce:  9999999.0,
				Le:  9999999.0,
			},
		}
		err := evt.Validate()
		if err == nil || !strings.Contains(err.Error(), "time must be within 24 hours of current time") {
			t.Error("Expected specific error message for past time validation")
		}
	})

	t.Run("cot_type_registry_duplication", func(t *testing.T) {
		// Test for duplicate type registration
		typ := "a-f-G-E-V-test"
		cotlib.RegisterCoTType(typ)
		cotlib.RegisterCoTType(typ)
		if err := cotlib.ValidateType(typ); err != nil {
			t.Error("Type should be registered")
		}
	})

	t.Run("wildcard_pattern_expansion", func(t *testing.T) {
		// Test wildcard pattern handling
		pattern := "a-.-G"
		if err := cotlib.ValidateType(pattern); err != nil {
			t.Error("Wildcard pattern should be valid")
		}
	})

	t.Run("point_validation_mutation", func(t *testing.T) {
		// Test Point validation mutation
		p := cotlib.Point{Lat: 45.0, Lon: -120.0, Hae: 100.0}
		original := p
		p.Validate()
		if p.Ce != original.Ce || p.Le != original.Le {
			t.Error("Point.Validate should not modify the receiver")
		}
	})

	t.Run("maxValueLen_race", func(t *testing.T) {
		// Test for race conditions in maxValueLen access
		go func() {
			cotlib.SetMaxValueLen(2000)
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
		_, err := cotlib.UnmarshalXMLEvent(context.Background(), xmlData)
		if !errors.Is(err, cotlib.ErrInvalidInput) {
			t.Error("Expected DOCTYPE to be rejected")
		}
	})

	t.Run("doctype_variations", func(t *testing.T) {
		// Test DOCTYPE rejection with case and spacing variations
		cases := []string{
			`<?xml version="1.0"?><!doctype foo><event></event>`,
			`<?xml version="1.0"?><!DoCtYpE foo><event></event>`,
			`<?xml version="1.0"?><!   DOCTYPE foo><event></event>`,
		}
		for _, xmlStr := range cases {
			_, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(xmlStr))
			if !errors.Is(err, cotlib.ErrInvalidInput) {
				t.Errorf("Expected DOCTYPE to be rejected for %q", xmlStr)
			}
		}
	})

	t.Run("logger_context", func(t *testing.T) {
		// Test logger context propagation
		ctx := context.Background()
		evt := &cotlib.Event{}
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
		_, err := cotlib.UnmarshalXMLEvent(context.Background(), xmlData)
		if err == nil {
			t.Error("Expected depth limit error")
		}
	})

	t.Run("exported_sentinels", func(t *testing.T) {
		// Test exported error sentinels
		err := cotlib.ValidateLatLon(91, 0)
		if !errors.Is(err, cotlib.ErrInvalidLatitude) {
			t.Error("Expected ErrInvalidLatitude")
		}

		err = cotlib.ValidateLatLon(0, 181)
		if !errors.Is(err, cotlib.ErrInvalidLongitude) {
			t.Error("Expected ErrInvalidLongitude")
		}
	})

	t.Run("uid_validation", func(t *testing.T) {
		// Test UID validation
		cases := []struct {
			uid     string
			wantErr bool
		}{
			{"7", false},                    // Single char
			{"A_", false},                   // Trailing underscore
			{"-a", true},                    // Leading hyphen
			{"a..b", true},                  // Double dot
			{"abc", false},                  // Normal case
			{strings.Repeat("x", 65), true}, // Too long
			{"has space", true},             // Contains whitespace
		}
		for _, tc := range cases {
			err := cotlib.ValidateUID(tc.uid)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateUID(%q) error = %v, wantErr %v", tc.uid, err, tc.wantErr)
			}
		}
	})

	t.Run("string_sanitizer", func(t *testing.T) {
		// Test string sanitizer
		input := "<![CDATA[test]]>"
		evt := &cotlib.Event{
			Detail: &cotlib.Detail{
				Contact: &cotlib.Contact{
					Callsign: input,
				},
			},
		}
		xmlData, err := evt.ToXML()
		if err != nil {
			t.Fatal(err)
		}
		escaped := "&lt;![CDATA[test]]&gt;"
		if !strings.Contains(string(xmlData), escaped) {
			t.Errorf("expected escaped CDATA, got %s", xmlData)
		}
	})

	t.Run("attribute_escaping", func(t *testing.T) {
		evt, err := cotlib.NewEvent("bad&\"id", "a-f-G", 10, 20, 0)
		if err != nil {
			t.Fatalf("NewEvent error: %v", err)
		}
		xmlData, err := evt.ToXML()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(xmlData), `uid="bad&amp;&quot;id"`) {
			t.Errorf("uid not escaped: %s", xmlData)
		}
	})

	t.Run("namespace_validation", func(t *testing.T) {
		// Test namespace validation
		xmlData := []byte(`<?xml version="1.0"?>
<event xmlns="` + strings.Repeat("x", 1025) + `">
</event>`)
		_, err := cotlib.UnmarshalXMLEvent(context.Background(), xmlData)
		if err == nil {
			t.Error("Expected namespace length error")
		}
	})

	t.Run("max_xml_size", func(t *testing.T) {
		big := strings.Repeat("a", (2<<20)+1)
		xmlData := []byte("<event>" + big + "</event>")
		_, err := cotlib.UnmarshalXMLEvent(context.Background(), xmlData)
		if !errors.Is(err, cotlib.ErrInvalidInput) {
			t.Error("Expected size limit error")
		}
	})

	t.Run("token_length_limit", func(t *testing.T) {
		name := strings.Repeat("x", 1025)
		xmlData := []byte("<event><" + name + "/></event>")
		_, err := cotlib.UnmarshalXMLEvent(context.Background(), xmlData)
		if !errors.Is(err, cotlib.ErrInvalidInput) {
			t.Error("Expected token length error")
		}
	})

	t.Run("element_count_limit", func(t *testing.T) {
		var b strings.Builder
		b.WriteString("<event>")
		for i := 0; i < 10001; i++ {
			b.WriteString("<a></a>")
		}
		b.WriteString("</event>")
		_, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(b.String()))
		if !errors.Is(err, cotlib.ErrInvalidInput) {
			t.Error("Expected element count error")
		}
	})
}

func TestWildcardPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{
			name:    "valid wildcard",
			pattern: "a-f-G-U-C-*",
			wantErr: false,
		},
		{
			name:    "invalid wildcard",
			pattern: "a-f-G-U-C-*-*",
			wantErr: true,
		},
		{
			name:    "invalid wildcard position",
			pattern: "a-f-G-U-*-C",
			wantErr: true,
		},
		{
			name:    "embedded wildcard segment",
			pattern: "a-f*G",
			wantErr: true,
		},
		{
			name:    "embedded wildcard final segment",
			pattern: "a-f-G*X",
			wantErr: true,
		},
		{
			name:    "suffix after wildcard",
			pattern: "a-*X",
			wantErr: true,
		},
		{
			name:    "mid wildcard suffix",
			pattern: "a-f-G*-Z",
			wantErr: true,
		},
		{
			name:    "leading wildcard",
			pattern: "*-a-f",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cotlib.ValidateType(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateType(%q) expected error", tt.pattern)
				} else if !errors.Is(err, cotlib.ErrInvalidType) {
					t.Errorf("ValidateType(%q) unexpected error = %v", tt.pattern, err)
				}
			} else if err != nil {
				t.Errorf("ValidateType(%q) unexpected error = %v", tt.pattern, err)
			}
		})
	}
}
