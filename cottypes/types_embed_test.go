package cottypes_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/NERVsystems/cotlib/cottypes"
	"github.com/NERVsystems/cotlib/ctxlog"
)

// TestRegisterXML tests that RegisterXML correctly logs a summary and not individual types
func TestRegisterXML(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	// Set the logger for the cottypes package
	cottypes.SetLogger(logger)

	// Read the XML file
	data, err := os.ReadFile("CoTtypes.xml")
	if err != nil {
		t.Fatalf("Failed to read XML file: %v", err)
	}

	// Register the XML types using a context with logger
	ctx := ctxlog.WithLogger(context.Background(), logger)
	err = cottypes.RegisterXML(ctx, data)
	if err != nil {
		t.Fatalf("Failed to register XML types: %v", err)
	}

	// Get log output
	logOutput := buf.String()
	t.Logf("Log output: %s", logOutput)

	// Verify debug logs are present
	if !strings.Contains(logOutput, "DEBUG") {
		t.Error("Expected DEBUG level logs in output, but none found")
	}

	// Verify that summary messages are logged
	if !strings.Contains(logOutput, "XML types registration complete") {
		t.Error("Expected 'XML types registration complete' summary message, but not found")
	}

	// Verify no INFO logs about individual type registrations
	if strings.Contains(logOutput, "INFO Added new type") {
		t.Error("Found 'INFO Added new type' messages, which should no longer be logged")
	}

	// Verify no logs about individual successful type additions
	if strings.Contains(logOutput, "level=INFO") && strings.Contains(logOutput, "Added") && strings.Contains(logOutput, "type") {
		t.Error("Found INFO level logs about individual type additions, which should be at DEBUG level")
	}
}

// TestRealWorldTypeRegistration simulates how the production environment might be using the code
func TestRealWorldTypeRegistration(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo, // Set to INFO level like in production
	})
	logger := slog.New(handler)

	// Set the logger for the cottypes package
	cottypes.SetLogger(logger)

	// Read the XML file
	data, err := os.ReadFile("CoTtypes.xml")
	if err != nil {
		t.Fatalf("Failed to read XML file: %v", err)
	}

	// Register the XML types using a context with logger
	ctx := ctxlog.WithLogger(context.Background(), logger)
	err = cottypes.RegisterXML(ctx, data)
	if err != nil {
		t.Fatalf("Failed to register XML types: %v", err)
	}

	// Get log output
	logOutput := buf.String()
	t.Logf("Log output: %s", logOutput)

	// Count how many "Added new type" messages are in the log
	addedTypeCount := strings.Count(logOutput, "Added new type")

	// We expect only a summary, not individual type registrations
	if addedTypeCount > 5 { // Allow a few possible messages for edge cases
		t.Errorf("Found %d 'Added new type' messages, but expected no more than 5", addedTypeCount)
	}
}
