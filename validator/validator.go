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

//go:embed schemas/details/environment.xsd
var takDetailsEnvironmentXSD []byte

//go:embed schemas/details/fileshare.xsd
var takDetailsFileshareXSD []byte

//go:embed schemas/details/precisionlocation.xsd
var takDetailsPrecisionLocationXSD []byte

//go:embed schemas/details/takv.xsd
var takDetailsTakvXSD []byte

//go:embed schemas/details/mission.xsd
var takDetailsMissionXSD []byte

//go:embed schemas/details/shape.xsd
var takDetailsShapeXSD []byte

//go:embed schemas/details/color.xsd
var takDetailsColorXSD []byte

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

	takDetailsEnvironment, err := Compile(takDetailsEnvironmentXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details environment schema: %w", err))
	}
	schemas["tak-details-environment"] = takDetailsEnvironment

	takDetailsFileshare, err := Compile(takDetailsFileshareXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details fileshare schema: %w", err))
	}
	schemas["tak-details-fileshare"] = takDetailsFileshare

	takDetailsPrecisionLocation, err := Compile(takDetailsPrecisionLocationXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details precisionlocation schema: %w", err))
	}
	schemas["tak-details-precisionlocation"] = takDetailsPrecisionLocation

	takDetailsTakv, err := Compile(takDetailsTakvXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details takv schema: %w", err))
	}
	schemas["tak-details-takv"] = takDetailsTakv

	takDetailsMission, err := Compile(takDetailsMissionXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details mission schema: %w", err))
	}
	schemas["tak-details-mission"] = takDetailsMission

	takDetailsShape, err := Compile(takDetailsShapeXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details shape schema: %w", err))
	}
	schemas["tak-details-shape"] = takDetailsShape

	takDetailsColor, err := Compile(takDetailsColorXSD)
	if err != nil {
		panic(fmt.Errorf("compile TAK details color schema: %w", err))
	}
	schemas["tak-details-color"] = takDetailsColor
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
