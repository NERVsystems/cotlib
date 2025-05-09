package cotlib

import (
	"fmt"
)

// Example_validateType demonstrates CoT type validation
func Example_validateType() {
	// Types that should be valid
	validTypes := []string{
		"a-f-G",     // Friendly ground
		"a-h-A",     // Hostile air
		"b-d",       // Detection
		"t-s",       // ISR tasking
		"a-f-G-E-V", // Friendly ground vehicle
		"a-h-G-I",   // Hostile installation
	}

	// Types that should be invalid
	invalidTypes := []string{
		"",             // Empty string
		"invalid-type", // Not a valid format
		"x-y-z",        // Unknown prefix
	}

	fmt.Println("Validating CoT types:")
	for _, typ := range validTypes {
		err := ValidateType(typ)
		if err != nil {
			fmt.Printf("Unexpectedly invalid: %s (%v)\n", typ, err)
		} else {
			fmt.Printf("Valid: %s\n", typ)
		}
	}

	fmt.Print("\nTesting invalid types:\n")
	for _, typ := range invalidTypes {
		err := ValidateType(typ)
		if err != nil {
			if typ == "" {
				fmt.Printf("Invalid as expected:\n")
			} else {
				fmt.Printf("Invalid as expected: %s\n", typ)
			}
		} else {
			fmt.Printf("Unexpectedly valid: %s\n", typ)
		}
	}

	// Output:
	// Validating CoT types:
	// Valid: a-f-G
	// Valid: a-h-A
	// Valid: b-d
	// Valid: t-s
	// Valid: a-f-G-E-V
	// Valid: a-h-G-I
	//
	// Testing invalid types:
	// Invalid as expected:
	// Invalid as expected: invalid-type
	// Invalid as expected: x-y-z
}

// Example_registerCustomCoTTypes demonstrates how to register custom CoT types
func Example_registerCustomCoTTypes() {
	// Custom organization-specific types
	customTypes := []string{
		"a-c-my-custom-type",      // Custom atom type
		"b-x-custom-detection",    // Custom detection
		"t-z-specialized-tasking", // Custom tasking
	}

	fmt.Println("Before registration:")
	for _, typ := range customTypes {
		err := ValidateType(typ)
		if err != nil {
			fmt.Printf("Invalid: %s\n", typ)
		} else {
			fmt.Printf("Valid: %s\n", typ)
		}
	}

	// Register custom types
	for _, typ := range customTypes {
		RegisterCoTType(typ)
	}

	fmt.Println("\nAfter registration:")
	for _, typ := range customTypes {
		err := ValidateType(typ)
		if err != nil {
			fmt.Printf("Invalid: %s\n", typ)
		} else {
			fmt.Printf("Valid: %s\n", typ)
		}
	}

	// Output:
	// Before registration:
	// Invalid: a-c-my-custom-type
	// Invalid: b-x-custom-detection
	// Invalid: t-z-specialized-tasking
	//
	// After registration:
	// Valid: a-c-my-custom-type
	// Valid: b-x-custom-detection
	// Valid: t-z-specialized-tasking
}

// Example_typePredicates demonstrates using type predicates
func Example_typePredicates() {
	// Create some example events
	events := []*Event{
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
