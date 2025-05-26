package cotlib_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/NERVsystems/cotlib"
)

func ExampleNewEvent() {
	// Create a new event with a friendly ground unit
	event, err := cotlib.NewEvent("test123", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Add some details
	event.Detail = &cotlib.Detail{
		Contact: &cotlib.Contact{
			Callsign: "TEST-1",
		},
	}

	// Print event details
	fmt.Printf("Event Type: %s\n", event.Type)
	fmt.Printf("Location: %.2f, %.2f\n", event.Point.Lat, event.Point.Lon)
	fmt.Printf("Callsign: %s\n", event.Detail.Contact.Callsign)

	// Output:
	// Event Type: a-f-G
	// Location: 30.00, -85.00
	// Callsign: TEST-1
}

func ExampleEvent_Is() {
	// Create a friendly ground unit event
	event, err := cotlib.NewEvent("test123", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Check various predicates
	fmt.Printf("Is friendly: %v\n", event.Is("friend"))
	fmt.Printf("Is hostile: %v\n", event.Is("hostile"))
	fmt.Printf("Is ground: %v\n", event.Is("ground"))
	fmt.Printf("Is air: %v\n", event.Is("air"))

	// Output:
	// Is friendly: true
	// Is hostile: false
	// Is ground: true
	// Is air: false
}

func ExampleEvent_AddLink() {
	// Create a main event
	event, err := cotlib.NewEvent("test123", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Add a link to another unit
	event.AddLink(&cotlib.Link{
		Uid:      "TARGET1",
		Type:     "a-f-G",
		Relation: "wingman",
	})

	// Print link details
	for _, link := range event.Links {
		fmt.Printf("Linked to: %s\n", link.Uid)
		fmt.Printf("Link type: %s\n", link.Type)
		fmt.Printf("Relation: %s\n", link.Relation)
	}

	// Output:
	// Linked to: TARGET1
	// Link type: a-f-G
	// Relation: wingman
}

func ExampleEvent_InjectIdentity() {
	// Create a new event
	event, err := cotlib.NewEvent("test123", "a-f-G", 30.0, -85.0, 0.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Inject identity information
	event.InjectIdentity("self123", "Blue", "HQ")

	// Print identity details
	if event.Detail != nil && event.Detail.Group != nil {
		fmt.Printf("Group: %s\n", event.Detail.Group.Name)
		fmt.Printf("Role: %s\n", event.Detail.Group.Role)
	}

	// Output:
	// Group: Blue
	// Role: HQ
}

func ExampleValidateType() {
	// Test various CoT types
	types := []string{
		"a-f-G",      // Friendly ground
		"a-h-A",      // Hostile air
		"b-d",        // Detection
		"t-x-takp-v", // TAK presence
		"invalid",    // Invalid type
	}

	for _, typ := range types {
		err := cotlib.ValidateType(typ)
		fmt.Printf("Type %s: %v\n", typ, err == nil)
	}

	// Output:
	// Type a-f-G: true
	// Type a-h-A: true
	// Type b-d: true
	// Type t-x-takp-v: true
	// Type invalid: false
}

// Example_typePredicates demonstrates using type predicates
func Example_typePredicates() {
	// Create some example events
	events := []*cotlib.Event{
		{Type: "a-f-G-U-C"}, // Friendly ground combat unit
		{Type: "a-h-A-M-F"}, // Hostile fixed wing aircraft
		{Type: "b-d-c-n-r"}, // NBC radiation detection
		{Type: "t-s-i-e"},   // ISR EO tasking
	}

	// Test various predicates
	predicates := []string{"atom", "friend", "hostile", "ground", "air"}

	for _, evt := range events {
		fmt.Printf("\nEvent type: %s\n", evt.Type)
		for _, pred := range predicates {
			if evt.Is(pred) {
				fmt.Printf("  Matches predicate: %s\n", pred)
			}
		}
	}

	// Output:
	// Event type: a-f-G-U-C
	//   Matches predicate: atom
	//   Matches predicate: friend
	//   Matches predicate: ground
	//
	// Event type: a-h-A-M-F
	//   Matches predicate: atom
	//   Matches predicate: hostile
	//   Matches predicate: air
	//
	// Event type: b-d-c-n-r
	//   Matches predicate: atom
	//
	// Event type: t-s-i-e
	//   Matches predicate: atom
}

func ExampleGetTypeFullName() {
	fullName, err := cotlib.GetTypeFullName("a-f-G-E-X-N")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Full name: %s\n", fullName)
	// Output: Full name: Gnd/Equip/Nbc Equipment
}

func ExampleGetTypeDescription() {
	desc, err := cotlib.GetTypeDescription("a-f-G-E-X-N")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Description: %s\n", desc)
	// Output: Description: NBC EQUIPMENT
}

func ExampleFindTypesByDescription() {
	types := cotlib.FindTypesByDescription("NBC EQUIPMENT")
	// Sort by name for consistent output
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})
	for _, t := range types {
		fmt.Printf("Found type: %s (%s)\n", t.Name, t.Description)
	}
	// Output:
	// Found type: a-f-G-E-X-N (NBC EQUIPMENT)
	// Found type: a-h-G-E-X-N (NBC EQUIPMENT)
	// Found type: a-n-G-E-X-N (NBC EQUIPMENT)
	// Found type: a-u-G-E-X-N (NBC EQUIPMENT)
}

func ExampleFindTypesByFullName() {
	types := cotlib.FindTypesByFullName("Gnd/Equip/Nbc Equipment")
	// Sort by name for consistent output
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})
	for _, t := range types {
		fmt.Printf("Found type: %s (%s)\n", t.Name, t.FullName)
	}
	// Output:
	// Found type: a-f-G-E-X-N (Gnd/Equip/Nbc Equipment)
	// Found type: a-h-G-E-X-N (Gnd/Equip/Nbc Equipment)
	// Found type: a-n-G-E-X-N (Gnd/Equip/Nbc Equipment)
	// Found type: a-u-G-E-X-N (Gnd/Equip/Nbc Equipment)
}

func Example_roundTripStrokeColorUsericon() {
	evt, _ := cotlib.NewEvent("EX-UI", "a-f-G", 0, 0, 0)
	evt.StrokeColor = "ff0000ff"
	evt.UserIcon = "icon.png"

	xmlData, _ := evt.ToXML()
	cotlib.ReleaseEvent(evt)

	out, _ := cotlib.UnmarshalXMLEvent(context.Background(), xmlData)
	fmt.Printf("strokeColor: %s\n", out.StrokeColor)
	fmt.Printf("usericon: %s\n", out.UserIcon)
	outXML, _ := out.ToXML()
	fmt.Printf("round-trip equal: %v\n", bytes.Equal(xmlData, outXML))
	cotlib.ReleaseEvent(out)
	// Output:
	// strokeColor: ff0000ff
	// usericon: icon.png
	// round-trip equal: true
}
