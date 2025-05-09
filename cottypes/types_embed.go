package cottypes

import (
	"bytes"
	"encoding/xml"
	"strings"
	"sync"
)

var (
	catalog     *Catalog
	catalogOnce sync.Once
)

// init initializes the CoT type catalog from the XML
func init() {
	// Initialize the catalog with the generated types
	cat := GetCatalog()
	for _, t := range expandedTypes {
		cat.types[t.Name] = Type{
			Name:        t.Name,
			FullName:    t.FullName,
			Description: t.Description,
		}
	}
}

// GetCatalog returns the singleton catalog instance
func GetCatalog() *Catalog {
	catalogOnce.Do(func() {
		catalog = &Catalog{
			types: make(map[string]Type),
		}
	})
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
		return err
	}

	cat := GetCatalog()
	for _, t := range types.Types {
		cotType := t.Cot

		// Handle wildcard expansion for affiliation
		if strings.Contains(cotType, "a-.-") {
			parts := strings.Split(cotType, "a-.-")
			if len(parts) == 2 {
				affiliations := []string{"f", "h", "n", "u"} // f=friendly, h=hostile, n=neutral, u=unknown
				for _, aff := range affiliations {
					expandedType := "a-" + aff + "-" + parts[1]
					cat.types[expandedType] = Type{
						Name:        expandedType,
						FullName:    t.Full,
						Description: t.Desc,
					}
				}
			}
		} else {
			cat.types[cotType] = Type{
				Name:        cotType,
				FullName:    t.Full,
				Description: t.Desc,
			}
		}
	}

	return nil
}
