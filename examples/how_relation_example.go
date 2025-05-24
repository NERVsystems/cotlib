package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/NERVsystems/cotlib"
	"github.com/NERVsystems/cotlib/cottypes"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Register all CoT types
	if err := cotlib.RegisterAllCoTTypes(); err != nil {
		log.Fatalf("Failed to register types: %v", err)
	}

	logger.Info("=== CoT How and Relation Values Example ===")

	// === PART 1: Working with How Values ===
	logger.Info("1. Working with How Values")

	// Create a CoT event
	event, err := cotlib.NewEvent("UNIT-GPS-001", "a-f-G-E-V", 37.4419, -122.1430, 100.0)
	if err != nil {
		log.Fatalf("Failed to create event: %v", err)
	}

	// Method 1: Set how value using descriptor (recommended)
	if err := cotlib.SetEventHowFromDescriptor(event, "gps"); err != nil {
		log.Fatalf("Failed to set how value: %v", err)
	}
	fmt.Printf("Set how using descriptor 'gps': %s\n", event.How)

	// Method 2: Set how value directly if you know the code
	event.How = "h-e" // manually entered
	fmt.Printf("Set how directly to manual entry: %s\n", event.How)

	// Validate how value
	if err := cotlib.ValidateHow(event.How); err != nil {
		log.Fatalf("How validation failed: %v", err)
	}
	fmt.Printf("How value '%s' is valid\n", event.How)

	// Get human-readable description of how value
	if desc, err := cotlib.GetHowDescriptor(event.How); err == nil {
		fmt.Printf("How '%s' means: %s\n", event.How, desc)
	}

	// === PART 2: Exploring Available How Values ===
	logger.Info("2. Exploring Available How Values")

	// Get all how values
	allHows := cottypes.GetAllHows()
	fmt.Printf("Total how values available: %d\n", len(allHows))

	// Show some key how values
	keyHows := []string{"h-g-i-g-o", "h-e", "m-g"}
	for _, how := range keyHows {
		if desc, err := cotlib.GetHowDescriptor(how); err == nil {
			fmt.Printf("  %s = %s\n", how, desc)
		}
	}

	// Search for GPS-related how values
	gpsHows := cottypes.FindHowsByDescriptor("gps")
	fmt.Printf("Found %d GPS-related how values:\n", len(gpsHows))
	for _, h := range gpsHows {
		if h.Nick != "" {
			fmt.Printf("  %s (%s) -> %s\n", h.What, h.Nick, h.Value)
		}
	}

	// === PART 3: Working with Relations ===
	logger.Info("3. Working with Relations")

	// Method 1: Add validated link (recommended)
	if err := event.AddValidatedLink("HQ-COMMAND", "a-f-G-U-C", "p-p"); err != nil {
		log.Fatalf("Failed to add validated link: %v", err)
	}
	fmt.Printf("Added validated parent-point link to HQ-COMMAND\n")

	// Method 2: Add link manually (validation happens during event.Validate())
	event.AddLink(&cotlib.Link{
		Uid:      "CHILD-UNIT-001",
		Type:     "a-f-G-E-V",
		Relation: "p-c", // parent-child
	})
	fmt.Printf("Added manual parent-child link to CHILD-UNIT-001\n")

	// Add connected relation
	event.AddLink(&cotlib.Link{
		Uid:      "SUPPORT-UNIT",
		Type:     "a-f-G-E-S",
		Relation: "c", // connected
	})
	fmt.Printf("Added connected link to SUPPORT-UNIT\n")

	// Validate all relations
	for i, link := range event.Links {
		if err := cotlib.ValidateRelation(link.Relation); err != nil {
			log.Fatalf("Link %d relation validation failed: %v", i, err)
		}

		// Get relation description
		if desc, err := cotlib.GetRelationDescription(link.Relation); err == nil {
			fmt.Printf("Link %d: %s -> %s (%s)\n", i+1, link.Uid, link.Relation, desc)
		}
	}

	// === PART 4: Exploring Available Relations ===
	logger.Info("4. Exploring Available Relations")

	// Get all relation values
	allRelations := cottypes.GetAllRelations()
	fmt.Printf("Total relation values available: %d\n", len(allRelations))

	// Show key relation values
	keyRelations := []string{"c", "p-p", "p-c", "p"}
	for _, rel := range keyRelations {
		if desc, err := cotlib.GetRelationDescription(rel); err == nil {
			fmt.Printf("  %s = %s\n", rel, desc)
		}
	}

	// Search for parent-related relations
	parentRelations := cottypes.FindRelationsByDescription("parent")
	fmt.Printf("Found %d parent-related relations:\n", len(parentRelations))
	for _, r := range parentRelations {
		fmt.Printf("  %s = %s\n", r.Cot, r.Description)
	}

	// === PART 5: Full Event Validation ===
	logger.Info("5. Full Event Validation")

	// Validate the complete event (includes how and relation validation)
	if err := event.Validate(); err != nil {
		log.Fatalf("Event validation failed: %v", err)
	}
	fmt.Printf("✅ Event with how='%s' and %d links validated successfully\n", event.How, len(event.Links))

	// === PART 6: Generate XML ===
	logger.Info("6. Generated XML Output")

	// Convert to XML
	xmlData, err := event.ToXML()
	if err != nil {
		log.Fatalf("Failed to generate XML: %v", err)
	}

	fmt.Println("Generated CoT XML:")
	fmt.Println(string(xmlData))

	// === PART 7: Error Handling Examples ===
	logger.Info("7. Error Handling Examples")

	// Test invalid how value
	if err := cotlib.ValidateHow("invalid-how-value"); err != nil {
		fmt.Printf("❌ Invalid how validation correctly failed: %v\n", err)
	}

	// Test invalid relation value
	if err := cotlib.ValidateRelation("invalid-relation"); err != nil {
		fmt.Printf("❌ Invalid relation validation correctly failed: %v\n", err)
	}

	// Test adding invalid link
	testEvent, _ := cotlib.NewEvent("test", "a-f-G", 0, 0, 0)
	if err := testEvent.AddValidatedLink("test", "invalid-type", "p-p"); err != nil {
		fmt.Printf("❌ Invalid link addition correctly failed: %v\n", err)
	}

	logger.Info("=== Example Complete ===")
	fmt.Printf("\nSummary:\n")
	fmt.Printf("- Event UID: %s\n", event.Uid)
	fmt.Printf("- Event Type: %s\n", event.Type)
	fmt.Printf("- How Value: %s (%s)\n", event.How, func() string {
		if desc, err := cotlib.GetHowDescriptor(event.How); err == nil {
			return desc
		}
		return "unknown"
	}())
	fmt.Printf("- Number of Links: %d\n", len(event.Links))
	for i, link := range event.Links {
		desc, _ := cotlib.GetRelationDescription(link.Relation)
		fmt.Printf("  Link %d: %s (%s -> %s)\n", i+1, link.Uid, link.Relation, desc)
	}
}
