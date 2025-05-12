package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"

	"github.com/NERVsystems/cotlib"
	"github.com/NERVsystems/cotlib/cottypes"
)

func main() {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo, // Set to INFO level like in production
	})
	logger := slog.New(handler)

	// Set up logging for both packages
	slog.SetDefault(logger)
	cotlib.SetLogger(logger)
	cottypes.SetLogger(logger)

	fmt.Println("Step 1: Using cottypes.RegisterXML directly")
	// Test loading with cottypes.RegisterXML
	data, err := os.ReadFile("cottypes/CoTtypes.xml")
	if err != nil {
		fmt.Printf("Error reading CoTtypes.xml: %v\n", err)
		os.Exit(1)
	}

	err = cottypes.RegisterXML(data)
	if err != nil {
		fmt.Printf("Error registering XML: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Log output from cottypes.RegisterXML:")
	fmt.Println(buf.String())

	// Clear the buffer for the next test
	buf.Reset()

	fmt.Println("\nStep 2: Using cotlib.RegisterCoTTypesFromFile")
	// Test loading with cotlib.RegisterCoTTypesFromFile
	err = cotlib.RegisterCoTTypesFromFile("cottypes/CoTtypes.xml")
	if err != nil {
		fmt.Printf("Error registering types from file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Log output from cotlib.RegisterCoTTypesFromFile:")
	fmt.Println(buf.String())

	// Clear the buffer for the next test
	buf.Reset()

	fmt.Println("\nStep 3: Using cotlib.LoadCoTTypesFromFile")
	// Test loading with cotlib.LoadCoTTypesFromFile
	err = cotlib.LoadCoTTypesFromFile("cottypes/CoTtypes.xml")
	if err != nil {
		fmt.Printf("Error loading types from file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Log output from cotlib.LoadCoTTypesFromFile:")
	fmt.Println(buf.String())

	fmt.Println("\nIf you still see 'Added new type' INFO logs in your production environment")
	fmt.Println("but they don't appear in any of the above outputs, it suggests that:")
	fmt.Println("1. Your application is using a custom way to register types")
	fmt.Println("2. Your application is using an older version of the library")
	fmt.Println("3. The logs are coming from a different part of your application")
}
