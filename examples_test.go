package cotlib_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/pdfinn/cotlib"
)

func Example() {
	// Setup logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	cotlib.SetLogger(logger)

	// Create a new friendly ground unit
	evt := cotlib.NewEvent("UNIT1", cotlib.TypePredFriend+"-G", 45.0, -120.0)
	evt.How = "m-g" // GPS measurement

	// Add detail extensions
	evt.DetailContent.UidAliases = struct {
		Aliases []struct {
			Value string `xml:",chardata"`
		} `xml:"uidAlias,omitempty"`
	}{
		Aliases: []struct {
			Value string `xml:",chardata"`
		}{
			{Value: "ALPHA1"},
			{Value: "HMMWV"},
		},
	}

	// Add a shape for area of operations
	evt.DetailContent.Shape = struct {
		Type   string  `xml:"type,attr,omitempty"`
		Points string  `xml:"points,attr,omitempty"`
		Radius float64 `xml:"radius,attr,omitempty"`
	}{
		Type:   "circle",
		Radius: 1000, // meters
	}

	// Marshal to XML
	xmlData, err := evt.ToXML()
	if err != nil {
		fmt.Printf("Error marshaling XML: %v\n", err)
		return
	}

	fmt.Printf("Generated XML:\n%s\n", xmlData)
}

func ExampleNewEvent() {
	// Create a hostile air track
	evt := cotlib.NewEvent("TRACK1", cotlib.TypePredHostile+"-A", 45.0, -120.0)

	// Validate the event
	if err := evt.Validate(); err != nil {
		fmt.Printf("Error validating event: %v\n", err)
		return
	}

	fmt.Printf("Created valid event with UID: %s\n", evt.Uid)
}

func ExampleEvent_AddLink() {
	// Create a flight lead and wingman
	lead := cotlib.NewEvent("LEAD", cotlib.TypePredFriend+"-A", 45.0, -120.0)
	wing := cotlib.NewEvent("WING1", cotlib.TypePredFriend+"-A", 45.1, -120.1)

	// Link the wingman to the lead
	lead.AddLink(wing.Uid, "member", "wingman1")

	fmt.Printf("Flight lead has %d links\n", len(lead.Links))
}

func ExampleEvent_Is() {
	// Create a friendly ground unit
	evt := cotlib.NewEvent("UNIT1", cotlib.TypePredFriend+"-G", 45.0, -120.0)

	// Check predicates
	fmt.Printf("Is friend: %v\n", evt.Is("friend"))
	fmt.Printf("Is ground: %v\n", evt.Is("ground"))
	fmt.Printf("Is air: %v\n", evt.Is("air"))
}

func TestExamples(t *testing.T) {
	// Create a flight lead
	lead := cotlib.NewEvent("LEAD1", "a-f-A", 30.0090027, -85.9578735)
	if !lead.Is("friend") || !lead.Is("air") {
		t.Error("Lead should be a friendly air track")
	}

	// Create wingman
	wing := cotlib.NewEvent("WING1", "a-f-A", 30.0090027, -85.9578735)

	// Link them
	lead.AddLink(wing.Uid, "member", "wingman")

	// Add some contact info
	lead.DetailContent.Contact = struct {
		Callsign string `xml:"callsign,attr,omitempty"`
	}{
		Callsign: "LEAD",
	}

	// Add some aliases
	lead.DetailContent.UidAliases = struct {
		Aliases []struct {
			Value string `xml:",chardata"`
		} `xml:"uidAlias,omitempty"`
	}{
		Aliases: []struct {
			Value string `xml:",chardata"`
		}{
			{Value: "EAGLE1"},
		},
	}

	// Add a shape for area of operations
	lead.DetailContent.Shape = struct {
		Type   string  `xml:"type,attr,omitempty"`
		Points string  `xml:"points,attr,omitempty"`
		Radius float64 `xml:"radius,attr,omitempty"`
	}{
		Type:   "circle",
		Radius: 1000, // meters
	}

	// Set valid times
	now := time.Now().UTC()
	lead.Time = now.Format(time.RFC3339)
	lead.Start = now.Add(-time.Minute).Format(time.RFC3339)
	lead.Stale = now.Add(time.Hour).Format(time.RFC3339)

	// Validate the event
	if err := lead.Validate(); err != nil {
		t.Errorf("Failed to validate lead event: %v", err)
	}

	// Convert to XML
	data, err := lead.ToXML()
	if err != nil {
		t.Errorf("Failed to convert lead to XML: %v", err)
	}

	// Parse back from XML
	parsed, err := cotlib.UnmarshalXMLEvent(data)
	if err != nil {
		t.Errorf("Failed to parse XML: %v", err)
	}

	// Verify the parsed event
	if parsed.Uid != lead.Uid {
		t.Errorf("UID mismatch: got %s, want %s", parsed.Uid, lead.Uid)
	}
	if parsed.Type != lead.Type {
		t.Errorf("Type mismatch: got %s, want %s", parsed.Type, lead.Type)
	}
	if parsed.DetailContent.Contact.Callsign != lead.DetailContent.Contact.Callsign {
		t.Errorf("Callsign mismatch: got %s, want %s",
			parsed.DetailContent.Contact.Callsign,
			lead.DetailContent.Contact.Callsign)
	}
}
