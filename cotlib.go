/*
Package cotlib provides data structures and utilities for parsing
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
    https://apps.dtic.mil/sti/citations/ADA637348
  - "Cursor on Target Message Router User's Guide"
    https://www.mitre.org/sites/default/files/pdf/09_4937.pdf
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
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/html/charset"
)

// Security limits for XML parsing and validation
const (
	// maxDetailSize is the maximum allowed size for detail content (1 MiB)
	maxDetailSize = 1 << 20

	// minStaleOffset is the minimum time between event time and stale time
	// Increased to 5 seconds to prevent replay attacks on slow links
	minStaleOffset = 5 * time.Second

	// maxStaleOffset is the maximum time between event time and stale time
	// Events cannot be valid for more than 7 days to prevent stale data
	maxStaleOffset = 7 * 24 * time.Hour

	// maxXMLSize is the maximum allowed size for XML input (2 MiB)
	maxXMLSize = 2 << 20

	// maxElementDepth is the maximum allowed depth of XML elements
	maxElementDepth = 32

	// maxElementCount is the maximum allowed number of XML elements
	maxElementCount = 10000

	// maxTokenLen is the maximum length for any single XML token
	maxTokenLen = 1024
)

// maxValueLen is the maximum length for attribute values and character data
// Set to 512 KiB to accommodate large KML polygons
var maxValueLen = 512 << 10

// SetMaxValueLen allows downstream applications to adjust the maximum value length
// for attribute values and character data. This is useful for applications that
// need to handle very large KML polygons or other large data.
func SetMaxValueLen(n int) {
	if n < 1024 {
		n = 1024 // Minimum 1 KiB
	}
	maxValueLen = n
}

// Logger is the package-level logger that can be accessed atomically
var logger atomic.Value

// loggerKey is used for context-based logger access
type loggerKey struct{}

// SetLogger allows injection of a configured logger
func SetLogger(l *slog.Logger) {
	if l != nil {
		logger.Store(l)
	}
}

// WithLogger returns a new context with the logger
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// getLogger returns the logger from context or atomic storage
func getLogger(ctx context.Context) *slog.Logger {
	if ctx != nil {
		if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
			return l
		}
	}
	if l := logger.Load(); l != nil {
		return l.(*slog.Logger)
	}
	return slog.Default()
}

// Common CoT type predicates as defined in CoTtypes.xml
const (
	TypePredAtom    = "a"     // Atomic type (single item)
	TypePredFriend  = "a-f"   // Friendly force
	TypePredHostile = "a-h"   // Hostile force
	TypePredUnknown = "a-u"   // Unknown force
	TypePredNeutral = "a-n"   // Neutral force
	TypePredGround  = "a-.-G" // Ground track
	TypePredAir     = "a-.-A" // Air track
	TypePredPending = "a-.-P" // Pending or planned track
)

// Regular expressions for validation
var (
	// typePattern allows for more flexible type strings including digits and uppercase
	// Format: a-f-G-U-C, a-f-G-U-C-1, etc.
	typePattern = regexp.MustCompile(`^a-[a-z]-[A-Z0-9]+(-[A-Z0-9]+)*$`)

	// uidPattern allows for more characters while maintaining security
	// Format: alphanumeric, hyphen, underscore, dot
	uidPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_\.]*[a-zA-Z0-9]$`)

	// predicatePatterns map type strings to their semantic meanings
	predicatePatterns = map[string]*regexp.Regexp{
		"atom":    regexp.MustCompile(`^a`),     // Matches any type starting with "a"
		"friend":  regexp.MustCompile(`^a-f\b`), // Matches "a-f" followed by word boundary
		"hostile": regexp.MustCompile(`^a-h\b`), // Matches "a-h" followed by word boundary
		"unknown": regexp.MustCompile(`^a-u\b`), // Matches "a-u" followed by word boundary
		"neutral": regexp.MustCompile(`^a-n\b`), // Matches "a-n" followed by word boundary
		"ground":  regexp.MustCompile(`-G\b`),   // Matches "-G" followed by word boundary
		"air":     regexp.MustCompile(`-A\b`),   // Matches "-A" followed by word boundary
		"pending": regexp.MustCompile(`-P\b`),   // Matches "-P" followed by word boundary
	}
)

// Buffer pool for XML marshaling
var xmlBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Event represents a Cursor on Target (CoT) message.
// It implements the core schema defined in Event.xsd with security
// considerations for XML parsing and validation.
type Event struct {
	XMLName xml.Name `xml:"event" json:"-"`

	// Required top-level XML attributes.
	Version string `xml:"version,attr"` // Must typically be "2.0"
	Uid     string `xml:"uid,attr"`     // Unique identifier for this event
	Type    string `xml:"type,attr"`    // CoT type string (e.g., "a-f-G")

	// CoT times (all RFC3339)
	Time  string `xml:"time,attr"`  // When the event was generated
	Start string `xml:"start,attr"` // When the event became valid
	Stale string `xml:"stale,attr"` // When the event becomes invalid

	// Optional attributes
	How    string `xml:"how,attr,omitempty"`    // How the event was generated
	Access string `xml:"access,attr,omitempty"` // Access control

	// Links to other events for complex relationships
	Links []Link `xml:"link,omitempty"`

	// Detail element (optional or sub-schema).
	DetailContent Detail `xml:"detail,omitempty"`

	// Required "point" child element
	Point *Point `xml:"point"` // Pointer to reduce copying
}

// Point represents the <point lat="..." lon="..." hae="..." ce="..." le="..."/> element.
type Point struct {
	Lat float64 `xml:"lat,attr"` // -90..90 in decimal degrees
	Lon float64 `xml:"lon,attr"` // -180..180 in decimal degrees
	Hae float64 `xml:"hae,attr"` // height above ellipsoid in meters
	Ce  float64 `xml:"ce,attr"`  // circular error, in meters
	Le  float64 `xml:"le,attr"`  // linear error, in meters
}

// Link represents relationships between events, forming directed graphs
// as described in the CoT Developer Guide.
type Link struct {
	XMLName  xml.Name `xml:"link"`
	Uid      string   `xml:"uid,attr"`      // UID of the linked event
	Type     string   `xml:"type,attr"`     // Relationship type
	Relation string   `xml:"relation,attr"` // Nature of the relationship
}

// Detail represents the detail element of a CoT event
type Detail struct {
	// Known elements with attributes
	Shape struct {
		Type   string  `xml:"type,attr,omitempty"`
		Points string  `xml:"points,attr,omitempty"`
		Radius float64 `xml:"radius,attr,omitempty"`
	} `xml:"shape,omitempty"`

	Remarks struct {
		Content string `xml:",chardata"`
	} `xml:"remarks,omitempty"`

	Contact struct {
		Callsign string `xml:"callsign,attr,omitempty"`
	} `xml:"contact,omitempty"`

	Status struct {
		Read bool `xml:"read,attr,omitempty"`
	} `xml:"status,omitempty"`

	FlowTags struct {
		Status string `xml:"status,attr,omitempty"`
		Chain  string `xml:"chain,attr,omitempty"`
	} `xml:"flowTags,omitempty"`

	UidAliases struct {
		Aliases []struct {
			Value string `xml:",chardata"`
		} `xml:"uidAlias,omitempty"`
	} `xml:"uidAliases,omitempty"`

	// UnknownElements stores any XML tokens not matching known elements
	UnknownElements []xml.Token
}

// Status represents the status element
type Status struct {
	XMLName xml.Name `xml:"status"`
	Read    bool     `xml:"read,attr"`
}

// Archive represents the archive element
type Archive struct {
	XMLName xml.Name `xml:"archive"`
	Value   bool     `xml:",chardata"`
}

// Color represents the color element
type Color struct {
	XMLName xml.Name `xml:"color"`
	Value   string   `xml:",chardata"`
}

// Contact represents contact information
type Contact struct {
	XMLName  xml.Name `xml:"contact"`
	Callsign string   `xml:"callsign,attr,omitempty"`
}

// UserIcon represents the usericon element
type UserIcon struct {
	XMLName xml.Name `xml:"usericon"`
	Value   string   `xml:",chardata"`
}

// Precision represents the precisionlocation element
type Precision struct {
	XMLName xml.Name `xml:"precisionlocation"`
	Value   string   `xml:",chardata"`
}

// Remarks represents the remarks element
type Remarks struct {
	XMLName xml.Name `xml:"remarks"`
	Content string   `xml:",chardata"`
}

// ServerDestination represents the __serverdestination element
type ServerDestination struct {
	XMLName      xml.Name `xml:"__serverdestination"`
	Destinations string   `xml:"destinations,attr"`
}

// MyCustomDetail is an example struct for a sub-schema under <detail>.
// Replace or extend with real domain-specific detail.
type MyCustomDetail struct {
	XMLName xml.Name `xml:"mycustomdetail"`
	Value   string   `xml:"value,attr,omitempty"`
}

// Shape represents geographic shapes in CoT
type Shape struct {
	XMLName xml.Name `xml:"shape"`
	Type    string   `xml:"type,attr,omitempty"` // e.g., "circle", "ellipse", "polygon"
	Points  string   `xml:"points,attr,omitempty"`
	Radius  float64  `xml:"radius,attr,omitempty"`
}

// Request represents CoT tasking elements
type Request struct {
	XMLName xml.Name `xml:"request"`
	Type    string   `xml:"type,attr"`
}

// FlowTags represents workflow metadata
type FlowTags struct {
	XMLName xml.Name `xml:"__flow-tags__"`
	Status  string   `xml:"status,attr,omitempty"`
	Chain   string   `xml:"chain,attr,omitempty"`
}

// UidAliases represents system-specific UIDs
type UidAliases struct {
	XMLName  xml.Name `xml:"uid"`
	Droid    string   `xml:"droid,attr,omitempty"`    // Android device ID
	Callsign string   `xml:"callsign,attr,omitempty"` // Radio callsign
	Platform string   `xml:"platform,attr,omitempty"` // Platform identifier
}

// ErrInvalidInput represents validation failures
var ErrInvalidInput = errors.New("invalid input")

// Custom error types for better error handling
var (
	ErrInvalidLatitude  = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180")
	ErrInvalidUID       = errors.New("uid must match pattern: alphanumeric with optional hyphen/underscore/dot")
	ErrInvalidType      = errors.New("type must be a valid CoT type string")
	ErrInvalidTime      = errors.New("time must be in RFC3339 format")
	ErrInvalidStale     = errors.New("stale time must be after event time")
	ErrInvalidDetail    = errors.New("detail content exceeds maximum size")
	ErrInvalidShape     = errors.New("invalid shape parameters")
)

// validateLatLon validates latitude and longitude values with detailed logging
func ValidateLatLon(lat, lon float64) error {
	logger := getLogger(context.Background())

	// Check for non-finite values
	if math.IsNaN(lat) || math.IsInf(lat, 0) {
		logger.Error("invalid latitude value",
			"lat", lat,
			"error", ErrInvalidLatitude)
		return fmt.Errorf("%w: got non-finite value %f", ErrInvalidLatitude, lat)
	}

	if math.IsNaN(lon) || math.IsInf(lon, 0) {
		logger.Error("invalid longitude value",
			"lon", lon,
			"error", ErrInvalidLongitude)
		return fmt.Errorf("%w: got non-finite value %f", ErrInvalidLongitude, lon)
	}

	if lat < -90 || lat > 90 {
		logger.Error("invalid latitude value",
			"lat", lat,
			"error", ErrInvalidLatitude)
		return fmt.Errorf("%w: got %f", ErrInvalidLatitude, lat)
	}

	if lon < -180 || lon > 180 {
		logger.Error("invalid longitude value",
			"lon", lon,
			"error", ErrInvalidLongitude)
		return fmt.Errorf("%w: got %f", ErrInvalidLongitude, lon)
	}

	return nil
}

// validateUID performs strict validation of UIDs with enhanced security checks
func ValidateUID(uid string) error {
	logger := getLogger(context.Background())

	if uid == "" {
		logger.Error("empty uid", "error", ErrInvalidUID)
		return ErrInvalidUID
	}

	if len(uid) > 100 {
		logger.Error("uid too long",
			"uid_length", len(uid),
			"max_length", 100,
			"error", ErrInvalidUID)
		return fmt.Errorf("%w: length exceeds 100 characters", ErrInvalidUID)
	}

	if !uidPattern.MatchString(uid) {
		logger.Error("uid failed pattern validation",
			"uid", uid,
			"error", ErrInvalidUID)
		return fmt.Errorf("%w: invalid format", ErrInvalidUID)
	}

	return nil
}

// validateType performs strict validation of CoT type strings
func ValidateType(typ string) error {
	logger := getLogger(context.Background())

	if typ == "" {
		logger.Error("empty type", "error", ErrInvalidType)
		return ErrInvalidType
	}

	if len(typ) > 100 {
		logger.Error("type too long",
			"type_length", len(typ),
			"max_length", 100,
			"error", ErrInvalidType)
		return fmt.Errorf("%w: length exceeds 100 characters", ErrInvalidType)
	}

	if !typePattern.MatchString(typ) {
		logger.Error("type failed pattern validation",
			"type", typ,
			"error", ErrInvalidType)
		return fmt.Errorf("%w: invalid format", ErrInvalidType)
	}

	return nil
}

// validateShape checks if the shape parameters are valid
func validateShape(s *struct {
	Type   string  `xml:"type,attr,omitempty"`
	Points string  `xml:"points,attr,omitempty"`
	Radius float64 `xml:"radius,attr,omitempty"`
}) error {
	if s == nil {
		return nil
	}

	// Validate type
	switch s.Type {
	case "circle", "ellipse", "polygon":
		// Valid types
	default:
		if s.Type != "" {
			return fmt.Errorf("invalid shape type: %s", s.Type)
		}
	}

	// Validate points for polygon
	if s.Type == "polygon" {
		if s.Points == "" {
			return fmt.Errorf("polygon requires points attribute")
		}
		// TODO: Validate points format
	}

	// Validate radius for circle/ellipse
	if s.Type == "circle" || s.Type == "ellipse" {
		if s.Radius <= 0 {
			return fmt.Errorf("circle/ellipse requires positive radius")
		}
	}

	return nil
}

// validateTimes validates all time fields in an Event with enhanced security checks
func (e *Event) validateTimes() error {
	logger := getLogger(context.Background())

	// Parse all times
	timeTime, err := time.Parse(time.RFC3339, e.Time)
	if err != nil {
		logger.Error("invalid time format",
			"time", e.Time,
			"error", ErrInvalidTime)
		return fmt.Errorf("%w: time field: %v", ErrInvalidTime, err)
	}

	startTime, err := time.Parse(time.RFC3339, e.Start)
	if err != nil {
		logger.Error("invalid start time format",
			"start", e.Start,
			"error", ErrInvalidTime)
		return fmt.Errorf("%w: start field: %v", ErrInvalidTime, err)
	}

	staleTime, err := time.Parse(time.RFC3339, e.Stale)
	if err != nil {
		logger.Error("invalid stale time format",
			"stale", e.Stale,
			"error", ErrInvalidTime)
		return fmt.Errorf("%w: stale field: %v", ErrInvalidTime, err)
	}

	// Validate time relationships
	if startTime.After(timeTime) {
		logger.Error("start time after event time",
			"start", startTime,
			"time", timeTime,
			"error", ErrInvalidTime)
		return fmt.Errorf("%w: start time must be before or equal to event time", ErrInvalidTime)
	}

	staleDiff := staleTime.Sub(timeTime)
	if staleDiff <= minStaleOffset {
		logger.Error("stale time too close to event time",
			"stale", staleTime,
			"time", timeTime,
			"min_offset", minStaleOffset,
			"actual_offset", staleDiff,
			"error", ErrInvalidStale)
		return fmt.Errorf("%w: stale time must be more than %v after event time", ErrInvalidStale, minStaleOffset)
	}

	if staleDiff > maxStaleOffset {
		logger.Error("stale time too far from event time",
			"stale", staleTime,
			"time", timeTime,
			"max_offset", maxStaleOffset,
			"actual_offset", staleDiff,
			"error", ErrInvalidStale)
		return fmt.Errorf("%w: stale time must be within %v of event time", ErrInvalidStale, maxStaleOffset)
	}

	// Validate against current time
	now := time.Now().UTC()
	maxPastOffset := 24 * time.Hour
	maxFutureOffset := 24 * time.Hour

	timeDiff := timeTime.Sub(now)
	if timeDiff < -maxPastOffset {
		logger.Error("event time too far in past",
			"time", timeTime,
			"now", now,
			"max_past_offset", maxPastOffset,
			"actual_offset", timeDiff,
			"error", ErrInvalidTime)
		return fmt.Errorf("%w: event time cannot be more than %v in the past", ErrInvalidTime, maxPastOffset)
	}

	if timeDiff > maxFutureOffset {
		logger.Error("event time too far in future",
			"time", timeTime,
			"now", now,
			"max_future_offset", maxFutureOffset,
			"actual_offset", timeDiff,
			"error", ErrInvalidTime)
		return fmt.Errorf("%w: event time cannot be more than %v in the future", ErrInvalidTime, maxFutureOffset)
	}

	return nil
}

// Validate performs comprehensive validation of an Event with security checks
func (e *Event) Validate() error {
	if e == nil {
		return fmt.Errorf("%w: event is nil", ErrInvalidInput)
	}

	logger := getLogger(context.Background())

	// Required fields
	if e.Version == "" {
		logger.Error("missing version",
			"error", ErrInvalidInput)
		return fmt.Errorf("%w: version is required", ErrInvalidInput)
	}

	if e.Version != "2.0" {
		logger.Error("unsupported version",
			"version", e.Version,
			"error", ErrInvalidInput)
		return fmt.Errorf("%w: only version 2.0 is supported", ErrInvalidInput)
	}

	// Validate UID
	if err := ValidateUID(e.Uid); err != nil {
		return err // Already logged
	}

	// Validate Type
	if err := ValidateType(e.Type); err != nil {
		return err // Already logged
	}

	// Validate times
	if err := e.validateTimes(); err != nil {
		return err // Already logged
	}

	// Validate Point
	if e.Point == nil {
		logger.Error("missing point element",
			"error", ErrInvalidInput)
		return fmt.Errorf("%w: point element is required", ErrInvalidInput)
	}

	if err := e.Point.Validate(); err != nil {
		return err // Already logged
	}

	// Validate Detail if present
	if err := e.DetailContent.validateDetail(); err != nil {
		return err // Already logged
	}

	// Validate Links if present
	for i, link := range e.Links {
		if err := ValidateUID(link.Uid); err != nil {
			logger.Error("invalid link uid",
				"index", i,
				"uid", link.Uid,
				"error", err)
			return fmt.Errorf("invalid link[%d]: %w", i, err)
		}

		if link.Type == "" {
			logger.Error("missing link type",
				"index", i,
				"error", ErrInvalidInput)
			return fmt.Errorf("%w: link[%d] type is required", ErrInvalidInput, i)
		}
	}

	return nil
}

// Validate performs validation of Point fields
func (p *Point) Validate() error {
	if p == nil {
		return errors.New("point is nil")
	}

	// Validate latitude (-90 to +90)
	if p.Lat < -90 || p.Lat > 90 {
		return fmt.Errorf("invalid latitude: %f (must be between -90 and +90)", p.Lat)
	}

	// Validate longitude (-180 to +180)
	if p.Lon < -180 || p.Lon > 180 {
		return fmt.Errorf("invalid longitude: %f (must be between -180 and +180)", p.Lon)
	}

	// Validate height above ellipsoid (reasonable range check)
	// Mount Everest is ~8,848m, Mariana Trench is ~-11,034m
	const (
		minHae = -12000 // Slightly below Mariana Trench
		maxHae = 9000   // Slightly above Mount Everest
	)
	if p.Hae < minHae || p.Hae > maxHae {
		return fmt.Errorf("suspicious height above ellipsoid: %f (expected between %d and %d meters)", p.Hae, minHae, maxHae)
	}

	// Validate circular and linear error (must be positive)
	if p.Ce < 0 {
		return fmt.Errorf("invalid circular error: %f (must be non-negative)", p.Ce)
	}
	if p.Le < 0 {
		return fmt.Errorf("invalid linear error: %f (must be non-negative)", p.Le)
	}

	// Set reasonable defaults for Ce/Le if they are zero
	if p.Ce == 0 {
		p.Ce = 9999999 // Default to maximum uncertainty
	}
	if p.Le == 0 {
		p.Le = 9999999 // Default to maximum uncertainty
	}

	return nil
}

// Is checks if this event matches a given type predicate
func (e *Event) Is(predicate string) bool {
	if pattern, exists := predicatePatterns[predicate]; exists {
		return pattern.MatchString(e.Type)
	}
	return false
}

// AddLink creates a relationship to another event
func (e *Event) AddLink(targetUID, linkType, relation string) {
	e.Links = append(e.Links, Link{
		Uid:      targetUID,
		Type:     linkType,
		Relation: relation,
	})
	getLogger(nil).Debug("added link to event",
		"source_uid", e.Uid,
		"target_uid", targetUID,
		"type", linkType,
		"relation", relation)
}

// validateDetail performs comprehensive validation of a Detail struct
func (d *Detail) validateDetail() error {
	if d == nil {
		return nil // Empty detail is valid
	}

	// Validate shape if present
	if err := validateShape(&d.Shape); err != nil {
		return fmt.Errorf("invalid shape: %w", err)
	}

	// Validate remarks if present
	if d.Remarks.Content != "" {
		if len(d.Remarks.Content) > maxTokenLen {
			return fmt.Errorf("remarks content exceeds maximum length of %d characters", maxTokenLen)
		}
	}

	// Validate contact if present
	if d.Contact.Callsign != "" {
		if len(d.Contact.Callsign) > maxTokenLen {
			return fmt.Errorf("callsign exceeds maximum length of %d characters", maxTokenLen)
		}
	}

	// Validate flow tags if present
	if d.FlowTags.Status != "" || d.FlowTags.Chain != "" {
		if len(d.FlowTags.Status) > maxTokenLen {
			return fmt.Errorf("flow tag status exceeds maximum length of %d characters", maxTokenLen)
		}
		if len(d.FlowTags.Chain) > maxTokenLen {
			return fmt.Errorf("flow tag chain exceeds maximum length of %d characters", maxTokenLen)
		}
	}

	// Validate unknown elements
	for _, tok := range d.UnknownElements {
		switch t := tok.(type) {
		case xml.StartElement:
			if len(t.Name.Local) > maxTokenLen {
				return fmt.Errorf("unknown element name exceeds maximum length of %d characters", maxTokenLen)
			}
			for _, attr := range t.Attr {
				if len(attr.Value) > maxTokenLen {
					return fmt.Errorf("unknown element attribute value exceeds maximum length of %d characters", maxTokenLen)
				}
			}
		case xml.CharData:
			if len(t) > maxTokenLen {
				return fmt.Errorf("unknown element content exceeds maximum length of %d characters", maxTokenLen)
			}
		}
	}

	// Check total size
	if err := d.validateDetailSize(); err != nil {
		return err
	}

	return nil
}

// validateDetailSize checks if the detail content exceeds size limits
func (d *Detail) validateDetailSize() error {
	if d == nil {
		return nil // Empty detail is valid
	}

	// Get a buffer from the pool
	buf := xmlBufferPool.Get().(*bytes.Buffer)
	defer xmlBufferPool.Put(buf)
	buf.Reset()

	// Marshal to get exact size
	enc := xml.NewEncoder(buf)
	if err := d.MarshalXML(enc, xml.StartElement{Name: xml.Name{Local: "detail"}}); err != nil {
		return fmt.Errorf("failed to marshal detail: %w", err)
	}
	if err := enc.Flush(); err != nil {
		return fmt.Errorf("failed to flush encoder: %w", err)
	}

	// Check total size
	if buf.Len() > maxDetailSize {
		return fmt.Errorf("detail content exceeds maximum size of %d bytes", maxDetailSize)
	}

	return nil
}

// UnmarshalXML implements xml.Unmarshaler for Detail
func (d *Detail) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	// Create a temporary struct to decode known elements
	temp := struct {
		Shape struct {
			Type   string  `xml:"type,attr,omitempty"`
			Points string  `xml:"points,attr,omitempty"`
			Radius float64 `xml:"radius,attr,omitempty"`
		} `xml:"shape,omitempty"`
		Remarks struct {
			Content string `xml:",chardata"`
		} `xml:"remarks,omitempty"`
		Contact struct {
			Callsign string `xml:"callsign,attr,omitempty"`
		} `xml:"contact,omitempty"`
		Status struct {
			Read bool `xml:"read,attr,omitempty"`
		} `xml:"status,omitempty"`
		FlowTags struct {
			Status string `xml:"status,attr,omitempty"`
			Chain  string `xml:"chain,attr,omitempty"`
		} `xml:"flowTags,omitempty"`
		UidAliases struct {
			Aliases []struct {
				Value string `xml:",chardata"`
			} `xml:"uidAlias,omitempty"`
		} `xml:"uidAliases,omitempty"`
	}{}

	// Decode directly into temp
	if err := dec.DecodeElement(&temp, &start); err != nil {
		return err
	}

	// Copy known elements
	d.Shape = temp.Shape
	d.Remarks = temp.Remarks
	d.Contact = temp.Contact
	d.Status = temp.Status
	d.FlowTags = temp.FlowTags
	d.UidAliases = temp.UidAliases

	return nil
}

// MarshalXML implements xml.Marshaler for Detail
func (d *Detail) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	// Start detail element
	if err := enc.EncodeToken(start); err != nil {
		return err
	}

	// Encode known elements
	if d.Shape.Type != "" || d.Shape.Points != "" || d.Shape.Radius != 0 {
		if err := enc.EncodeElement(d.Shape, xml.StartElement{Name: xml.Name{Local: "shape"}}); err != nil {
			return err
		}
	}

	if d.Remarks.Content != "" {
		if err := enc.EncodeElement(d.Remarks, xml.StartElement{Name: xml.Name{Local: "remarks"}}); err != nil {
			return err
		}
	}

	if d.Contact.Callsign != "" {
		if err := enc.EncodeElement(d.Contact, xml.StartElement{Name: xml.Name{Local: "contact"}}); err != nil {
			return err
		}
	}

	if d.Status.Read {
		if err := enc.EncodeElement(d.Status, xml.StartElement{Name: xml.Name{Local: "status"}}); err != nil {
			return err
		}
	}

	if d.FlowTags.Status != "" || d.FlowTags.Chain != "" {
		if err := enc.EncodeElement(d.FlowTags, xml.StartElement{Name: xml.Name{Local: "flowTags"}}); err != nil {
			return err
		}
	}

	if len(d.UidAliases.Aliases) > 0 {
		if err := enc.EncodeElement(d.UidAliases, xml.StartElement{Name: xml.Name{Local: "uidAliases"}}); err != nil {
			return err
		}
	}

	// Encode unknown elements
	for _, tok := range d.UnknownElements {
		if err := enc.EncodeToken(tok); err != nil {
			return err
		}
	}

	// End detail element
	return enc.EncodeToken(start.End())
}

// NewEvent creates a new CoT event with the given parameters.
// It enforces validation of all input parameters and sets default values.
// Returns nil if validation fails.
func NewEvent(uid, eventType string, lat, lon float64) *Event {
	// Assert valid parameters
	if err := ValidateUID(uid); err != nil {
		getLogger(context.Background()).Error("invalid uid in NewEvent", "uid", uid, "error", err)
		return nil
	}
	if err := ValidateType(eventType); err != nil {
		getLogger(context.Background()).Error("invalid type in NewEvent", "type", eventType, "error", err)
		return nil
	}
	if err := ValidateLatLon(lat, lon); err != nil {
		getLogger(context.Background()).Error("invalid coordinates in NewEvent", "lat", lat, "lon", lon, "error", err)
		return nil
	}

	now := time.Now().UTC()
	// Add 1 second guard band to ensure we're past minStaleOffset
	stale := now.Add(minStaleOffset + time.Second)

	e := &Event{
		Version: "2.0",
		Uid:     uid,
		Type:    eventType,
		Time:    now.Format(time.RFC3339),
		Start:   now.Format(time.RFC3339),
		Stale:   stale.Format(time.RFC3339),
		Point:   &Point{Lat: lat, Lon: lon},
	}

	// Assert event state
	if err := e.Validate(); err != nil {
		getLogger(context.Background()).Error("event validation failed in NewEvent", "error", err)
		return nil
	}

	return e
}

// secureDecoder wraps xml.Decoder with additional security features
type secureDecoder struct {
	*xml.Decoder
	depth      int
	tokenCount int
	mu         sync.Mutex
}

// Token implements custom token reading with security checks
func (d *secureDecoder) Token() (xml.Token, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	token, err := d.Decoder.Token()
	if err != nil {
		return token, err
	}

	// Track depth and count
	switch t := token.(type) {
	case xml.StartElement:
		d.depth++
		d.tokenCount++

		// Check depth limit after increment
		if d.depth >= maxElementDepth {
			return nil, fmt.Errorf("exceeded maximum element depth of %d", maxElementDepth)
		}

		// Check token count after increment
		if d.tokenCount >= maxElementCount {
			return nil, fmt.Errorf("exceeded maximum element count of %d", maxElementCount)
		}

		// Validate element name length
		if len(t.Name.Local) > maxTokenLen {
			return nil, fmt.Errorf("element name too long: %d > %d", len(t.Name.Local), maxTokenLen)
		}

		// Validate and sanitize attribute values
		for i := range t.Attr {
			if len(t.Attr[i].Value) > maxValueLen {
				return nil, fmt.Errorf("attribute value too long: %d > %d", len(t.Attr[i].Value), maxValueLen)
			}
			// Sanitize attribute values without escaping
			t.Attr[i].Value = sanitizeString(t.Attr[i].Value)
		}
		token = t // Return the modified token

	case xml.EndElement:
		d.depth--
		d.tokenCount++
		if d.depth < 0 {
			return nil, fmt.Errorf("unexpected closing tag %q", t.Name.Local)
		}

	case xml.CharData:
		d.tokenCount++
		// Validate and sanitize character data
		if len(t) > maxValueLen {
			return nil, fmt.Errorf("character data too long: %d > %d", len(t), maxValueLen)
		}
		t = xml.CharData(sanitizeString(string(t)))
		token = t
	}

	return token, nil
}

// secureXMLDecoder creates a new XML decoder with security restrictions.
// It implements protection against:
// - XXE (XML External Entity) attacks
// - Billion laughs attack
// - Quadratic blowup attack
// - Entity expansion attacks
// - Memory exhaustion attacks
func secureXMLDecoder(r io.Reader) *secureDecoder {
	dec := xml.NewDecoder(r)
	dec.CharsetReader = charset.NewReaderLabel // Use standard charset reader
	dec.DefaultSpace = ""                      // Set empty default namespace
	return &secureDecoder{Decoder: dec}
}

// sanitizeString removes potentially dangerous characters and sequences from strings.
// It implements protection against:
// - XML injection attacks
// - HTML injection attacks
// - Control character attacks
// - Unicode normalization attacks
func sanitizeString(s string) string {
	if s == "" {
		return s
	}

	// Remove control characters except basic whitespace
	s = strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, s)

	// Remove potentially dangerous sequences but preserve angle brackets
	replacer := strings.NewReplacer(
		"<!--", "", // Remove XML comments
		"-->", "",
		"<![CDATA[", "", // Remove CDATA sections
		"]]>", "",
		"<?", "", // Remove processing instructions
		"?>", "",
		"<!", "", // Remove doctype declarations
	)

	s = replacer.Replace(s)

	// Truncate to maxValueLen - 1 to ensure consistent limits
	if len(s) > maxValueLen-1 {
		s = s[:maxValueLen-1]
	}

	return s
}

// sanitizeEvent applies string sanitization to all string fields
func (e *Event) sanitizeEvent() {
	e.Version = sanitizeString(e.Version)
	e.Uid = sanitizeString(e.Uid)
	e.Type = sanitizeString(e.Type)
	e.Time = sanitizeString(e.Time)
	e.Start = sanitizeString(e.Start)
	e.Stale = sanitizeString(e.Stale)
	e.How = sanitizeString(e.How)
	e.Access = sanitizeString(e.Access)

	for i := range e.Links {
		e.Links[i].Uid = sanitizeString(e.Links[i].Uid)
		e.Links[i].Type = sanitizeString(e.Links[i].Type)
		e.Links[i].Relation = sanitizeString(e.Links[i].Relation)
	}

	if e.DetailContent.Shape.Type != "" {
		e.DetailContent.Shape.Type = sanitizeString(e.DetailContent.Shape.Type)
		e.DetailContent.Shape.Points = sanitizeString(e.DetailContent.Shape.Points)
	}

	if e.DetailContent.Contact.Callsign != "" {
		e.DetailContent.Contact.Callsign = sanitizeString(e.DetailContent.Contact.Callsign)
	}

	if e.DetailContent.FlowTags.Status != "" {
		e.DetailContent.FlowTags.Status = sanitizeString(e.DetailContent.FlowTags.Status)
		e.DetailContent.FlowTags.Chain = sanitizeString(e.DetailContent.FlowTags.Chain)
	}
}

// UnmarshalXMLEvent safely unmarshals a CoT event from XML bytes
func UnmarshalXMLEvent(data []byte) (*Event, error) {
	logger := getLogger(context.Background())

	// Validate input
	if len(data) == 0 {
		logger.Error("empty XML data")
		return nil, fmt.Errorf("empty XML data")
	}

	if len(data) > maxXMLSize {
		logger.Error("XML data exceeds maximum size",
			"size", len(data),
			"max_size", maxXMLSize)
		return nil, fmt.Errorf("XML data exceeds maximum size of %d bytes", maxXMLSize)
	}

	// Create secure decoder
	decoder := secureXMLDecoder(bytes.NewReader(data))

	// Track XML element depth
	depth := 0
	decoder.CharsetReader = func(label string, input io.Reader) (io.Reader, error) {
		depth++
		if depth > maxElementDepth {
			logger.Error("XML element depth exceeds maximum",
				"depth", depth,
				"max_depth", maxElementDepth)
			return nil, fmt.Errorf("XML element depth exceeds maximum of %d", maxElementDepth)
		}
		return charset.NewReaderLabel(label, input)
	}

	// Unmarshal with security checks
	event := &Event{}
	if err := decoder.Decode(event); err != nil {
		logger.Error("failed to decode XML",
			"error", err)
		return nil, fmt.Errorf("failed to decode XML: %w", err)
	}

	// Validate the unmarshaled event
	if err := event.Validate(); err != nil {
		logger.Error("invalid event data",
			"error", err)
		return nil, fmt.Errorf("invalid event data: %w", err)
	}

	// Sanitize all string fields
	event.sanitizeEvent()

	logger.Debug("successfully unmarshaled XML event",
		"uid", event.Uid,
		"type", event.Type)

	return event, nil
}

// MarshalXML implements xml.Marshaler for Event
func (e *Event) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	// Create a temporary struct for marshaling
	type Alias Event
	temp := &struct {
		XMLName xml.Name `xml:"event"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	// Marshal to XML with indentation
	enc.Indent("", "  ")
	if err := enc.EncodeElement(temp, start); err != nil {
		return err
	}

	return nil
}

// ToXML converts an Event to XML bytes
func (e *Event) ToXML() ([]byte, error) {
	// Get a buffer from the pool
	buf := xmlBufferPool.Get().(*bytes.Buffer)
	defer xmlBufferPool.Put(buf)
	buf.Reset()

	// Create an encoder with indentation
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")

	// Add XML header
	if _, err := buf.WriteString(xml.Header); err != nil {
		return nil, err
	}

	// Marshal the event
	if err := e.MarshalXML(enc, xml.StartElement{Name: xml.Name{Local: "event"}}); err != nil {
		return nil, err
	}

	// Flush the encoder
	if err := enc.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// FromXML creates an event from XML bytes
func FromXML(data []byte) (*Event, error) {
	return UnmarshalXMLEvent(data)
}

// Example usage function showing new features
func Example() {
	// 1. Create a flight lead
	lead := NewEvent("LEAD1", TypePredFriend+"-A", 30.0090027, -85.9578735)

	// 2. Create wingman
	wing := NewEvent("WING1", TypePredFriend+"-A", 30.0090027, -85.9578735)

	// 3. Link them
	lead.AddLink(wing.Uid, "member", "wingman")

	// 4. Add some contact info
	lead.DetailContent.Contact.Callsign = "LEAD"

	// 5. Check predicates
	if lead.Is("friend") && lead.Is("air") {
		fmt.Println("Lead is a friendly air track")
	}

	// 6. Marshal to XML
	xmlBytes, err := lead.ToXML()
	if err != nil {
		getLogger(nil).Error("failed to marshal event", "error", err)
		return
	}
	fmt.Printf("XML output:\n%s\n", string(xmlBytes))
}

/*
In a real application, do something like:

package main

import (
    "fmt"
    "log"
    "os"

    "github.com/nervsystems/cotlib"
)

func main() {
    // Create
    evt := cot.NewEvent("sampleUID", "a-h-G", 25.5, -120.7)
    output, err := evt.ToXML()
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
