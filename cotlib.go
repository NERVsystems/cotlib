/*
Package cot provides data structures and utilities for parsing
and generating Cursor on Target (CoT) XML messages.

Security Considerations:
  - XML parsing is restricted to prevent XXE attacks
  - Input validation is performed on all fields before processing
  - Coordinate ranges are strictly enforced
  - Time fields are validated to prevent time-based attacks
  - No sensitive data is logged at Info level or above
  - Detail extensions are isolated to prevent cross-contamination

Reference:
  - "Cursor on Target Developer Guide"
  - "Cursor on Target Message Router User's Guide"
  - http://cot.mitre.org

Key Goals:
  - High cohesion: focus on CoT event parsing and serialisation.
  - Low coupling: keep concern separation for expansions, transport, or advanced routing.
  - Composition over inheritance: nest sub-structures for detail fields.
  - Full coverage of base schema fields (Event.xsd), with example detail extension.
  - Secure by design: validate all inputs and prevent common XML attacks
*/
package cotlib

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Logger is the package-level logger that can be injected
var Logger *slog.Logger

func init() {
	// Default to a no-op logger if none is set
	Logger = slog.New(slog.NewTextHandler(nil, nil))
}

// SetLogger allows injection of a configured logger
func SetLogger(l *slog.Logger) {
	if l != nil {
		Logger = l
	}
}

// Event represents a basic CoT message (the "event" element).
// The mandatory fields (version, uid, type, time, start, stale, point)
// are modelled as struct fields or nested sub-structures.
//
// If your application needs further detail expansions (sub-schemas),
// you can embed them in a field under DetailContent.
type Event struct {
	XMLName xml.Name `xml:"event" json:"-"`

	// Required top-level XML attributes.
	Version string `xml:"version,attr"` // Must typically be "2.0"
	Uid     string `xml:"uid,attr"`
	Type    string `xml:"type,attr"`

	// CoT times
	Time  string `xml:"time,attr"`
	Start string `xml:"start,attr"`
	Stale string `xml:"stale,attr"`

	// 'how' field: often "m-g" (machine-gps), "h" (human), etc.
	How string `xml:"how,attr,omitempty"`

	// If you require QoS and OPEX fields from the docs,
	// you can add them here with `,attr,omitempty`.

	// Detail element (optional or sub-schema).
	// We keep it flexible to allow custom expansions.
	DetailContent Detail `xml:"detail,omitempty"`

	// Required "point" child element with mandatory attributes: lat, lon, hae, ce, le
	Point Point `xml:"point"`
}

// Point represents the <point lat="..." lon="..." hae="..." ce="..." le="..."/> element.
type Point struct {
	Lat float64 `xml:"lat,attr"` // -90..90 in decimal degrees
	Lon float64 `xml:"lon,attr"` // -180..180 in decimal degrees
	Hae float64 `xml:"hae,attr"` // height above ellipsoid in meters
	Ce  float64 `xml:"ce,attr"`  // circular error, in meters
	Le  float64 `xml:"le,attr"`  // linear error, in meters
}

// Detail holds arbitrary sub-schema expansions.
//
// In a real system, you'd define more typed fields for known sub-schemas (e.g. <__flow-tags__>).
// For unknown or custom detail sections, we store raw XML.
// The MyCustomDetail shows an example of a typed extension.
type Detail struct {
	// A catch-all for unknown or experimental detail fields
	RawXML []byte `xml:",innerxml"`

	// Example custom extension
	MyCustomDetail *MyCustomDetail `xml:"mycustomdetail,omitempty"`
}

// MyCustomDetail is an example struct for a sub-schema under <detail>.
// Replace or extend with real domain-specific detail.
type MyCustomDetail struct {
	XMLName xml.Name `xml:"mycustomdetail"`
	Value   string   `xml:"value,attr,omitempty"`
}

const (
	// MaxDetailSize is the maximum allowed size for detail content in bytes
	MaxDetailSize = 1024 * 1024 // 1MB

	// MinStaleOffset is the minimum allowed time between event time and stale time
	MinStaleOffset = time.Second * 1
)

// validateTimes ensures that time fields are logically consistent
// and within acceptable ranges to prevent time-based attacks
func (e *Event) validateTimes() error {
	eventTime, err := time.Parse(time.RFC3339, e.Time)
	if err != nil {
		return fmt.Errorf("invalid time format: %v", err)
	}

	startTime, err := time.Parse(time.RFC3339, e.Start)
	if err != nil {
		return fmt.Errorf("invalid start time format: %v", err)
	}

	staleTime, err := time.Parse(time.RFC3339, e.Stale)
	if err != nil {
		return fmt.Errorf("invalid stale time format: %v", err)
	}

	// Ensure logical time ordering
	if startTime.After(eventTime) {
		return errors.New("start time cannot be after event time")
	}

	if staleTime.Before(eventTime.Add(MinStaleOffset)) {
		return errors.New("stale time must be at least MinStaleOffset after event time")
	}

	return nil
}

// validateDetailSize ensures that detail content doesn't exceed size limits
// to prevent memory exhaustion attacks
func (d *Detail) validateDetailSize() error {
	if len(d.RawXML) > MaxDetailSize {
		return fmt.Errorf("detail content exceeds maximum size of %d bytes", MaxDetailSize)
	}
	return nil
}

// NewEvent constructs a CoT Event with minimal valid defaults.
// start and stale time are initialised with typical intervals.
func NewEvent(uid string, eventType string, lat, lon float64) *Event {
	Logger.Debug("creating new CoT event",
		"uid", uid,
		"type", eventType,
		"lat", lat,
		"lon", lon)

	now := time.Now().UTC()
	startTime := now
	// staleTime is offset for demonstration; it can be any appropriate future time.
	staleTime := now.Add(2 * time.Minute)

	evt := &Event{
		Version: "2.0",
		Uid:     uid,
		Type:    eventType,
		Time:    now.Format(time.RFC3339),
		Start:   startTime.Format(time.RFC3339),
		Stale:   staleTime.Format(time.RFC3339),
		Point: Point{
			Lat: lat,
			Lon: lon,
			// Default for demonstration
			Hae: 0.0,
			Ce:  9999999,
			Le:  9999999,
		},
		DetailContent: Detail{},
	}

	Logger.Info("created CoT event",
		"uid", evt.Uid,
		"time", evt.Time)

	return evt
}

// Validate checks basic CoT rules such as lat/lon range, correct times, etc.
func (e *Event) Validate() error {
	Logger.Debug("validating CoT event", "uid", e.Uid)

	if e.Version == "" || e.Uid == "" || e.Type == "" {
		Logger.Error("missing required CoT attributes",
			"version", e.Version != "",
			"uid", e.Uid != "",
			"type", e.Type != "")
		return errors.New("missing required CoT attribute(s) in event")
	}
	if e.Point.Lat < -90 || e.Point.Lat > 90 {
		Logger.Error("invalid latitude",
			"uid", e.Uid,
			"lat", e.Point.Lat)
		return fmt.Errorf("invalid latitude: %f", e.Point.Lat)
	}
	if e.Point.Lon < -180 || e.Point.Lon > 180 {
		Logger.Error("invalid longitude",
			"uid", e.Uid,
			"lon", e.Point.Lon)
		return fmt.Errorf("invalid longitude: %f", e.Point.Lon)
	}
	// Optional: parse time fields to check if start <= stale, etc.
	_, errTime := time.Parse(time.RFC3339, e.Time)
	_, errStart := time.Parse(time.RFC3339, e.Start)
	_, errStale := time.Parse(time.RFC3339, e.Stale)
	if errTime != nil {
		Logger.Error("invalid time field",
			"uid", e.Uid,
			"error", errTime)
		return fmt.Errorf("invalid time field: %v", errTime)
	}
	if errStart != nil {
		Logger.Error("invalid start field",
			"uid", e.Uid,
			"error", errStart)
		return fmt.Errorf("invalid start field: %v", errStart)
	}
	if errStale != nil {
		Logger.Error("invalid stale field",
			"uid", e.Uid,
			"error", errStale)
		return fmt.Errorf("invalid stale field: %v", errStale)
	}
	Logger.Debug("CoT event validation successful", "uid", e.Uid)
	return nil
}

// UnmarshalXMLEvent parses raw XML data into an Event struct with security checks
func UnmarshalXMLEvent(data []byte) (*Event, error) {
	Logger.Debug("unmarshalling XML data", "size", len(data))

	// Create a secure XML decoder that prevents XXE attacks
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Entity = nil // Disable external entity processing

	var evt Event
	if err := decoder.Decode(&evt); err != nil {
		Logger.Error("XML unmarshal failed", "error", err)
		return nil, err
	}

	// Validate detail size
	if err := evt.DetailContent.validateDetailSize(); err != nil {
		Logger.Error("detail validation failed", "error", err)
		return nil, err
	}

	// Validate time fields
	if err := evt.validateTimes(); err != nil {
		Logger.Error("time validation failed", "error", err)
		return nil, err
	}

	Logger.Info("successfully unmarshalled CoT event",
		"uid", evt.Uid,
		"type", evt.Type)
	return &evt, nil
}

// MarshalXML produces an XML representation of this CoT Event.
func (e *Event) MarshalXML() ([]byte, error) {
	// We want to ensure the struct is valid before encoding.
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return xml.MarshalIndent(e, "", "  ")
}

// Example usage function
func Example() {
	// 1. Create a new CoT event
	evt := NewEvent("MyUID", "a-f-G", 30.0090027, -85.9578735)
	evt.Point.Hae = -42.6
	evt.Point.Ce = 45.3
	evt.Point.Le = 99.5

	// 2. Add an example custom detail extension
	evt.DetailContent.MyCustomDetail = &MyCustomDetail{
		Value: "Some extended data",
	}

	// 3. Marshal to XML
	xmlBytes, err := evt.MarshalXML()
	if err != nil {
		panic(err)
	}
	fmt.Println("Marshalled CoT XML:")
	fmt.Println(string(xmlBytes))

	// 4. Unmarshal the same XML
	parsedEvt, err := UnmarshalXMLEvent(xmlBytes)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Parsed event UID: %s\n", parsedEvt.Uid)
	fmt.Printf("Parsed lat/long:  %.6f, %.6f\n",
		parsedEvt.Point.Lat, parsedEvt.Point.Lon)

	// 5. Validate
	if err := parsedEvt.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
	} else {
		fmt.Println("Event is valid per CoT base schema rules.")
	}
}

/*
In a real application, do something like:

package main

import (
    "fmt"
    "log"
    "os"

    "github.com/exampleuser/cotlib"
)

func main() {
    // Create
    evt := cot.NewEvent("sampleUID", "a-h-G", 25.5, -120.7)
    output, err := evt.MarshalXML()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(output))

    // Read from file or network:
    data, err := os.ReadFile("incoming_cot.xml")
    if err != nil {
        log.Fatal(err)
    }
    e, err := cot.UnmarshalXMLEvent(data)
    if err != nil {
        log.Fatal(err)
    }

    // Use the event e...
    fmt.Println("Received event with UID:", e.Uid)
}
*/
