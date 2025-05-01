package cotlib_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
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
	evt := cotlib.NewEvent("TEST1", "a-f-G", 45.0, -120.0)
	if evt == nil {
		t.Fatal("NewEvent() returned nil")
	}

	if evt.Uid != "TEST1" {
		t.Errorf("Uid = %v, want %v", evt.Uid, "TEST1")
	}
	if evt.Type != "a-f-G" {
		t.Errorf("Type = %v, want %v", evt.Type, "a-f-G")
	}
	if evt.Point == nil {
		t.Fatal("Point is nil")
	}
	if evt.Point.Lat != 45.0 {
		t.Errorf("Point.Lat = %v, want %v", evt.Point.Lat, 45.0)
	}
	if evt.Point.Lon != -120.0 {
		t.Errorf("Point.Lon = %v, want %v", evt.Point.Lon, -120.0)
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
			name: "valid times",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "TEST1",
				Type:    "a-f-G",
				Time:    now.Format(time.RFC3339),
				Start:   now.Add(-time.Hour).Format(time.RFC3339),
				Stale:   now.Add(time.Hour).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 45.0, Lon: -120.0},
			},
			wantErr: false,
		},
		{
			name: "start after time",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "TEST1",
				Type:    "a-f-G",
				Time:    now.Format(time.RFC3339),
				Start:   now.Add(time.Hour).Format(time.RFC3339),
				Stale:   now.Add(2 * time.Hour).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 45.0, Lon: -120.0},
			},
			wantErr: true,
		},
		{
			name: "stale too close to time",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "TEST1",
				Type:    "a-f-G",
				Time:    now.Format(time.RFC3339),
				Start:   now.Add(-time.Hour).Format(time.RFC3339),
				Stale:   now.Add(4 * time.Second).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 45.0, Lon: -120.0},
			},
			wantErr: true,
		},
		{
			name: "stale too far in future",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "TEST1",
				Type:    "a-f-G",
				Time:    now.Format(time.RFC3339),
				Start:   now.Add(-time.Hour).Format(time.RFC3339),
				Stale:   now.Add(8 * 24 * time.Hour).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 45.0, Lon: -120.0},
			},
			wantErr: true,
		},
		{
			name: "time too far in past",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "TEST1",
				Type:    "a-f-G",
				Time:    now.Add(-25 * time.Hour).Format(time.RFC3339),
				Start:   now.Add(-26 * time.Hour).Format(time.RFC3339),
				Stale:   now.Add(-24 * time.Hour).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 45.0, Lon: -120.0},
			},
			wantErr: true,
		},
		{
			name: "time too far in future",
			event: &cotlib.Event{
				Version: "2.0",
				Uid:     "TEST1",
				Type:    "a-f-G",
				Time:    now.Add(25 * time.Hour).Format(time.RFC3339),
				Start:   now.Add(24 * time.Hour).Format(time.RFC3339),
				Stale:   now.Add(26 * time.Hour).Format(time.RFC3339),
				Point:   &cotlib.Point{Lat: 45.0, Lon: -120.0},
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
	evt := cotlib.NewEvent("TEST1", "a-f-G", 45.0, -120.0)
	evt.Version = "2.0"
	evt.Time = now.Format(time.RFC3339)
	evt.Start = now.Add(-time.Hour).Format(time.RFC3339)
	evt.Stale = now.Add(time.Hour).Format(time.RFC3339)
	evt.How = "m-g"
	evt.DetailContent = cotlib.Detail{
		Remarks: struct {
			Content string `xml:",chardata"`
		}{
			Content: "Test remarks",
		},
		Contact: struct {
			Callsign string `xml:"callsign,attr,omitempty"`
		}{
			Callsign: "TEST",
		},
	}

	// Test XML serialization
	xmlData, err := evt.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}

	// Test XML deserialization
	newEvt, err := cotlib.FromXML(xmlData)
	if err != nil {
		t.Fatalf("FromXML() error = %v", err)
	}

	// Compare fields
	if newEvt.Uid != evt.Uid {
		t.Errorf("Uid = %v, want %v", newEvt.Uid, evt.Uid)
	}
	if newEvt.Type != evt.Type {
		t.Errorf("Type = %v, want %v", newEvt.Type, evt.Type)
	}
	if newEvt.Time != evt.Time {
		t.Errorf("Time = %v, want %v", newEvt.Time, evt.Time)
	}
	if newEvt.Start != evt.Start {
		t.Errorf("Start = %v, want %v", newEvt.Start, evt.Start)
	}
	if newEvt.Stale != evt.Stale {
		t.Errorf("Stale = %v, want %v", newEvt.Stale, evt.Stale)
	}
	if newEvt.How != evt.How {
		t.Errorf("How = %v, want %v", newEvt.How, evt.How)
	}
	if newEvt.Point.Lat != evt.Point.Lat {
		t.Errorf("Point.Lat = %v, want %v", newEvt.Point.Lat, evt.Point.Lat)
	}
	if newEvt.Point.Lon != evt.Point.Lon {
		t.Errorf("Point.Lon = %v, want %v", newEvt.Point.Lon, evt.Point.Lon)
	}
	if newEvt.DetailContent.Remarks.Content != evt.DetailContent.Remarks.Content {
		t.Errorf("DetailContent.Remarks.Content = %v, want %v", newEvt.DetailContent.Remarks.Content, evt.DetailContent.Remarks.Content)
	}
	if newEvt.DetailContent.Contact.Callsign != evt.DetailContent.Contact.Callsign {
		t.Errorf("DetailContent.Contact.Callsign = %v, want %v", newEvt.DetailContent.Contact.Callsign, evt.DetailContent.Contact.Callsign)
	}
}

func TestEventPredicate(t *testing.T) {
	evt := cotlib.NewEvent("TEST1", "a-f-G", 45.0, -120.0)

	tests := []struct {
		name      string
		predicate string
		want      bool
	}{
		{
			name:      "is atom",
			predicate: "atom",
			want:      true,
		},
		{
			name:      "is friend",
			predicate: "friend",
			want:      true,
		},
		{
			name:      "is ground",
			predicate: "ground",
			want:      true,
		},
		{
			name:      "not hostile",
			predicate: "hostile",
			want:      false,
		},
		{
			name:      "not air",
			predicate: "air",
			want:      false,
		},
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
	evt := cotlib.NewEvent("TEST1", "a-f-G", 45.0, -120.0)

	// Add a link
	evt.AddLink("TARGET1", "a-f-G", "child")

	// Verify link was added
	if len(evt.Links) != 1 {
		t.Errorf("AddLink() did not add link, got %d links", len(evt.Links))
	}

	link := evt.Links[0]
	if link.Uid != "TARGET1" {
		t.Errorf("Link.Uid = %v, want %v", link.Uid, "TARGET1")
	}
	if link.Type != "a-f-G" {
		t.Errorf("Link.Type = %v, want %v", link.Type, "a-f-G")
	}
	if link.Relation != "child" {
		t.Errorf("Link.Relation = %v, want %v", link.Relation, "child")
	}
}

func TestLoggerContext(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.Background()
	ctx = cotlib.WithLogger(ctx, logger)

	evt := cotlib.NewEvent("TEST1", "a-f-G", 45.0, -120.0)
	if err := evt.Validate(); err != nil {
		t.Errorf("Event.Validate() error = %v", err)
	}
}

func TestEventXML(t *testing.T) {
	now := time.Now().UTC()
	validTime := now.Format(time.RFC3339)
	validStart := now.Add(-time.Minute).Format(time.RFC3339)
	validStale := now.Add(time.Hour).Format(time.RFC3339)

	tests := []struct {
		name    string
		xml     string
		wantErr bool
	}{
		{
			name: "valid event",
			xml: fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<event version="2.0" uid="TEST1" type="a-f-G-U-C" time="%s" start="%s" stale="%s">
  <detail>
    <shape type="circle" radius="1000"></shape>
    <remarks>Test remarks</remarks>
    <contact callsign="TEST1"></contact>
    <status read="true"></status>
    <flowTags status="active" chain="command"></flowTags>
    <uidAliases>
      <uidAlias>ALIAS1</uidAlias>
      <uidAlias>ALIAS2</uidAlias>
    </uidAliases>
  </detail>
  <point lat="45" lon="-120" hae="0" ce="9.999999e+06" le="9.999999e+06"></point>
</event>`, validTime, validStart, validStale),
			wantErr: false,
		},
		{
			name: "invalid event with malformed XML",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<event version="2.0" uid="TEST1" type="a-f-G-U-C" time="2024-01-01T00:00:00Z" start="2024-01-01T00:00:00Z" stale="2024-01-01T01:00:00Z">
  <point lat="invalid" lon="-120" hae="0" ce="9.999999e+06" le="9.999999e+06"></point>
</event>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := cotlib.UnmarshalXMLEvent([]byte(tt.xml))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalXMLEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Marshal back to XML
			data, err := parsed.ToXML()
			if err != nil {
				t.Errorf("ToXML() error = %v", err)
				return
			}

			// Compare XML (ignoring whitespace)
			got := strings.Join(strings.Fields(string(data)), " ")
			want := strings.Join(strings.Fields(tt.xml), " ")
			if got != want {
				t.Errorf("XML mismatch:\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}
