package cotlib_test

import (
	"fmt"
	"log/slog"

	"github.com/pdfinn/cotlib"
)

func Example() {
	// Create a new event without hae (defaults to 0)
	evt, err := cotlib.NewEvent("sampleUID", "a-h-G", 25.5, -120.7)
	if err != nil {
		slog.Error("failed to create event", "error", err)
		return
	}

	// Marshal to XML
	output, err := evt.ToXML()
	if err != nil {
		slog.Error("failed to marshal event", "error", err)
		return
	}
	fmt.Println(string(output))

	// Create another event with hae specified
	evt2, err := cotlib.NewEvent("sampleUID2", "a-h-G", 25.5, -120.7, 100.0)
	if err != nil {
		slog.Error("failed to create event with hae", "error", err)
		return
	}

	// Marshal to XML
	output2, err := evt2.ToXML()
	if err != nil {
		slog.Error("failed to marshal event with hae", "error", err)
		return
	}
	fmt.Println(string(output2))
}

func ExampleNewEvent() {
	// Create a new event without hae (defaults to 0)
	evt, err := cotlib.NewEvent("sampleUID", "a-h-G", 25.5, -120.7)
	if err != nil {
		slog.Error("failed to create event", "error", err)
		return
	}

	// Print some fields
	fmt.Printf("UID: %s\n", evt.Uid)
	fmt.Printf("Type: %s\n", evt.Type)
	fmt.Printf("Location: %.6f, %.6f\n", evt.Point.Lat, evt.Point.Lon)
	fmt.Printf("HAE: %.1f\n", evt.Point.Hae)

	// Create another event with hae specified
	evt2, err := cotlib.NewEvent("sampleUID2", "a-h-G", 25.5, -120.7, 100.0)
	if err != nil {
		slog.Error("failed to create event with hae", "error", err)
		return
	}

	// Print some fields
	fmt.Printf("\nUID: %s\n", evt2.Uid)
	fmt.Printf("Type: %s\n", evt2.Type)
	fmt.Printf("Location: %.6f, %.6f\n", evt2.Point.Lat, evt2.Point.Lon)
	fmt.Printf("HAE: %.1f\n", evt2.Point.Hae)
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
