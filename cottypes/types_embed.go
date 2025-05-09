package cottypes

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

var (
	catalog     *Catalog
	catalogOnce sync.Once
	logger      = slog.Default()
)

// SetLogger sets the logger for the catalog package
func SetLogger(l *slog.Logger) {
	logger = l
}

// GetCatalog returns the singleton catalog instance
func GetCatalog() *Catalog {
	var initErr error
	catalogOnce.Do(func() {
		catalog = NewCatalog(logger)

		// Validate expanded types
		if len(expandedTypes) == 0 {
			initErr = fmt.Errorf("no types found in expandedTypes")
			logger.Error("Catalog initialization failed", "error", initErr)
			return
		}

		// Initialize the catalog with the generated types
		for _, t := range expandedTypes {
			if t.Name == "" {
				initErr = fmt.Errorf("invalid type found: empty name")
				logger.Error("Invalid type in expandedTypes", "error", initErr)
				return
			}

			if err := catalog.Upsert(t.Name, Type{
				Name:        t.Name,
				FullName:    t.FullName,
				Description: t.Description,
			}); err != nil {
				initErr = fmt.Errorf("failed to upsert type %s: %w", t.Name, err)
				logger.Error("Failed to upsert type",
					"error", initErr,
					"type", t.Name)
				return
			}
		}

		logger.Info("Catalog initialized successfully",
			"type_count", len(catalog.types),
			"first_type", expandedTypes[0].Name,
			"last_type", expandedTypes[len(expandedTypes)-1].Name)

		// Verify critical types exist
		criticalTypes := []string{"a-f-G-E-X-N", "a-h-G-E-X-N", "a-n-G-E-X-N", "a-u-G-E-X-N"}
		for _, typ := range criticalTypes {
			if _, err := catalog.GetType(typ); err != nil {
				logger.Warn("Critical type not found in catalog",
					"type", typ,
					"error", err)
			}
		}
	})

	if initErr != nil {
		logger.Error("Catalog initialization failed", "error", initErr)
		return nil
	}

	return catalog
}

// RegisterXML registers CoT types from XML content
func RegisterXML(data []byte) error {
	var types struct {
		Types []struct {
			Cot  string `xml:"cot,attr"`
			Full string `xml:"full,attr,omitempty"`
			Desc string `xml:"desc,attr,omitempty"`
		} `xml:"cot"`
	}

	if err := xml.NewDecoder(bytes.NewReader(data)).Decode(&types); err != nil {
		return fmt.Errorf("failed to decode XML: %w", err)
	}

	cat := GetCatalog()
	if cat == nil {
		return fmt.Errorf("catalog not initialized")
	}

	for _, t := range types.Types {
		cotType := t.Cot

		// Handle wildcard expansion for affiliation
		if strings.Contains(cotType, "a-.-") {
			parts := strings.Split(cotType, "a-.-")
			if len(parts) == 2 {
				affiliations := []string{"f", "h", "n", "u"} // f=friendly, h=hostile, n=neutral, u=unknown
				for _, aff := range affiliations {
					expandedType := "a-" + aff + "-" + parts[1]
					if err := cat.Upsert(expandedType, Type{
						Name:        expandedType,
						FullName:    t.Full,
						Description: t.Desc,
					}); err != nil {
						logger.Error("Failed to register expanded type",
							"error", err,
							"type", expandedType)
					}
				}
			}
		} else {
			if err := cat.Upsert(cotType, Type{
				Name:        cotType,
				FullName:    t.Full,
				Description: t.Desc,
			}); err != nil {
				logger.Error("Failed to register type",
					"error", err,
					"type", cotType)
			}
		}
	}

	return nil
}
