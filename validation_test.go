package cotlib_test

import (
	"testing"

	"github.com/NERVsystems/cotlib"
)

func TestWildcardPatterns(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected bool
	}{
		{"valid wildcard", "a-f-G-U-C-*", true},
		{"invalid wildcard", "a-f-G-U-C-", false},
		{"invalid wildcard position", "a-f-G-U-*-C", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cotlib.ValidateType(tt.pattern)
			if (err == nil) != tt.expected {
				t.Errorf("ValidateType(%q) error = %v, want error = %v", tt.pattern, err, !tt.expected)
			}
		})
	}
}

// ... existing code ...
