package cottypes

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/NERVsystems/cotlib/ctxlog"
)

// Type represents a CoT type with its metadata.
type Type struct {
	Name        string // The CoT type code (e.g., "a-f-G-E-X-N")
	FullName    string // The full name (e.g., "Gnd/Equip/Nbc Equipment")
	Description string // The description (e.g., "NBC EQUIPMENT")

	fullNameUpper    string
	descriptionUpper string
}

// Catalog maintains a registry of CoT types and provides lookup and search functions.
type Catalog struct {
	types map[string]Type
	mu    sync.RWMutex
}

// NewCatalog creates a new catalog instance.
func NewCatalog() *Catalog {
	return &Catalog{
		types: make(map[string]Type),
	}
}

// GetType returns the Type for the given name if it exists, or an error if not found.
func (c *Catalog) GetType(ctx context.Context, name string) (Type, error) {
	logger := ctxlog.LoggerFromContext(ctx)
	if name == "" {
		return Type{}, fmt.Errorf("empty type name")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	t, ok := c.types[name]
	if !ok {
		logger.Debug("Type not found", "name", name)
		return Type{}, fmt.Errorf("unknown type: %s", name)
	}

	return t, nil
}

// GetFullName returns the full name for a CoT type, or an error if not found.
func (c *Catalog) GetFullName(ctx context.Context, name string) (string, error) {
	logger := ctxlog.LoggerFromContext(ctx)
	if name == "" {
		return "", fmt.Errorf("empty type name")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	t, ok := c.types[name]
	if !ok {
		logger.Debug("Type not found", "name", name)
		return "", fmt.Errorf("unknown type: %s", name)
	}

	return t.FullName, nil
}

// GetDescription returns the description for a CoT type, or an error if not found.
func (c *Catalog) GetDescription(ctx context.Context, name string) (string, error) {
	logger := ctxlog.LoggerFromContext(ctx)
	if name == "" {
		return "", fmt.Errorf("empty type name")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	t, ok := c.types[name]
	if !ok {
		logger.Debug("Type not found", "name", name)
		return "", fmt.Errorf("unknown type: %s", name)
	}

	return t.Description, nil
}

// GetAllTypes returns all types in the catalog.
func (c *Catalog) GetAllTypes(ctx context.Context) []Type {
	logger := ctxlog.LoggerFromContext(ctx)

	c.mu.RLock()
	defer c.mu.RUnlock()

	types := make([]Type, 0, len(c.types))
	for _, t := range c.types {
		types = append(types, t)
	}

	logger.Debug("Retrieved all types", "count", len(types))
	return types
}

// FindByDescription searches for types matching the given description (case-insensitive, partial match).
// If desc is empty, returns all types.
func (c *Catalog) FindByDescription(ctx context.Context, desc string) []Type {
	logger := ctxlog.LoggerFromContext(ctx)
	if desc == "" {
		return c.GetAllTypes(ctx)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	desc = strings.ToUpper(desc)
	var matches []Type

	for _, t := range c.types {
		if strings.Contains(t.descriptionUpper, desc) {
			matches = append(matches, t)
		}
	}

	logger.Debug("Search by description",
		"query", desc,
		"matches", len(matches))
	return matches
}

// FindByFullName searches for types matching the given full name (case-insensitive, partial match).
// If name is empty, returns all types.
func (c *Catalog) FindByFullName(ctx context.Context, name string) []Type {
	logger := ctxlog.LoggerFromContext(ctx)
	if name == "" {
		return c.GetAllTypes(ctx)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	name = strings.ToUpper(name)
	var matches []Type

	for _, t := range c.types {
		if strings.Contains(t.fullNameUpper, name) {
			matches = append(matches, t)
		}
	}

	logger.Debug("Search by full name",
		"query", name,
		"matches", len(matches))
	return matches
}

// Upsert adds or updates a type in the catalog.
func (c *Catalog) Upsert(ctx context.Context, name string, t Type) error {
	logger := ctxlog.LoggerFromContext(ctx)
	if name == "" {
		return fmt.Errorf("empty type name")
	}
	if t.Name == "" {
		return fmt.Errorf("empty type name in Type struct")
	}

	t.fullNameUpper = strings.ToUpper(t.FullName)
	t.descriptionUpper = strings.ToUpper(t.Description)

	c.mu.Lock()
	defer c.mu.Unlock()

	existing, exists := c.types[name]
	c.types[name] = t

	// Always log at DEBUG level (never INFO) to prevent log spam
	// when adding thousands of types. The caller should log a summary instead.
	if exists {
		logger.Debug("Updated existing type",
			"name", name,
			"old_full_name", existing.FullName,
			"new_full_name", t.FullName)
	} else {
		logger.Debug("Added new type",
			"name", name,
			"full_name", t.FullName)
	}

	return nil
}

// Find returns all types that match the given pattern (exact or prefix match).
func (c *Catalog) Find(ctx context.Context, pattern string) []Type {
	logger := ctxlog.LoggerFromContext(ctx)
	if pattern == "" {
		return c.GetAllTypes(ctx)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// First try exact match
	if t, ok := c.types[pattern]; ok {
		logger.Debug("Found exact match", "pattern", pattern)
		return []Type{t}
	}

	// Then try prefix match
	var matches []Type
	for name, t := range c.types {
		if strings.HasPrefix(name, pattern) {
			matches = append(matches, t)
		}
	}

	logger.Debug("Search by pattern",
		"pattern", pattern,
		"matches", len(matches))
	return matches
}

// IsTAK returns true if the given type belongs to the TAK namespace.
// TAK types are identified by having a FullName that starts with "TAK/".
func IsTAK(t Type) bool {
	return strings.HasPrefix(t.FullName, "TAK/")
}

// GetHowValue returns the how value for a given "what" descriptor (e.g., "gps" -> "h-g-i-g-o").
// If multiple entries exist for the same descriptor, returns the last one (TAK overrides MITRE).
func GetHowValue(what string) (string, error) {
	if what == "" {
		return "", fmt.Errorf("empty what descriptor")
	}

	var lastValue string
	var found bool

	for _, h := range hows {
		if h.What == what {
			lastValue = h.Value
			found = true
		}
	}

	if !found {
		return "", fmt.Errorf("unknown how descriptor: %s", what)
	}

	return lastValue, nil
}

// GetHowNick returns the nickname for a given how code (e.g., "h-e" -> "manual").
func GetHowNick(cot string) (string, error) {
	if cot == "" {
		return "", fmt.Errorf("empty how code")
	}

	for _, h := range hows {
		if h.Cot == cot && h.Nick != "" {
			return h.Nick, nil
		}
	}

	return "", fmt.Errorf("unknown how code: %s", cot)
}

// FindHowsByDescriptor searches for how values by descriptor (case-insensitive).
func FindHowsByDescriptor(descriptor string) []HowInfo {
	if descriptor == "" {
		return hows
	}

	descriptor = strings.ToLower(descriptor)
	var matches []HowInfo

	for _, h := range hows {
		if strings.Contains(strings.ToLower(h.What), descriptor) ||
			strings.Contains(strings.ToLower(h.Nick), descriptor) {
			matches = append(matches, h)
		}
	}

	return matches
}

// GetRelationDescription returns the description for a given relation code (e.g., "c" -> "connected").
// If multiple entries exist for the same code, returns the last one (TAK overrides MITRE).
func GetRelationDescription(cot string) (string, error) {
	if cot == "" {
		return "", fmt.Errorf("empty relation code")
	}

	var lastDesc string
	var found bool

	for _, r := range relations {
		if r.Cot == cot {
			lastDesc = r.Description
			found = true
		}
	}

	if !found {
		return "", fmt.Errorf("unknown relation code: %s", cot)
	}

	return lastDesc, nil
}

// FindRelationsByDescription searches for relation values by description (case-insensitive).
func FindRelationsByDescription(description string) []RelationInfo {
	if description == "" {
		return relations
	}

	description = strings.ToLower(description)
	var matches []RelationInfo

	for _, r := range relations {
		if strings.Contains(strings.ToLower(r.Description), description) ||
			strings.Contains(strings.ToLower(r.Nick), description) {
			matches = append(matches, r)
		}
	}

	return matches
}

// GetAllHows returns all how value mappings.
func GetAllHows() []HowInfo {
	return hows
}

// GetAllRelations returns all relation value mappings.
func GetAllRelations() []RelationInfo {
	return relations
}
