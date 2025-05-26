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
			name:   "video",
			schema: "tak-details-__video",
			good:   []byte(`<__video url="http://x"/>`),
			bad:    []byte(`<__video/>`),
		},
		{
			name:   "fileshare",
			schema: "tak-details-fileshare",
			good:   []byte(`<fileshare filename="f" name="n" senderCallsign="A" senderUid="U" senderUrl="http://x" sha256="h" sizeInBytes="1"/>`),
			bad:    []byte(`<fileshare filename="f"/>`),
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
	schemas := validator.ListAvailableSchemas()
	if len(schemas) == 0 {
		t.Fatal("no schemas returned")
	}
}
