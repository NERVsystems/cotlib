//go:build !novalidator

package cotlib

import (
	"encoding/xml"
	"fmt"

	"github.com/NERVsystems/cotlib/validator"
)

// eventPointSchema holds the compiled schema for CoT event points.
var eventPointSchema *validator.Schema

func init() {
	var err error
	eventPointSchema, err = validator.Compile(validator.EventPointXSD())
	if err != nil {
		panic(fmt.Errorf("compile event point schema: %w", err))
	}
}

// ValidateAgainstSchema validates the given CoT event XML against the point schema.
func ValidateAgainstSchema(data []byte) error {
	var p struct {
		Point Point `xml:"point"`
	}
	if err := xml.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	if (p.Point == Point{}) {
		return fmt.Errorf("missing point element")
	}
	wrapper := struct {
		XMLName xml.Name `xml:"point"`
		Point
	}{Point: p.Point}
	ptXML, err := xml.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("marshal point: %w", err)
	}
	return eventPointSchema.Validate(ptXML)
}
