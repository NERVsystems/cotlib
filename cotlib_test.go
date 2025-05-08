package cotlib_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/pdfinn/cotlib"
)

// Constants for testing
const (
	cotTimeFormat  = "2006-01-02T15:04:05Z"
	minStaleOffset = 5 * time.Second
)

func TestMain(m *testing.M) {
	// Set up logger for tests
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cotlib.SetLogger(logger)
	os.Exit(m.Run())
}

func TestNewEvent(t *testing.T) {
	// Test creating an event without hae parameter (defaults to 0)
	evt, err := cotlib.NewEvent("test123", "a-f-G", 30.0, -85.0, 0.0)
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
	evt, err = cotlib.NewEvent("test456", "a-f-G", 30.0, -85.0, 100.0)
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
	evt, err := cotlib.NewPresenceEvent("test123", 30.0, -85.0, 0.0)
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
	if staleDiff <= minStaleOffset {
		t.Errorf("Stale time difference = %v, want > %v", staleDiff, minStaleOffset)
	}
}

func TestInjectIdentity(t *testing.T) {
	evt, err := cotlib.NewEvent("test123", "a-f-G", 30.0, -85.0, 0.0)
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
		event   *cotlib.Event
		wantErr bool
	}{
		{
			name: "valid Z format",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "test123",
				Type:    "a-f-G",
				Time:    cotlib.CoTTime(refTime),
				Start:   cotlib.CoTTime(refTime),
				Stale:   cotlib.CoTTime(refTime.Add(10 * time.Second)),
				Point:   cotlib.Point{Lat: 30.0, Lon: -85.0},
			},
			wantErr: false,
		},
		{
			name: "valid with offset",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "test124",
				Type:    "a-f-G",
				Time:    cotlib.CoTTime(refTime),
				Start:   cotlib.CoTTime(refTime),
				Stale:   cotlib.CoTTime(refTime.Add(10 * time.Second)),
				Point:   cotlib.Point{Lat: 30.0, Lon: -85.0},
			},
			wantErr: false,
		},
		{
			name: "valid with negative offset",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "test125",
				Type:    "a-f-G",
				Time:    cotlib.CoTTime(refTime),
				Start:   cotlib.CoTTime(refTime),
				Stale:   cotlib.CoTTime(refTime.Add(10 * time.Second)),
				Point:   cotlib.Point{Lat: 30.0, Lon: -85.0},
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "test126",
				Type:    "a-f-G",
				Time:    cotlib.CoTTime(refTime),
				Start:   cotlib.CoTTime(refTime.Add(time.Hour)), // Invalid: start after time
				Stale:   cotlib.CoTTime(refTime.Add(10 * time.Second)),
				Point:   cotlib.Point{Lat: 30.0, Lon: -85.0},
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
	tests := []struct {
		name    string
		event   *cotlib.Event
		wantErr bool
	}{
		{
			name: "valid event",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "a-f-G",
				Time:    cotlib.CoTTime(now),
				Start:   cotlib.CoTTime(now.Add(-time.Hour)),
				Stale:   cotlib.CoTTime(now.Add(time.Hour)),
				Point:   cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: false,
		},
		{
			name: "invalid start time",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "a-f-G",
				Time:    cotlib.CoTTime(now),
				Start:   cotlib.CoTTime(now.Add(time.Hour)),
				Stale:   cotlib.CoTTime(now.Add(2 * time.Hour)),
				Point:   cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: true,
		},
		{
			name: "invalid stale time",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "a-f-G",
				Time:    cotlib.CoTTime(now),
				Start:   cotlib.CoTTime(now.Add(-time.Hour)),
				Stale:   cotlib.CoTTime(now.Add(4 * time.Second)),
				Point:   cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: true,
		},
		{
			name: "stale too far in future",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "t-x-d-d", // Use a TAK system message type
				Time:    cotlib.CoTTime(now),
				Start:   cotlib.CoTTime(now.Add(-time.Hour)),
				Stale:   cotlib.CoTTime(now.Add(8 * 24 * time.Hour)),
				Point:   cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: false, // TAK system messages can have long stale times
		},
		{
			name: "event too far in past",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "a-f-G",
				Time:    cotlib.CoTTime(now.Add(-25 * time.Hour)),
				Start:   cotlib.CoTTime(now.Add(-26 * time.Hour)),
				Stale:   cotlib.CoTTime(now.Add(-24 * time.Hour)),
				Point:   cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: true,
		},
		{
			name: "event too far in future",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "a-f-G",
				Time:    cotlib.CoTTime(now.Add(25 * time.Hour)),
				Start:   cotlib.CoTTime(now.Add(24 * time.Hour)),
				Stale:   cotlib.CoTTime(now.Add(26 * time.Hour)),
				Point:   cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.event.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Event.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventXMLRoundTrip(t *testing.T) {
	now := time.Now().UTC()
	evt, err := cotlib.NewEvent("testUID", "a-f-G", 25.5, -120.7, 0.0)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}
	evt.Version = "2.0"
	evt.Time = cotlib.CoTTime(now)
	evt.Start = cotlib.CoTTime(now.Add(-time.Hour))
	evt.Stale = cotlib.CoTTime(now.Add(time.Hour))

	xmlData, err := evt.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}

	roundTrip, err := cotlib.FromXML(xmlData)
	if err != nil {
		t.Fatalf("FromXML() error = %v", err)
	}

	if roundTrip.Uid != evt.Uid {
		t.Errorf("Round trip UID = %v, want %v", roundTrip.Uid, evt.Uid)
	}
}

func TestEventPredicate(t *testing.T) {
	evt, err := cotlib.NewEvent("testUID", "a-f-G", 25.5, -120.7, 0.0)
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
	evt, err := cotlib.NewEvent("testUID", "a-f-G", 25.5, -120.7, 0.0)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}

	// Add a link
	evt.AddLink(&cotlib.Link{
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
	ctx = cotlib.WithLogger(ctx, logger)

	evt, err := cotlib.NewEvent("testUID", "a-f-G", 25.5, -120.7, 0.0)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}
	if err := evt.Validate(); err != nil {
		t.Errorf("Event.Validate() error = %v", err)
	}
}

func TestEventUnmarshalXML(t *testing.T) {
	validTime := time.Now().UTC().Format(time.RFC3339)
	validStart := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	validStale := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)

	tests := []struct {
		name    string
		xml     string
		wantErr bool
	}{
		{
			name: "valid event",
			xml: fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<event version="2.0" uid="testUID" type="a-f-G-U-C" time="%s" start="%s" stale="%s">
  <detail>
    <shape type="circle" radius="1000"></shape>
    <remarks>Test remarks</remarks>
    <contact callsign="TEST"></contact>
    <status read="true"></status>
    <flowTags status="active" chain="command"></flowTags>
    <uidAliases>
      <uidAlias>alias1</uidAlias>
      <uidAlias>alias2</uidAlias>
    </uidAliases>
  </detail>
  <point lat="25.5" lon="-120.7" hae="0" ce="9.999999e+06" le="9.999999e+06"></point>
</event>`, validTime, validStart, validStale),
			wantErr: false,
		},
		{
			name: "invalid event with malformed XML",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<event version="2.0" uid="testUID" type="a-f-G-U-C" time="2024-01-01T00:00:00Z" start="2024-01-01T00:00:00Z" stale="2024-01-01T01:00:00Z">
  <point lat="invalid" lon="-120.7" hae="0" ce="9.999999e+06" le="9.999999e+06"></point>
</event>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cotlib.UnmarshalXMLEvent([]byte(tt.xml))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalXMLEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
