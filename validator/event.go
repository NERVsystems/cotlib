package validator

import _ "embed"

//go:embed schemas/event/point.xsd
var eventPointXSD []byte

// EventPointXSD returns the raw XSD bytes for the event point schema.
func EventPointXSD() []byte { return eventPointXSD }
