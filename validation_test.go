package cotlib_test

import (
	"testing"

	"github.com/NERVsystems/cotlib"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cotlib.ValidateType(tt.pattern)
			if (err == nil) != tt.expected {
				t.Errorf("ValidateType(%q) error = %v, want error = %v", tt.pattern, err, !tt.expected)
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
			if tc.expectErr && err == nil {
				t.Errorf("Expected error for how value %s, but got none", tc.how)
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Expected no error for how value %s, but got: %v", tc.how, err)
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
			if tc.expectErr && err == nil {
				t.Errorf("Expected error for relation value %s, but got none", tc.relation)
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Expected no error for relation value %s, but got: %v", tc.relation, err)
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

func TestDetailExtensionsRoundTrip(t *testing.T) {
	evt, err := cotlib.NewEvent("X1", "a-f-G", 1, 2, 3)
	if err != nil {
		t.Fatalf("new event: %v", err)
	}
	evt.Detail = &cotlib.Detail{
		Chat:              &cotlib.Chat{Raw: []byte(`<__chat id="c1"/>`)},
		ChatReceipt:       &cotlib.ChatReceipt{Raw: []byte(`<__chatReceipt ack="y"/>`)},
		Geofence:          &cotlib.Geofence{Raw: []byte(`<__geofence radius="5"/>`)},
		ServerDestination: &cotlib.ServerDestination{Raw: []byte(`<__serverdestination host="srv"/>`)},
		Video:             &cotlib.Video{Raw: []byte(`<__video url="v"/>`)},
		GroupExtension:    &cotlib.GroupExtension{Raw: []byte(`<__group name="g"/>`)},
		Unknown:           []cotlib.RawMessage{[]byte(`<extra foo="bar"/>`)},
	}

	xmlData, err := evt.ToXML()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	cotlib.ReleaseEvent(evt)

	out, err := cotlib.UnmarshalXMLEvent(xmlData)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Detail == nil || out.Detail.Chat == nil || len(out.Detail.Chat.Raw) == 0 {
		t.Errorf("chat extension lost")
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
	evt.Detail = &cotlib.Detail{
		Archive:           &cotlib.Archive{Raw: []byte(`<archive id="a"/>`)},
		AttachmentList:    &cotlib.AttachmentList{Raw: []byte(`<attachmentList/>`)},
		Environment:       &cotlib.Environment{Raw: []byte(`<environment/>`)},
		FileShare:         &cotlib.FileShare{Raw: []byte(`<fileshare/>`)},
		PrecisionLocation: &cotlib.PrecisionLocation{Raw: []byte(`<precisionlocation/>`)},
		Takv:              &cotlib.Takv{Raw: []byte(`<takv/>`)},
		Track:             &cotlib.Track{Raw: []byte(`<track/>`)},
		Mission:           &cotlib.Mission{Raw: []byte(`<mission/>`)},
		Status:            &cotlib.Status{Raw: []byte(`<status/>`)},
		Shape:             &cotlib.Shape{Raw: []byte(`<shape/>`)},
	}

	xmlData, err := evt.ToXML()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	cotlib.ReleaseEvent(evt)

	out, err := cotlib.UnmarshalXMLEvent(xmlData)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Detail == nil || out.Detail.Archive == nil || len(out.Detail.Archive.Raw) == 0 {
		t.Errorf("archive extension lost")
	}
	cotlib.ReleaseEvent(out)
}
