package cottypes

import (
	"context"
	"testing"
)

// TestTypeValidation tests the GetType function for various valid and invalid CoT types.
func TestTypeValidation(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		typ     string
		wantErr bool
	}{
		{
			name:    "valid type",
			typ:     "a-f-G-U-C",
			wantErr: false,
		},
		{
			name:    "empty type",
			typ:     "",
			wantErr: true,
		},
		{
			name:    "incomplete type",
			typ:     "a-f-",
			wantErr: true,
		},
		{
			name:    "too long type",
			typ:     "a-f-G-U-C-" + string(make([]byte, 100)),
			wantErr: true,
		},
		{
			name:    "invalid format",
			typ:     "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetCatalog().GetType(ctx, tt.typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFindTypes tests the Find function for various query patterns.
func TestFindTypes(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		query    string
		wantLen  int
		contains string
	}{
		{
			name:     "find ground types",
			query:    "a-f-G",
			wantLen:  1,
			contains: "a-f-G",
		},
		{
			name:     "find air types",
			query:    "a-f-A",
			wantLen:  1,
			contains: "a-f-A",
		},
		{
			name:     "no matches",
			query:    "nonexistent",
			wantLen:  0,
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := GetCatalog().Find(ctx, tt.query)
			if len(matches) != tt.wantLen {
				t.Errorf("Find() returned %d matches, want %d", len(matches), tt.wantLen)
			}
			if tt.contains != "" {
				found := false
				for _, m := range matches {
					if m.Name == tt.contains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Find() did not contain expected type %q", tt.contains)
				}
			}
		})
	}
}
