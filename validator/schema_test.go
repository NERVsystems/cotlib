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
			good:   []byte(`<environment temperature="20" windDirection="10" windSpeed="5"/>`),
			bad:    []byte(`<environment temperature="20" windSpeed="5"/>`),
		},
		{
			name:   "fileshare",
			schema: "tak-details-fileshare",
			good:   []byte(`<fileshare filename="f" name="n" senderCallsign="A" senderUid="U" senderUrl="http://x" sha256="h" sizeInBytes="1"/>`),
			bad:    []byte(`<fileshare filename="f" name="n" senderCallsign="A" senderUid="U" senderUrl="http://x" sha256="h"/>`),
		},
		{
			name:   "precisionlocation",
			schema: "tak-details-precisionlocation",
			good:   []byte(`<precisionlocation altsrc="GPS"/>`),
			bad:    []byte(`<precisionlocation/>`),
		},
		{
			name:   "takv",
			schema: "tak-details-takv",
			good:   []byte(`<takv platform="Android" version="1"/>`),
			bad:    []byte(`<takv platform="Android"/>`),
		},
		{
			name:   "mission",
			schema: "tak-details-mission",
			good:   []byte(`<mission name="m" tool="t" type="x"/>`),
			bad:    []byte(`<mission tool="t" type="x"/>`),
		},
		{
			name:   "shape",
			schema: "tak-details-shape",
			good:   []byte(`<shape><polyline closed="true"><vertex hae="0" lat="1" lon="1"/></polyline></shape>`),
			bad:    []byte(`<shape><polyline closed="true"></polyline></shape>`),
		},
		{
			name:   "color",
			schema: "tak-details-color",
			good:   []byte(`<color argb="1"/>`),
			bad:    []byte(`<color/>`),
		},
		{
			name:   "__chat",
			schema: "tak-details-__chat",
			good:   []byte(`<__chat chatroom="c" groupOwner="false" id="1" senderCallsign="s"><chatgrp id="g" uid0="u"/></__chat>`),
			bad:    []byte(`<__chat chatroom="c"><chatgrp id="g" uid0="u"/></__chat>`),
		},
		{
			name:   "__chatreceipt",
			schema: "tak-details-__chatreceipt",
			good:   []byte(`<__chatreceipt chatroom="c" groupOwner="false" id="1" senderCallsign="s"><chatgrp id="g" uid0="u"/></__chatreceipt>`),
			bad:    []byte(`<__chatreceipt chatroom="c" groupOwner="false"><chatgrp id="g" uid0="u"/></__chatreceipt>`),
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
	schemas := validator.ListAvailableSchemas()
	if len(schemas) == 0 {
		t.Fatal("no schemas returned")
	}
}
