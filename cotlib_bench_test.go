package cotlib

import "testing"

func BenchmarkNewEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := NewEvent("bench", "a-f-G", 30.0, -85.0, 0.0); err != nil {
			b.Fatalf("NewEvent returned error: %v", err)
		}
	}
}
