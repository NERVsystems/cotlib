package cottypes_test

import (
	"strings"
	"testing"

	"github.com/NERVsystems/cotlib/cottypes"
)

// TestIntegrationRequirements verifies that all the implementation requirements are met.
func TestIntegrationRequirements(t *testing.T) {
	cat := cottypes.GetCatalog()
	if cat == nil {
		t.Fatal("GetCatalog() returned nil")
	}

	t.Run("generator_autonomy_preserved", func(t *testing.T) {
		// Verify that the generator loads from multiple XML files
		// by checking we have both MITRE and TAK types
		allTypes := cat.GetAllTypes()

		var mitreCount, takCount int
		for _, typ := range allTypes {
			if cottypes.IsTAK(typ) {
				takCount++
			} else {
				mitreCount++
			}
		}

		if mitreCount == 0 {
			t.Error("No MITRE types found - generator should load CoTtypes.xml")
		}
		if takCount == 0 {
			t.Error("No TAK types found - generator should load TAKtypes.xml")
		}

		t.Logf("Successfully loaded %d MITRE types and %d TAK types", mitreCount, takCount)
	})

	t.Run("dedicated_tak_namespace", func(t *testing.T) {
		// Verify all TAK types use the TAK/ namespace
		takTypes := cat.FindByFullName("TAK/")

		if len(takTypes) == 0 {
			t.Fatal("No types found with TAK/ namespace")
		}

		for _, typ := range takTypes {
			if !strings.HasPrefix(typ.FullName, "TAK/") {
				t.Errorf("Type %s found in TAK search but doesn't have TAK/ prefix: %s",
					typ.Name, typ.FullName)
			}

			// Verify TAK types don't use 'a-' prefix (MITRE affiliation)
			if strings.HasPrefix(typ.Name, "a-") {
				t.Errorf("TAK type %s should not use 'a-' prefix", typ.Name)
			}
		}

		t.Logf("Verified %d TAK types use proper TAK/ namespace", len(takTypes))
	})

	t.Run("enumerated_extension_set", func(t *testing.T) {
		// Test that key TAK extensions from the requirements are present
		requiredTAKTypes := []string{
			"u-d-f",     // Drawing FreeForm
			"b-t-f",     // Chat FreeText
			"b-r-f-h-c", // Medical CASEVAC
			"b-m-r",     // Route/Route
			"b-m-p-w",   // Route Waypoint
			"b-m-p-j",   // Route JumpPoint
			"b-e-r",     // Emergency Request
			"b-m-p-c-z", // Map BoundingBox
			"y-c-r",     // Reply Chat
		}

		for _, typeName := range requiredTAKTypes {
			typ, err := cat.GetType(typeName)
			if err != nil {
				t.Errorf("Required TAK type %s not found: %v", typeName, err)
				continue
			}

			if !cottypes.IsTAK(typ) {
				t.Errorf("Type %s should be identified as TAK type", typeName)
			}

			if typ.Description == "" {
				t.Errorf("TAK type %s missing description", typeName)
			}
		}
	})

	t.Run("catalog_integrity_retained", func(t *testing.T) {
		// Verify catalog operations work for both MITRE and TAK types

		// Test MITRE type lookup
		mitreType, err := cat.GetType("a-f-G-E-X-N")
		if err != nil {
			t.Errorf("MITRE type lookup failed: %v", err)
		} else if cottypes.IsTAK(mitreType) {
			t.Error("MITRE type incorrectly identified as TAK type")
		}

		// Test TAK type lookup
		takType, err := cat.GetType("b-t-f")
		if err != nil {
			t.Errorf("TAK type lookup failed: %v", err)
		} else if !cottypes.IsTAK(takType) {
			t.Error("TAK type not identified as TAK type")
		}

		// Test case-insensitive search works for both
		mitreResults := cat.FindByDescription("NBC")
		if len(mitreResults) == 0 {
			t.Error("Case-insensitive search failed for MITRE types")
		}

		takResults := cat.FindByDescription("Chat")
		foundTAKChat := false
		for _, result := range takResults {
			if cottypes.IsTAK(result) {
				foundTAKChat = true
				break
			}
		}
		if !foundTAKChat {
			t.Error("Case-insensitive search failed for TAK types")
		}
	})

	t.Run("prevent_regression", func(t *testing.T) {
		// This test itself serves as regression prevention
		// Verify we have a reasonable number of each type
		allTypes := cat.GetAllTypes()

		if len(allTypes) < 5000 {
			t.Errorf("Expected at least 5000 total types, got %d", len(allTypes))
		}

		takTypes := cat.FindByFullName("TAK/")
		if len(takTypes) < 50 {
			t.Errorf("Expected at least 50 TAK types, got %d", len(takTypes))
		}
	})

	t.Run("prefix_based_helper", func(t *testing.T) {
		// Test the IsTAK helper function
		testCases := []struct {
			typeName string
			expected bool
		}{
			{"b-t-f", true},        // TAK file transfer
			{"a-f-G-E-X-N", false}, // MITRE NBC Equipment
			{"t-x-c", true},        // TAK chat
			{"y-c-r", true},        // TAK reply
		}

		for _, tc := range testCases {
			typ, err := cat.GetType(tc.typeName)
			if err != nil {
				t.Errorf("Failed to get type %s: %v", tc.typeName, err)
				continue
			}

			result := cottypes.IsTAK(typ)
			if result != tc.expected {
				t.Errorf("IsTAK(%s) = %v, expected %v", tc.typeName, result, tc.expected)
			}
		}
	})
}
