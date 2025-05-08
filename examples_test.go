package cotlib_test

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"time"

	"github.com/pdfinn/cotlib"
)

func Example() {
	// Create an event without hae parameter
	event1, err := cotlib.NewEvent("test-uid-1", "a-f-G-U-C", 37.7749, -122.4194, 0.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Create an event with hae parameter
	event2, err := cotlib.NewEvent("test-uid-2", "a-f-G-U-C", 37.7749, -122.4194, 100.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Marshal events to XML
	xml1, err := xml.Marshal(event1)
	if err != nil {
		fmt.Printf("Error marshaling event1: %v\n", err)
		return
	}

	xml2, err := xml.Marshal(event2)
	if err != nil {
		fmt.Printf("Error marshaling event2: %v\n", err)
		return
	}

	fmt.Println(string(xml1))
	fmt.Println(string(xml2))
}

func ExampleNewEvent() {
	// Create an event without hae parameter
	event1, err := cotlib.NewEvent("test-uid-1", "a-f-G-U-C", 37.7749, -122.4194, 0.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Create an event with hae parameter
	event2, err := cotlib.NewEvent("test-uid-2", "a-f-G-U-C", 37.7749, -122.4194, 100.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	fmt.Printf("Event 1: UID=%s, Type=%s, Location=(%f,%f), HAE=%f\n",
		event1.Uid, event1.Type, event1.Point.Lat, event1.Point.Lon, event1.Point.Hae)
	fmt.Printf("Event 2: UID=%s, Type=%s, Location=(%f,%f), HAE=%f\n",
		event2.Uid, event2.Type, event2.Point.Lat, event2.Point.Lon, event2.Point.Hae)
}

func Example_robustTimeParsing() {
	// Create an event with a timezone offset
	now := time.Now().UTC()
	event := &cotlib.Event{
		Version: "2.0",
		Uid:     "test-uid",
		Type:    "a-f-G-U-C",
		Time:    cotlib.CoTTime(now),
		Start:   cotlib.CoTTime(now),
		Stale:   cotlib.CoTTime(now.Add(time.Hour)),
		Point:   cotlib.Point{Lat: 37.7749, Lon: -122.4194, Hae: 0.0},
	}

	// Validate the event
	if err := event.Validate(); err != nil {
		fmt.Printf("Error validating event: %v\n", err)
		return
	}

	// Print the normalized times
	fmt.Printf("Original time: %s\n", event.Time.Format(time.RFC3339))
	fmt.Printf("Normalized time: %s\n", event.Time.Format(cotlib.CotTimeFormat))
}

func ExampleEvent_Validate() {
	// Create a new event
	evt, err := cotlib.NewEvent("sampleUID", "a-h-G", 25.5, -120.7, 0.0)
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
	evt, err := cotlib.NewEvent("sampleUID", "a-h-G", 25.5, -120.7, 0.0)
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
	// Create an event
	event, err := cotlib.NewEvent("test-uid", "a-f-G-U-C", 37.7749, -122.4194, 0.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Add a link to another event
	link := &cotlib.Link{
		Uid:      "target-uid",
		Type:     "a-f-G-U-C",
		Relation: "p-p",
	}
	event.AddLink(link)

	// Marshal to XML to see the result
	xmlData, err := xml.Marshal(event)
	if err != nil {
		fmt.Printf("Error marshaling event: %v\n", err)
		return
	}
	fmt.Println(string(xmlData))
}
