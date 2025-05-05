package cotlib_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/pdfinn/cotlib"
)

// ExampleNewEvent demonstrates how to create a new CoT event.
func ExampleNewEvent() {
	// Create a new friendly ground unit
	evt := cotlib.NewEvent("UNIT1", cotlib.TypePredFriend+"-G", 45.0, -120.0)
	if evt == nil {
		log.Fatal("failed to create event")
	}

	// Add contact information
	evt.DetailContent.Contact.Callsign = "ALPHA1"

	// Add a shape
	evt.DetailContent.Shape.Type = "circle"
	evt.DetailContent.Shape.Radius = 1000 // meters

	// Marshal to XML
	xmlData, err := evt.ToXML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(xmlData))
	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	// <event version="2.0" uid="UNIT1" type="a-f-G" time="2024-03-14T12:00:00Z" start="2024-03-14T12:00:00Z" stale="2024-03-14T12:00:05Z">
	//   <point lat="45.000000" lon="-120.000000" hae="0.000000" ce="9999999.000000" le="9999999.000000"/>
	//   <detail>
	//     <contact callsign="ALPHA1"/>
	//     <shape type="circle" radius="1000.000000"/>
	//   </detail>
	// </event>
}

// ExampleEvent_Is demonstrates how to check event type predicates.
func ExampleEvent_Is() {
	evt := cotlib.NewEvent("UNIT1", cotlib.TypePredFriend+"-A", 45.0, -120.0)
	if evt == nil {
		log.Fatal("failed to create event")
	}

	fmt.Println("Is friendly:", evt.Is("friend"))
	fmt.Println("Is air:", evt.Is("air"))
	fmt.Println("Is ground:", evt.Is("ground"))
	// Output:
	// Is friendly: true
	// Is air: true
	// Is ground: false
}

// ExampleEvent_AddLink demonstrates how to create relationships between events.
func ExampleEvent_AddLink() {
	// Create a flight lead
	lead := cotlib.NewEvent("LEAD1", cotlib.TypePredFriend+"-A", 30.0, -85.0)
	if lead == nil {
		log.Fatal("failed to create lead event")
	}

	// Create a wingman
	wing := cotlib.NewEvent("WING1", cotlib.TypePredFriend+"-A", 30.0, -85.0)
	if wing == nil {
		log.Fatal("failed to create wing event")
	}

	// Link them
	lead.AddLink(wing.Uid, "member", "wingman")

	fmt.Println("Lead links:", len(lead.Links))
	fmt.Println("First link UID:", lead.Links[0].Uid)
	fmt.Println("First link type:", lead.Links[0].Type)
	// Output:
	// Lead links: 1
	// First link UID: WING1
	// First link type: member
}

// ExampleEvent_Validate demonstrates how to validate an event.
func ExampleEvent_Validate() {
	evt := cotlib.NewEvent("UNIT1", cotlib.TypePredFriend+"-G", 45.0, -120.0)
	if evt == nil {
		log.Fatal("failed to create event")
	}

	// Set invalid stale time (too close to event time)
	evt.Stale = evt.Time

	err := evt.Validate()
	fmt.Println("Validation error:", err)
	// Output:
	// Validation error: stale time must be more than 5s after event time
}

// ExampleUnmarshalXMLEvent demonstrates how to parse a CoT event from XML.
func ExampleUnmarshalXMLEvent() {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<event version="2.0" uid="UNIT1" type="a-f-G" time="2024-03-14T12:00:00Z" start="2024-03-14T12:00:00Z" stale="2024-03-14T12:00:05Z">
  <point lat="45.000000" lon="-120.000000" hae="0.000000" ce="9999999.000000" le="9999999.000000"/>
  <detail>
    <contact callsign="ALPHA1"/>
  </detail>
</event>`)

	evt, err := cotlib.UnmarshalXMLEvent(xmlData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("UID:", evt.Uid)
	fmt.Println("Type:", evt.Type)
	fmt.Println("Callsign:", evt.DetailContent.Contact.Callsign)
	// Output:
	// UID: UNIT1
	// Type: a-f-G
	// Callsign: ALPHA1
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
