package validator_test

import (
	"testing"

	"github.com/NERVsystems/cotlib/validator"
)

func TestValidateAgainstSchemaNonet(t *testing.T) {
	validator.ResetForTest()
	good := []byte(`<__chat sender="Alice" message="hi"/>`)
	if err := validator.ValidateAgainstSchema("chat", good); err != nil {
		t.Fatalf("valid chat rejected: %v", err)
	}

	bad := []byte(`<!DOCTYPE foo SYSTEM "http://example.com/foo.dtd"><__chat sender="Alice" message="hi">&ext;</__chat>`)
	if err := validator.ValidateAgainstSchema("chat", bad); err == nil {
		t.Fatal("expected error for external entity")
	}
}

func TestValidateAgainstTAKDetailSchemas(t *testing.T) {
	validator.ResetForTest()
	tests := []struct {
		name   string
		schema string
		good   []byte
		bad    []byte
	}{
		{
			name:   "contact",
			schema: "tak-details-contact",
			good:   []byte(`<contact callsign="A"/>`),
			bad:    []byte(`<contact/>`),
		},
		{
			name:   "track",
			schema: "tak-details-track",
			good:   []byte(`<track course="90" speed="10"/>`),
			bad:    []byte(`<track speed="10"/>`),
		},
		{
			name:   "status",
			schema: "tak-details-status",
			good:   []byte(`<status battery="80"/>`),
			bad:    []byte(`<status battery="bad"/>`),
		},
		{
			name:   "environment",
			schema: "tak-details-environment",
			good:   []byte(`<environment temperature="1" windDirection="2" windSpeed="3"/>`),
			bad:    []byte(`<environment temperature="1" windDirection="2"/>`),
		},
		{
			name:   "shape",
			schema: "tak-details-shape",
			good:   []byte(`<shape><polyline closed="false"><vertex lat="0" lon="0" hae="0"/></polyline></shape>`),
			bad:    []byte(`<shape><polyline><vertex lat="0" lon="0"/></polyline></shape>`),
		},
		{
			name:   "color",
			schema: "tak-details-color",
			good:   []byte(`<color argb="1"/>`),
			bad:    []byte(`<color/>`),
		},
		{
			name:   "usericon",
			schema: "tak-details-usericon",
			good:   []byte(`<usericon iconsetpath="foo"/>`),
			bad:    []byte(`<usericon/>`),
		},
		{
			name:   "mission",
			schema: "tak-details-mission",
			good:   []byte(`<mission name="op" tool="t" type="x"/>`),
			bad:    []byte(`<mission name="op" tool="t"/>`),
		},
		{
			name:   "attachment_list",
			schema: "tak-details-attachment_list",
			good:   []byte(`<attachment_list hashes="abc"/>`),
			bad:    []byte(`<attachment_list/>`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.ValidateAgainstSchema(tt.schema, tt.good); err != nil {
				t.Fatalf("valid %s rejected: %v", tt.name, err)
			}
			if err := validator.ValidateAgainstSchema(tt.schema, tt.bad); err == nil {
				t.Fatalf("expected error for invalid %s", tt.name)
			}
		})
	}
}

func TestListAvailableSchemas(t *testing.T) {
	validator.ResetForTest()
	schemas := validator.ListAvailableSchemas()
	if len(schemas) == 0 {
		t.Fatal("no schemas returned")
	}
}
