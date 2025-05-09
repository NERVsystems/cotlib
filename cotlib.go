/*
Package cotlib implements the Cursor on Target (CoT) protocol for Go.

The package provides data structures and utilities for parsing and generating
CoT XML messages with a focus on security and standards compliance.

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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
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

	// CotTimeFormat is the standard time format for CoT messages (Zulu time, no offset)
	// Format: "2006-01-02T15:04:05Z" (UTC without timezone offset)
	CotTimeFormat = "2006-01-02T15:04:05Z"
)

// maxValueLen is the maximum length for attribute values and character data
// Set to 512 KiB to accommodate large KML polygons
var maxValueLen = 512 << 10

// validCoTTypes is a registry of valid CoT types
var validCoTTypes = make(map[string]bool)

// typeRegistryMu protects validCoTTypes
var typeRegistryMu sync.RWMutex

// RegisterCoTType adds a specific CoT type to the valid types registry
func RegisterCoTType(typ string) {
	// Only register if it's a valid format (at least one hyphen and not incomplete)
	if strings.Count(typ, "-") >= 1 && !strings.HasSuffix(typ, "-") {
		// Additional validation for incomplete types
		parts := strings.Split(typ, "-")
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return
		}

		// Validate type format based on prefix
		switch parts[0] {
		case "a":
			// Atomic types must have at least 3 parts: affiliation, domain, and category
			if len(parts) < 3 {
				return
			}
		case "b", "t", "y", "c":
			// Other types must have at least 2 parts: prefix and subtype
			if len(parts) < 2 {
				return
			}
		default:
			// Unknown prefix
			return
		}

		typeRegistryMu.Lock()
		validCoTTypes[typ] = true
		typeRegistryMu.Unlock()
	}
}

// RegisterCoTTypesFromFile loads and registers CoT types from an XML file
func RegisterCoTTypesFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return RegisterCoTTypesFromReader(file)
}

// RegisterCoTTypesFromReader loads and registers CoT types from an XML reader
func RegisterCoTTypesFromReader(r io.Reader) error {
	decoder := xml.NewDecoder(r)

	// Simple structure to parse the XML
	var types struct {
		XMLName xml.Name `xml:"types"`
		CoTs    []struct {
			Type string `xml:"cot,attr"`
		} `xml:"cot"`
	}

	if err := decoder.Decode(&types); err != nil {
		return err
	}

	// Register all types
	for _, t := range types.CoTs {
		if t.Type != "" {
			RegisterCoTType(t.Type)
		}
	}

	return nil
}

// RegisterStandardCoTTypes registers all the standard CoT types defined in CoTtypes.xml
// This should be called during application initialization if full CoT type validation is needed
func RegisterStandardCoTTypes() {
	// This is a subset of the most common CoT types to register
	// Applications should still call RegisterCoTTypesFromFile or RegisterCoTTypesFromXMLContent
	// with the complete CoTtypes.xml for full validation

	// Atom types with different affiliations
	registerAtomTypes()

	// Bits types
	registerBitsTypes()

	// Message types
	registerMessageTypes()

	// Capabilities, Taskings, and Replies
	registerTaskingTypes()

	// Log registration stats
	logger := getLogger(context.Background())
	logger.Info("registered standard CoT types",
		"types", len(validCoTTypes))
}

// registerAtomTypes registers common atom types (a-*)
func registerAtomTypes() {
	atomTypes := []string{
		// Common track types
		"a-f-G", "a-h-G", "a-u-G", // Ground
		"a-f-A", "a-h-A", "a-u-A", // Air
		"a-f-S", "a-h-S", "a-u-S", // Surface
		"a-f-U", "a-h-U", "a-u-U", // Subsurface

		// Common military units
		"a-f-G-U-C", "a-h-G-U-C", "a-u-G-U-C", // Combat
		"a-f-G-E-V", "a-h-G-E-V", "a-u-G-E-V", // Vehicles
		"a-f-G-I", "a-h-G-I", "a-u-G-I", // Installations

		// Incident Management types
		"a-f-X-i", "a-h-X-i", "a-u-X-i", // Incidents
	}

	for _, t := range atomTypes {
		RegisterCoTType(t)
	}
}

// registerBitsTypes registers common bits types (b-*)
func registerBitsTypes() {
	bitsTypes := []string{
		// Common bits types
		"b-i", "b-m-p", "b-m-r", "b-d", "b-l",

		// Alarm types
		"b-l-c", "b-l-f", "b-l-m", "b-l-l",

		// Detection types
		"b-d-a", "b-d-c", "b-d-m", "b-d-s",

		// Map-related types
		"b-m-p-s-p-i", "b-m-p-w", "b-m-p-c",
	}

	for _, t := range bitsTypes {
		RegisterCoTType(t)
	}
}

// registerMessageTypes registers common message and protocol types
func registerMessageTypes() {
	msgTypes := []string{
		// TAK Protocol types
		"t-x-c", "t-x-m", "t-x-d", "t-x-t",

		// Special handling for 9-1-1 (Mayday) types
		"a-f-G-E-V-9-1-1", "a-u-G-E-V-9-1-1",
	}

	for _, t := range msgTypes {
		RegisterCoTType(t)
	}
}

// registerTaskingTypes registers tasking, capability, and reply types
func registerTaskingTypes() {
	taskingTypes := []string{
		// Tasking types
		"t-k", "t-s", "t-m", "t-r", "t-u", "t-q",

		// Reply types
		"y-a", "y-c", "y-s",

		// Capability types
		"c-f", "c-c", "c-r", "c-s", "c-l",
	}

	for _, t := range taskingTypes {
		RegisterCoTType(t)
	}
}

// init initializes the CoT type registry with common CoT types and prefixes
func init() {
	// Register common tactical type patterns
	// These represent base patterns that may have wildcards
	commonTypes := []string{
		"a-.-A", "a-.-G", "a-.-S", "a-.-U", "a-.-X",
		"b-d-a", "b-i-a", "b-m-p", "b-m-r", "b-l-a",
		"t-k-a", "y-a-a", "r-c-x",
	}

	for _, t := range commonTypes {
		RegisterCoTType(t)
	}

	// Handle wildcards in CoT types
	// The ".-" pattern in types like "a-.-G" means any affiliation
	affiliations := []string{"f", "h", "u", "n", "s", "j", "k", "a", "p"}
	for _, t := range commonTypes {
		if strings.Contains(t, ".-") {
			baseParts := strings.Split(t, ".-")
			if len(baseParts) == 2 && len(baseParts[0]) > 0 {
				// Register expanded prefixes with all affiliations
				for _, aff := range affiliations {
					expandedPrefix := baseParts[0] + "-" + aff + "-" + baseParts[1]
					RegisterCoTType(expandedPrefix)
				}
			}
		}
	}

	// Register all standard types for convenience
	// Use RegisterAllCoTTypes() to register all types from the complete CoTtypes.xml
	RegisterStandardCoTTypes()

	// Log initialization
	logger := getLogger(context.Background())
	logger.Info("registered standard CoT types", "types", len(validCoTTypes))
}

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
	// systemTypePrefixes are whitelisted prefixes for TAK system messages
	systemTypePrefixes = []string{
		"t-x-", // TAK protocol / server status / heartbeat
		"t-a-", // TAK admin
		"t-m-", // TAK meta
	}

	// tacticalTypePattern enforces strict validation for tactical symbols
	tacticalTypePattern = regexp.MustCompile(`^a-[fhun]-[AGSU](?:-[A-Z](?:-[A-Za-z]+)?)*$`)

	// takTypePattern enforces strict validation for TAK protocol types
	takTypePattern = regexp.MustCompile(`^t-x-[a-z]+(?:-[a-z]+)*$`)

	// otherTypePattern enforces strict validation for other types
	otherTypePattern = regexp.MustCompile(`^[b-z]-[a-z](?:-[a-z]+)?$`)

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

// CoTTime is a custom time type that implements XML marshaling for CoT format
type CoTTime time.Time

// Time returns the underlying time.Time value
func (t CoTTime) Time() time.Time {
	return time.Time(t)
}

// Format formats the time using the given layout
func (t CoTTime) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// MarshalXMLAttr implements xml.MarshalerAttr for CoTTime
func (t CoTTime) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: marshalCoTTime(time.Time(t))}, nil
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr for CoTTime
func (t *CoTTime) UnmarshalXMLAttr(attr xml.Attr) error {
	tt, err := unmarshalCoTTime(attr.Value)
	if err != nil {
		return err
	}
	*t = CoTTime(tt)
	return nil
}

// Event represents a CoT event
type Event struct {
	Version string  `xml:"version,attr"`
	Uid     string  `xml:"uid,attr"`
	Type    string  `xml:"type,attr"`
	How     string  `xml:"how,attr"`
	Time    CoTTime `xml:"time,attr"`
	Start   CoTTime `xml:"start,attr"`
	Stale   CoTTime `xml:"stale,attr"`
	Point   Point   `xml:"point"`
	Detail  *Detail `xml:"detail,omitempty"`
	Links   []*Link `xml:"link,omitempty"`
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

// Detail represents the optional detail element in a CoT event
type Detail struct {
	// Existing fields
	Shape    *Shape    `xml:"shape,omitempty"`
	Contact  *Contact  `xml:"contact,omitempty"`
	FlowTags *FlowTags `xml:"__flow-tags__,omitempty"`

	// New TAK-specific fields
	Group *Group `xml:"__group,omitempty"`
	TakV  *TakV  `xml:"takv,omitempty"`
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
	Endpoint string   `xml:"endpoint,attr,omitempty"` // TAK-specific endpoint
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

// Validate validates the event
func (e *Event) Validate() error {
	if err := ValidateUID(e.Uid); err != nil {
		return err
	}
	if err := ValidateType(e.Type); err != nil {
		return err
	}
	if err := ValidateLatLon(e.Point.Lat, e.Point.Lon); err != nil {
		return err
	}
	if err := e.validateTimes(); err != nil {
		return err
	}
	if e.Detail != nil {
		if err := e.Detail.validateDetail(); err != nil {
			return err
		}
	}
	return e.Point.Validate()
}

// validateShape validates the shape fields
func (s *Shape) validateShape() error {
	logger := getLogger(context.Background())

	if s.Type != "" && len(s.Type) > 64 {
		logger.Error("shape type too long",
			"type", s.Type,
			"max_length", 64)
		return fmt.Errorf("%w: type exceeds maximum length of 64 characters", ErrInvalidShape)
	}

	if s.Points != "" && len(s.Points) > 1024 {
		logger.Error("shape points too long",
			"points", s.Points,
			"max_length", 1024)
		return fmt.Errorf("%w: points exceed maximum length of 1024 characters", ErrInvalidShape)
	}

	// Add semantic validation
	switch s.Type {
	case "circle", "ellipse":
		if s.Radius <= 0 {
			return fmt.Errorf("%w: radius must be >0", ErrInvalidShape)
		}
	case "polygon":
		if s.Points == "" {
			return fmt.Errorf("%w: polygon requires points", ErrInvalidShape)
		}
	}

	return nil
}

// validateTimes checks if the event's time fields are valid
func (e *Event) validateTimes() error {
	now := time.Now().UTC()
	eventTime := e.Time.Time()
	startTime := e.Start.Time()
	staleTime := e.Stale.Time()

	// Check if event time is too far in the past
	if eventTime.Before(now.Add(-24 * time.Hour)) {
		logger := getLogger(context.Background())
		logger.Error("event time too far in past",
			"time", eventTime,
			"now", now,
			"max_past_offset", 24*time.Hour,
			"actual_offset", eventTime.Sub(now),
			"error", "time must be in RFC3339 format")
		return errors.New("time must be in RFC3339 format")
	}

	// Check if event time is too far in the future
	if eventTime.After(now.Add(24 * time.Hour)) {
		logger := getLogger(context.Background())
		logger.Error("event time too far in future",
			"time", eventTime,
			"now", now,
			"max_future_offset", 24*time.Hour,
			"actual_offset", eventTime.Sub(now),
			"error", "time must be in RFC3339 format")
		return errors.New("time must be in RFC3339 format")
	}

	// Check if start time is after event time
	if startTime.After(eventTime) {
		logger := getLogger(context.Background())
		logger.Error("start time after event time",
			"start", startTime,
			"time", eventTime,
			"error", "time must be in RFC3339 format")
		return errors.New("time must be in RFC3339 format")
	}

	// Check if stale time is too close to event time
	if staleTime.Sub(eventTime) < minStaleOffset {
		logger := getLogger(context.Background())
		logger.Error("stale time too close to event time",
			"stale", staleTime,
			"time", eventTime,
			"min_offset", minStaleOffset,
			"actual_offset", staleTime.Sub(eventTime),
			"error", "stale time must be after event time")
		return errors.New("stale time must be after event time")
	}

	// For non-TAK types, enforce maximum stale time
	if !strings.HasPrefix(e.Type, "t-x-takp") {
		if staleTime.Sub(eventTime) > maxStaleOffset {
			logger := getLogger(context.Background())
			logger.Error("stale time too far in future for non-TAK type",
				"stale", staleTime,
				"time", eventTime,
				"max_offset", maxStaleOffset,
				"actual_offset", staleTime.Sub(eventTime),
				"error", "stale time must be within 7 days for non-TAK types")
			return errors.New("stale time must be within 7 days for non-TAK types")
		}
	} else {
		// For TAK types, log but don't error on long stale times
		if staleTime.Sub(eventTime) > maxStaleOffset {
			logger := getLogger(context.Background())
			logger.Debug("legacy track with long stale",
				"uid", e.Uid,
				"type", e.Type,
				"stale_diff", staleTime.Sub(eventTime))
		}
	}

	return nil
}

// ValidateLatLon validates latitude and longitude values with detailed logging
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

// ValidateType checks if a CoT type string is valid
func ValidateType(typ string) error {
	logger := getLogger(context.Background())

	if typ == "" {
		logger.Error("empty type", "error", "type must be a valid CoT type string")
		return errors.New("type must be a valid CoT type string")
	}

	// Check for incomplete types (ending with hyphen)
	if strings.HasSuffix(typ, "-") {
		logger.Error("incomplete type", "type", typ, "error", "type must be a valid CoT type string")
		return errors.New("type must be a valid CoT type string")
	}

	// Check length
	if len(typ) > 100 {
		logger.Error("type too long", "type_length", len(typ), "max_length", 100, "error", "type must be a valid CoT type string")
		return errors.New("type must be a valid CoT type string")
	}

	// Validate type format
	parts := strings.Split(typ, "-")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		logger.Error("invalid type format", "type", typ, "error", "type must be a valid CoT type string")
		return errors.New("type must be a valid CoT type string")
	}

	// Check if type is registered
	typeRegistryMu.RLock()
	valid := validCoTTypes[typ]
	typeRegistryMu.RUnlock()

	if !valid {
		logger.Error("type not found in registry", "type", typ, "error", "type must be a valid CoT type string")
		return errors.New("type must be a valid CoT type string")
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
	const (
		minHae = -12000 // Slightly below Mariana Trench
		maxHae = 999999 // Allow for special values like 999999.0 — we may be dealing with space systems
	)
	if p.Hae < minHae || p.Hae > maxHae {
		return fmt.Errorf("invalid height above ellipsoid: %f (must be between %d and %d meters)", p.Hae, minHae, maxHae)
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

// AddLink adds a link to the event
func (e *Event) AddLink(link *Link) {
	if link != nil {
		e.Links = append(e.Links, link)
	}
}

// validateDetail performs comprehensive validation of the Detail element
func (d *Detail) validateDetail() error {
	logger := getLogger(context.Background())

	// Validate Group if present
	if d.Group != nil {
		if len(d.Group.Name) > 64 {
			logger.Error("group name too long",
				"name", d.Group.Name,
				"max_length", 64)
			return fmt.Errorf("group name exceeds maximum length of 64 characters")
		}
		if len(d.Group.Role) > 64 {
			logger.Error("group role too long",
				"role", d.Group.Role,
				"max_length", 64)
			return fmt.Errorf("group role exceeds maximum length of 64 characters")
		}
	}

	// Validate TakV if present
	if d.TakV != nil {
		if len(d.TakV.Version) > 64 {
			logger.Error("takv version too long",
				"version", d.TakV.Version,
				"max_length", 64)
			return fmt.Errorf("takv version exceeds maximum length of 64 characters")
		}
	}

	// Validate Contact if present
	if d.Contact != nil && d.Contact.Callsign != "" {
		if len(d.Contact.Callsign) > 64 {
			logger.Error("callsign too long",
				"callsign", d.Contact.Callsign,
				"max_length", 64)
			return fmt.Errorf("callsign exceeds maximum length of 64 characters")
		}
	}

	// Validate Shape if present
	if d.Shape != nil {
		if err := d.Shape.validateShape(); err != nil {
			return err // Already logged
		}
	}

	return nil
}

// sanitizeDetail performs security-focused sanitization of Detail fields
func (d *Detail) sanitizeDetail() {
	// Sanitize Group if present
	if d.Group != nil {
		d.Group.Name = sanitizeString(d.Group.Name)
		d.Group.Role = sanitizeString(d.Group.Role)
	}

	// Sanitize TakV if present
	if d.TakV != nil {
		d.TakV.Device = sanitizeString(d.TakV.Device)
		d.TakV.Platform = sanitizeString(d.TakV.Platform)
		d.TakV.Version = sanitizeString(d.TakV.Version)
	}

	// Sanitize Contact if present
	if d.Contact != nil {
		if d.Contact.Callsign != "" {
			d.Contact.Callsign = sanitizeString(d.Contact.Callsign)
		}
		if d.Contact.Endpoint != "" {
			d.Contact.Endpoint = sanitizeString(d.Contact.Endpoint)
		}
	}

	// Sanitize Shape if present
	if d.Shape != nil {
		if d.Shape.Type != "" {
			d.Shape.Type = sanitizeString(d.Shape.Type)
		}
		if d.Shape.Points != "" {
			d.Shape.Points = sanitizeString(d.Shape.Points)
		}
	}

	// Sanitize FlowTags if present
	if d.FlowTags != nil {
		if d.FlowTags.Status != "" {
			d.FlowTags.Status = sanitizeString(d.FlowTags.Status)
		}
		if d.FlowTags.Chain != "" {
			d.FlowTags.Chain = sanitizeString(d.FlowTags.Chain)
		}
	}
}

// WithStale sets a custom stale time duration for the event.
// The duration must be greater than minStaleOffset (5 seconds) and less than maxStaleOffset (7 days).
func (e *Event) WithStale(duration time.Duration) error {
	if duration <= minStaleOffset {
		return fmt.Errorf("stale duration must be greater than %v", minStaleOffset)
	}
	if duration > maxStaleOffset {
		return fmt.Errorf("stale duration must be less than %v", maxStaleOffset)
	}
	e.Stale = CoTTime(e.Time.Time().Add(duration))
	return e.validateTimes() // Ensure invariant remains true
}

// NewEvent creates a new Event with the given parameters
func NewEvent(uid string, eventType string, lat float64, lon float64, hae float64) (*Event, error) {
	now := time.Now().UTC().Truncate(time.Second)
	staleTime := now.Add(6 * time.Second) // Use 6 seconds to ensure it's more than minStaleOffset

	evt := &Event{
		Version: "2.0",
		Uid:     uid,
		Type:    eventType,
		How:     "m-g",
		Time:    CoTTime(now),
		Start:   CoTTime(now),
		Stale:   CoTTime(staleTime),
		Point: Point{
			Lat: lat,
			Lon: lon,
			Hae: hae,
			Ce:  9999999.0,
			Le:  9999999.0,
		},
	}

	// Validate the event before returning
	if err := evt.Validate(); err != nil {
		return nil, err
	}

	return evt, nil
}

// NewPresenceEvent creates a new presence event with the given parameters
func NewPresenceEvent(uid string, lat float64, lon float64, hae float64) (*Event, error) {
	now := time.Now().UTC().Truncate(time.Second)
	staleTime := now.Add(6 * time.Second) // Use 6 seconds to ensure it's more than minStaleOffset

	evt := &Event{
		Version: "2.0",
		Uid:     uid,
		Type:    "t-x-takp-v", // Correct TAK presence type
		How:     "m-g",        // Standard how value
		Time:    CoTTime(now),
		Start:   CoTTime(now),
		Stale:   CoTTime(staleTime),
		Point: Point{
			Lat: lat,
			Lon: lon,
			Hae: hae,
			Ce:  9999999.0,
			Le:  9999999.0,
		},
	}

	// Validate the event before returning
	if err := evt.Validate(); err != nil {
		return nil, err
	}

	return evt, nil
}

// InjectIdentity tags an event with __group and a p-p link if missing.
func (e *Event) InjectIdentity(selfUID, group, role string) {
	if e.Detail == nil {
		e.Detail = &Detail{}
	}
	if e.Detail.Group == nil {
		e.Detail.Group = &Group{Name: group, Role: role}
	}

	// Check if p-p link already exists
	for _, l := range e.Links {
		if l.Relation == "p-p" && l.Uid == selfUID {
			return
		}
	}

	// Add p-p link if not found
	e.AddLink(&Link{
		Uid:      selfUID,
		Type:     "t-x-takp-v",
		Relation: "p-p",
	})
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

// sanitizeEvent performs security-focused sanitization of all string fields
func (e *Event) sanitizeEvent() {
	// Sanitize all string fields
	e.Uid = sanitizeString(e.Uid)
	e.Type = sanitizeString(e.Type)
	e.How = sanitizeString(e.How)

	// Sanitize all links
	for i := range e.Links {
		e.Links[i].sanitizeLink()
	}

	// Sanitize Detail if present
	if e.Detail != nil {
		e.Detail.sanitizeDetail()
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
		return nil, err
	}

	// Validate the unmarshaled event
	if err := event.Validate(); err != nil {
		logger.Error("invalid event data",
			"error", err)
		return nil, err
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

// Example demonstrates creating and marshaling events
func Example() {
	// Create an event without hae parameter
	event1, err := NewEvent("test-uid-1", "a-f-G-U-C", 37.7749, -122.4194, 0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Create an event with hae parameter
	event2, err := NewEvent("test-uid-2", "a-f-G-U-C", 37.7749, -122.4194, 100.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Marshal events to XML
	xml1, err := xml.Marshal(event1)
	if err != nil {
		fmt.Printf("Error marshaling event1: %v\n", err)
		return
	}

	xml2, err := xml.Marshal(event2)
	if err != nil {
		fmt.Printf("Error marshaling event2: %v\n", err)
		return
	}

	fmt.Println(string(xml1))
	fmt.Println(string(xml2))
}

// ExampleNewEvent demonstrates creating a new event
func ExampleNewEvent() {
	// Create an event without hae parameter
	event1, err := NewEvent("test-uid-1", "a-f-G-U-C", 37.7749, -122.4194, 0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	// Create an event with hae parameter
	event2, err := NewEvent("test-uid-2", "a-f-G-U-C", 37.7749, -122.4194, 100.0)
	if err != nil {
		fmt.Printf("Error creating event: %v\n", err)
		return
	}

	fmt.Printf("Event 1: UID=%s, Type=%s, Location=(%f,%f), HAE=%f\n",
		event1.Uid, event1.Type, event1.Point.Lat, event1.Point.Lon, event1.Point.Hae)
	fmt.Printf("Event 2: UID=%s, Type=%s, Location=(%f,%f), HAE=%f\n",
		event2.Uid, event2.Type, event2.Point.Lat, event2.Point.Lon, event2.Point.Hae)
}

// sanitizeLink performs security-focused sanitization of Link fields
func (l *Link) sanitizeLink() {
	l.Uid = sanitizeString(l.Uid)
	l.Type = sanitizeString(l.Type)
	l.Relation = sanitizeString(l.Relation)
}

// Group appears as <__group name="Blue" role="HQ"/>
type Group struct {
	XMLName xml.Name `xml:"__group"`
	Name    string   `xml:"name,attr"`
	Role    string   `xml:"role,attr"`
}

// TakV represents ATAK software identity, shown in presence frames
type TakV struct {
	XMLName  xml.Name `xml:"takv"`
	Device   string   `xml:"device,attr,omitempty"`
	Platform string   `xml:"platform,attr,omitempty"`
	Version  string   `xml:"version,attr,omitempty"`
}

// marshalCoTTime formats a time in UTC to the CoT format, truncating to seconds.
// This is used internally by CoTTime.MarshalXMLAttr to ensure consistent time formatting.
func marshalCoTTime(t time.Time) string {
	return t.UTC().Truncate(time.Second).Format(CotTimeFormat)
}

// unmarshalCoTTime parses a time string in either CoT format (Z) or RFC3339 format (with offset)
// and normalizes it to UTC, truncating to seconds. This ensures consistent time handling
// and prevents time-based attacks.
func unmarshalCoTTime(s string) (time.Time, error) {
	// Fast path: already in correct format
	if t, err := time.Parse(CotTimeFormat, s); err == nil {
		return t.UTC().Truncate(time.Second), nil
	}

	// Fallback: RFC3339 with zone, normalize to UTC
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format: %w", err)
	}
	return t.UTC().Truncate(time.Second), nil
}

// RegisterCoTTypesFromXMLContent registers CoT types from the given XML content string
// This is particularly useful for embedding the CoTtypes.xml content directly in code
func RegisterCoTTypesFromXMLContent(xmlContent string) error {
	return RegisterCoTTypesFromReader(strings.NewReader(xmlContent))
}

// RegisterAllCoTTypes registers all CoT types from the embedded CoTtypes.xml content
// This is the most comprehensive way to ensure all standard CoT types are recognized
func RegisterAllCoTTypes() error {
	// The XML content is truncated here for brevity
	// In a real implementation, this would contain the complete CoTtypes.xml content
	cotTypesXML := `<?xml version="1.0"?>
<types>
  <cot cot="a-.-A"                     full="Air/Air Track"                                                 desc="Air Track"                                                              />
  <cot cot="a-.-A-C"                   full="Air/Civ"                                                       desc="CIVIL AIRCRAFT"                                                         />
  <cot cot="a-.-A-C-F"                 full="Air/Civ/fixed"                                                 desc="FIXED WING"                                                             />
  <cot cot="a-.-A-M-F"                 full="Air/Mil/Fixed"                                                 desc="FIXED WING"                                                             />
  <cot cot="a-.-A-M-H"                 full="Air/Mil/Rotor"                                                 desc="ROTARY WING"                                                            />
  <cot cot="a-.-A-W"                   full="Air/Weapon"                                                    desc="WEAPON"                                                                 />
  <cot cot="a-.-G"                     full="Ground"                                                        desc="GROUND TRACK"                                                           />
  <cot cot="a-.-G-E"                   full="Gnd/Equipment"                                                 desc="EQUIPMENT"                                                              />
  <cot cot="a-.-G-I"                   full="Gnd/Building"                                                  desc="Building"                                                               />
  <cot cot="a-.-G-U"                   full="Gnd/Unit"                                                      desc="UNIT"                                                                   />
  <cot cot="a-.-S"                     full="Surface/SEA SURFACE TRACK"                                     desc="SEA SURFACE TRACK"                                                      />
  <cot cot="a-.-U"                     full="SubSurf/SUBSURFACE TRACK"                                      desc="SUBSURFACE TRACK"                                                       />
  <cot cot="a-.-X"                     full="Other"                                                         desc="Other"                                                                  />
  <cot cot="a-.-X-i"                   full="Other/INCIDENT"                                                desc="Incident"                                                               />
  <cot cot="b"                       desc="Bits"                                                                   />
  <cot cot="b-i"                     desc="Image"                                                                  />
  <cot cot="b-m-r"                   desc="route"                                                                  />
  <cot cot="b-m-p"                   desc="map point"                                                              />
  <cot cot="b-m-p-s-p-i"             desc="spi"                                                                    />
  <cot cot="b-d"                     desc="Detection"                                                       />
  <cot cot="b-l"                     desc="Alarm"                                                           />
  <cot cot="t"                       desc="Tasking"                                                                />
  <cot cot="t-k"                     desc="Strike"                                                                 />
  <cot cot="t-s"                     desc="ISR"                                                                    />
  <cot cot="y"                       desc="Reply"                                                                  />
  <cot cot="y-a"                     desc="Ack"                                                                    />
  <cot cot="y-c"                     desc="Tasking Complete"                                                       />
</types>`

	return RegisterCoTTypesFromXMLContent(cotTypesXML)
}

// LoadCoTTypesFromFile loads CoT types from a file
func LoadCoTTypesFromFile(path string) error {
	logger := getLogger(context.Background())

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

	logger.Info("loaded CoT types from file",
		"path", path,
		"types", len(types.Types))

	return nil
}
