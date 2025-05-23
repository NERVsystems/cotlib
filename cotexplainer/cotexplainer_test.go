package cotexplainer_test

import (
	"reflect"
	"testing"

	"github.com/NERVsystems/cotlib/cotexplainer"
)

func TestExplainType(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got, err := cotexplainer.ExplainType("a-f-G-E-X-N")
		if err != nil {
			t.Fatalf("ExplainType() error = %v", err)
		}
		want := []string{"Atom", "Friendly", "Ground", "EQUIPMENT", "SPECIAL EQUIPMENT", "NBC EQUIPMENT"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ExplainType() = %v, want %v", got, want)
		}
	})

	t.Run("unknown_atom", func(t *testing.T) {
		if _, err := cotexplainer.ExplainType("z-f-G"); err == nil {
			t.Error("expected error for unknown atom")
		}
	})

	t.Run("unknown_affiliation", func(t *testing.T) {
		if _, err := cotexplainer.ExplainType("a-x-G"); err == nil {
			t.Error("expected error for unknown affiliation")
		}
	})

	t.Run("unknown_dimension", func(t *testing.T) {
		if _, err := cotexplainer.ExplainType("a-f-Z"); err == nil {
			t.Error("expected error for unknown battle dimension")
		}
	})

	t.Run("unknown_segment", func(t *testing.T) {
		if _, err := cotexplainer.ExplainType("a-f-G-unknown"); err == nil {
			t.Error("expected error for unknown segment")
		}
	})
}
