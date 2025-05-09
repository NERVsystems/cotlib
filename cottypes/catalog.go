package cottypes

import (
	"fmt"
	"strings"
	"sync"
)

// Type represents a CoT type with its metadata
type Type struct {
	Name        string // The CoT type code (e.g., "a-f-G-E-X-N")
	FullName    string // The full name (e.g., "Gnd/Equip/Nbc Equipment")
	Description string // The description (e.g., "NBC EQUIPMENT")
}

// Catalog maintains a registry of CoT types
type Catalog struct {
	types map[string]Type
	mu    sync.RWMutex
}

// GetType returns the Type for the given name if it exists
func (c *Catalog) GetType(name string) (Type, bool) {
	if name == "" {
		return Type{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.types[name]
	return t, ok
}

// GetFullName returns the full name for a CoT type
func (c *Catalog) GetFullName(name string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if t, ok := c.types[name]; ok {
		return t.FullName, nil
	}
	return "", fmt.Errorf("unknown type: %s", name)
}

// GetDescription returns the description for a CoT type
func (c *Catalog) GetDescription(name string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if t, ok := c.types[name]; ok {
		return t.Description, nil
	}
	return "", fmt.Errorf("unknown type: %s", name)
}

// FindByDescription searches for types matching the given description
// The search is case-insensitive and matches partial descriptions
func (c *Catalog) FindByDescription(desc string) []Type {
	c.mu.RLock()
	defer c.mu.RUnlock()

	desc = strings.ToLower(desc)
	var matches []Type

	for _, t := range c.types {
		if strings.Contains(strings.ToLower(t.Description), desc) {
			matches = append(matches, t)
		}
	}

	return matches
}

// FindByFullName searches for types matching the given full name
// The search is case-insensitive and matches partial names
func (c *Catalog) FindByFullName(name string) []Type {
	c.mu.RLock()
	defer c.mu.RUnlock()

	name = strings.ToLower(name)
	var matches []Type

	for _, t := range c.types {
		if strings.Contains(strings.ToLower(t.FullName), name) {
			matches = append(matches, t)
		}
	}

	return matches
}

// Upsert adds or updates a type in the catalog
func (c *Catalog) Upsert(name string, t Type) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.types[name] = t
}

// Find returns all types that match the given pattern
func (c *Catalog) Find(pattern string) []Type {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if pattern == "" {
		return nil
	}

	// First try exact match
	if t, ok := c.types[pattern]; ok {
		return []Type{t}
	}

	// Then try prefix match
	var matches []Type
	for name, t := range c.types {
		if strings.HasPrefix(name, pattern) {
			matches = append(matches, t)
		}
	}
	return matches
}
