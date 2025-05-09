package cottypes_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/NERVsystems/cotlib/cottypes"
)

func TestTypeMetadata(t *testing.T) {
	cat := cottypes.GetCatalog()

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

func TestTypeCatalogFunctions(t *testing.T) {
	cat := cottypes.GetCatalog()

	// Test GetTypeFullName
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

	// Test GetTypeDescription
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

func ExampleCatalog_GetFullName() {
	cat := cottypes.GetCatalog()
	fullName, err := cat.GetFullName("a-f-G-E-X-N")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Full name: %s\n", fullName)
	// Output: Full name: Gnd/Equip/Nbc Equipment
}

func ExampleCatalog_GetDescription() {
	cat := cottypes.GetCatalog()
	desc, err := cat.GetDescription("a-f-G-E-X-N")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Description: %s\n", desc)
	// Output: Description: NBC EQUIPMENT
}

func ExampleCatalog_FindByDescription() {
	cat := cottypes.GetCatalog()
	types := cat.FindByDescription("NBC EQUIPMENT")
	// Sort by name for consistent output
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})
	for _, t := range types {
		fmt.Printf("Found type: %s (%s)\n", t.Name, t.Description)
	}
	// Output:
	// Found type: a-f-G-E-X-N (NBC EQUIPMENT)
	// Found type: a-h-G-E-X-N (NBC EQUIPMENT)
	// Found type: a-n-G-E-X-N (NBC EQUIPMENT)
	// Found type: a-u-G-E-X-N (NBC EQUIPMENT)
}

func ExampleCatalog_FindByFullName() {
	cat := cottypes.GetCatalog()
	types := cat.FindByFullName("Gnd/Equip/Nbc Equipment")
	// Sort by name for consistent output
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})
	for _, t := range types {
		fmt.Printf("Found type: %s (%s)\n", t.Name, t.FullName)
	}
	// Output:
	// Found type: a-f-G-E-X-N (Gnd/Equip/Nbc Equipment)
	// Found type: a-h-G-E-X-N (Gnd/Equip/Nbc Equipment)
	// Found type: a-n-G-E-X-N (Gnd/Equip/Nbc Equipment)
	// Found type: a-u-G-E-X-N (Gnd/Equip/Nbc Equipment)
}
