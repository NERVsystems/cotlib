package cotlib_test

import (
	"encoding/xml"
	"fmt"
	"log/slog"

	"github.com/pdfinn/cotlib"
)

func Example() {
	// Create a new event without hae parameter (defaults to 0)
	evt, err := cotlib.NewEvent("test123", "a-f-G", 30.0, -85.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(evt, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling to XML: %v\n", err)
		return
	}
	fmt.Printf("Event without hae:\n%s\n\n", xmlData)

	// Create another event with hae parameter
	evt, err = cotlib.NewEvent("test456", "a-f-G", 30.0, -85.0, 100.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Marshal to XML
	xmlData, err = xml.MarshalIndent(evt, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling to XML: %v\n", err)
		return
	}
	fmt.Printf("Event with hae:\n%s\n", xmlData)
}

func ExampleNewEvent() {
	// Create a new event without hae parameter (defaults to 0)
	evt, err := cotlib.NewEvent("test123", "a-f-G", 30.0, -85.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	fmt.Printf("Event without hae:\n")
	fmt.Printf("UID: %s\n", evt.Uid)
	fmt.Printf("Type: %s\n", evt.Type)
	fmt.Printf("Location: %.6f, %.6f\n", evt.Point.Lat, evt.Point.Lon)
	fmt.Printf("HAE: %.1f\n", evt.Point.Hae)
	fmt.Printf("Time: %s\n", evt.Time)
	fmt.Printf("Start: %s\n", evt.Start)
	fmt.Printf("Stale: %s\n\n", evt.Stale)

	// Create another event with hae parameter
	evt, err = cotlib.NewEvent("test456", "a-f-G", 30.0, -85.0, 100.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	fmt.Printf("Event with hae:\n")
	fmt.Printf("UID: %s\n", evt.Uid)
	fmt.Printf("Type: %s\n", evt.Type)
	fmt.Printf("Location: %.6f, %.6f\n", evt.Point.Lat, evt.Point.Lon)
	fmt.Printf("HAE: %.1f\n", evt.Point.Hae)
	fmt.Printf("Time: %s\n", evt.Time)
	fmt.Printf("Start: %s\n", evt.Start)
	fmt.Printf("Stale: %s\n", evt.Stale)
}

func Example_robustTimeParsing() {
	// Create an event with a time that includes a timezone offset
	evt := &cotlib.Event{
		Version: "2.0",
		Uid:     "test123",
		Type:    "a-f-G",
		Time:    "2024-03-14T12:00:00+07:00", // Time with offset
		Start:   "2024-03-14T12:00:00+07:00", // Start with offset
		Stale:   "2024-03-14T12:00:00+07:00", // Stale with offset
		Point:   &cotlib.Point{Lat: 30.0, Lon: -85.0},
	}

	// Validate the event (this will normalize the times to UTC)
	err := evt.Validate()
	if err != nil {
		fmt.Printf("Error validating event: %v\n", err)
		return
	}

	fmt.Printf("Original time with offset: 2024-03-14T12:00:00+07:00\n")
	fmt.Printf("Normalized to UTC: %s\n", evt.Time)
	fmt.Printf("Start normalized to UTC: %s\n", evt.Start)
	fmt.Printf("Stale normalized to UTC: %s\n", evt.Stale)
}

func ExampleEvent_Validate() {
	// Create a new event
	evt, err := cotlib.NewEvent("sampleUID", "a-h-G", 25.5, -120.7)
	if err != nil {
		slog.Error("failed to create event", "error", err)
		return
	}

	// Validate the event
	if err := evt.Validate(); err != nil {
		slog.Error("event validation failed", "error", err)
		return
	}

	fmt.Println("Event is valid")
}

func ExampleEvent_Is() {
	// Create a new event
	evt, err := cotlib.NewEvent("sampleUID", "a-h-G", 25.5, -120.7)
	if err != nil {
		slog.Error("failed to create event", "error", err)
		return
	}

	// Check if it's a hostile ground track
	if evt.Is("hostile") && evt.Is("ground") {
		fmt.Println("This is a hostile ground track")
	}
}

func ExampleEvent_AddLink() {
	// Create two events
	evt1, err := cotlib.NewEvent("sampleUID1", "a-h-G", 25.5, -120.7)
	if err != nil {
		slog.Error("failed to create event 1", "error", err)
		return
	}

	evt2, err := cotlib.NewEvent("sampleUID2", "a-h-G", 25.6, -120.8)
	if err != nil {
		slog.Error("failed to create event 2", "error", err)
		return
	}

	// Link them
	evt1.AddLink(evt2.Uid, "member", "wingman")

	// Print the link
	for _, link := range evt1.Links {
		fmt.Printf("Link: %s -> %s (%s)\n", evt1.Uid, link.Uid, link.Relation)
	}
}
