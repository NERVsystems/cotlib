package cotlib

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

// Constants for testing
const (
	cotTimeFormat   = "2006-01-02T15:04:05Z"
	testStaleOffset = 5 * time.Second
)

func TestMain(m *testing.M) {
	// Set up logger for tests
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	SetLogger(logger)
	os.Exit(m.Run())
}

func TestNewEvent(t *testing.T) {
	// Test creating an event without hae parameter (defaults to 0)
	evt, err := NewEvent("test123", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		t.Fatalf("NewEvent() error = %v", err)
	}

	// Verify event properties
	if evt.Uid != "test123" {
		t.Errorf("Uid = %v, want %v", evt.Uid, "test123")
	}
	if evt.Type != "a-f-G" {
		t.Errorf("Type = %v, want %v", evt.Type, "a-f-G")
	}
	if evt.Point.Lat != 30.0 {
		t.Errorf("Point.Lat = %v, want %v", evt.Point.Lat, 30.0)
	}
	if evt.Point.Lon != -85.0 {
		t.Errorf("Point.Lon = %v, want %v", evt.Point.Lon, -85.0)
	}
	if evt.Point.Hae != 0.0 {
		t.Errorf("Point.Hae = %v, want %v", evt.Point.Hae, 0.0)
	}

	// Test creating an event with hae parameter
	evt, err = NewEvent("test456", "a-f-G", 30.0, -85.0, 100.0)
	if err != nil {
		t.Fatalf("NewEvent() error = %v", err)
	}

	// Verify event properties
	if evt.Uid != "test456" {
		t.Errorf("Uid = %v, want %v", evt.Uid, "test456")
	}
	if evt.Type != "a-f-G" {
		t.Errorf("Type = %v, want %v", evt.Type, "a-f-G")
	}
	if evt.Point.Lat != 30.0 {
		t.Errorf("Point.Lat = %v, want %v", evt.Point.Lat, 30.0)
	}
	if evt.Point.Lon != -85.0 {
		t.Errorf("Point.Lon = %v, want %v", evt.Point.Lon, -85.0)
	}
	if evt.Point.Hae != 100.0 {
		t.Errorf("Point.Hae = %v, want %v", evt.Point.Hae, 100.0)
	}

	// Verify time fields
	now := time.Now().UTC().Truncate(time.Second)
	if !evt.Time.Time().Equal(now) {
		t.Errorf("Time = %v, want %v", evt.Time.Time(), now)
	}
	if !evt.Start.Time().Equal(now) {
		t.Errorf("Start = %v, want %v", evt.Start.Time(), now)
	}

	// Verify stale time is set correctly (more than 5 seconds after event time)
	staleDiff := evt.Stale.Time().Sub(evt.Time.Time())
	if staleDiff <= 5*time.Second {
		t.Errorf("Stale time difference = %v, want > %v", staleDiff, 5*time.Second)
	}
}

func TestNewPresenceEvent(t *testing.T) {
	evt, err := NewPresenceEvent("test123", 30.0, -85.0, 0.0)
	if err != nil {
		t.Fatalf("NewPresenceEvent() error = %v", err)
	}

	// Verify event properties
	if evt.Uid != "test123" {
		t.Errorf("Uid = %v, want %v", evt.Uid, "test123")
	}
	if evt.Type != "t-x-takp-v" {
		t.Errorf("Type = %v, want %v", evt.Type, "t-x-takp-v")
	}
	if evt.How != "m-g" {
		t.Errorf("How = %v, want %v", evt.How, "m-g")
	}

	// Verify stale time is set correctly (more than 5 seconds after event time)
	staleDiff := evt.Stale.Time().Sub(evt.Time.Time())
	if staleDiff <= testStaleOffset {
		t.Errorf("Stale time difference = %v, want > %v", staleDiff, testStaleOffset)
	}
}

func TestInjectIdentity(t *testing.T) {
	evt, err := NewEvent("test123", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		t.Fatalf("NewEvent() error = %v", err)
	}

	// Test injecting identity
	evt.InjectIdentity("self123", "Blue", "HQ")

	// Verify group information
	if evt.Detail == nil {
		t.Fatal("Detail is nil")
	}
	if evt.Detail.Group == nil {
		t.Fatal("Group is nil")
	}
	if evt.Detail.Group.Name != "Blue" {
		t.Errorf("Group.Name = %v, want %v", evt.Detail.Group.Name, "Blue")
	}
	if evt.Detail.Group.Role != "HQ" {
		t.Errorf("Group.Role = %v, want %v", evt.Detail.Group.Role, "HQ")
	}

	// Verify p-p link
	found := false
	for _, l := range evt.Links {
		if l.Relation == "p-p" && l.Uid == "self123" && l.Type == "t-x-takp-v" {
			found = true
			break
		}
	}
	if !found {
		t.Error("p-p link not found")
	}

	// Test injecting identity again (should not add duplicate link)
	evt.InjectIdentity("self123", "Blue", "HQ")
	linkCount := 0
	for _, l := range evt.Links {
		if l.Relation == "p-p" && l.Uid == "self123" {
			linkCount++
		}
	}
	if linkCount != 1 {
		t.Errorf("Found %d p-p links, want 1", linkCount)
	}
}

func TestTimeParsing(t *testing.T) {
	// Create a reference time within the valid range (within 24 hours)
	now := time.Now().UTC().Truncate(time.Second)
	refTime := now.Add(-time.Hour) // 1 hour ago

	tests := []struct {
		name    string
		event   *Event
		wantErr bool
	}{
		{
			name: "valid Z format",
			event: &Event{
				Version: "2.0",
				Uid:     "test123",
				Type:    "a-f-G",
				Time:    CoTTime(refTime),
				Start:   CoTTime(refTime),
				Stale:   CoTTime(refTime.Add(10 * time.Second)),
				Point:   Point{Lat: 30.0, Lon: -85.0, Ce: 9999999.0, Le: 9999999.0},
			},
			wantErr: false,
		},
		{
			name: "valid with offset",
			event: &Event{
				Version: "2.0",
				Uid:     "test124",
				Type:    "a-f-G",
				Time:    CoTTime(refTime),
				Start:   CoTTime(refTime),
				Stale:   CoTTime(refTime.Add(10 * time.Second)),
				Point:   Point{Lat: 30.0, Lon: -85.0, Ce: 9999999.0, Le: 9999999.0},
			},
			wantErr: false,
		},
		{
			name: "valid with negative offset",
			event: &Event{
				Version: "2.0",
				Uid:     "test125",
				Type:    "a-f-G",
				Time:    CoTTime(refTime),
				Start:   CoTTime(refTime),
				Stale:   CoTTime(refTime.Add(10 * time.Second)),
				Point:   Point{Lat: 30.0, Lon: -85.0, Ce: 9999999.0, Le: 9999999.0},
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			event: &Event{
				Version: "2.0",
				Uid:     "test126",
				Type:    "a-f-G",
				Time:    CoTTime(refTime),
				Start:   CoTTime(refTime.Add(time.Hour)), // Invalid: start after time
				Stale:   CoTTime(refTime.Add(10 * time.Second)),
				Point:   Point{Lat: 30.0, Lon: -85.0, Ce: 9999999.0, Le: 9999999.0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate the event
			err := tt.event.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
				return
			}

			// Compare times truncated to seconds since that's our precision
			if got := tt.event.Time.Time().Truncate(time.Second); !got.Equal(refTime.Truncate(time.Second)) {
				t.Errorf("Time = %v, want %v", got, refTime.Truncate(time.Second))
			}
			if got := tt.event.Start.Time().Truncate(time.Second); !got.Equal(refTime.Truncate(time.Second)) {
				t.Errorf("Start = %v, want %v", got, refTime.Truncate(time.Second))
			}
			if got := tt.event.Stale.Time().Truncate(time.Second); !got.Equal(refTime.Add(10 * time.Second).Truncate(time.Second)) {
				t.Errorf("Stale = %v, want %v", got, refTime.Add(10*time.Second).Truncate(time.Second))
			}
		})
	}
}

func TestEventValidation(t *testing.T) {
	now := time.Now().UTC()
	validEvent := &Event{
		Version: "2.0",
		Uid:     "test-uid",
		Type:    "a-f-G",
		How:     "m-g",
		Time:    CoTTime(now),
		Start:   CoTTime(now),
		Stale:   CoTTime(now.Add(6 * time.Second)),
		Point: Point{
			Lat: 0,
			Lon: 0,
			Hae: 0,
			Ce:  9999999.0,
			Le:  9999999.0,
		},
	}

	t.Run("valid_event", func(t *testing.T) {
		if err := validEvent.Validate(); err != nil {
			t.Errorf("Event.Validate() error = %v, wantErr false", err)
		}
	})

	t.Run("invalid_start_time", func(t *testing.T) {
		event := *validEvent
		event.Start = CoTTime(event.Time.Time().Add(time.Hour))
		if err := event.Validate(); err == nil {
			t.Error("Event.Validate() error = nil, wantErr true")
		}
	})

	t.Run("invalid_stale_time", func(t *testing.T) {
		event := *validEvent
		event.Stale = CoTTime(event.Time.Time().Add(4 * time.Second))
		if err := event.Validate(); err == nil {
			t.Error("Event.Validate() error = nil, wantErr true")
		}
	})

	t.Run("stale_too_far_in_future", func(t *testing.T) {
		event := *validEvent
		event.Type = "a-f-G"
		event.Stale = CoTTime(event.Time.Time().Add(8 * 24 * time.Hour))
		if err := event.Validate(); err != nil {
			t.Errorf("Event.Validate() error = %v, wantErr false for long stale", err)
		}

		event.Type = "t-x-takp-v"
		if err := event.Validate(); err != nil {
			t.Errorf("Event.Validate() error = %v, wantErr false for long stale", err)
		}
	})

	t.Run("event_too_far_in_past", func(t *testing.T) {
		event := *validEvent
		event.Time = CoTTime(now.Add(-25 * time.Hour))
		event.Start = CoTTime(event.Time.Time())
		event.Stale = CoTTime(event.Time.Time().Add(6 * time.Second))
		if err := event.Validate(); err == nil {
			t.Error("Event.Validate() error = nil, wantErr true")
		}
	})

	t.Run("event_too_far_in_future", func(t *testing.T) {
		event := *validEvent
		event.Time = CoTTime(now.Add(25 * time.Hour))
		event.Start = CoTTime(event.Time.Time())
		event.Stale = CoTTime(event.Time.Time().Add(6 * time.Second))
		if err := event.Validate(); err == nil {
			t.Error("Event.Validate() error = nil, wantErr true")
		}
	})
}

func TestEventPredicate(t *testing.T) {
	evt, err := NewEvent("testUID", "a-f-G", 25.5, -120.7, 0.0)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}

	tests := []struct {
		name      string
		predicate string
		want      bool
	}{
		{"is atom", "atom", true},
		{"is friend", "friend", true},
		{"is hostile", "hostile", false},
		{"is ground", "ground", true},
		{"is air", "air", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evt.Is(tt.predicate); got != tt.want {
				t.Errorf("Event.Is(%q) = %v, want %v", tt.predicate, got, tt.want)
			}
		})
	}
}

func TestEventLinks(t *testing.T) {
	evt, err := NewEvent("testUID", "a-f-G", 25.5, -120.7, 0.0)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}

	// Add a link
	evt.AddLink(&Link{
		Uid:      "TARGET1",
		Type:     "member",
		Relation: "wingman",
	})

	if len(evt.Links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(evt.Links))
	}

	link := evt.Links[0]
	if link.Uid != "TARGET1" {
		t.Errorf("Link UID = %v, want TARGET1", link.Uid)
	}
	if link.Type != "member" {
		t.Errorf("Link type = %v, want member", link.Type)
	}
	if link.Relation != "wingman" {
		t.Errorf("Link relation = %v, want wingman", link.Relation)
	}
}

func TestEventLogging(t *testing.T) {
	// Create a context with a test logger
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	// Store logger in context, but we're not using ctx directly in this test
	// This demonstrates how to set up a logger in context
	_ = WithLogger(ctx, logger)

	evt, err := NewEvent("testUID", "a-f-G", 25.5, -120.7, 0.0)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}
	if err := evt.Validate(); err != nil {
		t.Errorf("Event.Validate() error = %v", err)
	}
}

func TestValidateType(t *testing.T) {
	tests := []struct {
		name     string
		typ      string
		expected bool
	}{
		{
			name:     "empty type",
			typ:      "",
			expected: false,
		},
		{
			name:     "invalid type",
			typ:      "x",
			expected: false,
		},
		{
			name:     "invalid format",
			typ:      "a_b_c",
			expected: false,
		},
		{
			name:     "unknown prefix",
			typ:      "x-f-G",
			expected: false,
		},
		{
			name:     "too long",
			typ:      strings.Repeat("a-f-G-", 50),
			expected: false,
		},
		{
			name:     "valid friend ground",
			typ:      "a-f-G",
			expected: true,
		},
		{
			name:     "valid hostile air",
			typ:      "a-h-A",
			expected: true,
		},
		{
			name:     "valid detection",
			typ:      "d-f-C",
			expected: false,
		},
		{
			name:     "valid tasking",
			typ:      "t-x-c",
			expected: true,
		},
		{
			name:     "unknown but valid format",
			typ:      "a-f-Z-Q-X",
			expected: false,
		},
		{
			name:     "wildcard at end",
			typ:      "a-f-G-*",
			expected: true,
		},
		{
			name:     "wildcard in middle",
			typ:      "a-f-*-G",
			expected: false,
		},
		{
			name:     "atomic wildcard",
			typ:      "a-*",
			expected: true,
		},
		{
			name:     "catalog type",
			typ:      "a-f-G-E-X-N",
			expected: true,
		},
		{
			name:     "valid TAK chat",
			typ:      "t-x-c",
			expected: true,
		},
		{
			name:     "valid TAK drawing",
			typ:      "u-d-f",
			expected: true,
		},
		{
			name:     "valid TAK bits file",
			typ:      "b-t-f",
			expected: true,
		},
		{
			name:     "valid TAK reply",
			typ:      "y-c-r",
			expected: true,
		},
		{
			name:     "valid TAK route checkpoint",
			typ:      "b-r-f-h-c",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateType(tt.typ)
			if (err == nil) != tt.expected {
				t.Errorf("ValidateType(%q) = %v, want %v", tt.typ, err == nil, tt.expected)
				if err != nil {
					t.Logf("Error: %v", err)
				}
			}
			if err != nil && !errors.Is(err, ErrInvalidType) {
				t.Errorf("error does not wrap ErrInvalidType: %v", err)
			}
		})
	}
}

func TestWildcardExpansion(t *testing.T) {
	// Test that expanded types are valid
	expandedTypes := []string{
		"a-f-G",
		"a-h-G",
		"a-u-G",
		"a-f-G-U-C", // Further extension should be valid
		"a-h-G-E-V",
	}

	for _, typ := range expandedTypes {
		if err := ValidateType(typ); err != nil {
			t.Errorf("Expected expanded type %s to be valid, got error: %v", typ, err)
		}
	}
}

func TestRegisterCoTType(t *testing.T) {
	// Test registering a valid custom type that extends a standard prefix
	customType := "a-f-G-E-V-custom"
	RegisterCoTType(customType)
	if err := ValidateType(customType); err != nil {
		t.Errorf("Expected type %s with standard prefix to be valid after registration, got error: %v", customType, err)
	}

	// Test that invalid types cannot be registered
	invalidType := "a-f"
	RegisterCoTType(invalidType)
	if err := ValidateType(invalidType); err == nil {
		t.Error("Expected incomplete type to remain invalid even after registration")
	}
}

func TestValidateTypeDotWildcard(t *testing.T) {
	// Register a type that uses '.' as a wildcard for affiliation
	wildcardType := "a-.-Z"
	RegisterCoTType(wildcardType)

	if err := ValidateType("a-f-Z"); err != nil {
		t.Errorf("ValidateType failed wildcard resolution: %v", err)
	}
}

func TestValidateTypeDotWildcardNegative(t *testing.T) {
	if err := ValidateType("a-f-nonexistent"); err == nil {
		t.Error("expected validation to fail for missing wildcard match")
	}
}

func TestEmbeddedTypesValidation(t *testing.T) {
	// Test common tactical types
	tacticalTypes := []string{
		"a-f-G",     // Friendly ground
		"a-h-A",     // Hostile air
		"a-u-S",     // Unknown surface
		"a-n-U",     // Neutral subsurface
		"a-f-G-E-V", // Friendly ground vehicle
		"a-h-G-I",   // Hostile installation
	}

	for _, typ := range tacticalTypes {
		if err := ValidateType(typ); err != nil {
			t.Errorf("Embedded tactical type %q failed validation: %v", typ, err)
		}
	}

	// Test common bits types
	bitsTypes := []string{
		"b-i",   // Image
		"b-m-p", // Map point
		"b-m-r", // Route
		"b-d",   // Detection
		"b-l",   // Alarm
	}

	for _, typ := range bitsTypes {
		if err := ValidateType(typ); err != nil {
			t.Errorf("Embedded bits type %q failed validation: %v", typ, err)
		}
	}

	// Test common TAK protocol types
	takTypes := []string{
		"t-x-c",      // TAK server status
		"t-x-d",      // TAK data package
		"t-x-m",      // TAK mission package
		"t-x-t",      // TAK text message
		"t-x-takp-v", // TAK presence
	}

	for _, typ := range takTypes {
		if err := ValidateType(typ); err != nil {
			t.Errorf("Embedded TAK type %q failed validation: %v", typ, err)
		}
	}

	// Test common tasking types
	taskingTypes := []string{
		"t-k", // Strike
		"t-s", // ISR
		"t-m", // Mission
		"t-r", // Recon
		"t-u", // Update
		"t-q", // Query
	}

	for _, typ := range taskingTypes {
		if err := ValidateType(typ); err != nil {
			t.Errorf("Embedded tasking type %q failed validation: %v", typ, err)
		}
	}

	// Test common reply types
	replyTypes := []string{
		"y-a", // Ack
		"y-c", // Complete
		"y-s", // Status
	}

	for _, typ := range replyTypes {
		if err := ValidateType(typ); err != nil {
			t.Errorf("Embedded reply type %q failed validation: %v", typ, err)
		}
	}

	// Test common capability types
	capabilityTypes := []string{
		"c-f", // Fire support
		"c-c", // Command
		"c-r", // Recon
		"c-s", // Support
		"c-l", // Logistics
	}

	for _, typ := range capabilityTypes {
		if err := ValidateType(typ); err != nil {
			t.Errorf("Embedded capability type %q failed validation: %v", typ, err)
		}
	}
}

func TestPointValidation(t *testing.T) {
	tests := []struct {
		name    string
		point   *Point
		wantErr bool
	}{
		{
			name: "valid point",
			point: &Point{
				Lat: 37.7749,
				Lon: -122.4194,
				Hae: 100.0,
				Ce:  9999999.0,
				Le:  9999999.0,
			},
			wantErr: false,
		},
		{
			name: "sentinel HAE allowed",
			point: &Point{
				Lat: 37.7749,
				Lon: -122.4194,
				Hae: 9999999.0,
				Ce:  9999999.0,
				Le:  9999999.0,
			},
			wantErr: false,
		},
		{
			name: "HAE above limit",
			point: &Point{
				Lat: 37.7749,
				Lon: -122.4194,
				Hae: 40000001.0,
				Ce:  9999999.0,
				Le:  9999999.0,
			},
			wantErr: true,
		},
		{
			name: "invalid latitude",
			point: &Point{
				Lat: 91.0,
				Lon: -122.4194,
				Hae: 100.0,
				Ce:  9999999.0,
				Le:  9999999.0,
			},
			wantErr: true,
		},
		{
			name: "invalid longitude",
			point: &Point{
				Lat: 37.7749,
				Lon: 181.0,
				Hae: 100.0,
				Ce:  9999999.0,
				Le:  9999999.0,
			},
			wantErr: true,
		},
		{
			name: "invalid HAE",
			point: &Point{
				Lat: 37.7749,
				Lon: -122.4194,
				Hae: -13000.0,
				Ce:  9999999.0,
				Le:  9999999.0,
			},
			wantErr: true,
		},
		{
			name: "invalid CE",
			point: &Point{
				Lat: 37.7749,
				Lon: -122.4194,
				Hae: 100.0,
				Ce:  0.0,
				Le:  9999999.0,
			},
			wantErr: true,
		},
		{
			name: "invalid LE",
			point: &Point{
				Lat: 37.7749,
				Lon: -122.4194,
				Hae: 100.0,
				Ce:  9999999.0,
				Le:  0.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.point.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Point.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTypeCatalogFunctions(t *testing.T) {
	// Test GetTypeFullName
	t.Run("get_full_name", func(t *testing.T) {
		fullName, err := GetTypeFullName("a-f-G-E-X-N")
		if err != nil {
			t.Fatalf("GetTypeFullName() error = %v", err)
		}
		if fullName != "Gnd/Equip/Nbc Equipment" {
			t.Errorf("GetTypeFullName() = %v, want %v", fullName, "Gnd/Equip/Nbc Equipment")
		}

		// Test non-existent type
		_, err = GetTypeFullName("nonexistent-type")
		if err == nil {
			t.Error("GetTypeFullName() expected error for non-existent type")
		}
	})

	// Test GetTypeDescription
	t.Run("get_description", func(t *testing.T) {
		desc, err := GetTypeDescription("a-f-G-E-X-N")
		if err != nil {
			t.Fatalf("GetTypeDescription() error = %v", err)
		}
		if desc != "NBC EQUIPMENT" {
			t.Errorf("GetTypeDescription() = %v, want %v", desc, "NBC EQUIPMENT")
		}

		// Test non-existent type
		_, err = GetTypeDescription("nonexistent-type")
		if err == nil {
			t.Error("GetTypeDescription() expected error for non-existent type")
		}
	})

	// Test FindTypesByDescription
	t.Run("find_by_description", func(t *testing.T) {
		types := FindTypesByDescription("NBC")
		if len(types) == 0 {
			t.Error("FindTypesByDescription() returned no matches")
		}
		found := false
		for _, typ := range types {
			if typ.Name == "a-f-G-E-X-N" {
				found = true
				break
			}
		}
		if !found {
			t.Error("FindTypesByDescription() did not find expected type")
		}

		// Test no matches
		types = FindTypesByDescription("nonexistent")
		if len(types) != 0 {
			t.Error("FindTypesByDescription() returned matches for nonexistent description")
		}
	})

	// Test FindTypesByFullName
	t.Run("find_by_full_name", func(t *testing.T) {
		types := FindTypesByFullName("Nbc Equipment")
		if len(types) == 0 {
			t.Error("FindTypesByFullName() returned no matches")
		}
		found := false
		for _, typ := range types {
			if typ.Name == "a-f-G-E-X-N" {
				found = true
				break
			}
		}
		if !found {
			t.Error("FindTypesByFullName() did not find expected type")
		}

		// Test no matches
		types = FindTypesByFullName("nonexistent")
		if len(types) != 0 {
			t.Error("FindTypesByFullName() returned matches for nonexistent name")
		}
	})
}

func TestToXMLIncludesPointWithZeroCoordinates(t *testing.T) {
	now := time.Now().UTC()
	evt := &Event{
		Version: "2.0",
		Uid:     "zero",
		Type:    "a-f-G",
		Time:    CoTTime(now),
		Start:   CoTTime(now),
		Stale:   CoTTime(now.Add(10 * time.Second)),
		Point: Point{
			Lat: 0,
			Lon: 0,
			Ce:  9999999.0,
			Le:  9999999.0,
		},
	}
	xmlData, err := evt.ToXML()
	if err != nil {
		t.Fatalf("ToXML returned error: %v", err)
	}
	if !strings.Contains(string(xmlData), `version="2.0"`) {
		t.Error("version attribute missing in XML output")
	}
	if !strings.Contains(string(xmlData), "<point") {
		t.Error("point element missing in XML output")
	}
}

func TestToXMLEscapesControlChars(t *testing.T) {
	evt, err := NewEvent("base", "a-f-G", 0, 0, 0)
	if err != nil {
		t.Fatalf("NewEvent returned error: %v", err)
	}

	tests := []struct {
		name   string
		modify func(*Event)
		expect string
	}{
		{
			name:   "uid_newline",
			modify: func(e *Event) { e.Uid = "id\nend" },
			expect: "id&#xA;end",
		},
		{
			name:   "uid_cr",
			modify: func(e *Event) { e.Uid = "id\rend" },
			expect: "id&#xD;end",
		},
		{
			name:   "version_tab",
			modify: func(e *Event) { e.Version = "2.0\tbeta" },
			expect: "2.0&#x9;beta",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := *evt
			tc.modify(&e)
			xmlData, err := e.ToXML()
			if err != nil {
				t.Fatalf("ToXML returned error: %v", err)
			}
			if !strings.Contains(string(xmlData), tc.expect) {
				t.Errorf("expected %q in output: %s", tc.expect, string(xmlData))
			}
		})
	}
}

func TestValidateTypeWildcardResolution(t *testing.T) {
	tests := []struct {
		name    string
		typ     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "friendly pickup zone should resolve to neutral",
			typ:     "b-g-f-G-G-A-P",
			wantErr: false,
		},
		{
			name:    "neutral pickup zone should be valid",
			typ:     "b-g-.-G-G-A-P",
			wantErr: false,
		},
		{
			name:    "neutral landing zone should be valid",
			typ:     "b-g-.-G-G-A-L",
			wantErr: false,
		},
		{
			name:    "hostile pickup zone should resolve to neutral",
			typ:     "b-g-h-G-G-A-P",
			wantErr: false,
		},
		{
			name:    "unknown pickup zone should resolve to neutral",
			typ:     "b-g-u-G-G-A-P",
			wantErr: false,
		},
		{
			name:    "atomic wildcard should still work",
			typ:     "a-.-G",
			wantErr: false,
		},
		{
			name:    "invalid atomic wildcard should fail",
			typ:     "b-.-G",
			wantErr: true,
			errMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateType(tt.typ)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateType(%q) expected error but got none", tt.typ)
				} else if !errors.Is(err, ErrInvalidType) {
					t.Errorf("error does not wrap ErrInvalidType: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateType(%q) unexpected error = %v", tt.typ, err)
				}
			}
		})
	}
}

func TestLookupTypeWildcardResolution(t *testing.T) {
	typ, ok := LookupType("b-g-f-G-G-A-P")
	if !ok {
		t.Fatal("LookupType failed to resolve wildcard variation")
	}
	if typ.Description != "PICKUP ZONE (PZ)" {
		t.Errorf("unexpected description: %s", typ.Description)
	}

	if _, ok := LookupType("b-g-f-G-G-A-Q"); ok {
		t.Error("unexpected success for non-existent type")
	}
}

func TestEventBuilder(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	evt, err := NewEventBuilder("B1", "a-f-G", 10.0, -20.0, 0).
		WithContact(&Contact{Callsign: "ALPHA"}).
		WithGroup(&Group{Name: "Blue", Role: "Inf"}).
		WithStaleTime(now.Add(10 * time.Second)).
		Build()
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if evt.Detail == nil || evt.Detail.Contact == nil || evt.Detail.Contact.Callsign != "ALPHA" {
		t.Error("contact not set correctly")
	}
	if evt.Detail.Group == nil || evt.Detail.Group.Name != "Blue" {
		t.Error("group not set correctly")
	}
	if !evt.Stale.Time().Equal(now.Add(10 * time.Second)) {
		t.Error("stale time not set")
	}
	ReleaseEvent(evt)
}

func TestEventBuilderInvalid(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	_, err := NewEventBuilder("B2", "a-f-G", 10.0, -20.0, 0).
		WithStaleTime(now).
		Build()
	if err == nil {
		t.Error("expected error for stale time too close to event time")
	}
}

func TestCoTTimeRFC3339NanoRoundTrip(t *testing.T) {
	const ts = "2025-06-03T17:31:12.013Z"
	want, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t.Fatalf("parse ref time: %v", err)
	}

	t.Run("attr", func(t *testing.T) {
		input := fmt.Sprintf(`<t time="%s"/>`, ts)
		var out struct {
			XMLName xml.Name `xml:"t"`
			Time    CoTTime  `xml:"time,attr"`
		}
		if err := xml.Unmarshal([]byte(input), &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !out.Time.Time().Equal(want) {
			t.Errorf("parsed: got %v want %v", out.Time.Time(), want)
		}
		data, err := xml.Marshal(&out)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var r struct {
			XMLName xml.Name `xml:"t"`
			Time    CoTTime  `xml:"time,attr"`
		}
		if err := xml.Unmarshal(data, &r); err != nil {
			t.Fatalf("re-unmarshal: %v", err)
		}
		if !r.Time.Time().Equal(want.Truncate(time.Second)) {
			t.Errorf("round-trip: got %v want %v", r.Time.Time(), want.Truncate(time.Second))
		}
	})

	t.Run("element", func(t *testing.T) {
		input := fmt.Sprintf(`<t><time>%s</time></t>`, ts)
		var out struct {
			XMLName xml.Name `xml:"t"`
			Time    CoTTime  `xml:"time"`
		}
		if err := xml.Unmarshal([]byte(input), &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !out.Time.Time().Equal(want) {
			t.Errorf("parsed: got %v want %v", out.Time.Time(), want)
		}
		data, err := xml.Marshal(&out)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var r struct {
			XMLName xml.Name `xml:"t"`
			Time    CoTTime  `xml:"time"`
		}
		if err := xml.Unmarshal(data, &r); err != nil {
			t.Fatalf("re-unmarshal: %v", err)
		}
		if !r.Time.Time().Equal(want.Truncate(time.Second)) {
			t.Errorf("round-trip: got %v want %v", r.Time.Time(), want.Truncate(time.Second))
		}
	})
}
