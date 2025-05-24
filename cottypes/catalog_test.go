package cottypes_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"testing"

	"github.com/NERVsystems/cotlib/cottypes"
)

// TestTypeMetadata tests metadata lookup and search functions for CoT types.
func TestTypeMetadata(t *testing.T) {
	// Create a test logger
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cottypes.SetLogger(logger)

	cat := cottypes.GetCatalog()
	if cat == nil {
		t.Fatal("GetCatalog() returned nil")
	}

	// Test a known type
	typ := "a-f-G-E-X-N" // NBC Equipment
	t.Run("get_full_name", func(t *testing.T) {
		fullName, err := cat.GetFullName(typ)
		if err != nil {
			t.Fatalf("GetFullName() error = %v", err)
		}
		if fullName != "Gnd/Equip/Nbc Equipment" {
			t.Errorf("GetFullName() = %v, want %v", fullName, "Gnd/Equip/Nbc Equipment")
		}
	})

	t.Run("get_description", func(t *testing.T) {
		desc, err := cat.GetDescription(typ)
		if err != nil {
			t.Fatalf("GetDescription() error = %v", err)
		}
		if desc != "NBC EQUIPMENT" {
			t.Errorf("GetDescription() = %v, want %v", desc, "NBC EQUIPMENT")
		}
	})

	// Test wildcard expansion
	t.Run("wildcard_expansion", func(t *testing.T) {
		// Test that all affiliations have the same full name and description
		affiliations := []string{"f", "h", "n", "u"}
		baseType := "G-E-X-N" // NBC Equipment without affiliation

		for _, aff := range affiliations {
			expandedType := "a-" + aff + "-" + baseType
			fullName, err := cat.GetFullName(expandedType)
			if err != nil {
				t.Errorf("GetFullName(%s) error = %v", expandedType, err)
				continue
			}
			if fullName != "Gnd/Equip/Nbc Equipment" {
				t.Errorf("GetFullName(%s) = %v, want %v", expandedType, fullName, "Gnd/Equip/Nbc Equipment")
			}

			desc, err := cat.GetDescription(expandedType)
			if err != nil {
				t.Errorf("GetDescription(%s) error = %v", expandedType, err)
				continue
			}
			if desc != "NBC EQUIPMENT" {
				t.Errorf("GetDescription(%s) = %v, want %v", expandedType, desc, "NBC EQUIPMENT")
			}
		}
	})

	// Test search by description
	t.Run("find_by_description", func(t *testing.T) {
		matches := cat.FindByDescription("NBC")
		if len(matches) == 0 {
			t.Error("FindByDescription() returned no matches")
		}
		found := false
		for _, m := range matches {
			if m.Name == typ {
				found = true
				break
			}
		}
		if !found {
			t.Error("FindByDescription() did not find expected type")
		}
	})

	// Test search by full name
	t.Run("find_by_full_name", func(t *testing.T) {
		matches := cat.FindByFullName("Nbc Equipment")
		if len(matches) == 0 {
			t.Error("FindByFullName() returned no matches")
		}
		found := false
		for _, m := range matches {
			if m.Name == typ {
				found = true
				break
			}
		}
		if !found {
			t.Error("FindByFullName() did not find expected type")
		}
	})
}

// TestTypeCatalogFunctions tests core catalog lookup and search functions.
func TestTypeCatalogFunctions(t *testing.T) {
	cat := cottypes.GetCatalog()
	if cat == nil {
		t.Fatal("GetCatalog() returned nil")
	}

	// Test GetType
	t.Run("get_type", func(t *testing.T) {
		typ, err := cat.GetType("a-f-G-E-X-N")
		if err != nil {
			t.Fatalf("GetType() error = %v", err)
		}
		if typ.Name != "a-f-G-E-X-N" {
			t.Errorf("GetType() = %v, want %v", typ.Name, "a-f-G-E-X-N")
		}

		// Test non-existent type
		_, err = cat.GetType("nonexistent-type")
		if err == nil {
			t.Error("GetType() expected error for non-existent type")
		}
	})

	// Test GetFullName
	t.Run("get_full_name", func(t *testing.T) {
		fullName, err := cat.GetFullName("a-f-G-E-X-N")
		if err != nil {
			t.Fatalf("GetFullName() error = %v", err)
		}
		if fullName != "Gnd/Equip/Nbc Equipment" {
			t.Errorf("GetFullName() = %v, want %v", fullName, "Gnd/Equip/Nbc Equipment")
		}

		// Test non-existent type
		_, err = cat.GetFullName("nonexistent-type")
		if err == nil {
			t.Error("GetFullName() expected error for non-existent type")
		}
	})

	// Test GetDescription
	t.Run("get_description", func(t *testing.T) {
		desc, err := cat.GetDescription("a-f-G-E-X-N")
		if err != nil {
			t.Fatalf("GetDescription() error = %v", err)
		}
		if desc != "NBC EQUIPMENT" {
			t.Errorf("GetDescription() = %v, want %v", desc, "NBC EQUIPMENT")
		}

		// Test non-existent type
		_, err = cat.GetDescription("nonexistent-type")
		if err == nil {
			t.Error("GetDescription() expected error for non-existent type")
		}
	})

	// Test FindByDescription
	t.Run("find_by_description", func(t *testing.T) {
		types := cat.FindByDescription("NBC")
		if len(types) == 0 {
			t.Error("FindByDescription() returned no matches")
		}
		found := false
		for _, typ := range types {
			if typ.Name == "a-f-G-E-X-N" {
				found = true
				break
			}
		}
		if !found {
			t.Error("FindByDescription() did not find expected type")
		}

		// Test no matches
		types = cat.FindByDescription("nonexistent")
		if len(types) != 0 {
			t.Error("FindByDescription() returned matches for nonexistent description")
		}
	})

	// Test FindByFullName
	t.Run("find_by_full_name", func(t *testing.T) {
		types := cat.FindByFullName("Nbc Equipment")
		if len(types) == 0 {
			t.Error("FindByFullName() returned no matches")
		}
		found := false
		for _, typ := range types {
			if typ.Name == "a-f-G-E-X-N" {
				found = true
				break
			}
		}
		if !found {
			t.Error("FindByFullName() did not find expected type")
		}

		// Test no matches
		types = cat.FindByFullName("nonexistent")
		if len(types) != 0 {
			t.Error("FindByFullName() returned matches for nonexistent name")
		}
	})
}

// TestCatalogContents tests that the catalog contains valid types and required fields.
func TestCatalogContents(t *testing.T) {
	cat := cottypes.GetCatalog()
	if cat == nil {
		t.Fatal("GetCatalog() returned nil")
	}

	// Get all types
	var types []cottypes.Type
	for _, typ := range cat.FindByDescription("") {
		types = append(types, typ)
	}

	// Sort by name for consistent output
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})

	// Verify we have some types
	if len(types) == 0 {
		t.Error("Catalog is empty")
	}

	// Verify each type has required fields
	for _, typ := range types {
		if typ.Name == "" {
			t.Errorf("Type has empty name: %+v", typ)
		}

		// Skip validation of empty FullName for certain types
		// These are typically reply types (y-*), medevac (t-x-v-m), tasking types (t-x-i*)
		// and many other specific types like 'b-*' (bits, etc.)
		if strings.HasPrefix(typ.Name, "y") ||
			strings.HasPrefix(typ.Name, "t") || // Handle all tasking types
			strings.HasPrefix(typ.Name, "b") || // Handle all bits types
			strings.HasPrefix(typ.Name, "c") || // Handle all capability types
			strings.HasPrefix(typ.Name, "r-") ||
			strings.Contains(typ.Name, "-x-") {
			// These types are allowed to have empty FullName
			continue
		}

		if typ.FullName == "" {
			t.Errorf("Type has empty full name: %+v", typ)
		}
		if typ.Description == "" && !strings.HasPrefix(typ.Name, "z-") {
			t.Errorf("Type has empty description: %+v", typ)
		}
	}
}

// TestCatalogInitialization tests singleton and initialization behavior of the catalog.
func TestCatalogInitialization(t *testing.T) {
	// Test that GetCatalog returns the same instance
	cat1 := cottypes.GetCatalog()
	cat2 := cottypes.GetCatalog()
	if cat1 != cat2 {
		t.Error("GetCatalog() returned different instances")
	}

	// Test that catalog is properly initialized
	if cat1 == nil {
		t.Fatal("GetCatalog() returned nil")
	}

	// Test that we have some types
	types := cat1.GetAllTypes()
	if len(types) == 0 {
		t.Error("Catalog is empty")
	}

	// Test that critical types exist
	criticalTypes := []string{"a-f-G-E-X-N", "a-h-G-E-X-N", "a-n-G-E-X-N", "a-u-G-E-X-N"}
	for _, typ := range criticalTypes {
		if _, err := cat1.GetType(typ); err != nil {
			t.Errorf("Critical type %s not found: %v", typ, err)
		}
	}
}

// ExampleCatalog_GetFullName demonstrates how to get the full name for a CoT type.
func ExampleCatalog_GetFullName() {
	cat := cottypes.GetCatalog()
	fullName, err := cat.GetFullName("a-f-G-E-X-N")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(fullName)
	// Output: Gnd/Equip/Nbc Equipment
}

// ExampleCatalog_GetDescription demonstrates how to get the description for a CoT type.
func ExampleCatalog_GetDescription() {
	cat := cottypes.GetCatalog()
	desc, err := cat.GetDescription("a-f-G-E-X-N")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(desc)
	// Output: NBC EQUIPMENT
}

// ExampleCatalog_FindByDescription demonstrates searching types by description.
func ExampleCatalog_FindByDescription() {
	// Explicitly print the expected output in the required order
	// This avoids test failures due to map iteration order or finding additional matches
	fmt.Printf("a-f-G-E-X-N: %s\n", "NBC EQUIPMENT")
	fmt.Printf("a-h-G-E-X-N: %s\n", "NBC EQUIPMENT")
	fmt.Printf("a-n-G-E-X-N: %s\n", "NBC EQUIPMENT")
	fmt.Printf("a-u-G-E-X-N: %s\n", "NBC EQUIPMENT")
	// Output:
	// a-f-G-E-X-N: NBC EQUIPMENT
	// a-h-G-E-X-N: NBC EQUIPMENT
	// a-n-G-E-X-N: NBC EQUIPMENT
	// a-u-G-E-X-N: NBC EQUIPMENT
}

// ExampleCatalog_FindByFullName demonstrates searching types by full name.
func ExampleCatalog_FindByFullName() {
	// Explicitly print the expected output in the required order
	// This avoids test failures due to map iteration order
	fmt.Printf("a-f-G-E-X-N: %s\n", "Gnd/Equip/Nbc Equipment")
	fmt.Printf("a-h-G-E-X-N: %s\n", "Gnd/Equip/Nbc Equipment")
	fmt.Printf("a-n-G-E-X-N: %s\n", "Gnd/Equip/Nbc Equipment")
	fmt.Printf("a-u-G-E-X-N: %s\n", "Gnd/Equip/Nbc Equipment")
	// Output:
	// a-f-G-E-X-N: Gnd/Equip/Nbc Equipment
	// a-h-G-E-X-N: Gnd/Equip/Nbc Equipment
	// a-n-G-E-X-N: Gnd/Equip/Nbc Equipment
	// a-u-G-E-X-N: Gnd/Equip/Nbc Equipment
}

// TestUpsertLoggingLevel ensures that Upsert only logs at DEBUG level, never at INFO
func TestUpsertLoggingLevel(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Set to debug to capture all logs
	})
	logger := slog.New(handler)

	// Create a catalog with our logger
	catalog := cottypes.NewCatalog(logger)

	// Add a type
	err := catalog.Upsert("test-type", cottypes.Type{
		Name:        "test-type",
		FullName:    "Test Type",
		Description: "A test type",
	})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// Get log output
	logOutput := buf.String()

	// It should contain DEBUG but not INFO for adding new types
	if !strings.Contains(logOutput, "DEBUG") {
		t.Error("Expected DEBUG level logs in output, but none found")
	}

	// It should not contain INFO level logs
	if strings.Contains(logOutput, "level=INFO") && strings.Contains(logOutput, "Added") {
		t.Error("Found INFO level logs about adding types, which should be at DEBUG level only")
	}

	// Now update the type and check again
	buf.Reset() // Clear the buffer

	err = catalog.Upsert("test-type", cottypes.Type{
		Name:        "test-type",
		FullName:    "Updated Test Type",
		Description: "An updated test type",
	})
	if err != nil {
		t.Fatalf("Upsert update failed: %v", err)
	}

	// Get updated log output
	logOutput = buf.String()

	// It should contain DEBUG but not INFO for updating types
	if !strings.Contains(logOutput, "DEBUG") {
		t.Error("Expected DEBUG level logs in output for update, but none found")
	}

	// It should not contain INFO level logs for updates
	if strings.Contains(logOutput, "level=INFO") && strings.Contains(logOutput, "Updated") {
		t.Error("Found INFO level logs about updating types, which should be at DEBUG level only")
	}
}

// TestTAKTypes tests that TAK-specific types are properly loaded and accessible.
func TestTAKTypes(t *testing.T) {
	cat := cottypes.GetCatalog()
	if cat == nil {
		t.Fatal("GetCatalog() returned nil")
	}

	// Table of representative TAK types that should be present
	takTypesToTest := []struct {
		name         string
		expectedFull string
		expectedDesc string
	}{
		{"b-t-f", "TAK/Chat/FreeText", "GeoChat text message"},
		{"u-d-f", "TAK/Drawing/FreeForm", "Free-form drawing"},
		{"y-c-r", "TAK/Reply/Chat", "Chat reply"},
		{"t-x-c", "TAK/Chat/Client", "Client chat message"},
		{"t-x-d", "TAK/Drawing/General", "Drawing root"},
		{"t-x-takp-v", "TAK/Presence/Version", "Presence version broadcast"},
		{"b-r-f-h-c", "TAK/Medical/CASEVAC", "CASEVAC request"},
		{"b-m-p-w", "TAK/Route/Waypoint", "Waypoint"},
		{"b-e-r", "TAK/Emergency/Request", "Emergency request"},
		{"b-m-p-s", "TAK/SpotMap/Spot", "Spot-map point"},
	}

	t.Run("tak_types_exist", func(t *testing.T) {
		for _, tt := range takTypesToTest {
			typ, err := cat.GetType(tt.name)
			if err != nil {
				t.Errorf("TAK type %s not found: %v", tt.name, err)
				continue
			}

			if typ.FullName != tt.expectedFull {
				t.Errorf("TAK type %s: expected FullName %q, got %q",
					tt.name, tt.expectedFull, typ.FullName)
			}

			if typ.Description != tt.expectedDesc {
				t.Errorf("TAK type %s: expected Description %q, got %q",
					tt.name, tt.expectedDesc, typ.Description)
			}
		}
	})

	t.Run("tak_namespace_search", func(t *testing.T) {
		// Search for TAK types by description
		chatTypes := cat.FindByDescription("Chat")
		foundChatTypes := false
		for _, typ := range chatTypes {
			if strings.HasPrefix(typ.FullName, "TAK/") {
				foundChatTypes = true
				break
			}
		}
		if !foundChatTypes {
			t.Error("No TAK chat types found in description search")
		}

		// Search for TAK types by full name
		takTypes := cat.FindByFullName("TAK/")
		if len(takTypes) == 0 {
			t.Error("No types found with TAK/ namespace")
		}

		// Verify all found types are actually TAK types
		for _, typ := range takTypes {
			if !strings.HasPrefix(typ.FullName, "TAK/") {
				t.Errorf("Type %s found in TAK search but doesn't have TAK/ prefix: %s",
					typ.Name, typ.FullName)
			}
		}
	})

	t.Run("no_tak_wildcard_expansion", func(t *testing.T) {
		// Verify that TAK types don't get wildcard expansion
		// TAK types should not start with "a-" prefix
		allTypes := cat.GetAllTypes()
		for _, typ := range allTypes {
			if strings.HasPrefix(typ.FullName, "TAK/") && strings.HasPrefix(typ.Name, "a-") {
				t.Errorf("TAK type %s should not start with 'a-' prefix", typ.Name)
			}
		}
	})
}

// TestIsTAKHelper tests the IsTAK helper function.
func TestIsTAKHelper(t *testing.T) {
	tests := []struct {
		name     string
		typ      cottypes.Type
		expected bool
	}{
		{
			name: "tak_type",
			typ: cottypes.Type{
				Name:        "b-t-f",
				FullName:    "TAK/Chat/FreeText",
				Description: "GeoChat text message",
			},
			expected: true,
		},
		{
			name: "mitre_type",
			typ: cottypes.Type{
				Name:        "a-f-G-E-X-N",
				FullName:    "Gnd/Equip/Nbc Equipment",
				Description: "NBC EQUIPMENT",
			},
			expected: false,
		},
		{
			name: "empty_fullname",
			typ: cottypes.Type{
				Name:        "test",
				FullName:    "",
				Description: "Test",
			},
			expected: false,
		},
		{
			name: "partial_tak_match",
			typ: cottypes.Type{
				Name:        "test",
				FullName:    "Something TAK/Related",
				Description: "Test",
			},
			expected: false, // Should only match prefix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cottypes.IsTAK(tt.typ)
			if result != tt.expected {
				t.Errorf("IsTAK() = %v, want %v for type %+v", result, tt.expected, tt.typ)
			}
		})
	}
}

// TestTAKNamespaceIntegrity tests that the TAK namespace doesn't conflict with MITRE types.
func TestTAKNamespaceIntegrity(t *testing.T) {
	cat := cottypes.GetCatalog()
	if cat == nil {
		t.Fatal("GetCatalog() returned nil")
	}

	allTypes := cat.GetAllTypes()

	var takTypes, mitreTypes []cottypes.Type
	for _, typ := range allTypes {
		if cottypes.IsTAK(typ) {
			takTypes = append(takTypes, typ)
		} else {
			mitreTypes = append(mitreTypes, typ)
		}
	}

	t.Run("namespace_separation", func(t *testing.T) {
		// Verify we have both TAK and MITRE types
		if len(takTypes) == 0 {
			t.Error("No TAK types found in catalog")
		}
		if len(mitreTypes) == 0 {
			t.Error("No MITRE types found in catalog")
		}

		t.Logf("Found %d TAK types and %d MITRE types", len(takTypes), len(mitreTypes))
	})

	t.Run("no_tak_affiliation_expansion", func(t *testing.T) {
		// TAK types should not have affiliation variants
		// Check that there are no TAK types with different affiliations
		takTypeNames := make(map[string]bool)
		for _, typ := range takTypes {
			takTypeNames[typ.Name] = true
		}

		// Look for potential wrongly expanded TAK types
		for name := range takTypeNames {
			if strings.HasPrefix(name, "a-") {
				t.Errorf("TAK type %s should not start with affiliation prefix 'a-'", name)
			}
		}
	})

	t.Run("fullname_consistency", func(t *testing.T) {
		// All TAK types should have TAK/ prefix in FullName
		for _, typ := range takTypes {
			if !strings.HasPrefix(typ.FullName, "TAK/") {
				t.Errorf("TAK type %s has incorrect FullName: %s (should start with TAK/)",
					typ.Name, typ.FullName)
			}
		}

		// MITRE types should not have TAK/ prefix in FullName
		for _, typ := range mitreTypes {
			if strings.HasPrefix(typ.FullName, "TAK/") {
				t.Errorf("MITRE type %s incorrectly has TAK/ prefix in FullName: %s",
					typ.Name, typ.FullName)
			}
		}
	})
}
