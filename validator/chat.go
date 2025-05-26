package validator

import (
	"encoding/xml"
	"fmt"
)

// ChatSchema defines the expected attributes for the __chat extension.
type ChatSchema struct {
	XMLName xml.Name `xml:"__chat"`
	ID      string   `xml:"id,attr,omitempty"`
	Message string   `xml:"message,attr"`
	Sender  string   `xml:"sender,attr"`
}

// ValidateChat parses and validates a __chat extension.
func ValidateChat(data []byte) error {
	var c ChatSchema
	if err := xml.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	if c.Message == "" {
		return fmt.Errorf("missing message: %w", ErrInvalidChat)
	}
	if c.Sender == "" {
		return fmt.Errorf("missing sender: %w", ErrInvalidChat)
	}
	return nil
}
