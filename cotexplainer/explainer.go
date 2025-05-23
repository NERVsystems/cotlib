package cotexplainer

import (
	"fmt"
	"strings"

	"github.com/NERVsystems/cotlib/cottypes"
)

var atomMap = map[string]string{
	"a": "Atom",
	"b": "Bits",
	"c": "Capability",
	"t": "Tasking",
	"y": "Reply",
}

var affiliationMap = map[string]string{
	"f": "Friendly",
	"h": "Hostile",
	"n": "Neutral",
	"u": "Unknown",
	"p": "Pending",
	"a": "Assumed Friend",
	"s": "Suspect",
}

var battleDimensionMap = map[string]string{
	"A": "Air",
	"G": "Ground",
	"S": "Surface",
	"U": "Subsurface",
	"X": "Other",
	"P": "Space",
}

// ExplainType resolves a CoT type code into its component meanings.
//
// It returns a slice of descriptions for each level of the type hierarchy.
func ExplainType(code string) ([]string, error) {
	if code == "" {
		return nil, fmt.Errorf("empty type")
	}

	parts := strings.Split(code, "-")
	if len(parts) < 3 { // need at least atom, affiliation and dimension
		return nil, fmt.Errorf("invalid type format")
	}

	res := make([]string, 0, len(parts))

	atom, ok := atomMap[parts[0]]
	if !ok {
		return nil, fmt.Errorf("unknown atom prefix: %s", parts[0])
	}
	res = append(res, atom)

	aff, ok := affiliationMap[parts[1]]
	if !ok {
		return nil, fmt.Errorf("unknown affiliation: %s", parts[1])
	}
	res = append(res, aff)

	dim, ok := battleDimensionMap[parts[2]]
	if !ok {
		return nil, fmt.Errorf("unknown battle dimension: %s", parts[2])
	}
	res = append(res, dim)

	prefix := strings.Join(parts[:3], "-")
	for _, seg := range parts[3:] {
		prefix += "-" + seg
		typ, err := cottypes.GetCatalog().GetType(prefix)
		if err != nil {
			return nil, fmt.Errorf("unknown type segment: %s", seg)
		}
		res = append(res, typ.Description)
	}

	return res, nil
}
