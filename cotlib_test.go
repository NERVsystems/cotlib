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

func TestMain(m *testing.M) {
	// Set up logger for tests
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	cotlib.SetLogger(logger)
	os.Exit(m.Run())
}

func TestNewEvent(t *testing.T) {
	evt, err := cotlib.NewEvent("testUID", "a-f-G", 25.5, -120.7)
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}
	if evt == nil {
		t.Fatal("NewEvent returned nil event")
	}

	if evt.Uid != "testUID" {
		t.Errorf("Uid = %v, want %v", evt.Uid, "testUID")
	}
	if evt.Type != "a-f-G" {
		t.Errorf("Type = %v, want %v", evt.Type, "a-f-G")
	}
	if evt.Point == nil {
		t.Fatal("Point is nil")
	}
	if evt.Point.Lat != 25.5 {
		t.Errorf("Point.Lat = %v, want %v", evt.Point.Lat, 25.5)
	}
	if evt.Point.Lon != -120.7 {
		t.Errorf("Point.Lon = %v, want %v", evt.Point.Lon, -120.7)
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
