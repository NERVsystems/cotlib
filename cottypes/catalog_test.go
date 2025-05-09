package cottypes_test

import (
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
