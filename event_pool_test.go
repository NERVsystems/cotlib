package cotlib

import (
	"runtime/debug"
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
