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
	evt, err := cotlib.NewEvent("test123", "a-f-G", 30.0, -85.0)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}

	// Verify event properties
	if evt.Uid != "test123" {
		t.Errorf("Uid = %v, want test123", evt.Uid)
	}
	if evt.Type != "a-f-G" {
		t.Errorf("Type = %v, want a-f-G", evt.Type)
	}
	if evt.Point.Lat != 30.0 {
		t.Errorf("Point.Lat = %v, want 30.0", evt.Point.Lat)
	}
	if evt.Point.Lon != -85.0 {
		t.Errorf("Point.Lon = %v, want -85.0", evt.Point.Lon)
	}
	if evt.Point.Hae != 0.0 {
		t.Errorf("Point.Hae = %v, want 0.0", evt.Point.Hae)
	}

	// Verify time format
	now := time.Now().UTC().Truncate(time.Second)
	if evt.Time != now.Format(cotTimeFormat) {
		t.Errorf("Time = %v, want %v", evt.Time, now.Format(cotTimeFormat))
	}
	if evt.Start != now.Format(cotTimeFormat) {
		t.Errorf("Start = %v, want %v", evt.Start, now.Format(cotTimeFormat))
	}
	stale := now.Add(minStaleOffset + time.Second)
	if evt.Stale != stale.Format(cotTimeFormat) {
		t.Errorf("Stale = %v, want %v", evt.Stale, stale.Format(cotTimeFormat))
	}

	// Test creating an event with hae parameter
	evt, err = cotlib.NewEvent("test456", "a-f-G", 30.0, -85.0, 100.0)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}
	if evt.Point.Hae != 100.0 {
		t.Errorf("Point.Hae = %v, want 100.0", evt.Point.Hae)
	}
}

func TestTimeParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid Z format",
			input:    "2024-03-14T12:00:00Z",
			expected: "2024-03-14T12:00:00Z",
			wantErr:  false,
		},
		{
			name:     "valid with offset",
			input:    "2024-03-14T12:00:00+07:00",
			expected: "2024-03-14T05:00:00Z",
			wantErr:  false,
		},
		{
			name:     "valid with negative offset",
			input:    "2024-03-14T12:00:00-05:00",
			expected: "2024-03-14T17:00:00Z",
			wantErr:  false,
		},
		{
			name:     "invalid format",
			input:    "2024-03-14 12:00:00",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an event with the test time
			evt := &cotlib.Event{
				Version: "2.0",
				Uid:     "test123",
				Type:    "a-f-G",
				Time:    tt.input,
				Start:   tt.input,
				// Set stale time to be 1 minute after the event time
				Stale: time.Now().UTC().Add(time.Minute).Format(cotTimeFormat),
				Point: &cotlib.Point{Lat: 30.0, Lon: -85.0},
			}

			// Validate the event
			err := evt.Validate()

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

			// Check that times were normalized to Z format
			if evt.Time != tt.expected {
				t.Errorf("Time = %v, want %v", evt.Time, tt.expected)
			}
			if evt.Start != tt.expected {
				t.Errorf("Start = %v, want %v", evt.Start, tt.expected)
			}
			// Don't check Stale time as it's set to current time + 1 minute
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
				Time:    now.Format(time.RFC3339),
				Start:   now.Add(-time.Hour).Format(time.RFC3339),
				Stale:   now.Add(time.Hour).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: false,
		},
		{
			name: "invalid start time",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "a-f-G",
				Time:    now.Format(time.RFC3339),
				Start:   now.Add(time.Hour).Format(time.RFC3339),
				Stale:   now.Add(2 * time.Hour).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: true,
		},
		{
			name: "invalid stale time",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "a-f-G",
				Time:    now.Format(time.RFC3339),
				Start:   now.Add(-time.Hour).Format(time.RFC3339),
				Stale:   now.Add(4 * time.Second).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: true,
		},
		{
			name: "stale too far in future",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "t-x-d-d", // Use a TAK system message type
				Time:    now.Format(time.RFC3339),
				Start:   now.Add(-time.Hour).Format(time.RFC3339),
				Stale:   now.Add(8 * 24 * time.Hour).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: false, // TAK system messages can have long stale times
		},
		{
			name: "event too far in past",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "a-f-G",
				Time:    now.Add(-25 * time.Hour).Format(time.RFC3339),
				Start:   now.Add(-26 * time.Hour).Format(time.RFC3339),
				Stale:   now.Add(-24 * time.Hour).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 25.5, Lon: -120.7},
			},
			wantErr: true,
		},
		{
			name: "event too far in future",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "testUID",
				Type:    "a-f-G",
				Time:    now.Add(25 * time.Hour).Format(time.RFC3339),
				Start:   now.Add(24 * time.Hour).Format(time.RFC3339),
				Stale:   now.Add(26 * time.Hour).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 25.5, Lon: -120.7},
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
	evt, err := cotlib.NewEvent("testUID", "a-f-G", 25.5, -120.7)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}
	evt.Version = "2.0"
	evt.Time = now.Format(time.RFC3339)
	evt.Start = now.Add(-time.Hour).Format(time.RFC3339)
	evt.Stale = now.Add(time.Hour).Format(time.RFC3339)

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
	evt, err := cotlib.NewEvent("testUID", "a-f-G", 25.5, -120.7)
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
	evt, err := cotlib.NewEvent("testUID", "a-f-G", 25.5, -120.7)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}

	// Add a link
	evt.AddLink("TARGET1", "member", "wingman")

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

	evt, err := cotlib.NewEvent("testUID", "a-f-G", 25.5, -120.7)
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
