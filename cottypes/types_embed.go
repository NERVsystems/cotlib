package cottypes

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"github.com/NERVsystems/cotlib/ctxlog"
)

var (
	catalog     *Catalog
	catalogOnce sync.Once
	logger      = slog.Default()
)

var doctypePattern = regexp.MustCompile(`(?i)<!\s*DOCTYPE`)

// limitTokenReader enforces basic XML token limits during decoding.
type limitTokenReader struct {
	dec   *xml.Decoder
	depth int
	count int
}

func (l *limitTokenReader) Token() (xml.Token, error) {
	off := l.dec.InputOffset()
	tok, err := l.dec.RawToken()
	if err != nil {
		return tok, err
	}
	if l.dec.InputOffset()-off > 4096 {
		return nil, fmt.Errorf("invalid input")
	}
	switch t := tok.(type) {
	case xml.StartElement:
		l.depth++
		l.count++
		if l.depth > 32 || l.count > 10000 {
			return nil, fmt.Errorf("invalid input")
		}
		for _, a := range t.Attr {
			if len(a.Value) > 512*1024 {
				return nil, fmt.Errorf("invalid input")
			}
		}
	case xml.EndElement:
		if l.depth > 0 {
			l.depth--
		}
	case xml.CharData:
		if len(t) > 512*1024 {
			return nil, fmt.Errorf("invalid input")
		}
	}
	return tok, nil
}

func decodeWithLimits(dec *xml.Decoder, v any) error {
	ltd := &limitTokenReader{dec: dec}
	return xml.NewTokenDecoder(ltd).Decode(v)
}

// SetLogger sets the logger for the catalog package.
func SetLogger(l *slog.Logger) {
	logger = l
}

// GetCatalog returns the singleton catalog instance, initializing it if necessary.
func GetCatalog() *Catalog {
	var initErr error
	catalogOnce.Do(func() {
		catalog = NewCatalog()
		ctx := ctxlog.WithLogger(context.Background(), logger)

		// Validate expanded types
		if len(expandedTypes) == 0 {
			initErr = fmt.Errorf("no types found in expandedTypes")
			logger.Error("Catalog initialization failed", "error", initErr)
			return
		}

		// Initialize the catalog with the generated types
		successCount := 0
		var failedTypes []string

		for _, t := range expandedTypes {
			if t.Name == "" {
				initErr = fmt.Errorf("invalid type found: empty name")
				logger.Error("Invalid type in expandedTypes", "error", initErr)
				return
			}

			if err := catalog.Upsert(ctx, t.Name, Type{
				Name:        t.Name,
				FullName:    t.FullName,
				Description: t.Description,
			}); err != nil {
				failedTypes = append(failedTypes, t.Name)
				logger.Error("Failed to upsert type",
					"error", err,
					"type", t.Name)
			} else {
				successCount++
			}
		}

		// Log summary instead of individual successes
		logger.Debug("Catalog initialized successfully",
			"total_types", len(expandedTypes),
			"loaded_types", successCount,
			"failed_types", len(failedTypes))

		if len(failedTypes) > 0 && len(failedTypes) <= 10 {
			logger.Warn("Some types failed to load", "failed_types", failedTypes)
		} else if len(failedTypes) > 10 {
			logger.Warn("Multiple types failed to load",
				"failed_count", len(failedTypes),
				"sample", failedTypes[:10])
		}

		// Verify critical types exist
		criticalTypes := []string{"a-f-G-E-X-N", "a-h-G-E-X-N", "a-n-G-E-X-N", "a-u-G-E-X-N"}
		var missingCritical []string
		for _, typ := range criticalTypes {
			if _, err := catalog.GetType(ctx, typ); err != nil {
				missingCritical = append(missingCritical, typ)
			}
		}

		if len(missingCritical) > 0 {
			logger.Warn("Critical types not found in catalog", "types", missingCritical)
		}
	})

	if initErr != nil {
		logger.Error("Catalog initialization failed", "error", initErr)
		return nil
	}

	return catalog
}

// RegisterXML registers CoT types from XML content into the catalog.
func RegisterXML(ctx context.Context, data []byte) error {
	logger := ctxlog.LoggerFromContext(ctx)

	if doctypePattern.Match(data) {
		logger.Error("invalid doctype detected")
		return fmt.Errorf("invalid input")
	}

	var types struct {
		Types []struct {
			Cot  string `xml:"cot,attr"`
			Full string `xml:"full,attr,omitempty"`
			Desc string `xml:"desc,attr,omitempty"`
		} `xml:"cot"`
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.CharsetReader = nil
	dec.Entity = nil
	if err := decodeWithLimits(dec, &types); err != nil {
		return fmt.Errorf("failed to decode XML: %w", err)
	}

	cat := GetCatalog()
	if cat == nil {
		return fmt.Errorf("catalog not initialized")
	}

	successCount := 0
	failedCount := 0
	var failedTypes []string

	for _, t := range types.Types {
		cotType := t.Cot

		// Handle wildcard expansion for affiliation
		if strings.Contains(cotType, "a-.-") {
			parts := strings.Split(cotType, "a-.-")
			if len(parts) == 2 {
				expandedCount := 0
				affiliations := []string{"f", "h", "n", "u"} // f=friendly, h=hostile, n=neutral, u=unknown
				for _, aff := range affiliations {
					expandedType := "a-" + aff + "-" + parts[1]
					if err := cat.Upsert(ctx, expandedType, Type{
						Name:        expandedType,
						FullName:    t.Full,
						Description: t.Desc,
					}); err != nil {
						failedTypes = append(failedTypes, expandedType)
						failedCount++
						logger.Error("Failed to register expanded type",
							"error", err,
							"type", expandedType)
					} else {
						expandedCount++
					}
				}
				successCount += expandedCount
			}
		} else {
			if err := cat.Upsert(ctx, cotType, Type{
				Name:        cotType,
				FullName:    t.Full,
				Description: t.Desc,
			}); err != nil {
				failedTypes = append(failedTypes, cotType)
				failedCount++
				logger.Error("Failed to register type",
					"error", err,
					"type", cotType)
			} else {
				successCount++
			}
		}
	}

	// Log summary of registration
	logger.Debug("XML types registration complete",
		"total_processed", len(types.Types),
		"success_count", successCount,
		"failed_count", failedCount)

	if failedCount > 0 && failedCount <= 10 {
		logger.Warn("Some XML types failed to register", "failed_types", failedTypes)
	} else if failedCount > 10 {
		logger.Warn("Multiple XML types failed to register",
			"failed_count", failedCount,
			"sample", failedTypes[:10])
	}

	return nil
}
