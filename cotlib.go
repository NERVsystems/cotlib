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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/NERVsystems/cotlib/ctxlog"

	"github.com/NERVsystems/cotlib/cottypes"
	"github.com/NERVsystems/cotlib/validator"
)

// Security limits for XML parsing and validation
const (
	// minStaleOffset is the minimum time between event time and stale time
	// Increased to 5 seconds to prevent replay attacks on slow links
	minStaleOffset = 5 * time.Second

	// maxStaleOffset is the maximum time between event time and stale time
	// Events cannot be valid for more than 7 days to prevent stale data
	maxStaleOffset = 7 * 24 * time.Hour

	// CotTimeFormat is the standard time format for CoT messages (Zulu time, no offset)
	// Format: "2006-01-02T15:04:05Z" (UTC without timezone offset)
	CotTimeFormat = "2006-01-02T15:04:05Z"
)

// Default security limits. These can be adjusted with the setter functions
var (
	maxXMLSize      atomic.Int64
	maxElementDepth atomic.Int64
	maxElementCount atomic.Int64
	maxTokenLen     atomic.Int64

	// maxValueLen is the maximum length for attribute values and character data
	// Set to 512 KiB to accommodate large KML polygons
	maxValueLen atomic.Int64
)

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

// currentMaxXMLSize returns the configured maximum XML size
func currentMaxXMLSize() int64 {
	return maxXMLSize.Load()
}

// SetMaxXMLSize sets the maximum allowed size for XML input
func SetMaxXMLSize(max int64) {
	if max < 0 {
		max = 0
	}
	maxXMLSize.Store(max)
}

// currentMaxElementDepth returns the maximum XML element depth
func currentMaxElementDepth() int64 {
	return maxElementDepth.Load()
}

// SetMaxElementDepth sets the maximum depth of XML elements
func SetMaxElementDepth(max int64) {
	if max < 0 {
		max = 0
	}
	maxElementDepth.Store(max)
}

// currentMaxElementCount returns the maximum allowed number of XML elements
func currentMaxElementCount() int64 {
	return maxElementCount.Load()
}

// SetMaxElementCount sets the maximum allowed number of XML elements
func SetMaxElementCount(max int64) {
	if max < 0 {
		max = 0
	}
	maxElementCount.Store(max)
}

// currentMaxTokenLen returns the maximum allowed token length
func currentMaxTokenLen() int64 {
	return maxTokenLen.Load()
}

// SetMaxTokenLen sets the maximum length for any single XML token
func SetMaxTokenLen(max int64) {
	if max < 0 {
		max = 0
	}
	maxTokenLen.Store(max)
}

// attrEscaper escapes XML special characters in attribute values.
// It also encodes carriage returns, newlines, and tabs using
// their numeric character references.
var attrEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"\"", "&quot;",
	"'", "&apos;",
	"\r", "&#xD;",
	"\n", "&#xA;",
	"\t", "&#x9;",
)

// textEscaper escapes XML special characters in element text.
var textEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"\r", "&#xD;",
)

// escapeAttr returns the escaped version of s for use in XML attributes.
func escapeAttr(s string) string {
	if s == "" {
		return s
	}
	return attrEscaper.Replace(s)
}

// escapeText writes the escaped version of s to the buffer.
func escapeText(buf *bytes.Buffer, s string) {
	if s == "" {
		return
	}
	buf.WriteString(textEscaper.Replace(s))
}

// RegisterCoTType adds a specific CoT type to the valid types registry
// It does not log individual type registrations to avoid log spam
func RegisterCoTType(name string) {
	if !basicSyntaxOK(name) {
		return
	}
	cat := cottypes.GetCatalog()
	if err := cat.Upsert(context.Background(), name, cottypes.Type{Name: name}); err != nil {
		slog.Error("failed to register CoT type",
			"name", name,
			"error", err)
	}
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
func RegisterCoTTypesFromFile(ctx context.Context, filename string) error {
	logger := LoggerFromContext(ctx)

	clean := filepath.Clean(filename)
	if strings.Contains(clean, "..") {
		logger.Error("invalid path", "path", filename)
		return ErrInvalidInput
	}

	data, err := os.ReadFile(clean)
	if err != nil {
		logger.Error("failed to read file",
			"path", filename,
			"error", err)
		return err
	}

	if doctypePattern.Match(data) {
		logger.Error("invalid doctype detected",
			"path", filename)
		return ErrInvalidInput
	}

	var types struct {
		XMLName xml.Name `xml:"types"`
		CoTs    []struct {
			Type string `xml:"cot,attr"`
		} `xml:"cot"`
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.CharsetReader = nil
	dec.Entity = nil
	if err := decodeWithLimits(dec, &types); err != nil {
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
func RegisterCoTTypesFromReader(ctx context.Context, r io.Reader) error {
	logger := LoggerFromContext(ctx)

	data, err := io.ReadAll(r)
	if err != nil {
		logger.Error("failed to read from reader", "error", err)
		return err
	}

	if doctypePattern.Match(data) {
		logger.Error("invalid doctype detected")
		return ErrInvalidInput
	}

	var types struct {
		XMLName xml.Name `xml:"types"`
		CoTs    []struct {
			Type string `xml:"cot,attr"`
		} `xml:"cot"`
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.CharsetReader = nil
	dec.Entity = nil
	if err := decodeWithLimits(dec, &types); err != nil {
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
func RegisterCoTTypesFromXMLContent(ctx context.Context, xmlContent string) error {
	logger := LoggerFromContext(ctx)

	data := []byte(xmlContent)

	if doctypePattern.Match(data) {
		logger.Error("invalid doctype detected")
		return ErrInvalidInput
	}

	var types struct {
		XMLName xml.Name `xml:"types"`
		CoTs    []struct {
			Type string `xml:"cot,attr"`
		} `xml:"cot"`
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.CharsetReader = nil
	dec.Entity = nil
	if err := decodeWithLimits(dec, &types); err != nil {
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
func LoadCoTTypesFromFile(ctx context.Context, path string) error {
	logger := LoggerFromContext(ctx)

	clean := filepath.Clean(path)
	if strings.Contains(clean, "..") {
		logger.Error("invalid path", "path", path)
		return ErrInvalidInput
	}

	data, err := os.ReadFile(clean)
	if err != nil {
		logger.Error("failed to read file",
			"path", path,
			"error", err)
		return fmt.Errorf("failed to read file: %w", err)
	}

	if doctypePattern.Match(data) {
		logger.Error("invalid doctype detected",
			"path", path)
		return ErrInvalidInput
	}

	var types struct {
		Types []string `xml:"type"`
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.CharsetReader = nil
	dec.Entity = nil
	if err := decodeWithLimits(dec, &types); err != nil {
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
// LookupType returns the Type for the given name if it exists.
// If the exact type is not found, the function attempts wildcard resolution
// by substituting affiliation segments (f/h/n/u) with '.' and retrying the
// lookup. This mirrors ValidateType's wildcard handling to ensure lookups
// succeed for types that only exist in their wildcard form in the catalog.
func LookupType(name string) (cottypes.Type, bool) {
	cat := cottypes.GetCatalog()
	if cat == nil {
		return cottypes.Type{}, false
	}

	t, err := cat.GetType(context.Background(), name)
	if err == nil {
		return t, true
	}

	parts := strings.Split(name, "-")
	for i, seg := range parts {
		switch seg {
		case "f", "h", "n", "u":
			orig := parts[i]
			parts[i] = "."
			if t2, err2 := cat.GetType(context.Background(), strings.Join(parts, "-")); err2 == nil {
				return t2, true
			}
			parts[i] = orig
		}
	}

	return cottypes.Type{}, false
}

// FindTypes returns all types matching the given query
func FindTypes(query string) []cottypes.Type {
	return cottypes.GetCatalog().Find(context.Background(), query)
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
	// Lower bound: Mariana Trench depth (-12000m) with safety margin
	// Upper bound: Geostationary orbit (40000km) for space assets
	// Special case: 9999999.0 is allowed as traditional TAK "unknown altitude"
	if p.Hae < -12000 || (p.Hae > 40000000 && p.Hae != 9999999.0) {
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
	// Message is populated for GeoChat events from the <remarks> element.
	Message string `xml:"-"`
	// StrokeColor is an ARGB hex color used for drawing events.
	StrokeColor string `xml:"strokeColor,attr,omitempty"`
	// UserIcon specifies a custom icon URL or resource for the event.
	UserIcon string `xml:"usericon,attr,omitempty"`
}

// Error sentinels for validation
var (
	ErrInvalidInput     = fmt.Errorf("invalid input")
	ErrInvalidLatitude  = fmt.Errorf("invalid latitude")
	ErrInvalidLongitude = fmt.Errorf("invalid longitude")
	ErrInvalidUID       = fmt.Errorf("invalid UID")
	// ErrInvalidType is returned when a CoT type fails validation.
	ErrInvalidType = fmt.Errorf("invalid type")
	// ErrInvalidHow is returned when a how value is not recognised.
	ErrInvalidHow = fmt.Errorf("invalid how")
	// ErrInvalidRelation is returned when a relation value is not recognised.
	ErrInvalidRelation = fmt.Errorf("invalid relation")
)

// doctypePattern matches XML DOCTYPE declarations case-insensitively
var doctypePattern = regexp.MustCompile(`(?i)<!\s*DOCTYPE`)

// Contact represents contact information
type Contact struct {
	XMLName  xml.Name `xml:"contact"`
	Callsign string   `xml:"callsign,attr,omitempty"`
}

// Detail contains additional information about an event
type Detail struct {
	Group             *Group             `xml:"group,omitempty"`
	Contact           *Contact           `xml:"contact,omitempty"`
	Chat              *Chat              `xml:"__chat,omitempty"`
	ChatReceipt       *ChatReceipt       `xml:"__chatReceipt,omitempty"`
	Emergency         *Emergency         `xml:"emergency,omitempty"`
	Geofence          *Geofence          `xml:"__geofence,omitempty"`
	ServerDestination *ServerDestination `xml:"__serverdestination,omitempty"`
	Video             *Video             `xml:"__video,omitempty"`
	GroupExtension    *GroupExtension    `xml:"__group,omitempty"`
	Archive           *Archive           `xml:"archive,omitempty"`
	AttachmentList    *AttachmentList    `xml:"attachmentList,omitempty"`
	Environment       *Environment       `xml:"environment,omitempty"`
	FileShare         *FileShare         `xml:"fileshare,omitempty"`
	PrecisionLocation *PrecisionLocation `xml:"precisionlocation,omitempty"`
	Takv              *Takv              `xml:"takv,omitempty"`
	Track             *Track             `xml:"track,omitempty"`
	Mission           *Mission           `xml:"mission,omitempty"`
	Status            *Status            `xml:"status,omitempty"`
	Shape             *Shape             `xml:"shape,omitempty"`
	StrokeColor       *StrokeColor       `xml:"strokecolor,omitempty"`
	StrokeWeight      *StrokeWeight      `xml:"strokeweight,omitempty"`
	FillColor         *FillColor         `xml:"fillcolor,omitempty"`
	Height            *Height            `xml:"height,omitempty"`
	HeightUnit        *HeightUnit        `xml:"height_unit,omitempty"`
	LabelsOn          *LabelsOn          `xml:"labelson,omitempty"`
	ColorExtension    *ColorExtension    `xml:"color,omitempty"`
	Hierarchy         *Hierarchy         `xml:"hierarchy,omitempty"`
	LinkDetail        *DetailLink        `xml:"link,omitempty"`
	UserIcon          *UserIcon          `xml:"usericon,omitempty"`
	UID               *UID               `xml:"uid,omitempty"`
	Bullseye          *Bullseye          `xml:"bullseye,omitempty"`
	RouteInfo         *RouteInfo         `xml:"routeInfo,omitempty"`
	Marti             *Marti             `xml:"marti,omitempty"`
	Remarks           *Remarks           `xml:"remarks,omitempty"`
	Unknown           []RawMessage       `xml:"-"`
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

// UnmarshalXML implements xml.Unmarshaler for Detail.
func (d *Detail) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	*d = Detail{}
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "group":
				var g Group
				if err := dec.DecodeElement(&g, &t); err != nil {
					return err
				}
				d.Group = &g
			case "contact":
				var c Contact
				if err := dec.DecodeElement(&c, &t); err != nil {
					return err
				}
				d.Contact = &c
			case "__chat":
				var c Chat
				if err := dec.DecodeElement(&c, &t); err != nil {
					return err
				}
				d.Chat = &c
			case "__chatReceipt", "__chatreceipt":
				var c ChatReceipt
				if err := dec.DecodeElement(&c, &t); err != nil {
					return err
				}
				d.ChatReceipt = &c
			case "emergency":
				var em Emergency
				if err := dec.DecodeElement(&em, &t); err != nil {
					return err
				}
				d.Emergency = &em
			case "__geofence":
				var gf Geofence
				if err := dec.DecodeElement(&gf, &t); err != nil {
					return err
				}
				d.Geofence = &gf
			case "__serverdestination":
				var sd ServerDestination
				if err := dec.DecodeElement(&sd, &t); err != nil {
					return err
				}
				d.ServerDestination = &sd
			case "__video":
				var v Video
				if err := dec.DecodeElement(&v, &t); err != nil {
					return err
				}
				d.Video = &v
			case "__group":
				var gext GroupExtension
				if err := dec.DecodeElement(&gext, &t); err != nil {
					return err
				}
				d.GroupExtension = &gext
			case "archive":
				var a Archive
				if err := dec.DecodeElement(&a, &t); err != nil {
					return err
				}
				d.Archive = &a
			case "attachmentList":
				var al AttachmentList
				if err := dec.DecodeElement(&al, &t); err != nil {
					return err
				}
				d.AttachmentList = &al
			case "environment":
				var env Environment
				if err := dec.DecodeElement(&env, &t); err != nil {
					return err
				}
				d.Environment = &env
			case "fileshare":
				var fs FileShare
				if err := dec.DecodeElement(&fs, &t); err != nil {
					return err
				}
				d.FileShare = &fs
			case "precisionlocation":
				var pl PrecisionLocation
				if err := dec.DecodeElement(&pl, &t); err != nil {
					return err
				}
				d.PrecisionLocation = &pl
			case "takv":
				var tv Takv
				if err := dec.DecodeElement(&tv, &t); err != nil {
					return err
				}
				d.Takv = &tv
			case "track":
				var tr Track
				if err := dec.DecodeElement(&tr, &t); err != nil {
					return err
				}
				d.Track = &tr
			case "mission":
				var m Mission
				if err := dec.DecodeElement(&m, &t); err != nil {
					return err
				}
				d.Mission = &m
			case "status":
				var s Status
				if err := dec.DecodeElement(&s, &t); err != nil {
					return err
				}
				d.Status = &s
			case "shape":
				var sh Shape
				if err := dec.DecodeElement(&sh, &t); err != nil {
					return err
				}
				d.Shape = &sh
			case "strokecolor":
				var sc StrokeColor
				if err := dec.DecodeElement(&sc, &t); err != nil {
					return err
				}
				d.StrokeColor = &sc
			case "strokeweight":
				var sw StrokeWeight
				if err := dec.DecodeElement(&sw, &t); err != nil {
					return err
				}
				d.StrokeWeight = &sw
			case "fillcolor":
				var fc FillColor
				if err := dec.DecodeElement(&fc, &t); err != nil {
					return err
				}
				d.FillColor = &fc
			case "height":
				var h Height
				if err := dec.DecodeElement(&h, &t); err != nil {
					return err
				}
				d.Height = &h
			case "height_unit":
				var hu HeightUnit
				if err := dec.DecodeElement(&hu, &t); err != nil {
					return err
				}
				d.HeightUnit = &hu
			case "labelson":
				var lo LabelsOn
				if err := dec.DecodeElement(&lo, &t); err != nil {
					return err
				}
				d.LabelsOn = &lo
			case "color":
				var co ColorExtension
				if err := dec.DecodeElement(&co, &t); err != nil {
					return err
				}
				d.ColorExtension = &co
			case "hierarchy":
				var h Hierarchy
				if err := dec.DecodeElement(&h, &t); err != nil {
					return err
				}
				d.Hierarchy = &h
			case "link":
				var dl DetailLink
				if err := dec.DecodeElement(&dl, &t); err != nil {
					return err
				}
				d.LinkDetail = &dl
			case "usericon":
				var ui UserIcon
				if err := dec.DecodeElement(&ui, &t); err != nil {
					return err
				}
				d.UserIcon = &ui
			case "uid":
				var u UID
				if err := dec.DecodeElement(&u, &t); err != nil {
					return err
				}
				d.UID = &u
			case "bullseye":
				var b Bullseye
				if err := dec.DecodeElement(&b, &t); err != nil {
					return err
				}
				d.Bullseye = &b
			case "routeInfo":
				var ri RouteInfo
				if err := dec.DecodeElement(&ri, &t); err != nil {
					return err
				}
				d.RouteInfo = &ri
			case "marti":
				var m Marti
				if err := dec.DecodeElement(&m, &t); err != nil {
					return err
				}
				d.Marti = &m
			case "remarks":
				var r Remarks
				if err := dec.DecodeElement(&r, &t); err != nil {
					return err
				}
				d.Remarks = &r
			default:
				raw, err := captureRaw(dec, t)
				if err != nil {
					return err
				}
				d.Unknown = append(d.Unknown, raw)
			}
		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
	return nil
}

// MarshalXML implements xml.Marshaler for Detail.
func (d *Detail) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	if d.Contact != nil {
		if err := enc.Encode(d.Contact); err != nil {
			return err
		}
	}
	if d.Group != nil {
		if err := enc.Encode(d.Group); err != nil {
			return err
		}
	}
	if d.Chat != nil {
		if err := enc.Encode(d.Chat); err != nil {
			return err
		}
	}
	if d.ChatReceipt != nil {
		if err := enc.Encode(d.ChatReceipt); err != nil {
			return err
		}
	}
	if d.Emergency != nil {
		if err := encodeRaw(enc, d.Emergency.Raw); err != nil {
			return err
		}
	}
	if d.Geofence != nil {
		if err := encodeRaw(enc, d.Geofence.Raw); err != nil {
			return err
		}
	}
	if d.ServerDestination != nil {
		if err := encodeRaw(enc, d.ServerDestination.Raw); err != nil {
			return err
		}
	}
	if d.Video != nil {
		if err := encodeRaw(enc, d.Video.Raw); err != nil {
			return err
		}
	}
	if d.GroupExtension != nil {
		if err := encodeRaw(enc, d.GroupExtension.Raw); err != nil {
			return err
		}
	}
	if d.Archive != nil {
		if err := encodeRaw(enc, d.Archive.Raw); err != nil {
			return err
		}
	}
	if d.AttachmentList != nil {
		if err := encodeRaw(enc, d.AttachmentList.Raw); err != nil {
			return err
		}
	}
	if d.Environment != nil {
		if err := encodeRaw(enc, d.Environment.Raw); err != nil {
			return err
		}
	}
	if d.FileShare != nil {
		if err := encodeRaw(enc, d.FileShare.Raw); err != nil {
			return err
		}
	}
	if d.PrecisionLocation != nil {
		if err := encodeRaw(enc, d.PrecisionLocation.Raw); err != nil {
			return err
		}
	}
	if d.Takv != nil {
		if err := encodeRaw(enc, d.Takv.Raw); err != nil {
			return err
		}
	}
	if d.Track != nil {
		if err := encodeRaw(enc, d.Track.Raw); err != nil {
			return err
		}
	}
	if d.Mission != nil {
		if err := encodeRaw(enc, d.Mission.Raw); err != nil {
			return err
		}
	}
	if d.Status != nil {
		if err := encodeRaw(enc, d.Status.Raw); err != nil {
			return err
		}
	}
	if d.Shape != nil {
		if err := encodeRaw(enc, d.Shape.Raw); err != nil {
			return err
		}
	}
	if d.StrokeColor != nil {
		if err := encodeRaw(enc, d.StrokeColor.Raw); err != nil {
			return err
		}
	}
	if d.StrokeWeight != nil {
		if err := encodeRaw(enc, d.StrokeWeight.Raw); err != nil {
			return err
		}
	}
	if d.FillColor != nil {
		if err := encodeRaw(enc, d.FillColor.Raw); err != nil {
			return err
		}
	}
	if d.Height != nil {
		if err := encodeRaw(enc, d.Height.Raw); err != nil {
			return err
		}
	}
	if d.HeightUnit != nil {
		if err := encodeRaw(enc, d.HeightUnit.Raw); err != nil {
			return err
		}
	}
	if d.LabelsOn != nil {
		if err := encodeRaw(enc, d.LabelsOn.Raw); err != nil {
			return err
		}
	}
	if d.ColorExtension != nil {
		if err := encodeRaw(enc, d.ColorExtension.Raw); err != nil {
			return err
		}
	}
	if d.Hierarchy != nil {
		if err := encodeRaw(enc, d.Hierarchy.Raw); err != nil {
			return err
		}
	}
	if d.LinkDetail != nil {
		if err := encodeRaw(enc, d.LinkDetail.Raw); err != nil {
			return err
		}
	}
	if d.UserIcon != nil {
		if err := encodeRaw(enc, d.UserIcon.Raw); err != nil {
			return err
		}
	}
	if d.UID != nil {
		if err := encodeRaw(enc, d.UID.Raw); err != nil {
			return err
		}
	}
	if d.Bullseye != nil {
		if err := encodeRaw(enc, d.Bullseye.Raw); err != nil {
			return err
		}
	}
	if d.RouteInfo != nil {
		if err := encodeRaw(enc, d.RouteInfo.Raw); err != nil {
			return err
		}
	}
	if d.Marti != nil {
		if err := enc.Encode(d.Marti); err != nil {
			return err
		}
	}
	if d.Remarks != nil {
		if len(d.Remarks.Raw) > 0 && d.Remarks.Text == "" &&
			d.Remarks.Source == "" && d.Remarks.SourceID == "" &&
			d.Remarks.To == "" && d.Remarks.Time.Time().IsZero() {
			if err := encodeRaw(enc, d.Remarks.Raw); err != nil {
				return err
			}
		} else {
			if err := enc.Encode(d.Remarks); err != nil {
				return err
			}
		}
	}
	for _, raw := range d.Unknown {
		if err := encodeRaw(enc, raw); err != nil {
			return err
		}
	}
	return enc.EncodeToken(start.End())
}

// NewEvent creates a new CoT event with the given parameters
func NewEvent(uid, typ string, lat, lon, hae float64) (*Event, error) {
	now := time.Now().UTC().Truncate(time.Second)
	evt := getEvent()
	*evt = Event{
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
	if err := evt.ValidateAt(now); err != nil {
		ReleaseEvent(evt)
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
		return fmt.Errorf("empty type: %w", ErrInvalidType)
	}
	if len(typ) > 100 {
		return fmt.Errorf("type too long: %w", ErrInvalidType)
	}

	if strings.HasSuffix(typ, "-") {
		return fmt.Errorf("type cannot end with dash: %w", ErrInvalidType)
	}

	// Fast path for wildcard patterns that don't need catalog lookup
	if strings.Contains(typ, "*") {
		parts := strings.Split(typ, "-")
		if len(parts) < 2 {
			return fmt.Errorf("invalid type format: %w", ErrInvalidType)
		}

		// Only allow a trailing segment consisting solely of '*'
		for i, p := range parts {
			if strings.Contains(p, "*") {
				if p != "*" {
					return fmt.Errorf("wildcard must be standalone segment: %w", ErrInvalidType)
				}
				if i != len(parts)-1 {
					return fmt.Errorf("wildcard only allowed at end of type: %w", ErrInvalidType)
				}
			}
		}

		// Validate the prefix
		if parts[0] != "a" && parts[0] != "b" && parts[0] != "t" {
			return fmt.Errorf("invalid type prefix: %w", ErrInvalidType)
		}
		return nil
	}

	// Fast path for atomic type wildcards (a-.-X)
	if strings.HasPrefix(typ, "a-.") {
		parts := strings.Split(typ, "-")
		if len(parts) < 2 {
			return fmt.Errorf("invalid type format: %w", ErrInvalidType)
		}
		if parts[0] != "a" {
			return fmt.Errorf("wildcard only allowed in atomic types: %w", ErrInvalidType)
		}
		if parts[1] != "." {
			return fmt.Errorf("invalid wildcard format: %w", ErrInvalidType)
		}
		return nil
	}

	// Use the catalog for validation of non-wildcard types
	cat := cottypes.GetCatalog()
	_, err := cat.GetType(context.Background(), typ)
	if err != nil {
		invalidErr := fmt.Errorf("invalid type: %w", ErrInvalidType)

		// Attempt wildcard resolution by replacing f/h/n/u segments with '.'
		parts := strings.Split(typ, "-")
		for i, seg := range parts {
			switch seg {
			case "f", "h", "n", "u":
				orig := parts[i]
				parts[i] = "."
				if _, err2 := cat.GetType(context.Background(), strings.Join(parts, "-")); err2 == nil {
					return nil
				}
				parts[i] = orig
			}
		}

		return invalidErr
	}

	return nil
}

// ValidateHow checks if a how value is valid according to the CoT catalog.
// How values indicate the source or method of position determination.
func ValidateHow(how string) error {
	if how == "" {
		return nil // How is optional
	}

	// Check if it's a valid how code in the catalog
	allHows := cottypes.GetAllHows()
	for _, h := range allHows {
		if h.Value == how || h.Cot == how {
			return nil // Found it
		}
	}

	return fmt.Errorf("invalid how value %s: %w", how, ErrInvalidHow)
}

// ValidateRelation checks if a relation value is valid according to the CoT catalog.
// Relation values indicate the relationship type in link elements.
func ValidateRelation(relation string) error {
	if relation == "" {
		return fmt.Errorf("empty relation: %w", ErrInvalidRelation)
	}

	// Check if relation exists in catalog
	_, err := cottypes.GetRelationDescription(relation)
	if err != nil {
		return fmt.Errorf("invalid relation value %s: %w", relation, ErrInvalidRelation)
	}

	return nil
}

// Validate checks if the event is valid
func (e *Event) Validate() error {
	return e.ValidateAt(time.Now().UTC())
}

// ValidateAt checks if the event is valid using the provided reference time
func (e *Event) ValidateAt(now time.Time) error {
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

	// Validate how field if present
	if err := ValidateHow(e.How); err != nil {
		return fmt.Errorf("invalid how: %w", err)
	}

	// Validate link relations
	for i, link := range e.Links {
		if err := ValidateRelation(link.Relation); err != nil {
			return fmt.Errorf("invalid relation in link %d: %w", i, err)
		}
		// Also validate link type
		if err := ValidateType(link.Type); err != nil {
			return fmt.Errorf("invalid link type in link %d: %w", i, err)
		}
	}

	// Validate time fields
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
	// Skip maximum stale offset checks to allow extended validity

	// Validate point
	if err := e.Point.Validate(); err != nil {
		return err
	}

	// Validate chat-related extensions if present
	if e.Detail != nil {
		if e.Detail.Chat != nil {
			data, err := xml.Marshal(e.Detail.Chat)
			if err != nil {
				return fmt.Errorf("marshal chat: %w", err)
			}
			if err := validator.ValidateAgainstSchema("chat", data); err != nil {
				if err2 := validator.ValidateAgainstSchema("tak-details-__chat", data); err2 != nil {
					return fmt.Errorf("chat validation failed: %w", err)
				}
			} else {
				if err := validator.ValidateChat(data); err != nil {
					return fmt.Errorf("chat validation failed: %w", err)
				}
			}
		}
		if e.Detail.ChatReceipt != nil {
			var data []byte
			if len(e.Detail.ChatReceipt.Raw) > 0 {
				data = e.Detail.ChatReceipt.Raw
			} else {
				var err error
				data, err = xml.Marshal(e.Detail.ChatReceipt)
				if err != nil {
					return fmt.Errorf("marshal chatReceipt: %w", err)
				}
			}
			if err := validator.ValidateAgainstSchema("chatReceipt", data); err != nil {
				if err := validator.ValidateAgainstSchema("tak-details-__chatreceipt", data); err != nil {
					return fmt.Errorf("chatReceipt validation failed: %w", err)
				}
			}
		}

		if err := e.validateDetailSchemas(); err != nil {
			return err
		}
	}

	return nil
}

func (e *Event) validateDetailSchemas() error {
	if e.Detail == nil {
		return nil
	}

	type field struct {
		name   string
		schema string
		data   func() ([]byte, bool, error)
	}

	fields := []field{
		{
			name:   "emergency",
			schema: "tak-details-emergency",
			data: func() ([]byte, bool, error) {
				if e.Detail.Emergency == nil {
					return nil, false, nil
				}
				return e.Detail.Emergency.Raw, true, nil
			},
		},
		{
			name:   "__serverdestination",
			schema: "tak-details-__serverdestination",
			data: func() ([]byte, bool, error) {
				if e.Detail.ServerDestination == nil {
					return nil, false, nil
				}
				return e.Detail.ServerDestination.Raw, true, nil
			},
		},
		{
			name:   "__group",
			schema: "tak-details-__group",
			data: func() ([]byte, bool, error) {
				if e.Detail.GroupExtension == nil {
					return nil, false, nil
				}
				return e.Detail.GroupExtension.Raw, true, nil
			},
		},
		{
			name:   "contact",
			schema: "tak-details-contact",
			data: func() ([]byte, bool, error) {
				if e.Detail.Contact == nil {
					return nil, false, nil
				}
				b, err := xml.Marshal(e.Detail.Contact)
				return b, true, err
			},
		},
		{
			name:   "track",
			schema: "tak-details-track",
			data: func() ([]byte, bool, error) {
				if e.Detail.Track == nil {
					return nil, false, nil
				}
				return e.Detail.Track.Raw, true, nil
			},
		},
		{
			name:   "status",
			schema: "tak-details-status",
			data: func() ([]byte, bool, error) {
				if e.Detail.Status == nil {
					return nil, false, nil
				}
				return e.Detail.Status.Raw, true, nil
			},
		},
		{
			name:   "archive",
			schema: "tak-details-archive",
			data: func() ([]byte, bool, error) {
				if e.Detail.Archive == nil {
					return nil, false, nil
				}
				return e.Detail.Archive.Raw, true, nil
			},
		},
		{
			name:   "__video",
			schema: "tak-details-__video",
			data: func() ([]byte, bool, error) {
				if e.Detail.Video == nil {
					return nil, false, nil
				}
				return e.Detail.Video.Raw, true, nil
			},
		},
		{
			name:   "attachment_list",
			schema: "tak-details-attachment_list",
			data: func() ([]byte, bool, error) {
				if e.Detail.AttachmentList == nil {
					return nil, false, nil
				}
				return e.Detail.AttachmentList.Raw, true, nil
			},
		},
		{
			name:   "uid",
			schema: "tak-details-uid",
			data: func() ([]byte, bool, error) {
				if e.Detail.UID == nil {
					return nil, false, nil
				}
				return e.Detail.UID.Raw, true, nil
			},
		},
		{
			name:   "bullseye",
			schema: "tak-details-bullseye",
			data: func() ([]byte, bool, error) {
				if e.Detail.Bullseye == nil {
					return nil, false, nil
				}
				return e.Detail.Bullseye.Raw, true, nil
			},
		},
		{
			name:   "routeinfo",
			schema: "tak-details-routeinfo",
			data: func() ([]byte, bool, error) {
				if e.Detail.RouteInfo == nil {
					return nil, false, nil
				}
				return e.Detail.RouteInfo.Raw, true, nil
			},
		},
		{
			name:   "marti",
			schema: "tak-details-marti",
			data: func() ([]byte, bool, error) {
				if e.Detail.Marti == nil {
					return nil, false, nil
				}
				b, err := xml.Marshal(e.Detail.Marti)
				return b, true, err
			},
		},
		{
			name:   "environment",
			schema: "tak-details-environment",
			data: func() ([]byte, bool, error) {
				if e.Detail.Environment == nil {
					return nil, false, nil
				}
				return e.Detail.Environment.Raw, true, nil
			},
		},
		{
			name:   "fileshare",
			schema: "tak-details-fileshare",
			data: func() ([]byte, bool, error) {
				if e.Detail.FileShare == nil {
					return nil, false, nil
				}
				return e.Detail.FileShare.Raw, true, nil
			},
		},
		{
			name:   "precisionlocation",
			schema: "tak-details-precisionlocation",
			data: func() ([]byte, bool, error) {
				if e.Detail.PrecisionLocation == nil {
					return nil, false, nil
				}
				return e.Detail.PrecisionLocation.Raw, true, nil
			},
		},
		{
			name:   "takv",
			schema: "tak-details-takv",
			data: func() ([]byte, bool, error) {
				if e.Detail.Takv == nil {
					return nil, false, nil
				}
				return e.Detail.Takv.Raw, true, nil
			},
		},
		{
			name:   "mission",
			schema: "tak-details-mission",
			data: func() ([]byte, bool, error) {
				if e.Detail.Mission == nil {
					return nil, false, nil
				}
				return e.Detail.Mission.Raw, true, nil
			},
		},
		{
			name:   "shape",
			schema: "tak-details-shape",
			data: func() ([]byte, bool, error) {
				if e.Detail.Shape == nil {
					return nil, false, nil
				}
				return e.Detail.Shape.Raw, true, nil
			},
		},
		{
			name:   "__geofence",
			schema: "tak-details-__geofence",
			data: func() ([]byte, bool, error) {
				if e.Detail.Geofence == nil {
					return nil, false, nil
				}
				return e.Detail.Geofence.Raw, true, nil
			},
		},
		{
			name:   "strokeColor",
			schema: "tak-details-strokeColor",
			data: func() ([]byte, bool, error) {
				if e.Detail.StrokeColor == nil {
					return nil, false, nil
				}
				return e.Detail.StrokeColor.Raw, true, nil
			},
		},
		{
			name:   "strokeWeight",
			schema: "tak-details-strokeWeight",
			data: func() ([]byte, bool, error) {
				if e.Detail.StrokeWeight == nil {
					return nil, false, nil
				}
				return e.Detail.StrokeWeight.Raw, true, nil
			},
		},
		{
			name:   "fillColor",
			schema: "tak-details-fillColor",
			data: func() ([]byte, bool, error) {
				if e.Detail.FillColor == nil {
					return nil, false, nil
				}
				return e.Detail.FillColor.Raw, true, nil
			},
		},
		{
			name:   "height",
			schema: "tak-details-height",
			data: func() ([]byte, bool, error) {
				if e.Detail.Height == nil {
					return nil, false, nil
				}
				return e.Detail.Height.Raw, true, nil
			},
		},
		{
			name:   "height_unit",
			schema: "tak-details-height_unit",
			data: func() ([]byte, bool, error) {
				if e.Detail.HeightUnit == nil {
					return nil, false, nil
				}
				return e.Detail.HeightUnit.Raw, true, nil
			},
		},
		{
			name:   "labels_on",
			schema: "tak-details-labels_on",
			data: func() ([]byte, bool, error) {
				if e.Detail.LabelsOn == nil {
					return nil, false, nil
				}
				return e.Detail.LabelsOn.Raw, true, nil
			},
		},
		{
			name:   "color",
			schema: "tak-details-color",
			data: func() ([]byte, bool, error) {
				if e.Detail.ColorExtension == nil {
					return nil, false, nil
				}
				return e.Detail.ColorExtension.Raw, true, nil
			},
		},
		{
			name:   "hierarchy",
			schema: "tak-details-hierarchy",
			data: func() ([]byte, bool, error) {
				if e.Detail.Hierarchy == nil {
					return nil, false, nil
				}
				return e.Detail.Hierarchy.Raw, true, nil
			},
		},
		{
			name:   "link",
			schema: "tak-details-link",
			data: func() ([]byte, bool, error) {
				if e.Detail.LinkDetail == nil {
					return nil, false, nil
				}
				return e.Detail.LinkDetail.Raw, true, nil
			},
		},
		{
			name:   "usericon",
			schema: "tak-details-usericon",
			data: func() ([]byte, bool, error) {
				if e.Detail.UserIcon == nil {
					return nil, false, nil
				}
				return e.Detail.UserIcon.Raw, true, nil
			},
		},
		{
			name:   "remarks",
			schema: "tak-details-remarks",
			data: func() ([]byte, bool, error) {
				if e.Detail.Remarks == nil {
					return nil, false, nil
				}
				if len(e.Detail.Remarks.Raw) > 0 && e.Detail.Remarks.Text == "" &&
					e.Detail.Remarks.Source == "" && e.Detail.Remarks.SourceID == "" &&
					e.Detail.Remarks.To == "" && e.Detail.Remarks.Time.Time().IsZero() {
					return e.Detail.Remarks.Raw, true, nil
				}
				data, err := xml.Marshal(e.Detail.Remarks)
				return data, true, err
			},
		},
	}

	for _, f := range fields {
		data, ok, err := f.data()
		if err != nil {
			return fmt.Errorf("marshal %s: %w", f.name, err)
		}
		if !ok {
			continue
		}
		if err := validator.ValidateAgainstSchema(f.schema, data); err != nil {
			return fmt.Errorf("invalid %s: %w", f.name, err)
		}
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
	return ctxlog.WithLogger(ctx, l)
}

// LoggerFromContext retrieves the logger from context or returns slog.Default.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	return ctxlog.LoggerFromContext(ctx)
}

// GetTypeFullName returns the full hierarchical name for a CoT type.
// For example, "a-f-G-E-X-N" returns "Gnd/Equip/Nbc Equipment".
//
// The full name represents the type's position in the CoT type hierarchy,
// making it useful for building user interfaces and documentation.
//
// Returns an error if the type is not registered in the catalog.
func GetTypeFullName(name string) (string, error) {
	return cottypes.GetCatalog().GetFullName(context.Background(), name)
}

// GetTypeDescription returns the human-readable description for a CoT type.
// For example, "a-f-G-E-X-N" returns "NBC EQUIPMENT".
//
// The description is a concise explanation of what the type represents,
// suitable for display in user interfaces and logs.
//
// Returns an error if the type is not registered in the catalog.
func GetTypeDescription(name string) (string, error) {
	return cottypes.GetCatalog().GetDescription(context.Background(), name)
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
	return cottypes.GetCatalog().FindByDescription(context.Background(), desc)
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
	return cottypes.GetCatalog().FindByFullName(context.Background(), name)
}

// UnmarshalXMLEvent parses an XML byte slice into an Event. The returned Event
// is obtained from an internal pool; callers should release it with
// ReleaseEvent when finished.
// The function uses the standard library's encoding/xml Decoder under the hood.
func UnmarshalXMLEvent(ctx context.Context, data []byte) (*Event, error) {
	logger := LoggerFromContext(ctx)

	if len(data) > int(currentMaxXMLSize()) {
		logger.Error("xml size exceeds limit",
			"size", len(data),
			"limit", currentMaxXMLSize())
		return nil, ErrInvalidInput
	}

	// Check for DOCTYPE in a case-insensitive manner
	if doctypePattern.Match(data) {
		logger.Error("invalid doctype detected")
		return nil, ErrInvalidInput
	}

	// Check namespace length
	if idx := bytes.Index(data, []byte(`xmlns="`)); idx >= 0 {
		end := bytes.Index(data[idx+7:], []byte(`"`))
		if end > 1024 {
			logger.Error("namespace value too long")
			return nil, ErrInvalidInput
		}
	}

	pd := getDecoder(data)
	defer putDecoder(pd)

	evt := getEvent()
	if err := decodeWithLimits(pd.dec, evt); err != nil {
		ReleaseEvent(evt)
		logger.Error("failed to decode XML", "error", err)
		return nil, fmt.Errorf("failed to decode XML: %w", err)
	}

	if evt.Type == "b-t-f" && evt.Detail != nil && evt.Detail.Remarks != nil {
		if evt.Detail.Remarks.Text == "" {
			_ = evt.Detail.Remarks.Parse()
		}
		evt.Message = evt.Detail.Remarks.Text
	}

	if err := evt.ValidateAt(time.Now().UTC()); err != nil {
		ReleaseEvent(evt)
		logger.Error("event validation failed", "error", err)
		return nil, err
	}

	return evt, nil
}

// UnmarshalXMLEventCtx parses an XML byte slice into an Event using the
// provided context for logging. The returned Event is obtained from an
// internal pool and must be released with ReleaseEvent when finished.
func UnmarshalXMLEventCtx(ctx context.Context, data []byte) (*Event, error) {
	return UnmarshalXMLEvent(ctx, data)
}

// ValidateLatLon checks if latitude and longitude are within valid ranges
func ValidateLatLon(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return ErrInvalidLatitude
	}
	if lon < -180 || lon > 180 {
		return ErrInvalidLongitude
	}
	return nil
}

// ValidateUID checks if a UID is valid.
// It rejects empty values, leading hyphens, double dots,
// whitespace, and UIDs longer than 64 characters.
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
	if len(uid) > 64 {
		return ErrInvalidUID
	}
	if strings.ContainsAny(uid, " \t\n\r") {
		return ErrInvalidUID
	}
	return nil
}

// ToXML serialises an Event to CoT-compliant XML.
// Attribute values are escaped to prevent XML-injection.
// The <point> element is always emitted so that the
// zero coordinate (0° N 0° E) is representable.
func (e *Event) ToXML() ([]byte, error) {
	buf := getBuffer()
	defer putBuffer(buf)
	buf.Grow(256)
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	var tmp [32]byte

	// <event>
	buf.WriteString("<event")
	if e.Version != "" {
		buf.WriteString(` version="`)
		buf.WriteString(escapeAttr(e.Version))
		buf.WriteByte('"')
	}
	if e.Type != "" {
		buf.WriteString(` type="`)
		buf.WriteString(escapeAttr(e.Type))
		buf.WriteByte('"')
	}
	if e.How != "" {
		buf.WriteString(` how="`)
		buf.WriteString(escapeAttr(e.How))
		buf.WriteByte('"')
	}
	if e.Uid != "" {
		buf.WriteString(` uid="`)
		buf.WriteString(escapeAttr(e.Uid))
		buf.WriteByte('"')
	}
	if !e.Time.Time().IsZero() {
		buf.WriteString(` time="`)
		buf.WriteString(e.Time.Time().UTC().Format(CotTimeFormat))
		buf.WriteByte('"')
	}
	if !e.Start.Time().IsZero() {
		buf.WriteString(` start="`)
		buf.WriteString(e.Start.Time().UTC().Format(CotTimeFormat))
		buf.WriteByte('"')
	}
	if !e.Stale.Time().IsZero() {
		buf.WriteString(` stale="`)
		buf.WriteString(e.Stale.Time().UTC().Format(CotTimeFormat))
		buf.WriteByte('"')
	}
	if e.StrokeColor != "" {
		buf.WriteString(` strokeColor="`)
		buf.WriteString(escapeAttr(e.StrokeColor))
		buf.WriteByte('"')
	}
	if e.UserIcon != "" {
		buf.WriteString(` usericon="`)
		buf.WriteString(escapeAttr(e.UserIcon))
		buf.WriteByte('"')
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
				buf.WriteString(` callsign="`)
				buf.WriteString(escapeAttr(c.Callsign))
				buf.WriteByte('"')
			}
			buf.WriteString("/>\n")
		}
		if g := e.Detail.Group; g != nil {
			buf.WriteString("    <group")
			if g.Name != "" {
				buf.WriteString(` name="`)
				buf.WriteString(escapeAttr(g.Name))
				buf.WriteByte('"')
			}
			if g.Role != "" {
				buf.WriteString(` role="`)
				buf.WriteString(escapeAttr(g.Role))
				buf.WriteByte('"')
			}
			buf.WriteString("/>\n")
		}
		if e.Detail.Chat != nil {
			if len(e.Detail.Chat.Raw) > 0 {
				buf.WriteString("    ")
				buf.Write(e.Detail.Chat.Raw)
				buf.WriteByte('\n')
			} else {
				buf.WriteString("    <__chat")
				if e.Detail.Chat.ID != "" {
					buf.WriteString(` id="`)
					buf.WriteString(escapeAttr(e.Detail.Chat.ID))
					buf.WriteByte('"')
				}
				if e.Detail.Chat.Message != "" {
					buf.WriteString(` message="`)
					buf.WriteString(escapeAttr(e.Detail.Chat.Message))
					buf.WriteByte('"')
				}
				if e.Detail.Chat.Sender != "" {
					buf.WriteString(` sender="`)
					buf.WriteString(escapeAttr(e.Detail.Chat.Sender))
					buf.WriteByte('"')
				}
				if e.Detail.Chat.Chatroom != "" {
					buf.WriteString(` chatroom="`)
					buf.WriteString(escapeAttr(e.Detail.Chat.Chatroom))
					buf.WriteByte('"')
				}
				if e.Detail.Chat.GroupOwner != "" {
					buf.WriteString(` groupOwner="`)
					buf.WriteString(escapeAttr(e.Detail.Chat.GroupOwner))
					buf.WriteByte('"')
				}
				if e.Detail.Chat.SenderCallsign != "" {
					buf.WriteString(` senderCallsign="`)
					buf.WriteString(escapeAttr(e.Detail.Chat.SenderCallsign))
					buf.WriteByte('"')
				}
				if e.Detail.Chat.Parent != "" {
					buf.WriteString(` parent="`)
					buf.WriteString(escapeAttr(e.Detail.Chat.Parent))
					buf.WriteByte('"')
				}
				if e.Detail.Chat.MessageID != "" {
					buf.WriteString(` messageId="`)
					buf.WriteString(escapeAttr(e.Detail.Chat.MessageID))
					buf.WriteByte('"')
				}
				if e.Detail.Chat.DeleteChild != "" {
					buf.WriteString(` deleteChild="`)
					buf.WriteString(escapeAttr(e.Detail.Chat.DeleteChild))
					buf.WriteByte('"')
				}
				if len(e.Detail.Chat.ChatGrps) == 0 && e.Detail.Chat.Hierarchy == nil {
					buf.WriteString("/>\n")
				} else {
					buf.WriteString(">\n")
					for _, g := range e.Detail.Chat.ChatGrps {
						buf.WriteString("      <chatgrp")
						if g.ID != "" {
							buf.WriteString(` id="`)
							buf.WriteString(escapeAttr(g.ID))
							buf.WriteByte('"')
						}
						if g.UID0 != "" {
							buf.WriteString(` uid0="`)
							buf.WriteString(escapeAttr(g.UID0))
							buf.WriteByte('"')
						}
						if g.UID1 != "" {
							buf.WriteString(` uid1="`)
							buf.WriteString(escapeAttr(g.UID1))
							buf.WriteByte('"')
						}
						if g.UID2 != "" {
							buf.WriteString(` uid2="`)
							buf.WriteString(escapeAttr(g.UID2))
							buf.WriteByte('"')
						}
						buf.WriteString("/>")
						buf.WriteByte('\n')
					}
					if e.Detail.Chat.Hierarchy != nil {
						buf.WriteString("      ")
						buf.Write(e.Detail.Chat.Hierarchy.Raw)
						buf.WriteByte('\n')
					}
					buf.WriteString("    </__chat>\n")
				}
			}
		}
		if e.Detail.ChatReceipt != nil {
			cr := e.Detail.ChatReceipt
			if len(cr.Raw) > 0 && cr.Ack == "" && cr.ID == "" && cr.Chatroom == "" && cr.GroupOwner == "" && cr.SenderCallsign == "" && cr.MessageID == "" && cr.Parent == "" && cr.ChatGrp == nil {
				buf.WriteString("    ")
				buf.Write(cr.Raw)
				buf.WriteByte('\n')
			} else {
				name := cr.XMLName.Local
				if name == "" {
					name = "__chatReceipt"
				}
				buf.WriteString("    <" + name)
				if cr.Ack != "" {
					buf.WriteString(` ack="`)
					buf.WriteString(escapeAttr(cr.Ack))
					buf.WriteByte('"')
				}
				if cr.ID != "" {
					buf.WriteString(` id="`)
					buf.WriteString(escapeAttr(cr.ID))
					buf.WriteByte('"')
				}
				if cr.Chatroom != "" {
					buf.WriteString(` chatroom="`)
					buf.WriteString(escapeAttr(cr.Chatroom))
					buf.WriteByte('"')
				}
				if cr.GroupOwner != "" {
					buf.WriteString(` groupOwner="`)
					buf.WriteString(escapeAttr(cr.GroupOwner))
					buf.WriteByte('"')
				}
				if cr.SenderCallsign != "" {
					buf.WriteString(` senderCallsign="`)
					buf.WriteString(escapeAttr(cr.SenderCallsign))
					buf.WriteByte('"')
				}
				if cr.MessageID != "" {
					buf.WriteString(` messageId="`)
					buf.WriteString(escapeAttr(cr.MessageID))
					buf.WriteByte('"')
				}
				if cr.Parent != "" {
					buf.WriteString(` parent="`)
					buf.WriteString(escapeAttr(cr.Parent))
					buf.WriteByte('"')
				}
				if cr.ChatGrp != nil {
					buf.WriteString(">\n")
					buf.WriteString("      <chatgrp")
					if cr.ChatGrp.ID != "" {
						buf.WriteString(` id="`)
						buf.WriteString(escapeAttr(cr.ChatGrp.ID))
						buf.WriteByte('"')
					}
					if cr.ChatGrp.UID0 != "" {
						buf.WriteString(` uid0="`)
						buf.WriteString(escapeAttr(cr.ChatGrp.UID0))
						buf.WriteByte('"')
					}
					if cr.ChatGrp.UID1 != "" {
						buf.WriteString(` uid1="`)
						buf.WriteString(escapeAttr(cr.ChatGrp.UID1))
						buf.WriteByte('"')
					}
					if cr.ChatGrp.UID2 != "" {
						buf.WriteString(` uid2="`)
						buf.WriteString(escapeAttr(cr.ChatGrp.UID2))
						buf.WriteByte('"')
					}
					buf.WriteString("/>")
					buf.WriteString("\n    </" + name + ">\n")
				} else {
					buf.WriteString("/>\n")
				}
			}
		}
		if e.Detail.RouteInfo != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.RouteInfo.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.Marti != nil {
			buf.WriteString("    <marti>\n")
			for _, d := range e.Detail.Marti.Dest {
				buf.WriteString("      <dest")
				if d.Callsign != "" {
					buf.WriteString(` callsign="`)
					buf.WriteString(escapeAttr(d.Callsign))
					buf.WriteByte('"')
				}
				buf.WriteString("/>")
				buf.WriteByte('\n')
			}
			buf.WriteString("    </marti>\n")
		}
		if e.Detail.Geofence != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.Geofence.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.ServerDestination != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.ServerDestination.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.Video != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.Video.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.GroupExtension != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.GroupExtension.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.Archive != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.Archive.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.AttachmentList != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.AttachmentList.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.Environment != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.Environment.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.FileShare != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.FileShare.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.PrecisionLocation != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.PrecisionLocation.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.Takv != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.Takv.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.Track != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.Track.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.Mission != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.Mission.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.Status != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.Status.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.Shape != nil {
			buf.WriteString("    ")
			buf.Write(e.Detail.Shape.Raw)
			buf.WriteByte('\n')
		}
		if e.Detail.Remarks != nil {
			r := e.Detail.Remarks
			if len(r.Raw) > 0 && r.Text == "" && r.Source == "" && r.SourceID == "" && r.To == "" && r.Time.Time().IsZero() {
				buf.WriteString("    ")
				buf.Write(r.Raw)
				buf.WriteByte('\n')
			} else {
				buf.WriteString("    <remarks")
				if r.Source != "" {
					buf.WriteString(` source="`)
					buf.WriteString(escapeAttr(r.Source))
					buf.WriteByte('"')
				}
				if r.SourceID != "" {
					buf.WriteString(` sourceID="`)
					buf.WriteString(escapeAttr(r.SourceID))
					buf.WriteByte('"')
				}
				if r.To != "" {
					buf.WriteString(` to="`)
					buf.WriteString(escapeAttr(r.To))
					buf.WriteByte('"')
				}
				if !r.Time.Time().IsZero() {
					buf.WriteString(` time="`)
					buf.WriteString(r.Time.Time().UTC().Format(CotTimeFormat))
					buf.WriteByte('"')
				}
				if r.Text == "" {
					buf.WriteString("/>")
					buf.WriteByte('\n')
				} else {
					buf.WriteString(">")
					escapeText(buf, r.Text)
					buf.WriteString("</remarks>\n")
				}
			}
		}
		for _, raw := range e.Detail.Unknown {
			buf.WriteString("    ")
			buf.Write(raw)
			buf.WriteByte('\n')
		}
		buf.WriteString("  </detail>\n")
	}

	// <link> (0..n)
	for _, l := range e.Links {
		buf.WriteString("  <link")
		if l.Uid != "" {
			buf.WriteString(` uid="`)
			buf.WriteString(escapeAttr(l.Uid))
			buf.WriteByte('"')
		}
		if l.Type != "" {
			buf.WriteString(` type="`)
			buf.WriteString(escapeAttr(l.Type))
			buf.WriteByte('"')
		}
		if l.Relation != "" {
			buf.WriteString(` relation="`)
			buf.WriteString(escapeAttr(l.Relation))
			buf.WriteByte('"')
		}
		buf.WriteString("/>\n")
	}

	buf.WriteString("</event>")
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

// SetEventHowFromDescriptor sets the how field on an event using a descriptor.
// For example: SetEventHowFromDescriptor(event, "gps") sets how to "h-g-i-g-o".
// It returns an error if event is nil or the descriptor is invalid.
func SetEventHowFromDescriptor(event *Event, descriptor string) error {
	if event == nil {
		return fmt.Errorf("nil event")
	}

	howValue, err := cottypes.GetHowValue(descriptor)
	if err != nil {
		return fmt.Errorf("failed to get how value for descriptor %s: %w", descriptor, err)
	}
	event.How = howValue
	return nil
}

// AddValidatedLink adds a link to the event after validating the relation and type.
// It returns an error if called on a nil Event.
func (e *Event) AddValidatedLink(uid, linkType, relation string) error {
	if e == nil {
		return fmt.Errorf("nil event")
	}

	if err := ValidateType(linkType); err != nil {
		return fmt.Errorf("invalid link type: %w", err)
	}
	if err := ValidateRelation(relation); err != nil {
		return fmt.Errorf("invalid relation: %w", err)
	}

	e.AddLink(&Link{
		Uid:      uid,
		Type:     linkType,
		Relation: relation,
	})
	return nil
}

// GetHowDescriptor returns a human-readable description of the how value.
// For example: "h-g-i-g-o" returns "gps".
func GetHowDescriptor(how string) (string, error) {
	return cottypes.GetHowNick(how)
}

// GetRelationDescription returns a human-readable description of the relation value.
// For example: "p-p" returns "parent-point".
func GetRelationDescription(relation string) (string, error) {
	return cottypes.GetRelationDescription(relation)
}
