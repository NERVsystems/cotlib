package cottypes

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

// Type represents a CoT type with its metadata.
type Type struct {
	Name        string // The CoT type code (e.g., "a-f-G-E-X-N")
	FullName    string // The full name (e.g., "Gnd/Equip/Nbc Equipment")
	Description string // The description (e.g., "NBC EQUIPMENT")
}

// Catalog maintains a registry of CoT types and provides lookup and search functions.
type Catalog struct {
	types  map[string]Type
	mu     sync.RWMutex
	logger *slog.Logger
}

// NewCatalog creates a new catalog instance with the given logger.
func NewCatalog(logger *slog.Logger) *Catalog {
	return &Catalog{
		types:  make(map[string]Type),
		logger: logger,
	}
}

// GetType returns the Type for the given name if it exists, or an error if not found.
func (c *Catalog) GetType(name string) (Type, error) {
	if name == "" {
		return Type{}, fmt.Errorf("empty type name")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	t, ok := c.types[name]
	if !ok {
		c.logger.Debug("Type not found", "name", name)
		return Type{}, fmt.Errorf("unknown type: %s", name)
	}

	return t, nil
}

// GetFullName returns the full name for a CoT type, or an error if not found.
func (c *Catalog) GetFullName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty type name")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	t, ok := c.types[name]
	if !ok {
		c.logger.Debug("Type not found", "name", name)
		return "", fmt.Errorf("unknown type: %s", name)
	}

	return t.FullName, nil
}

// GetDescription returns the description for a CoT type, or an error if not found.
func (c *Catalog) GetDescription(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty type name")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	t, ok := c.types[name]
	if !ok {
		c.logger.Debug("Type not found", "name", name)
		return "", fmt.Errorf("unknown type: %s", name)
	}

	return t.Description, nil
}

// GetAllTypes returns all types in the catalog.
func (c *Catalog) GetAllTypes() []Type {
	c.mu.RLock()
	defer c.mu.RUnlock()

	types := make([]Type, 0, len(c.types))
	for _, t := range c.types {
		types = append(types, t)
	}

	c.logger.Debug("Retrieved all types", "count", len(types))
	return types
}

// FindByDescription searches for types matching the given description (case-insensitive, partial match).
// If desc is empty, returns all types.
func (c *Catalog) FindByDescription(desc string) []Type {
	if desc == "" {
		return c.GetAllTypes()
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	desc = strings.ToUpper(desc)
	var matches []Type

	for _, t := range c.types {
		if strings.Contains(strings.ToUpper(t.Description), desc) {
			matches = append(matches, t)
		}
	}

	c.logger.Debug("Search by description",
		"query", desc,
		"matches", len(matches))
	return matches
}

// FindByFullName searches for types matching the given full name (case-insensitive, partial match).
// If name is empty, returns all types.
func (c *Catalog) FindByFullName(name string) []Type {
	if name == "" {
		return c.GetAllTypes()
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	name = strings.ToUpper(name)
	var matches []Type

	for _, t := range c.types {
		if strings.Contains(strings.ToUpper(t.FullName), name) {
			matches = append(matches, t)
		}
	}

	c.logger.Debug("Search by full name",
		"query", name,
		"matches", len(matches))
	return matches
}

// Upsert adds or updates a type in the catalog.
func (c *Catalog) Upsert(name string, t Type) error {
	if name == "" {
		return fmt.Errorf("empty type name")
	}
	if t.Name == "" {
		return fmt.Errorf("empty type name in Type struct")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	existing, exists := c.types[name]
	c.types[name] = t

	if exists {
		c.logger.Debug("Updated existing type",
			"name", name,
			"old_full_name", existing.FullName,
			"new_full_name", t.FullName)
	} else {
		c.logger.Debug("Added new type",
			"name", name,
			"full_name", t.FullName)
	}

	return nil
}

// Find returns all types that match the given pattern (exact or prefix match).
func (c *Catalog) Find(pattern string) []Type {
	if pattern == "" {
		return c.GetAllTypes()
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// First try exact match
	if t, ok := c.types[pattern]; ok {
		c.logger.Debug("Found exact match", "pattern", pattern)
		return []Type{t}
	}

	// Then try prefix match
	var matches []Type
	for name, t := range c.types {
		if strings.HasPrefix(name, pattern) {
			matches = append(matches, t)
		}
	}

	c.logger.Debug("Search by pattern",
		"pattern", pattern,
		"matches", len(matches))
	return matches
}
