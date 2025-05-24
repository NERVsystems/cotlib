package cottypes_test

import (
	"testing"

	"github.com/NERVsystems/cotlib/cottypes"
)

// TestHowFunctionality tests how value lookup and search functions.
func TestHowFunctionality(t *testing.T) {
	t.Run("get_how_value", func(t *testing.T) {
		// Test TAK-specific how values
		testCases := []struct {
			what     string
			expected string
		}{
			{"gps", "h-g-i-g-o"},
			{"entered", "h-e"},
			{"manual", "h-e"},
		}

		for _, tc := range testCases {
			value, err := cottypes.GetHowValue(tc.what)
			if err != nil {
				t.Errorf("GetHowValue(%s) error = %v", tc.what, err)
				continue
			}
			if value != tc.expected {
				t.Errorf("GetHowValue(%s) = %s, want %s", tc.what, value, tc.expected)
			}
		}

		// Test error case
		_, err := cottypes.GetHowValue("nonexistent")
		if err == nil {
			t.Error("GetHowValue(nonexistent) expected error")
		}

		// Test empty case
		_, err = cottypes.GetHowValue("")
		if err == nil {
			t.Error("GetHowValue('') expected error")
		}
	})

	t.Run("get_how_nick", func(t *testing.T) {
		// Test TAK-specific how nicknames
		testCases := []struct {
			cot      string
			expected string
		}{
			{"h-e", "manual"},
			{"h-g-i-g-o", "gps"},
		}

		for _, tc := range testCases {
			nick, err := cottypes.GetHowNick(tc.cot)
			if err != nil {
				t.Errorf("GetHowNick(%s) error = %v", tc.cot, err)
				continue
			}
			if nick != tc.expected {
				t.Errorf("GetHowNick(%s) = %s, want %s", tc.cot, nick, tc.expected)
			}
		}

		// Test error case
		_, err := cottypes.GetHowNick("nonexistent")
		if err == nil {
			t.Error("GetHowNick(nonexistent) expected error")
		}
	})

	t.Run("find_hows_by_descriptor", func(t *testing.T) {
		// Test finding how values by descriptor
		gpsHows := cottypes.FindHowsByDescriptor("gps")
		if len(gpsHows) == 0 {
			t.Error("FindHowsByDescriptor('gps') returned no results")
		}

		// Check that we find the right ones
		found := false
		for _, h := range gpsHows {
			if h.What == "gps" && h.Value == "h-g-i-g-o" {
				found = true
				break
			}
		}
		if !found {
			t.Error("FindHowsByDescriptor('gps') did not find expected how value")
		}

		// Test case insensitive search
		manualHows := cottypes.FindHowsByDescriptor("MANUAL")
		if len(manualHows) == 0 {
			t.Error("FindHowsByDescriptor('MANUAL') returned no results")
		}

		// Test empty returns all
		allHows := cottypes.FindHowsByDescriptor("")
		if len(allHows) == 0 {
			t.Error("FindHowsByDescriptor('') should return all how values")
		}
	})

	t.Run("get_all_hows", func(t *testing.T) {
		allHows := cottypes.GetAllHows()
		if len(allHows) == 0 {
			t.Error("GetAllHows() returned empty slice")
		}

		// Verify we have TAK-specific how values
		foundTAKHows := 0
		for _, h := range allHows {
			if h.What == "gps" || h.What == "entered" || h.What == "manual" {
				foundTAKHows++
			}
		}

		if foundTAKHows < 3 {
			t.Errorf("Expected at least 3 TAK how values, found %d", foundTAKHows)
		}
	})
}

// TestRelationFunctionality tests relation value lookup and search functions.
func TestRelationFunctionality(t *testing.T) {
	t.Run("get_relation_description", func(t *testing.T) {
		// Test TAK-specific relation values
		testCases := []struct {
			cot      string
			expected string
		}{
			{"c", "connected"},
			{"p-p", "parent-point"},
			{"p-c", "parent-child"},
		}

		for _, tc := range testCases {
			desc, err := cottypes.GetRelationDescription(tc.cot)
			if err != nil {
				t.Errorf("GetRelationDescription(%s) error = %v", tc.cot, err)
				continue
			}
			if desc != tc.expected {
				t.Errorf("GetRelationDescription(%s) = %s, want %s", tc.cot, desc, tc.expected)
			}
		}

		// Test error case
		_, err := cottypes.GetRelationDescription("nonexistent")
		if err == nil {
			t.Error("GetRelationDescription(nonexistent) expected error")
		}

		// Test empty case
		_, err = cottypes.GetRelationDescription("")
		if err == nil {
			t.Error("GetRelationDescription('') expected error")
		}
	})

	t.Run("find_relations_by_description", func(t *testing.T) {
		// Test finding relation values by description
		parentRelations := cottypes.FindRelationsByDescription("parent")
		if len(parentRelations) == 0 {
			t.Error("FindRelationsByDescription('parent') returned no results")
		}

		// Check that we find the right ones
		found := false
		for _, r := range parentRelations {
			if r.Cot == "p-p" && r.Description == "parent-point" {
				found = true
				break
			}
		}
		if !found {
			t.Error("FindRelationsByDescription('parent') did not find expected relation")
		}

		// Test case insensitive search
		connectedRelations := cottypes.FindRelationsByDescription("CONNECTED")
		if len(connectedRelations) == 0 {
			t.Error("FindRelationsByDescription('CONNECTED') returned no results")
		}

		// Test empty returns all
		allRelations := cottypes.FindRelationsByDescription("")
		if len(allRelations) == 0 {
			t.Error("FindRelationsByDescription('') should return all relations")
		}
	})

	t.Run("get_all_relations", func(t *testing.T) {
		allRelations := cottypes.GetAllRelations()
		if len(allRelations) == 0 {
			t.Error("GetAllRelations() returned empty slice")
		}

		// Verify we have TAK-specific relation values
		foundTAKRelations := 0
		for _, r := range allRelations {
			if r.Cot == "c" || r.Cot == "p-p" || r.Cot == "p-c" {
				foundTAKRelations++
			}
		}

		// Note: We might have duplicates (c appears twice in MITRE and TAK)
		if foundTAKRelations < 2 {
			t.Errorf("Expected at least 2 TAK relation values, found %d", foundTAKRelations)
		}
	})
}

// TestTAKSpecificHowRelations tests specifically the TAK values we requested.
func TestTAKSpecificHowRelations(t *testing.T) {
	t.Run("requested_how_values", func(t *testing.T) {
		// Test the specific how values from the original requirements
		requiredHows := []struct {
			desc string
			code string
		}{
			{"h-e", "entered manually"},
			{"h-g-i-g-o", "GPS"},
		}

		for _, req := range requiredHows {
			// Test lookup by what descriptor
			if req.code == "h-e" {
				value, err := cottypes.GetHowValue("entered")
				if err != nil {
					t.Errorf("GetHowValue('entered') error = %v", err)
				} else if value != "h-e" {
					t.Errorf("GetHowValue('entered') = %s, want h-e", value)
				}

				value, err = cottypes.GetHowValue("manual")
				if err != nil {
					t.Errorf("GetHowValue('manual') error = %v", err)
				} else if value != "h-e" {
					t.Errorf("GetHowValue('manual') = %s, want h-e", value)
				}
			}

			if req.code == "h-g-i-g-o" {
				value, err := cottypes.GetHowValue("gps")
				if err != nil {
					t.Errorf("GetHowValue('gps') error = %v", err)
				} else if value != "h-g-i-g-o" {
					t.Errorf("GetHowValue('gps') = %s, want h-g-i-g-o", value)
				}
			}
		}
	})

	t.Run("requested_relation_values", func(t *testing.T) {
		// Test the specific relation values from the original requirements
		requiredRelations := []struct {
			code string
			desc string
		}{
			{"c", "connected"},
			{"p-p", "parent-point"},
			{"p-c", "parent-child"},
		}

		for _, req := range requiredRelations {
			desc, err := cottypes.GetRelationDescription(req.code)
			if err != nil {
				t.Errorf("GetRelationDescription(%s) error = %v", req.code, err)
				continue
			}
			if desc != req.desc {
				t.Errorf("GetRelationDescription(%s) = %s, want %s", req.code, desc, req.desc)
			}
		}
	})
}

// TestHowRelationIntegration tests integration with the overall system.
func TestHowRelationIntegration(t *testing.T) {
	t.Run("data_consistency", func(t *testing.T) {
		// Verify that how and relation data is properly loaded
		allHows := cottypes.GetAllHows()
		allRelations := cottypes.GetAllRelations()

		if len(allHows) < 30 {
			t.Errorf("Expected at least 30 how values, got %d", len(allHows))
		}

		if len(allRelations) < 20 {
			t.Errorf("Expected at least 20 relation values, got %d", len(allRelations))
		}

		// Verify no duplicate how entries with same what/value
		seenHow := make(map[string]string)
		for _, h := range allHows {
			if h.What != "" && h.Value != "" {
				if existing, found := seenHow[h.What]; found && existing != h.Value {
					t.Logf("Multiple how values for '%s': %s and %s", h.What, existing, h.Value)
				}
				seenHow[h.What] = h.Value
			}
		}
	})

	t.Run("search_functionality", func(t *testing.T) {
		// Test that search functions work correctly
		gpsHows := cottypes.FindHowsByDescriptor("gps")
		if len(gpsHows) == 0 {
			t.Error("No GPS how values found")
		}

		parentRelations := cottypes.FindRelationsByDescription("parent")
		if len(parentRelations) == 0 {
			t.Error("No parent relation values found")
		}

		// Verify case insensitive search
		manualHows := cottypes.FindHowsByDescriptor("MANUAL")
		if len(manualHows) == 0 {
			t.Error("Case insensitive search failed for how values")
		}
	})
}
