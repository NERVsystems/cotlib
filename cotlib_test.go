package cotlib

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Setup test logging
	testLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	SetLogger(testLogger)

	// Run tests
	os.Exit(m.Run())
}

// TestNewEvent verifies that NewEvent creates valid events with correct defaults
func TestNewEvent(t *testing.T) {
	tests := []struct {
		name      string
		uid       string
		eventType string
		lat       float64
		lon       float64
		wantErr   bool
	}{
		{
			name:      "valid event",
			uid:       "TEST1",
			eventType: TypePredFriend + "-G",
			lat:       45.0,
			lon:       -120.0,
			wantErr:   false,
		},
		{
			name:      "invalid latitude",
			uid:       "TEST2",
			eventType: TypePredHostile + "-A",
			lat:       91.0, // Invalid
			lon:       0.0,
			wantErr:   true,
		},
		{
			name:      "invalid longitude",
			uid:       "TEST3",
			eventType: TypePredNeutral + "-G",
			lat:       0.0,
			lon:       181.0, // Invalid
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := NewEvent(tt.uid, tt.eventType, tt.lat, tt.lon)
			err := evt.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("NewEvent() validation error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check required fields
				if evt.Version != "2.0" {
					t.Errorf("NewEvent() version = %v, want %v", evt.Version, "2.0")
				}
				if evt.Uid != tt.uid {
					t.Errorf("NewEvent() uid = %v, want %v", evt.Uid, tt.uid)
				}
				if evt.Type != tt.eventType {
					t.Errorf("NewEvent() type = %v, want %v", evt.Type, tt.eventType)
				}

				// Verify time fields are properly set
				now := time.Now().UTC()
				eventTime, err := time.Parse(time.RFC3339, evt.Time)
				if err != nil {
					t.Errorf("NewEvent() invalid time format: %v", err)
				}
				if eventTime.Sub(now) > time.Second {
					t.Errorf("NewEvent() time not within 1 second of creation")
				}
			}
		})
	}
}

// TestEventValidation tests the Validate function thoroughly
func TestEventValidation(t *testing.T) {
	tests := []struct {
		name    string
		event   *Event
		wantErr bool
	}{
		{
			name: "valid event",
			event: &Event{
				Version: "2.0",
				Uid:     "TEST1",
				Type:    TypePredFriend + "-G",
				Time:    time.Now().UTC().Format(time.RFC3339),
				Start:   time.Now().UTC().Format(time.RFC3339),
				Stale:   time.Now().UTC().Add(time.Minute).Format(time.RFC3339),
				Point:   Point{Lat: 45.0, Lon: -120.0, Hae: 0.0, Ce: 9999999, Le: 9999999},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			event: &Event{
				Uid:   "TEST2",
				Type:  TypePredHostile + "-A",
				Point: Point{Lat: 45.0, Lon: -120.0},
			},
			wantErr: true,
		},
		{
			name: "invalid time format",
			event: &Event{
				Version: "2.0",
				Uid:     "TEST3",
				Type:    TypePredNeutral + "-G",
				Time:    "invalid",
				Point:   Point{Lat: 45.0, Lon: -120.0},
			},
			wantErr: true,
		},
		{
			name: "stale before time",
			event: &Event{
				Version: "2.0",
				Uid:     "TEST4",
				Type:    TypePredFriend + "-G",
				Time:    time.Now().UTC().Format(time.RFC3339),
				Start:   time.Now().UTC().Format(time.RFC3339),
				Stale:   time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
				Point:   Point{Lat: 45.0, Lon: -120.0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestEventPredicate tests the Is function for type predicates
func TestEventPredicate(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		predicate string
		want      bool
	}{
		{"is friend", TypePredFriend + "-G", "friend", true},
		{"is hostile", TypePredHostile + "-A", "hostile", true},
		{"is air", TypePredFriend + "-A", "air", true},
		{"is ground", TypePredHostile + "-G", "ground", true},
		{"not friend", TypePredHostile + "-G", "friend", false},
		{"not air", TypePredFriend + "-G", "air", false},
		{"invalid predicate", TypePredFriend + "-G", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := NewEvent("TEST", tt.eventType, 45.0, -120.0)
			if got := evt.Is(tt.predicate); got != tt.want {
				t.Errorf("Event.Is(%v) = %v, want %v", tt.predicate, got, tt.want)
			}
		})
	}
}

// TestEventLinks tests the linking functionality
func TestEventLinks(t *testing.T) {
	lead := NewEvent("LEAD", TypePredFriend+"-A", 45.0, -120.0)
	wing1 := NewEvent("WING1", TypePredFriend+"-A", 45.1, -120.1)
	wing2 := NewEvent("WING2", TypePredFriend+"-A", 44.9, -119.9)

	// Test adding links
	lead.AddLink(wing1.Uid, "member", "wingman1")
	lead.AddLink(wing2.Uid, "member", "wingman2")

	if len(lead.Links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(lead.Links))
	}

	// Verify link contents
	for _, link := range lead.Links {
		if link.Type != "member" {
			t.Errorf("Expected link type 'member', got %s", link.Type)
		}
		if !strings.HasPrefix(link.Relation, "wingman") {
			t.Errorf("Expected relation to start with 'wingman', got %s", link.Relation)
		}
	}
}

// TestXMLMarshaling tests XML marshaling and unmarshaling
func TestXMLMarshaling(t *testing.T) {
	// Create a complex event with various features
	evt := NewEvent("TEST", TypePredFriend+"-A", 45.0, -120.0)
	evt.How = "m-g"
	evt.DetailContent.Shape = &Shape{
		Type:   "circle",
		Radius: 1000,
	}
	evt.DetailContent.UidAliases = &UidAliases{
		Callsign: "EAGLE1",
		Platform: "F16",
	}
	evt.AddLink("WING1", "member", "wingman")

	// Marshal to XML
	xmlData, err := evt.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}

	// Verify XML structure
	if !bytes.Contains(xmlData, []byte(`<event version="2.0"`)) {
		t.Error("XML missing event element with version")
	}
	if !bytes.Contains(xmlData, []byte(`<point lat="45" lon="-120"`)) {
		t.Error("XML missing point element with coordinates")
	}
	if !bytes.Contains(xmlData, []byte(`<shape type="circle" radius="1000"`)) {
		t.Error("XML missing shape element")
	}

	// Unmarshal back
	parsed, err := UnmarshalXMLEvent(xmlData)
	if err != nil {
		t.Fatalf("UnmarshalXMLEvent() error = %v", err)
	}

	// Verify unmarshaled data
	if parsed.Uid != evt.Uid {
		t.Errorf("Unmarshaled UID = %v, want %v", parsed.Uid, evt.Uid)
	}
	if parsed.DetailContent.Shape.Type != "circle" {
		t.Errorf("Unmarshaled shape type = %v, want circle", parsed.DetailContent.Shape.Type)
	}
	if parsed.DetailContent.UidAliases.Callsign != "EAGLE1" {
		t.Errorf("Unmarshaled callsign = %v, want EAGLE1", parsed.DetailContent.UidAliases.Callsign)
	}
}

// TestSecurityFeatures tests security-related functionality
func TestSecurityFeatures(t *testing.T) {
	// Test XXE prevention
	xmlWithXXE := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<!DOCTYPE event [
			<!ENTITY xxe SYSTEM "file:///etc/passwd">
		]>
		<event version="2.0" uid="TEST" type="a-f-G" time="2024-01-01T00:00:00Z">
			&xxe;
			<point lat="45" lon="-120" hae="0" ce="9999999" le="9999999"/>
		</event>`)

	_, err := UnmarshalXMLEvent(xmlWithXXE)
	if err == nil {
		t.Error("Expected error for XXE attack, got nil")
	}

	// Test detail size limit
	evt := NewEvent("TEST", TypePredFriend+"-G", 45.0, -120.0)
	evt.DetailContent.RawXML = make([]byte, MaxDetailSize+1)
	if err := evt.DetailContent.validateDetailSize(); err == nil {
		t.Error("Expected error for oversized detail content")
	}

	// Test time-based attack prevention
	evt = NewEvent("TEST", TypePredFriend+"-G", 45.0, -120.0)
	evt.Time = time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339) // Future time
	evt.Start = time.Now().UTC().Format(time.RFC3339)
	evt.Stale = time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	if err := evt.validateTimes(); err == nil {
		t.Error("Expected error for future event time")
	}
}

// TestDetailExtensions tests the detail extension functionality
func TestDetailExtensions(t *testing.T) {
	evt := NewEvent("TEST", TypePredFriend+"-G", 45.0, -120.0)

	// Test Shape extension
	evt.DetailContent.Shape = &Shape{
		Type:   "circle",
		Radius: 1000,
	}

	// Test FlowTags extension
	evt.DetailContent.FlowTags = &FlowTags{
		Status: "pending",
		Chain:  "command",
	}

	// Test UidAliases extension
	evt.DetailContent.UidAliases = &UidAliases{
		Callsign: "EAGLE1",
		Platform: "F16",
		Droid:    "android123",
	}

	// Marshal to verify extensions are properly encoded
	xmlData, err := evt.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}

	// Verify all extensions are present in XML
	if !bytes.Contains(xmlData, []byte(`<shape type="circle" radius="1000"`)) {
		t.Error("XML missing shape extension")
	}
	if !bytes.Contains(xmlData, []byte(`<__flow-tags__ status="pending" chain="command"`)) {
		t.Error("XML missing flow-tags extension")
	}
	if !bytes.Contains(xmlData, []byte(`<uid callsign="EAGLE1" platform="F16" droid="android123"`)) {
		t.Error("XML missing uid aliases extension")
	}

	// Unmarshal to verify extensions are properly decoded
	parsed, err := UnmarshalXMLEvent(xmlData)
	if err != nil {
		t.Fatalf("UnmarshalXMLEvent() error = %v", err)
	}

	// Verify unmarshaled extensions
	if parsed.DetailContent.Shape.Radius != 1000 {
		t.Errorf("Unmarshaled shape radius = %v, want 1000", parsed.DetailContent.Shape.Radius)
	}
	if parsed.DetailContent.FlowTags.Status != "pending" {
		t.Errorf("Unmarshaled flow-tags status = %v, want pending", parsed.DetailContent.FlowTags.Status)
	}
	if parsed.DetailContent.UidAliases.Callsign != "EAGLE1" {
		t.Errorf("Unmarshaled uid callsign = %v, want EAGLE1", parsed.DetailContent.UidAliases.Callsign)
	}
}
