package validator

import (
	_ "embed"
	"fmt"
	"sync"
)

//go:embed schemas/chat.xsd
var chatXSD []byte

//go:embed schemas/chatReceipt.xsd
var chatReceiptXSD []byte

// TAKCoT detail schemas (self-contained)
//
//go:embed schemas/details/contact.xsd
var takDetailsContactXSD []byte

//go:embed schemas/details/emergency.xsd
var takDetailsEmergencyXSD []byte

//go:embed schemas/details/status.xsd
var takDetailsStatusXSD []byte

//go:embed schemas/details/track.xsd
var takDetailsTrackXSD []byte

//go:embed schemas/details/remarks.xsd
var takDetailsRemarksXSD []byte

var (
	schemas map[string]*Schema
	once    sync.Once
)

func initSchemas() {
	schemas = make(map[string]*Schema)

	// Original schemas
	chat, err := Compile(chatXSD)
	if err != nil {
		panic(fmt.Errorf("compile chat schema: %w", err))
	}
	schemas["chat"] = chat

	receipt, err := Compile(chatReceiptXSD)
	if err != nil {
		panic(fmt.Errorf("compile chatReceipt schema: %w", err))
	}
	schemas["chatReceipt"] = receipt

	// TAKCoT detail schemas
	takDetailsContact, err := Compile(takDetailsContactXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details contact schema: %w", err))
	}
	schemas["tak-details-contact"] = takDetailsContact

	takDetailsEmergency, err := Compile(takDetailsEmergencyXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details emergency schema: %w", err))
	}
	schemas["tak-details-emergency"] = takDetailsEmergency

	takDetailsStatus, err := Compile(takDetailsStatusXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details status schema: %w", err))
	}
	schemas["tak-details-status"] = takDetailsStatus

	takDetailsTrack, err := Compile(takDetailsTrackXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details track schema: %w", err))
	}
	schemas["tak-details-track"] = takDetailsTrack

	takDetailsRemarks, err := Compile(takDetailsRemarksXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details remarks schema: %w", err))
	}
	schemas["tak-details-remarks"] = takDetailsRemarks
}

// ValidateAgainstSchema validates XML against a named schema.
func ValidateAgainstSchema(name string, xml []byte) error {
	once.Do(initSchemas)
	s, ok := schemas[name]
	if !ok {
		return fmt.Errorf("unknown schema %s", name)
	}
	return s.Validate(xml)
}

// ListAvailableSchemas returns a list of all available schema names.
func ListAvailableSchemas() []string {
	once.Do(initSchemas)
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	return names
}
