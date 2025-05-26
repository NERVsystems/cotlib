package validator

import _ "embed"

//go:embed schemas/event/point.xsd
var eventPointXSD []byte

// defaultEventPointXSD holds the original event point schema bytes for testing.
var defaultEventPointXSD = eventPointXSD

// EventPointXSD returns the raw XSD bytes for the event point schema.
func EventPointXSD() []byte { return eventPointXSD }
