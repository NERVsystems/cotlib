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

	// Prime pool with a single event
	base := getEvent()
	ReleaseEvent(base)

	// Parse invalid XML to trigger error after pool allocation
	invalid := []byte("<event><bad></event>")
	if _, err := UnmarshalXMLEvent(invalid); err == nil {
		t.Fatal("expected error from invalid XML")
	}

	// Get event from pool; should be the same object
	e := getEvent()
	if e != base {
		t.Error("event was not returned to pool after failure")
	}
	ReleaseEvent(e)
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

	base := getEvent()
	ReleaseEvent(base)

	if _, err := NewEvent("test", "a-f-G", 95, 0, 0); err == nil {
		t.Fatal("expected validation error")
	}

	e := getEvent()
	if e != base {
		t.Error("event was not returned to pool after validation failure")
	}
	ReleaseEvent(e)
}
