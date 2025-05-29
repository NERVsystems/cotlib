package cotlib

import (
	"bytes"
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"testing"
)

func TestEventPoolReuseOnInvalidXML(t *testing.T) {
	// Disable GC to prevent sync.Pool cleanup
	pct := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(pct)

	// Count the number of events we can get from the pool initially
	var initialEvents []*Event
	for i := 0; i < 10; i++ {
		e := getEvent()
		initialEvents = append(initialEvents, e)
	}

	// Return them all to the pool
	for _, e := range initialEvents {
		ReleaseEvent(e)
	}

	// Parse invalid XML to trigger error after pool allocation
	invalid := []byte("<event><bad></event>")
	if _, err := UnmarshalXMLEvent(context.Background(), invalid); err == nil {
		t.Fatal("expected error from invalid XML")
	}

	// Get events from pool; at least one should be from our initial set
	// This tests that events are being returned to the pool after failures
	var foundReused bool
	for i := 0; i < 20; i++ {
		e := getEvent()
		for _, initial := range initialEvents {
			if e == initial {
				foundReused = true
				break
			}
		}
		ReleaseEvent(e)
		if foundReused {
			break
		}
	}

	if !foundReused {
		t.Error("event was not returned to pool after failure")
	}
}

func TestUnmarshalXMLEventCtxLogsError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)
	ctx := WithLogger(context.Background(), logger)

	if _, err := UnmarshalXMLEventCtx(ctx, []byte("<event><bad></event>")); err == nil {
		t.Fatal("expected error from invalid XML")
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "level=ERROR") {
		t.Errorf("expected error log, got: %s", logOutput)
	}
}

func TestNewEventPoolReuseOnValidationError(t *testing.T) {
	pct := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(pct)

	// Count the number of events we can get from the pool initially
	var initialEvents []*Event
	for i := 0; i < 10; i++ {
		e := getEvent()
		initialEvents = append(initialEvents, e)
	}

	// Return them all to the pool
	for _, e := range initialEvents {
		ReleaseEvent(e)
	}

	if _, err := NewEvent("test", "a-f-G", 95, 0, 0); err == nil {
		t.Fatal("expected validation error")
	}

	// Get events from pool; at least one should be from our initial set
	// This tests that events are being returned to the pool after failures
	var foundReused bool
	for i := 0; i < 20; i++ {
		e := getEvent()
		for _, initial := range initialEvents {
			if e == initial {
				foundReused = true
				break
			}
		}
		ReleaseEvent(e)
		if foundReused {
			break
		}
	}

	if !foundReused {
		t.Error("event was not returned to pool after validation failure")
	}
}
