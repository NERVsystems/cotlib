package cotlib_test

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/NERVsystems/cotlib"
	"github.com/NERVsystems/cotlib/validator"
)

func TestWildcardPatterns(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected bool
	}{
		{"valid wildcard", "a-f-G-U-C-*", true},
		{"invalid wildcard", "a-f-G-U-C-", false},
		{"invalid wildcard position", "a-f-G-U-*-C", false},
		{"embedded wildcard segment", "a-f*G", false},
		{"embedded wildcard final segment", "a-f-G*X", false},
		{"multiple wildcard segments", "a-f-G-*-*", false},
		{"double asterisk segment", "a-f-G**", false},
		{"multi-embedded asterisks", "a*f*G", false},
		{"trailing dash", "a-f-G-", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cotlib.ValidateType(tt.pattern)
			if tt.expected {
				if err != nil {
					t.Errorf("ValidateType(%q) unexpected error = %v", tt.pattern, err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateType(%q) expected error", tt.pattern)
				} else if !errors.Is(err, cotlib.ErrInvalidType) {
					t.Errorf("ValidateType(%q) unexpected error = %v", tt.pattern, err)
				}
			}
		})
	}
}

// TestValidateHow tests the ValidateHow function with various inputs.
func TestValidateHow(t *testing.T) {
	testCases := []struct {
		name      string
		how       string
		expectErr bool
	}{
		{"empty_how_valid", "", false},
		{"valid_tak_gps", "h-g-i-g-o", false},
		{"valid_tak_manual", "h-e", false},
		{"valid_mitre_gps", "m-g", false},
		{"invalid_how", "invalid-how", true},
		{"nonexistent_how", "x-x-x", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cotlib.ValidateHow(tc.how)
			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error for how value %s, but got none", tc.how)
				} else if !errors.Is(err, cotlib.ErrInvalidHow) {
					t.Errorf("error does not wrap ErrInvalidHow: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for how value %s, but got: %v", tc.how, err)
				}
			}
		})
	}
}

// TestValidateRelation tests the ValidateRelation function with various inputs.
func TestValidateRelation(t *testing.T) {
	testCases := []struct {
		name      string
		relation  string
		expectErr bool
	}{
		{"empty_relation_invalid", "", true},
		{"valid_connected", "c", false},
		{"valid_parent_point", "p-p", false},
		{"valid_parent_child", "p-c", false},
		{"valid_mitre_parent", "p", false},
		{"invalid_relation", "invalid-rel", true},
		{"nonexistent_relation", "x-x-x", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cotlib.ValidateRelation(tc.relation)
			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error for relation value %s, but got none", tc.relation)
				} else if !errors.Is(err, cotlib.ErrInvalidRelation) {
					t.Errorf("error does not wrap ErrInvalidRelation: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for relation value %s, but got: %v", tc.relation, err)
				}
			}
		})
	}
}

// TestEventValidationWithHowAndRelation tests event validation including how and relation fields.
func TestEventValidationWithHowAndRelation(t *testing.T) {
	t.Run("valid_event_with_how", func(t *testing.T) {
		event, err := cotlib.NewEvent("test-123", "a-f-G", 30.0, -85.0, 0.0)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}

		// Set valid how value
		event.How = "h-g-i-g-o"

		// Should validate successfully
		if err := event.Validate(); err != nil {
			t.Errorf("Event validation failed: %v", err)
		}
	})

	t.Run("invalid_event_with_bad_how", func(t *testing.T) {
		event, err := cotlib.NewEvent("test-123", "a-f-G", 30.0, -85.0, 0.0)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}

		// Set invalid how value
		event.How = "invalid-how"

		// Should fail validation
		if err := event.Validate(); err == nil {
			t.Error("Expected validation to fail with invalid how value")
		}
	})

	t.Run("valid_event_with_links", func(t *testing.T) {
		event, err := cotlib.NewEvent("test-123", "a-f-G", 30.0, -85.0, 0.0)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}

		// Add valid link
		err = event.AddValidatedLink("parent-123", "a-f-G-U-C", "p-p")
		if err != nil {
			t.Fatalf("Failed to add validated link: %v", err)
		}

		// Should validate successfully
		if err := event.Validate(); err != nil {
			t.Errorf("Event validation failed: %v", err)
		}
	})

	t.Run("invalid_event_with_bad_relation", func(t *testing.T) {
		event, err := cotlib.NewEvent("test-123", "a-f-G", 30.0, -85.0, 0.0)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}

		// Add link with invalid relation
		event.AddLink(&cotlib.Link{
			Uid:      "test-uid",
			Type:     "a-f-G",
			Relation: "invalid-relation",
		})

		// Should fail validation
		if err := event.Validate(); err == nil {
			t.Error("Expected validation to fail with invalid relation")
		}
	})
}

// TestSetEventHowFromDescriptor tests the convenience function for setting how values.
func TestSetEventHowFromDescriptor(t *testing.T) {
	event, err := cotlib.NewEvent("test-123", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	testCases := []struct {
		name        string
		descriptor  string
		expectedHow string
		expectErr   bool
	}{
		{"gps_descriptor", "gps", "h-g-i-g-o", false},
		{"manual_descriptor", "manual", "h-e", false},
		{"entered_descriptor", "entered", "h-e", false},
		{"invalid_descriptor", "nonexistent", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cotlib.SetEventHowFromDescriptor(event, tc.descriptor)

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error for descriptor %s, but got none", tc.descriptor)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for descriptor %s, but got: %v", tc.descriptor, err)
				}
				if event.How != tc.expectedHow {
					t.Errorf("Expected how value %s, got %s", tc.expectedHow, event.How)
				}
			}
		})
	}

	t.Run("nil_event", func(t *testing.T) {
		if err := cotlib.SetEventHowFromDescriptor(nil, "gps"); err == nil {
			t.Error("Expected error for nil event, but got none")
		}
	})
}

// TestAddValidatedLink tests the AddValidatedLink method.
func TestAddValidatedLink(t *testing.T) {
	event, err := cotlib.NewEvent("test-123", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	t.Run("valid_link", func(t *testing.T) {
		err := event.AddValidatedLink("parent-123", "a-f-G-U-C", "p-p")
		if err != nil {
			t.Errorf("Expected no error for valid link, but got: %v", err)
		}

		// Check that link was added
		if len(event.Links) == 0 {
			t.Error("Link was not added to event")
		}
	})

	t.Run("invalid_type", func(t *testing.T) {
		err := event.AddValidatedLink("test-uid", "invalid-type", "p-p")
		if err == nil {
			t.Error("Expected error for invalid link type, but got none")
		}
	})

	t.Run("invalid_relation", func(t *testing.T) {
		err := event.AddValidatedLink("test-uid", "a-f-G", "invalid-relation")
		if err == nil {
			t.Error("Expected error for invalid relation, but got none")
		}
	})

	t.Run("nil_event", func(t *testing.T) {
		var nilEvent *cotlib.Event
		if err := nilEvent.AddValidatedLink("uid", "a-f-G", "c"); err == nil {
			t.Error("Expected error for nil event, but got none")
		}
	})
}

// TestHowRelationDescriptors tests the descriptor helper functions.
func TestHowRelationDescriptors(t *testing.T) {
	t.Run("get_how_descriptor", func(t *testing.T) {
		desc, err := cotlib.GetHowDescriptor("h-g-i-g-o")
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		if desc != "gps" {
			t.Errorf("Expected 'gps', got '%s'", desc)
		}

		desc, err = cotlib.GetHowDescriptor("h-e")
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		if desc != "manual" {
			t.Errorf("Expected 'manual', got '%s'", desc)
		}

		_, err = cotlib.GetHowDescriptor("invalid")
		if err == nil {
			t.Error("Expected error for invalid how code, but got none")
		}
	})

	t.Run("get_relation_description", func(t *testing.T) {
		desc, err := cotlib.GetRelationDescription("c")
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		if desc != "connected" {
			t.Errorf("Expected 'connected', got '%s'", desc)
		}

		desc, err = cotlib.GetRelationDescription("p-p")
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		if desc != "parent-point" {
			t.Errorf("Expected 'parent-point', got '%s'", desc)
		}

		_, err = cotlib.GetRelationDescription("invalid")
		if err == nil {
			t.Error("Expected error for invalid relation code, but got none")
		}
	})
}

func TestChatValidationMissingFields(t *testing.T) {
	evt, err := cotlib.NewEvent("C1", "t-x-c", 0, 0, 0)
	if err != nil {
		t.Fatalf("new event: %v", err)
	}
	evt.Detail = &cotlib.Detail{Chat: &cotlib.Chat{Message: "", Sender: ""}}
	if err := evt.Validate(); err == nil {
		t.Fatal("expected error for empty chat sender and message")
	}
	evt.Detail.Chat.Message = "hi"
	if err := evt.Validate(); err == nil {
		t.Fatal("expected error for empty chat sender")
	}
	evt.Detail.Chat.Sender = "A"
	if err := evt.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cotlib.ReleaseEvent(evt)
}

func TestDetailExtensionsRoundTrip(t *testing.T) {
	evt, err := cotlib.NewEvent("X1", "a-f-G", 1, 2, 3)
	if err != nil {
		t.Fatalf("new event: %v", err)
	}
	evt.Detail = &cotlib.Detail{
		Chat:              &cotlib.Chat{ID: "", Message: "m", Sender: "s"},
		ChatReceipt:       &cotlib.ChatReceipt{Ack: "y"},
		Geofence:          &cotlib.Geofence{Raw: []byte(`<__geofence elevationMonitored="false" minElevation="0" monitor="in" trigger="enter" tracking="true" maxElevation="1" boundingSphere="1"/>`)},
		ServerDestination: &cotlib.ServerDestination{Raw: []byte(`<__serverdestination destinations="srv"/>`)},
		Video:             &cotlib.Video{Raw: []byte(`<__video url="v"/>`)},
		GroupExtension:    &cotlib.GroupExtension{Raw: []byte(`<__group name="g" role="member"/>`)},

		Unknown: []cotlib.RawMessage{[]byte(`<extra foo="bar"/>`)},
	}

	xmlData, err := evt.ToXML()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	cotlib.ReleaseEvent(evt)

	out, err := cotlib.UnmarshalXMLEvent(context.Background(), xmlData)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Detail == nil || out.Detail.Chat == nil || out.Detail.Chat.Message == "" {
		t.Errorf("chat extension lost")
	}
	if want := []byte(`<__chat message="m" sender="s"></__chat>`); !bytes.Equal(out.Detail.Chat.Raw, want) {
		t.Errorf("chat raw mismatch: got %s want %s", string(out.Detail.Chat.Raw), string(want))
	}
	if len(out.Detail.Unknown) != 1 {
		t.Errorf("expected 1 unknown element, got %d", len(out.Detail.Unknown))
	}

	cotlib.ReleaseEvent(out)
}

func TestAdditionalDetailExtensionsRoundTrip(t *testing.T) {
	evt, err := cotlib.NewEvent("X2", "a-f-G", 1, 2, 3)
	if err != nil {
		t.Fatalf("new event: %v", err)
	}

	envXML := []byte(`<environment temperature="20" windDirection="10" windSpeed="5"></environment>`)
	fileShareXML := []byte(`<fileshare filename="f" name="n" senderCallsign="A" senderUid="U" senderUrl="http://x" sha256="h" sizeInBytes="1"></fileshare>`)
	precisionXML := []byte(`<precisionlocation altsrc="GPS"></precisionlocation>`)
	takvXML := []byte(`<takv platform="Android" version="1"></takv>`)
	trackXML := []byte(`<track course="90" speed="10"></track>`)
	missionXML := []byte(`<mission name="op" tool="t" type="x"></mission>`)
	statusXML := []byte(`<status battery="80"></status>`)
	shapeXML := []byte(`<shape><polyline closed="true"><vertex hae="0" lat="1" lon="1"></vertex></polyline></shape>`)

	evt.Detail = &cotlib.Detail{
		Environment:       &cotlib.Environment{Raw: envXML},
		FileShare:         &cotlib.FileShare{Raw: fileShareXML},
		PrecisionLocation: &cotlib.PrecisionLocation{Raw: precisionXML},
		Takv:              &cotlib.Takv{Raw: takvXML},
		Track:             &cotlib.Track{Raw: trackXML},
		Mission:           &cotlib.Mission{Raw: missionXML},
		Status:            &cotlib.Status{Raw: statusXML},
		Shape:             &cotlib.Shape{Raw: shapeXML},
	}

	xmlData, err := evt.ToXML()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	cotlib.ReleaseEvent(evt)

	out, err := cotlib.UnmarshalXMLEvent(context.Background(), xmlData)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Detail == nil {
		t.Fatalf("detail missing after round trip")
	}

	checks := []struct {
		name string
		got  []byte
		want []byte
	}{
		{"environment", out.Detail.Environment.Raw, envXML},
		{"fileshare", out.Detail.FileShare.Raw, fileShareXML},
		{"precisionlocation", out.Detail.PrecisionLocation.Raw, precisionXML},
		{"takv", out.Detail.Takv.Raw, takvXML},
		{"track", out.Detail.Track.Raw, trackXML},
		{"mission", out.Detail.Mission.Raw, missionXML},
		{"status", out.Detail.Status.Raw, statusXML},
		{"shape", out.Detail.Shape.Raw, shapeXML},
	}

	for _, c := range checks {
		if !bytes.Equal(c.got, c.want) {
			t.Errorf("%s round trip mismatch: got %s want %s", c.name, string(c.got), string(c.want))
		}
	}
	cotlib.ReleaseEvent(out)
}

func TestChatSchemaValidation(t *testing.T) {
	validator.ResetForTest()
	valid := []byte(`<__chat sender="A" message="hi"/>`)
	if err := validator.ValidateAgainstSchema("chat", valid); err != nil {
		t.Fatalf("valid chat rejected: %v", err)
	}

	invalid := []byte(`<__chat unknown="x"/>`)
	if err := validator.ValidateAgainstSchema("chat", invalid); err == nil {
		t.Fatal("expected error for invalid chat")
	}

	validReceipt := []byte(`<__chatReceipt ack="y"/>`)
	if err := validator.ValidateAgainstSchema("chatReceipt", validReceipt); err != nil {
		t.Fatalf("valid chatReceipt rejected: %v", err)
	}

	invalidReceipt := []byte(`<__chatReceipt/>`)
	if err := validator.ValidateAgainstSchema("chatReceipt", invalidReceipt); err == nil {
		t.Fatal("expected error for invalid chatReceipt")
	}
}

func TestTAKChatFallback(t *testing.T) {
	now := time.Now().UTC()
	chatRaw := `<__chat chatroom="room" groupOwner="false" id="1" senderCallsign="A"><chatgrp id="room" uid0="u"/></__chat>`
	expectedRaw := `<__chat chatroom="room" groupOwner="false" id="1" senderCallsign="A"><chatgrp id="room" uid0="u"></chatgrp></__chat>`
	xmlData := fmt.Sprintf(`<event version="2.0" uid="U" type="a-f-G" time="%[1]s" start="%[1]s" stale="%[2]s">`+
		`<point lat="0" lon="0" hae="0" ce="1" le="1"/>`+
		`<detail>%s</detail></event>`,
		now.Format(cotlib.CotTimeFormat),
		now.Add(10*time.Second).Format(cotlib.CotTimeFormat),
		chatRaw)

	evt, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(xmlData))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := evt.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if evt.Detail == nil || evt.Detail.Chat == nil {
		t.Fatalf("chat detail missing")
	}
	if !bytes.Equal(evt.Detail.Chat.Raw, []byte(expectedRaw)) {
		t.Errorf("chat raw mismatch: got %s want %s", string(evt.Detail.Chat.Raw), expectedRaw)
	}
	cotlib.ReleaseEvent(evt)
}

func TestUnmarshalInvalidChatExtensions(t *testing.T) {
	now := time.Now().UTC()
	base := `<event version="2.0" uid="U" type="a-f-G" time="%[1]s" start="%[1]s" stale="%[2]s">` +
		`<point lat="0" lon="0" hae="0" ce="1" le="1"/>` +
		`<detail>%s</detail></event>`

	t.Run("invalid_chat", func(t *testing.T) {
		xmlData := fmt.Sprintf(base,
			now.Format(cotlib.CotTimeFormat),
			now.Add(10*time.Second).Format(cotlib.CotTimeFormat),
			`<__chat unknown="x"/>`,
		)
		if _, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(xmlData)); err == nil {
			t.Error("expected error for invalid chat")
		}
	})

	t.Run("invalid_chatReceipt", func(t *testing.T) {
		xmlData := fmt.Sprintf(base,
			now.Format(cotlib.CotTimeFormat),
			now.Add(10*time.Second).Format(cotlib.CotTimeFormat),
			`<__chatReceipt/>`,
		)
		if _, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(xmlData)); err == nil {
			t.Error("expected error for invalid chatReceipt")
		}
	})
}

func TestChatReceiptTwoSchemas(t *testing.T) {
	ack := []byte(`<__chatReceipt ack="y"/>`)
	var r1 cotlib.ChatReceipt
	if err := xml.Unmarshal(ack, &r1); err != nil {
		t.Fatalf("unmarshal ack: %v", err)
	}

	detail := []byte(`<__chatreceipt chatroom="c" groupOwner="false" id="1" senderCallsign="A"><chatgrp id="g" uid0="u0"/></__chatreceipt>`)
	var r2 cotlib.ChatReceipt
	if err := xml.Unmarshal(detail, &r2); err != nil {
		t.Fatalf("unmarshal detail: %v", err)
	}
	if len(r2.Raw) == 0 {
		t.Error("expected raw preserved for detail chatreceipt")
	}
	if r2.ID != "1" || r2.Chatroom != "c" || r2.GroupOwner != "false" || r2.SenderCallsign != "A" {
		t.Errorf("unexpected parsed chatreceipt: %+v", r2)
	}
	if r2.ChatGrp == nil || r2.ChatGrp.ID != "g" || r2.ChatGrp.UID0 != "u0" {
		t.Errorf("unexpected chatgrp: %+v", r2.ChatGrp)
	}
	expectedRaw := `<__chatreceipt id="1" chatroom="c" groupOwner="false" senderCallsign="A"><chatgrp id="g" uid0="u0"></chatgrp></__chatreceipt>`
	if b, err := xml.Marshal(&r2); err != nil {
		t.Fatalf("marshal detail: %v", err)
	} else if string(b) != expectedRaw {
		t.Errorf("marshal mismatch: got %s want %s", string(b), expectedRaw)
	}

	now := time.Now().UTC()
	evt1, _ := cotlib.NewEvent("CR1", "a-f-G", 0, 0, 1)
	evt1.Detail = &cotlib.Detail{ChatReceipt: &r1}
	if err := evt1.ValidateAt(now); err != nil {
		t.Fatalf("validate ack: %v", err)
	}
	cotlib.ReleaseEvent(evt1)

	evt2, _ := cotlib.NewEvent("CR2", "a-f-G", 0, 0, 1)
	evt2.Detail = &cotlib.Detail{ChatReceipt: &r2}
	if err := evt2.ValidateAt(now); err != nil {
		t.Fatalf("validate detail: %v", err)
	}
	cotlib.ReleaseEvent(evt2)

	// receipt without id should still be valid
	noID := []byte(`<__chatreceipt chatroom="c" groupOwner="false" senderCallsign="A"><chatgrp id="g" uid0="u0"/></__chatreceipt>`)
	var r3 cotlib.ChatReceipt
	if err := xml.Unmarshal(noID, &r3); err != nil {
		t.Fatalf("unmarshal no id: %v", err)
	}
	if r3.ID != "" {
		t.Errorf("expected empty ID, got %q", r3.ID)
	}
	evt3, _ := cotlib.NewEvent("CR3", "a-f-G", 0, 0, 1)
	evt3.Detail = &cotlib.Detail{ChatReceipt: &r3}
	if err := evt3.ValidateAt(now); err != nil {
		t.Fatalf("validate no id: %v", err)
	}
	cotlib.ReleaseEvent(evt3)
}

func TestChatIsGroupChat(t *testing.T) {
	cases := []struct {
		name    string
		xmlData string
		want    bool
	}{
		{
			name:    "direct",
			xmlData: `<__chat sender="A" message="hi"/>`,
			want:    false,
		},
		{
			name:    "group",
			xmlData: `<__chat chatroom="c" groupOwner="false" id="1" senderCallsign="A"><chatgrp id="c" uid0="u0"/></__chat>`,
			want:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c cotlib.Chat
			if err := xml.Unmarshal([]byte(tc.xmlData), &c); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := c.IsGroupChat(); got != tc.want {
				t.Errorf("IsGroupChat()=%v want %v", got, tc.want)
			}
		})
	}
}

func TestTAKDetailSchemaValidation(t *testing.T) {
	t.Run("contact", func(t *testing.T) {
		evt, err := cotlib.NewEvent("C1", "t-x-d", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			Contact: &cotlib.Contact{Callsign: "A"},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid contact rejected: %v", err)
		}
		evt.Detail.Contact.Callsign = ""
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for missing callsign")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("track", func(t *testing.T) {
		evt, err := cotlib.NewEvent("T1", "t-x-t", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			Track: &cotlib.Track{Raw: []byte(`<track course="90" speed="10"/>`)},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid track rejected: %v", err)
		}
		evt.Detail.Track.Raw = []byte(`<track speed="10"/>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid track")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("status", func(t *testing.T) {
		evt, err := cotlib.NewEvent("S1", "a-f-G", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			Status: &cotlib.Status{Raw: []byte(`<status battery="80"/>`)},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid status rejected: %v", err)
		}
		evt.Detail.Status.Raw = []byte(`<status battery="bad"/>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid status")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("geofence_and_drawing", func(t *testing.T) {
		evt, err := cotlib.NewEvent("G1", "a-f-G", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			Geofence:     &cotlib.Geofence{Raw: []byte(`<__geofence elevationMonitored="false" minElevation="0" monitor="in" trigger="enter" tracking="true" maxElevation="10" boundingSphere="1"/>`)},
			StrokeColor:  &cotlib.StrokeColor{Raw: []byte(`<strokecolor value="1"/>`)},
			StrokeWeight: &cotlib.StrokeWeight{Raw: []byte(`<strokeweight value="1"/>`)},
			FillColor:    &cotlib.FillColor{Raw: []byte(`<fillcolor value="1"/>`)},
			LabelsOn:     &cotlib.LabelsOn{Raw: []byte(`<labelson value="true"/>`)},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid drawing extensions rejected: %v", err)
		}
		evt.Detail.StrokeColor.Raw = []byte(`<strokecolor/>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid strokecolor")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("emergency", func(t *testing.T) {
		evt, err := cotlib.NewEvent("EM1", "a-f-G", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			Emergency: &cotlib.Emergency{Raw: []byte(`<emergency>help</emergency>`)},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid emergency rejected: %v", err)
		}
		evt.Detail.Emergency.Raw = []byte(`<emergency>bad value</emergency>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid emergency")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("height", func(t *testing.T) {
		evt, err := cotlib.NewEvent("H1", "a-f-G", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			Height:     &cotlib.Height{Raw: []byte(`<height>1</height>`)},
			HeightUnit: &cotlib.HeightUnit{Raw: []byte(`<height_unit>1</height_unit>`)},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid height rejected: %v", err)
		}
		evt.Detail.Height.Raw = []byte(`<height>bad</height>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid height")
		}
		evt.Detail.Height.Raw = []byte(`<height>1</height>`)
		evt.Detail.HeightUnit.Raw = []byte(`<height_unit>x</height_unit>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid height_unit")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("hierarchy", func(t *testing.T) {
		evt, err := cotlib.NewEvent("H2", "a-f-G", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			Hierarchy: &cotlib.Hierarchy{Raw: []byte(`<hierarchy><group uid="g" name="n"/></hierarchy>`)},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid hierarchy rejected: %v", err)
		}
		evt.Detail.Hierarchy.Raw = []byte(`<hierarchy/>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid hierarchy")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("link_detail", func(t *testing.T) {
		evt, err := cotlib.NewEvent("L1", "a-f-G", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			LinkDetail: &cotlib.DetailLink{Raw: []byte(`<link uid="u" type="a-f-G" relation="p-p"/>`)},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid link detail rejected: %v", err)
		}
		evt.Detail.LinkDetail.Raw = []byte(`<link uid="u"/>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid link detail")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("group_and_user", func(t *testing.T) {
		evt, err := cotlib.NewEvent("G2", "a-f-G", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			GroupExtension:    &cotlib.GroupExtension{Raw: []byte(`<__group name="g" role="member"/>`)},
			ServerDestination: &cotlib.ServerDestination{Raw: []byte(`<__serverdestination destinations="srv"/>`)},
			UserIcon:          &cotlib.UserIcon{Raw: []byte(`<usericon iconsetpath="icons"/>`)},
			LabelsOn:          &cotlib.LabelsOn{Raw: []byte(`<labelson value="true"/>`)},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid group/user extensions rejected: %v", err)
		}

		evt.Detail.GroupExtension.Raw = []byte(`<__group name="g"/>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid __group")
		}
		evt.Detail.GroupExtension.Raw = []byte(`<__group name="g" role="member"/>`)

		evt.Detail.ServerDestination.Raw = []byte(`<__serverdestination/>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid __serverdestination")
		}
		evt.Detail.ServerDestination.Raw = []byte(`<__serverdestination destinations="srv"/>`)

		evt.Detail.UserIcon.Raw = []byte(`<usericon/>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid usericon")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("fileshare", func(t *testing.T) {
		evt, err := cotlib.NewEvent("FS1", "a-f-G", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			FileShare: &cotlib.FileShare{Raw: []byte(`<fileshare filename="f" name="n" senderCallsign="A" senderUid="U" senderUrl="http://x" sha256="h" sizeInBytes="1"/>`)},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid fileshare rejected: %v", err)
		}
		evt.Detail.FileShare.Raw = []byte(`<fileshare filename="f"/>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid fileshare")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("routeinfo", func(t *testing.T) {
		evt, err := cotlib.NewEvent("RI1", "a-f-G", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			RouteInfo: &cotlib.RouteInfo{Raw: []byte(`<__routeinfo><__navcues/></__routeinfo>`)},
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid routeinfo rejected: %v", err)
		}
		evt.Detail.RouteInfo.Raw = []byte(`<__routeinfo foo="bar"/>`)
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid routeinfo")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("marti", func(t *testing.T) {
		evt, err := cotlib.NewEvent("M2", "b-t-f", 1, 1, 0)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{Marti: &cotlib.Marti{Dest: []cotlib.MartiDest{{Callsign: "A"}}}}
		if err := evt.Validate(); err != nil {
			t.Fatalf("valid marti rejected: %v", err)
		}
		evt.Detail.Marti.Dest = []cotlib.MartiDest{{}}
		if err := evt.Validate(); err == nil {
			t.Fatal("expected error for invalid marti")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("uid_schema", func(t *testing.T) {
		good := []byte(`<uid Droid="droid://123" nett="net"/>`)
		if err := validator.ValidateAgainstSchema("tak-details-uid", good); err != nil {
			t.Fatalf("valid uid rejected: %v", err)
		}
		bad := []byte(`<uid nett="net"/>`)
		if err := validator.ValidateAgainstSchema("tak-details-uid", bad); err == nil {
			t.Fatal("expected error for invalid uid")
		}
	})

	t.Run("remarks_schema", func(t *testing.T) {
		good := []byte(`<remarks source="src" sourceID="id" time="2023-01-02T15:04:05Z" to="dest">hi</remarks>`)
		if err := validator.ValidateAgainstSchema("tak-details-remarks", good); err != nil {
			t.Fatalf("valid remarks rejected: %v", err)
		}
		bad := []byte(`<remarks foo="bar"/>`)
		if err := validator.ValidateAgainstSchema("tak-details-remarks", bad); err == nil {
			t.Fatal("expected error for invalid remarks")
		}
	})

	t.Run("remarks_case_mismatch", func(t *testing.T) {
		now := time.Now().UTC()
		xmlData := fmt.Sprintf(`<event version="2.0" uid="U" type="a-f-G" time="%[1]s" start="%[1]s" stale="%[2]s">`+
			`<point lat="0" lon="0" hae="0" ce="1" le="1"/>`+
			`<detail><Remarks>hi</Remarks></detail></event>`,
			now.Format(cotlib.CotTimeFormat),
			now.Add(10*time.Second).Format(cotlib.CotTimeFormat))
		if _, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(xmlData)); err == nil {
			t.Fatal("expected error for unrecognized Remarks element")
		}
	})

	t.Run("tak_chat_with_chatgrp", func(t *testing.T) {
		now := time.Now().UTC()
		xmlData := fmt.Sprintf(`<event version="2.0" uid="U" type="a-f-G" time="%[1]s" start="%[1]s" stale="%[2]s">`+
			`<point lat="0" lon="0" hae="0" ce="1" le="1"/>`+
			`<detail><__chat chatroom="c" groupOwner="false" id="1" senderCallsign="A"><chatgrp id="g" uid0="u0"/></__chat></detail>`+
			`</event>`,
			now.Format(cotlib.CotTimeFormat),
			now.Add(10*time.Second).Format(cotlib.CotTimeFormat))
		evt, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(xmlData))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("validate: %v", err)
		}
		if evt.Detail == nil || evt.Detail.Chat == nil {
			t.Fatalf("chat detail missing")
		}
		if evt.Detail.Chat.Chatroom != "c" {
			t.Errorf("chatroom parsed incorrectly: %s", evt.Detail.Chat.Chatroom)
		}
		if len(evt.Detail.Chat.ChatGrps) != 1 || evt.Detail.Chat.ChatGrps[0].ID != "g" {
			t.Errorf("chatgrp parsed incorrectly: %+v", evt.Detail.Chat.ChatGrps)
		}
		out, err := evt.ToXML()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if !bytes.Contains(out, []byte(`chatroom="c"`)) {
			t.Errorf("expected chatroom attribute in output")
		}
		if !bytes.Contains(out, []byte(`groupOwner="false"`)) {
			t.Errorf("expected groupOwner attribute in output")
		}
		if !bytes.Contains(out, []byte(`senderCallsign="A"`)) {
			t.Errorf("expected senderCallsign attribute in output")
		}
		if !bytes.Contains(out, []byte(`<chatgrp`)) {
			t.Errorf("expected chatgrp element in output")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("tak_chat_parent_messageid", func(t *testing.T) {
		now := time.Now().UTC()
		xmlData := fmt.Sprintf(`<event version="2.0" uid="U" type="a-f-G" time="%[1]s" start="%[1]s" stale="%[2]s">`+
			`<point lat="0" lon="0" hae="0" ce="1" le="1"/>`+
			`<detail><__chat chatroom="c" groupOwner="false" id="1" senderCallsign="A" parent="p" messageId="m"><chatgrp id="g" uid0="u0"/></__chat></detail>`+
			`</event>`,
			now.Format(cotlib.CotTimeFormat),
			now.Add(10*time.Second).Format(cotlib.CotTimeFormat))
		evt, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(xmlData))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if evt.Detail.Chat.Parent != "p" || evt.Detail.Chat.MessageID != "m" {
			t.Errorf("chat parent/messageId parsed incorrectly: %+v", evt.Detail.Chat)
		}
		out, err := evt.ToXML()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if !bytes.Contains(out, []byte(`parent="p"`)) {
			t.Errorf("expected parent attribute in output")
		}
		if !bytes.Contains(out, []byte(`messageId="m"`)) {
			t.Errorf("expected messageId attribute in output")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("tak_chat_deletechild", func(t *testing.T) {
		now := time.Now().UTC()
		xmlData := fmt.Sprintf(`<event version="2.0" uid="U" type="a-f-G" time="%[1]s" start="%[1]s" stale="%[2]s">`+
			`<point lat="0" lon="0" hae="0" ce="1" le="1"/>`+
			`<detail><__chat chatroom="c" groupOwner="false" id="1" senderCallsign="A" deleteChild="child"><chatgrp id="g" uid0="u0"/></__chat></detail>`+
			`</event>`,
			now.Format(cotlib.CotTimeFormat),
			now.Add(10*time.Second).Format(cotlib.CotTimeFormat))
		evt, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(xmlData))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if evt.Detail.Chat.DeleteChild != "child" {
			t.Errorf("deleteChild parsed incorrectly: %s", evt.Detail.Chat.DeleteChild)
		}
		out, err := evt.ToXML()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if !bytes.Contains(out, []byte(`deleteChild="child"`)) {
			t.Errorf("expected deleteChild attribute in output")
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("tak_chat_no_id", func(t *testing.T) {
		now := time.Now().UTC()
		xmlData := fmt.Sprintf(`<event version="2.0" uid="U" type="a-f-G" time="%[1]s" start="%[1]s" stale="%[2]s">`+
			`<point lat="0" lon="0" hae="0" ce="1" le="1"/>`+
			`<detail><__chat chatroom="c" groupOwner="false" senderCallsign="A"><chatgrp id="g" uid0="u0"/></__chat></detail>`+
			`</event>`,
			now.Format(cotlib.CotTimeFormat),
			now.Add(10*time.Second).Format(cotlib.CotTimeFormat))
		evt, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(xmlData))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if err := evt.Validate(); err != nil {
			t.Fatalf("validate: %v", err)
		}
		if evt.Detail == nil || evt.Detail.Chat == nil {
			t.Fatalf("chat detail missing")
		}
		if evt.Detail.Chat.ID != "" {
			t.Errorf("expected empty ID, got %q", evt.Detail.Chat.ID)
		}
		cotlib.ReleaseEvent(evt)
	})

	t.Run("marti_geochat_roundtrip", func(t *testing.T) {
		evt, err := cotlib.NewEvent("M1", "b-t-f", 1, 2, 3)
		if err != nil {
			t.Fatalf("new event: %v", err)
		}
		evt.Detail = &cotlib.Detail{
			Marti: &cotlib.Marti{Dest: []cotlib.MartiDest{{Callsign: "A"}, {Callsign: "B"}}},
		}
		xmlData, err := evt.ToXML()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		cotlib.ReleaseEvent(evt)

		out, err := cotlib.UnmarshalXMLEvent(context.Background(), xmlData)
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if err := out.Validate(); err != nil {
			t.Fatalf("validate: %v", err)
		}
		if out.Detail == nil || out.Detail.Marti == nil {
			t.Fatalf("marti detail missing")
		}
		if len(out.Detail.Marti.Dest) != 2 || out.Detail.Marti.Dest[0].Callsign != "A" || out.Detail.Marti.Dest[1].Callsign != "B" {
			t.Errorf("marti dest round trip mismatch: %#v", out.Detail.Marti.Dest)
		}
		cotlib.ReleaseEvent(out)
	})

	t.Run("event_unknown_attr_roundtrip", func(t *testing.T) {
		now := time.Now().UTC()
		xmlData := fmt.Sprintf(`<event version="2.0" uid="U" type="a-f-G" access="U" extra="Undefined" time="%[1]s" start="%[1]s" stale="%[2]s">`+
			`<point lat="0" lon="0" ce="1" le="1"/>`+
			`</event>`,
			now.Format(cotlib.CotTimeFormat),
			now.Add(10*time.Second).Format(cotlib.CotTimeFormat))

		evt, err := cotlib.UnmarshalXMLEvent(context.Background(), []byte(xmlData))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if evt.Access != "U" {
			t.Fatalf("expected access with value \"U\", but got %q", evt.Access)
		}
		if len(evt.UnknownAttrs) != 1 {
			t.Fatalf("expected 1 unknown attr, got %d", len(evt.UnknownAttrs))
		}
		ua := evt.UnknownAttrs[0]
		if ua.Name.Local != "extra" || ua.Value != "Undefined" {
			t.Errorf("unexpected unknown attr: %+v", ua)
		}
		out, err := evt.ToXML()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if !bytes.Contains(out, []byte(`extra="Undefined"`)) {
			t.Errorf("extra attribute lost in output")
		}
		cotlib.ReleaseEvent(evt)
	})

}
