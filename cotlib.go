/*
Package cotlib implements the Cursor on Target (CoT) protocol for Go.

The package provides data structures and utilities for parsing and generating
CoT messages, as well as a comprehensive type catalog system for working with
CoT type codes.

# Type Catalog

The type catalog system provides a way to work with CoT type codes and their
metadata. Each type code (e.g., "a-f-G-E-X-N") has associated metadata:

  - Full Name: A hierarchical name (e.g., "Gnd/Equip/Nbc Equipment")
  - Description: A human-readable description (e.g., "NBC EQUIPMENT")

The catalog supports several operations:

  - Looking up metadata for a specific type code
  - Searching for types by description or full name
  - Validating type codes
  - Registering custom type codes

Example usage:

	// Look up type metadata
	fullName, err := cotlib.GetTypeFullName("a-f-G-E-X-N")
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Printf("Full name: %s\n", fullName)

	// Search for types
	types := cotlib.FindTypesByDescription("NBC")
	for _, t := range types {
	    fmt.Printf("Found type: %s (%s)\n", t.Name, t.Description)
	}

# Thread Safety

All operations on the type catalog are thread-safe. The catalog uses internal
synchronization to ensure safe concurrent access.

# Custom Types

Applications can register custom type codes using RegisterCoTType. These custom
types must follow the standard CoT type format and will be validated before
registration.

For more information about CoT types and their format, see:
https://www.mitre.org/sites/default/files/pdf/09_4937.pdf

Security features include:
  - XML parsing restrictions to prevent XXE attacks
  - Input validation on all fields
  - Strict coordinate range enforcement
  - Time field validation to prevent time-based attacks
  - Secure logging practices
  - Detail extension isolation

For more information about CoT, see:
  - https://apps.dtic.mil/sti/citations/ADA637348 (Developer Guide)
  - https://www.mitre.org/sites/default/files/pdf/09_4937.pdf (Message Router Guide)
  - http://cot.mitre.org

The package follows these design principles:
  - High cohesion: focused on CoT event parsing and serialization
  - Low coupling: separated concerns for expansions and transport
  - Composition over inheritance: nested sub-structures for detail fields
  - Full schema coverage: implements Event.xsd with example extensions
  - Secure by design: validates inputs and prevents common attacks
*/
package cotlib

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/NERVsystems/cotlib/cottypes"
)

// Security limits for XML parsing and validation
const (
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

	// CotTimeFormat is the standard time format for CoT messages (Zulu time, no offset)
	// Format: "2006-01-02T15:04:05Z" (UTC without timezone offset)
	CotTimeFormat = "2006-01-02T15:04:05Z"
)

// maxValueLen is the maximum length for attribute values and character data
// Set to 512 KiB to accommodate large KML polygons
var maxValueLen atomic.Int64

// currentMaxValueLen returns the current maximum value length
func currentMaxValueLen() int64 {
	return maxValueLen.Load()
}

// SetMaxValueLen sets the maximum allowed length for XML attribute values and character data
// This is used to prevent memory exhaustion attacks via large XML payloads
func SetMaxValueLen(max int64) {
	if max < 0 {
		max = 0
	}
	maxValueLen.Store(max)
}

// checkXMLLimits performs a lightweight scan of the XML data to ensure it does
// not exceed configured security limits.
func checkXMLLimits(data []byte) error {
	if len(data) > maxXMLSize {
		return ErrInvalidInput
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.CharsetReader = nil
	dec.Entity = nil

	depth := 0
	count := 0

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return ErrInvalidInput
		}
		count++
		if count > maxElementCount {
			return ErrInvalidInput
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if depth > maxElementDepth {
				return ErrInvalidInput
			}
			for _, a := range t.Attr {
				if len(a.Value) > int(currentMaxValueLen()) {
					return ErrInvalidInput
				}
			}
		case xml.EndElement:
			if depth > 0 {
				depth--
			}
		case xml.CharData:
			if len(t) > int(currentMaxValueLen()) {
				return ErrInvalidInput
			}
		}
	}

	return nil
}

// attrEscaper escapes XML special characters in attribute values.
var attrEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"\"", "&quot;",
	"'", "&apos;",
)

// escapeAttr returns the escaped version of s for use in XML attributes.
func escapeAttr(s string) string {
	if s == "" {
		return s
	}
	return attrEscaper.Replace(s)
}

// RegisterCoTType adds a specific CoT type to the valid types registry
// It does not log individual type registrations to avoid log spam
func RegisterCoTType(name string) {
	if !basicSyntaxOK(name) {
		return
	}
	cat := cottypes.GetCatalog()
	cat.Upsert(name, cottypes.Type{Name: name})
}

// basicSyntaxOK performs basic syntax validation on a CoT type
func basicSyntaxOK(name string) bool {
	// Only register if it's a valid format (at least one hyphen and not incomplete)
	if strings.Count(name, "-") < 1 || strings.HasSuffix(name, "-") {
		return false
	}

	// Additional validation for incomplete types
	parts := strings.Split(name, "-")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return false
	}

	// Validate type format based on prefix
	switch parts[0] {
	case "a":
		// Atomic types must have at least 3 parts: affiliation, domain, and category
		if len(parts) < 3 {
			return false
		}
	case "b", "t", "y", "c":
		// Other types must have at least 2 parts: prefix and subtype
		if len(parts) < 2 {
			return false
		}
	default:
		// Unknown prefix
		return false
	}

	return true
}

// RegisterCoTTypesFromFile loads and registers CoT types from an XML file
func RegisterCoTTypesFromFile(filename string) error {
	logger := slog.Default()

	file, err := os.Open(filename)
	if err != nil {
		logger.Error("failed to open file",
			"path", filename,
			"error", err)
		return err
	}
	defer file.Close()

	// Simple structure to parse the XML
	var types struct {
		XMLName xml.Name `xml:"types"`
		CoTs    []struct {
			Type string `xml:"cot,attr"`
		} `xml:"cot"`
	}

	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&types); err != nil {
		logger.Error("failed to decode XML",
			"path", filename,
			"error", err)
		return err
	}

	// Register all types without individual logging
	typeCount := 0
	for _, t := range types.CoTs {
		if t.Type != "" {
			RegisterCoTType(t.Type)
			typeCount++
		}
	}

	// Log a summary at DEBUG level only
	logger.Debug("registered CoT types from file",
		"path", filename,
		"type_count", typeCount)

	return nil
}

// RegisterCoTTypesFromReader loads and registers CoT types from an XML reader
func RegisterCoTTypesFromReader(r io.Reader) error {
	logger := slog.Default()
	decoder := xml.NewDecoder(r)

	// Simple structure to parse the XML
	var types struct {
		XMLName xml.Name `xml:"types"`
		CoTs    []struct {
			Type string `xml:"cot,attr"`
		} `xml:"cot"`
	}

	if err := decoder.Decode(&types); err != nil {
		logger.Error("failed to decode XML from reader",
			"error", err)
		return err
	}

	// Register all types without individual logging
	typeCount := 0
	for _, t := range types.CoTs {
		if t.Type != "" {
			RegisterCoTType(t.Type)
			typeCount++
		}
	}

	// Log a summary at DEBUG level only
	logger.Debug("registered CoT types from reader",
		"type_count", typeCount)

	return nil
}

// RegisterCoTTypesFromXMLContent registers CoT types from the given XML content string
// This is particularly useful for embedding the CoTtypes.xml content directly in code
func RegisterCoTTypesFromXMLContent(xmlContent string) error {
	logger := slog.Default()

	// Use a Reader for the XML content
	reader := strings.NewReader(xmlContent)

	// Use the standard decoder
	decoder := xml.NewDecoder(reader)

	// Simple structure to parse the XML
	var types struct {
		XMLName xml.Name `xml:"types"`
		CoTs    []struct {
			Type string `xml:"cot,attr"`
		} `xml:"cot"`
	}

	if err := decoder.Decode(&types); err != nil {
		logger.Error("failed to decode XML content",
			"error", err)
		return err
	}

	// Register all types without individual logging
	typeCount := 0
	for _, t := range types.CoTs {
		if t.Type != "" {
			RegisterCoTType(t.Type)
			typeCount++
		}
	}

	// Log a summary at DEBUG level only
	logger.Debug("registered CoT types from XML content",
		"type_count", typeCount)

	return nil
}

// RegisterAllCoTTypes is a no-op since XML is already embedded
func RegisterAllCoTTypes() error {
	return nil
}

// LoadCoTTypesFromFile loads CoT types from a file
func LoadCoTTypesFromFile(path string) error {
	logger := slog.Default()

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Error("failed to read file",
			"path", path,
			"error", err)
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the XML
	var types struct {
		Types []string `xml:"type"`
	}
	if err := xml.Unmarshal(data, &types); err != nil {
		logger.Error("failed to parse XML",
			"path", path,
			"error", err)
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	// Register each type
	for _, typ := range types.Types {
		RegisterCoTType(typ)
	}

	logger.Debug("loaded CoT types from file",
		"path", path,
		"types", len(types.Types))

	return nil
}

// LookupType returns the Type for the given name if it exists
func LookupType(name string) (cottypes.Type, bool) {
	t, err := cottypes.GetCatalog().GetType(name)
	return t, err == nil
}

// FindTypes returns all types matching the given query
func FindTypes(query string) []cottypes.Type {
	return cottypes.GetCatalog().Find(query)
}

// isRegisteredType is an internal helper that checks if a type is registered
func isRegisteredType(typ string) bool {
	_, ok := LookupType(typ)
	return ok
}

// CoTTime represents a time in CoT format (UTC without timezone offset)
type CoTTime time.Time

// Time returns the underlying time.Time value
func (t CoTTime) Time() time.Time {
	return time.Time(t)
}

// MarshalXML implements xml.Marshaler
func (t CoTTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(t.Time().UTC().Format(CotTimeFormat), start)
}

// UnmarshalXML implements xml.Unmarshaler
func (t *CoTTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	parsed, err := time.Parse(CotTimeFormat, s)
	if err != nil {
		return fmt.Errorf("invalid time format: %w", err)
	}
	*t = CoTTime(parsed)
	return nil
}

// MarshalXMLAttr implements xml.MarshalerAttr
func (t CoTTime) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{
		Name:  name,
		Value: t.Time().UTC().Format(CotTimeFormat),
	}, nil
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr
func (t *CoTTime) UnmarshalXMLAttr(attr xml.Attr) error {
	parsed, err := time.Parse(CotTimeFormat, attr.Value)
	if err != nil {
		return fmt.Errorf("invalid time format: %w", err)
	}
	*t = CoTTime(parsed)
	return nil
}

// Point represents a location in 3D space with error estimates
type Point struct {
	Lat float64 `xml:"lat,attr"` // Latitude in degrees
	Lon float64 `xml:"lon,attr"` // Longitude in degrees
	Hae float64 `xml:"hae,attr"` // Height above ellipsoid in meters
	Ce  float64 `xml:"ce,attr"`  // Circular error in meters
	Le  float64 `xml:"le,attr"`  // Linear error in meters
}

// Validate checks if the point coordinates and errors are valid
func (p *Point) Validate() error {
	// Validate latitude and longitude
	if err := ValidateLatLon(p.Lat, p.Lon); err != nil {
		return err
	}

	// Validate height above ellipsoid (HAE)
	// HAE must be between -10000 (Dead Sea) and 100000 (edge of space)
	if p.Hae < -10000 || p.Hae > 100000 {
		return fmt.Errorf("invalid HAE: %f", p.Hae)
	}

	// Validate circular error (CE)
	// CE must be strictly positive and less than or equal to 9999999 meters
	if p.Ce <= 0 || p.Ce > 9999999.0 {
		return fmt.Errorf("invalid CE: %f", p.Ce)
	}

	// Validate linear error (LE)
	// LE must be strictly positive and less than or equal to 9999999 meters
	if p.Le <= 0 || p.Le > 9999999.0 {
		return fmt.Errorf("invalid LE: %f", p.Le)
	}

	return nil
}

// Event represents a CoT event message
type Event struct {
	XMLName xml.Name `xml:"event"`
	Version string   `xml:"version,attr"`
	Uid     string   `xml:"uid,attr"`
	Type    string   `xml:"type,attr"`
	How     string   `xml:"how,attr,omitempty"`
	Time    CoTTime  `xml:"time,attr"`
	Start   CoTTime  `xml:"start,attr"`
	Stale   CoTTime  `xml:"stale,attr"`
	Point   Point    `xml:"point"`
	Detail  *Detail  `xml:"detail,omitempty"`
	Links   []Link   `xml:"link,omitempty"`
}

// Error sentinels for validation
var (
	ErrInvalidInput    = fmt.Errorf("invalid input")
	ErrInvalidLatitude = fmt.Errorf("invalid latitude")
	ErrInvalidUID      = fmt.Errorf("invalid UID")
)

// doctypePattern matches XML DOCTYPE declarations case-insensitively
var doctypePattern = regexp.MustCompile(`(?i)<!\s*DOCTYPE`)

// Contact represents contact information
type Contact struct {
	Callsign string `xml:"callsign,attr,omitempty"`
}

// Detail contains additional information about an event
type Detail struct {
	Group   *Group   `xml:"group,omitempty"`
	Contact *Contact `xml:"contact,omitempty"`
}

// Group represents a group affiliation
type Group struct {
	Name string `xml:"name,attr"`
	Role string `xml:"role,attr"`
}

// Link represents a relationship to another event
type Link struct {
	Uid      string `xml:"uid,attr"`
	Type     string `xml:"type,attr"`
	Relation string `xml:"relation,attr"`
}

// NewEvent creates a new CoT event with the given parameters
func NewEvent(uid, typ string, lat, lon, hae float64) (*Event, error) {
	now := time.Now().UTC().Truncate(time.Second)
	evt := &Event{
		Version: "2.0",
		Uid:     uid,
		Type:    typ,
		How:     "m-g",
		Time:    CoTTime(now),
		Start:   CoTTime(now),
		Stale:   CoTTime(now.Add(6 * time.Second)),
		Point: Point{
			Lat: lat,
			Lon: lon,
			Hae: hae,
			Ce:  9999999.0,
			Le:  9999999.0,
		},
	}
	if err := evt.Validate(); err != nil {
		return nil, err
	}
	return evt, nil
}

// NewPresenceEvent creates a new presence event (t-x-takp-v)
func NewPresenceEvent(uid string, lat, lon, hae float64) (*Event, error) {
	return NewEvent(uid, "t-x-takp-v", lat, lon, hae)
}

// ValidateType checks if a CoT type is valid
func ValidateType(typ string) error {
	if typ == "" {
		return fmt.Errorf("empty type")
	}
	if len(typ) > 100 {
		return fmt.Errorf("type too long")
	}

	// Fast path for wildcard patterns that don't need catalog lookup
	if strings.Contains(typ, "*") {
		parts := strings.Split(typ, "-")
		if len(parts) < 2 {
			return fmt.Errorf("invalid type format")
		}
		// Only allow wildcards at the end of the type string
		for i := 0; i < len(parts)-1; i++ {
			if parts[i] == "*" {
				return fmt.Errorf("wildcard only allowed at end of type")
			}
		}
		// Validate the prefix
		if parts[0] != "a" && parts[0] != "b" && parts[0] != "t" {
			return fmt.Errorf("invalid type prefix")
		}
		return nil
	}

	// Fast path for atomic type wildcards (a-.-X)
	if strings.Contains(typ, ".-") {
		parts := strings.Split(typ, "-")
		if len(parts) < 2 {
			return fmt.Errorf("invalid type format")
		}
		if parts[0] != "a" {
			return fmt.Errorf("wildcard only allowed in atomic types")
		}
		if parts[1] != "." {
			return fmt.Errorf("invalid wildcard format")
		}
		return nil
	}

	// Use the catalog for validation of non-wildcard types
	_, err := cottypes.GetCatalog().GetType(typ)
	if err != nil {
		return fmt.Errorf("invalid type: %w", err)
	}

	return nil
}

// Validate checks if the event is valid
func (e *Event) Validate() error {
	// Check required fields
	if e.Version == "" {
		return fmt.Errorf("missing version")
	}
	if e.Uid == "" {
		return fmt.Errorf("missing uid")
	}
	if e.Type == "" {
		return fmt.Errorf("missing type")
	}

	// Validate type
	if err := ValidateType(e.Type); err != nil {
		return err
	}

	// Validate time fields
	now := time.Now().UTC()
	eventTime := e.Time.Time()
	startTime := e.Start.Time()
	staleTime := e.Stale.Time()

	// Check time ranges
	if eventTime.Before(now.Add(-24 * time.Hour)) {
		return fmt.Errorf("time must be within 24 hours of current time")
	}
	if eventTime.After(now.Add(24 * time.Hour)) {
		return fmt.Errorf("time must be within 24 hours of current time")
	}

	// Check start time
	if startTime.After(eventTime) {
		return fmt.Errorf("start time after event time")
	}

	// Check stale time
	staleDiff := staleTime.Sub(eventTime)
	if staleDiff < minStaleOffset {
		return fmt.Errorf("stale time too close to event time")
	}
	// Allow longer stale times for TAK types
	maxStale := maxStaleOffset
	if strings.HasPrefix(e.Type, "t-x-") {
		maxStale = 30 * 24 * time.Hour // 30 days for TAK types
	}
	if staleDiff > maxStale {
		return fmt.Errorf("stale time too far from event time")
	}

	// Validate point
	if err := e.Point.Validate(); err != nil {
		return err
	}

	return nil
}

// AddLink adds a link to the event
func (e *Event) AddLink(link *Link) {
	e.Links = append(e.Links, *link)
}

// InjectIdentity adds identity information to the event
func (e *Event) InjectIdentity(selfUid, groupName, groupRole string) {
	// Add group information
	if e.Detail == nil {
		e.Detail = &Detail{}
	}
	if e.Detail.Group == nil {
		e.Detail.Group = &Group{}
	}
	e.Detail.Group.Name = groupName
	e.Detail.Group.Role = groupRole

	// Add self link if not already present
	for _, l := range e.Links {
		if l.Relation == "p-p" && l.Uid == selfUid {
			return
		}
	}
	e.AddLink(&Link{
		Uid:      selfUid,
		Type:     "t-x-takp-v",
		Relation: "p-p",
	})
}

// Is checks if the event matches a predicate
func (e *Event) Is(pred string) bool {
	parts := strings.Split(e.Type, "-")
	if len(parts) < 2 {
		return false
	}

	switch pred {
	case "atom":
		// All types with valid prefixes are considered atoms
		return parts[0] == "a" || parts[0] == "b" || parts[0] == "t"
	case "friend":
		return len(parts) > 1 && parts[1] == "f"
	case "hostile":
		return len(parts) > 1 && parts[1] == "h"
	case "ground":
		return len(parts) > 2 && parts[2] == "G"
	case "air":
		return len(parts) > 2 && parts[2] == "A"
	default:
		return false
	}
}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

type loggerKey struct{}

// GetTypeFullName returns the full hierarchical name for a CoT type.
// For example, "a-f-G-E-X-N" returns "Gnd/Equip/Nbc Equipment".
//
// The full name represents the type's position in the CoT type hierarchy,
// making it useful for building user interfaces and documentation.
//
// Returns an error if the type is not registered in the catalog.
func GetTypeFullName(name string) (string, error) {
	return cottypes.GetCatalog().GetFullName(name)
}

// GetTypeDescription returns the human-readable description for a CoT type.
// For example, "a-f-G-E-X-N" returns "NBC EQUIPMENT".
//
// The description is a concise explanation of what the type represents,
// suitable for display in user interfaces and logs.
//
// Returns an error if the type is not registered in the catalog.
func GetTypeDescription(name string) (string, error) {
	return cottypes.GetCatalog().GetDescription(name)
}

// FindTypesByDescription searches for types matching the given description.
// The search is case-insensitive and matches partial descriptions.
//
// For example:
//   - "NBC" finds all types containing "NBC" in their description
//   - "EQUIPMENT" finds all equipment-related types
//   - "COMBAT" finds all combat-related types
//
// This is useful for building search interfaces and type discovery tools.
// Returns an empty slice if no matches are found.
func FindTypesByDescription(desc string) []cottypes.Type {
	return cottypes.GetCatalog().FindByDescription(desc)
}

// FindTypesByFullName searches for types matching the given full name.
// The search is case-insensitive and matches partial names.
//
// For example:
//   - "Nbc Equipment" finds all NBC equipment types
//   - "Ground" finds all ground-based types
//   - "Vehicle" finds all vehicle types
//
// This is useful for finding types based on their hierarchical classification.
// Returns an empty slice if no matches are found.
func FindTypesByFullName(name string) []cottypes.Type {
	return cottypes.GetCatalog().FindByFullName(name)
}

// UnmarshalXMLEvent parses an XML byte slice into an Event
func UnmarshalXMLEvent(data []byte) (*Event, error) {
	// Check for DOCTYPE in a case-insensitive manner
	if doctypePattern.Match(data) {
		return nil, ErrInvalidInput
	}

	// Check namespace length
	if idx := bytes.Index(data, []byte(`xmlns="`)); idx >= 0 {
		end := bytes.Index(data[idx+7:], []byte(`"`))
		if end > 1024 {
			return nil, ErrInvalidInput
		}
	}

	// Enforce parser limits before decoding
	if err := checkXMLLimits(data); err != nil {
		return nil, err
	}

	// Create a secure decoder with limits
	decoder := xml.NewDecoder(io.LimitReader(bytes.NewReader(data), int64(len(data))))
	decoder.CharsetReader = nil // Disable charset conversion
	decoder.Entity = nil        // Disable entity expansion

	var evt Event
	if err := decoder.Decode(&evt); err != nil {
		return nil, fmt.Errorf("failed to decode XML: %w", err)
	}

	if err := evt.Validate(); err != nil {
		return nil, err
	}

	return &evt, nil
}

// ValidateLatLon checks if latitude and longitude are within valid ranges
func ValidateLatLon(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return ErrInvalidLatitude
	}
	if lon < -180 || lon > 180 {
		return fmt.Errorf("invalid longitude")
	}
	return nil
}

// ValidateUID checks if a UID is valid
func ValidateUID(uid string) error {
	if uid == "" {
		return ErrInvalidUID
	}
	if strings.HasPrefix(uid, "-") {
		return ErrInvalidUID
	}
	if strings.Contains(uid, "..") {
		return ErrInvalidUID
	}
	return nil
}

// ToXML serialises an Event to CoT-compliant XML.
// Attribute values are escaped to prevent XML-injection.
// The <point> element is always emitted so that the
// zero coordinate (0° N 0° E) is representable.
func (e *Event) ToXML() ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(256)
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	var tmp [32]byte

	// <event>
	buf.WriteString("<event")
	if e.Version != "" {
		fmt.Fprintf(&buf, ` version="%s"`, escapeAttr(e.Version))
	}
	if e.Type != "" {
		fmt.Fprintf(&buf, ` type="%s"`, escapeAttr(e.Type))
	}
	if e.How != "" {
		fmt.Fprintf(&buf, ` how="%s"`, escapeAttr(e.How))
	}
	if e.Uid != "" {
		fmt.Fprintf(&buf, ` uid="%s"`, escapeAttr(e.Uid))
	}
	if !e.Time.Time().IsZero() {
		fmt.Fprintf(&buf, ` time="%s"`, e.Time.Time().UTC().Format(CotTimeFormat))
	}
	if !e.Start.Time().IsZero() {
		fmt.Fprintf(&buf, ` start="%s"`, e.Start.Time().UTC().Format(CotTimeFormat))
	}
	if !e.Stale.Time().IsZero() {
		fmt.Fprintf(&buf, ` stale="%s"`, e.Stale.Time().UTC().Format(CotTimeFormat))
	}
	buf.WriteString(">\n")

	// <point>
	buf.WriteString("  <point")
	buf.WriteString(` lat="`)
	buf.Write(strconv.AppendFloat(tmp[:0], e.Point.Lat, 'f', 6, 64))
	buf.WriteString(`" lon="`)
	buf.Write(strconv.AppendFloat(tmp[:0], e.Point.Lon, 'f', 6, 64))
	buf.WriteByte('"')
	if e.Point.Hae != 0 {
		buf.WriteString(` hae="`)
		buf.Write(strconv.AppendFloat(tmp[:0], e.Point.Hae, 'f', 1, 64))
		buf.WriteByte('"')
	}
	if e.Point.Ce != 0 {
		buf.WriteString(` ce="`)
		buf.Write(strconv.AppendFloat(tmp[:0], e.Point.Ce, 'f', 1, 64))
		buf.WriteByte('"')
	}
	if e.Point.Le != 0 {
		buf.WriteString(` le="`)
		buf.Write(strconv.AppendFloat(tmp[:0], e.Point.Le, 'f', 1, 64))
		buf.WriteByte('"')
	}
	buf.WriteString("/>\n")

	// <detail> (optional)
	if e.Detail != nil {
		buf.WriteString("  <detail>\n")
		if c := e.Detail.Contact; c != nil {
			buf.WriteString("    <contact")
			if c.Callsign != "" {
				fmt.Fprintf(&buf, ` callsign="%s"`, escapeAttr(c.Callsign))
			}
			buf.WriteString("/>\n")
		}
		if g := e.Detail.Group; g != nil {
			buf.WriteString("    <group")
			if g.Name != "" {
				fmt.Fprintf(&buf, ` name="%s"`, escapeAttr(g.Name))
			}
			if g.Role != "" {
				fmt.Fprintf(&buf, ` role="%s"`, escapeAttr(g.Role))
			}
			buf.WriteString("/>\n")
		}
		buf.WriteString("  </detail>\n")
	}

	// <link> (0..n)
	for _, l := range e.Links {
		buf.WriteString("  <link")
		if l.Uid != "" {
			fmt.Fprintf(&buf, ` uid="%s"`, escapeAttr(l.Uid))
		}
		if l.Type != "" {
			fmt.Fprintf(&buf, ` type="%s"`, escapeAttr(l.Type))
		}
		if l.Relation != "" {
			fmt.Fprintf(&buf, ` relation="%s"`, escapeAttr(l.Relation))
		}
		buf.WriteString("/>\n")
	}

	buf.WriteString("</event>")
	return buf.Bytes(), nil
}
