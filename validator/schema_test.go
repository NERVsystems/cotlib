package validator_test

import (
	"testing"

	"github.com/NERVsystems/cotlib/validator"
)

func TestValidateAgainstSchemaNonet(t *testing.T) {
	validator.ResetForTest()
	good := []byte(`<__chat sender="Alice" message="hi"/>`)
	if err := validator.ValidateAgainstSchema("chat", good); err != nil {
		t.Fatalf("valid chat rejected: %v", err)
	}

	bad := []byte(`<!DOCTYPE foo SYSTEM "http://example.com/foo.dtd"><__chat sender="Alice" message="hi">&ext;</__chat>`)
	if err := validator.ValidateAgainstSchema("chat", bad); err == nil {
		t.Fatal("expected error for external entity")
	}
}

func TestValidateAgainstTAKDetailSchemas(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		good   []byte
		bad    []byte
	}{
		{
			name:   "contact",
			schema: "tak-details-contact",
			good:   []byte(`<contact callsign="A"/>`),
			bad:    []byte(`<contact/>`),
		},
		{
			name:   "track",
			schema: "tak-details-track",
			good:   []byte(`<track course="90" speed="10"/>`),
			bad:    []byte(`<track speed="10"/>`),
		},
		{
			name:   "status",
			schema: "tak-details-status",
			good:   []byte(`<status battery="80"/>`),
			bad:    []byte(`<status battery="bad"/>`),
		},
		{
			name:   "video",
			schema: "tak-details-__video",
			good:   []byte(`<__video url="http://x"/>`),
			bad:    []byte(`<__video/>`),
		},
		{
			name:   "fileshare",
			schema: "tak-details-fileshare",
			good:   []byte(`<fileshare filename="f" name="n" senderCallsign="A" senderUid="U" senderUrl="http://x" sha256="h" sizeInBytes="1"/>`),
			bad:    []byte(`<fileshare filename="f"/>`),
		},
		{
			name:   "attachment_list",
			schema: "tak-details-attachment_list",
			good:   []byte(`<attachment_list hashes="abc"/>`),
			bad:    []byte(`<attachment_list/>`),
		},
		{
			name:   "bullseye",
			schema: "tak-details-bullseye",
			good:   []byte(`<bullseye mils="true" distance="10" bearingRef="T" bullseyeUID="b" distanceUnits="u-r-b-bullseye" edgeToCenter="false" rangeRingVisible="true" title="t" hasRangeRings="false"/>`),
			bad:    []byte(`<bullseye/>`),
		},
		{
			name:   "routeinfo",
			schema: "tak-details-routeinfo",
			good:   []byte(`<__routeinfo><__navcues/></__routeinfo>`),
			bad:    []byte(`<__routeinfo foo="bar"/>`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.ValidateAgainstSchema(tt.schema, tt.good); err != nil {
				t.Fatalf("valid %s rejected: %v", tt.name, err)
			}
			if err := validator.ValidateAgainstSchema(tt.schema, tt.bad); err == nil {
				t.Fatalf("expected error for invalid %s", tt.name)
			}
		})
	}
}

func TestListAvailableSchemas(t *testing.T) {
	schemas := validator.ListAvailableSchemas()
	if len(schemas) == 0 {
		t.Fatal("no schemas returned")
	}
}

func TestValidateAdditionalDetailSchemas(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		good   []byte
		bad    []byte
	}{
		{
			name:   "environment",
			schema: "tak-details-environment",
			good:   []byte(`<environment temperature="20" windDirection="10" windSpeed="5"/>`),
			bad:    []byte(`<environment temperature="20" windDirection="10"/>`),
		},
		{
			name:   "mission",
			schema: "tak-details-mission",
			good:   []byte(`<mission name="op" tool="t" type="x"/>`),
			bad:    []byte(`<mission tool="t" type="x"/>`),
		},
		{
			name:   "precisionlocation",
			schema: "tak-details-precisionlocation",
			good:   []byte(`<precisionlocation altsrc="GPS"/>`),
			bad:    []byte(`<precisionlocation/>`),
		},
		{
			name:   "takv",
			schema: "tak-details-takv",
			good:   []byte(`<takv platform="Android" version="1"/>`),
			bad:    []byte(`<takv platform="Android"/>`),
		},
		{
			name:   "shape",
			schema: "tak-details-shape",
			good:   []byte(`<shape><polyline closed="true"><vertex lat="1" lon="1" hae="0"/></polyline></shape>`),
			bad:    []byte(`<shape><polyline><vertex lat="1" lon="1" hae="0"/></polyline></shape>`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.ValidateAgainstSchema(tt.schema, tt.good); err != nil {
				t.Fatalf("valid %s rejected: %v", tt.name, err)
			}
			if err := validator.ValidateAgainstSchema(tt.schema, tt.bad); err == nil {
				t.Fatalf("expected error for invalid %s", tt.name)
			}
		})
	}
}

func TestValidateDrawingShapeSchemas(t *testing.T) {
	basePoint := `<point lat="0" lon="0" hae="0" ce="0" le="0"/>`
	tests := []struct {
		name   string
		schema string
		good   string
		bad    string
	}{
		{
			name:   "circle",
			schema: "Drawing_Shapes_-_Circle",
			good: `<event version="2.0" uid="C1" type="u-d-c-c" time="2000-01-01T00:00:00Z" start="2000-01-01T00:00:00Z" stale="2000-01-02T00:00:00Z" how="m-g">` +
				basePoint + `<detail><shape><ellipse angle="0" major="1" minor="1"/><link relation="p-p" type="a-f-G" uid="X"/></shape>` +
				`<strokeColor value="1"/><strokeWeight value="1"/><fillColor value="1"/><contact callsign="A"/><remarks/>` +
				`<archive/><labels_on value="true"/><precisionlocation altsrc="GPS"/></detail></event>`,
			bad: `<event version="2.0" uid="C1" type="u-d-c-c" time="2000-01-01T00:00:00Z" start="2000-01-01T00:00:00Z" stale="2000-01-02T00:00:00Z" how="m-g">` +
				basePoint + `<detail><strokeColor value="1"/></detail></event>`,
		},
		{
			name:   "free_form",
			schema: "Drawing_Shapes_-_Free_Form",
			good: `<event version="2.0" uid="F1" type="u-d-f" time="2000-01-01T00:00:00Z" start="2000-01-01T00:00:00Z" stale="2000-01-02T00:00:00Z" how="m-g">` +
				basePoint + `<detail><link point="0,0"/><strokeColor value="1"/><strokeWeight value="1"/><fillColor value="1"/><contact callsign="A"/><remarks/>` +
				`<archive/><labels_on value="true"/><color value="1"/><precisionlocation altsrc="GPS"/></detail></event>`,
			bad: `<event version="2.0" uid="F1" type="u-d-f" time="2000-01-01T00:00:00Z" start="2000-01-01T00:00:00Z" stale="2000-01-02T00:00:00Z" how="m-g">` +
				basePoint + `<detail><strokeColor value="1"/></detail></event>`,
		},
		{
			name:   "rectangle",
			schema: "Drawing_Shapes_-_Rectangle",
			good: `<event version="2.0" uid="R1" type="u-d-r" time="2000-01-01T00:00:00Z" start="2000-01-01T00:00:00Z" stale="2000-01-02T00:00:00Z" how="m-g">` +
				basePoint + `<detail><link point="0,0"/><strokeColor value="1"/><strokeWeight value="1"/><fillColor value="1"/><contact callsign="A"/>` +
				`<tog enabled="1"/><remarks/><archive/><labels_on value="true"/><precisionlocation altsrc="GPS"/></detail></event>`,
			bad: `<event version="2.0" uid="R1" type="u-d-r" time="2000-01-01T00:00:00Z" start="2000-01-01T00:00:00Z" stale="2000-01-02T00:00:00Z" how="m-g">` +
				basePoint + `<detail><strokeColor value="1"/></detail></event>`,
		},
		{
			name:   "telestration",
			schema: "Drawing_Shapes_-_Telestration",
			good: `<event version="2.0" uid="T1" type="u-d-f-m" time="2000-01-01T00:00:00Z" start="2000-01-01T00:00:00Z" stale="2000-01-02T00:00:00Z" how="m-g">` +
				basePoint + `<detail><link line="0,0"/><strokeColor value="1"/><strokeWeight value="1"/><contact callsign="A"/><remarks/><archive/><labels_on value="true"/><color value="1"/></detail></event>`,
			bad: `<event version="2.0" uid="T1" type="u-d-f-m" time="2000-01-01T00:00:00Z" start="2000-01-01T00:00:00Z" stale="2000-01-02T00:00:00Z" how="m-g">` +
				basePoint + `<detail><strokeColor value="1"/></detail></event>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.ValidateAgainstSchema(tt.schema, []byte(tt.good)); err != nil {
				t.Fatalf("valid %s rejected: %v", tt.name, err)
			}
			if err := validator.ValidateAgainstSchema(tt.schema, []byte(tt.bad)); err == nil {
				t.Fatalf("expected error for invalid %s", tt.name)
			}
		})
	}
}

func TestValidateRemainingDetailSchemas(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		good   []byte
		bad    []byte
	}{
		{
			name:   "color",
			schema: "tak-details-color",
			good:   []byte(`<color argb="1"/>`),
			bad:    []byte(`<color/>`),
		},
		{
			name:   "fillColor",
			schema: "tak-details-fillColor",
			good:   []byte(`<fillcolor value="1"/>`),
			bad:    []byte(`<fillcolor/>`),
		},
		{
			name:   "strokeColor",
			schema: "tak-details-strokeColor",
			good:   []byte(`<strokecolor value="1"/>`),
			bad:    []byte(`<strokecolor/>`),
		},
		{
			name:   "strokeWeight",
			schema: "tak-details-strokeWeight",
			good:   []byte(`<strokeweight value="1"/>`),
			bad:    []byte(`<strokeweight/>`),
		},
		{
			name:   "height",
			schema: "tak-details-height",
			good:   []byte(`<height>10</height>`),
			bad:    []byte(`<height>abc</height>`),
		},
		{
			name:   "height_unit",
			schema: "tak-details-height_unit",
			good:   []byte(`<height_unit>1</height_unit>`),
			bad:    []byte(`<height_unit>abc</height_unit>`),
		},
		{
			name:   "uid",
			schema: "tak-details-uid",
			good:   []byte(`<uid Droid="A" nett="N"/>`),
			bad:    []byte(`<uid/>`),
		},
		{
			name:   "usericon",
			schema: "tak-details-usericon",
			good:   []byte(`<usericon iconsetpath="p"/>`),
			bad:    []byte(`<usericon/>`),
		},
		{
			name:   "chatgrp",
			schema: "tak-details-chatgrp",
			good:   []byte(`<chatgrp id="c" uid0="u"/>`),
			bad:    []byte(`<chatgrp id="c"/>`),
		},
		{
			name:   "group",
			schema: "tak-details-__group",
			good:   []byte(`<__group name="n" role="r"/>`),
			bad:    []byte(`<__group name="n"/>`),
		},
		{
			name:   "serverdestination",
			schema: "tak-details-__serverdestination",
			good:   []byte(`<__serverdestination destinations="d"/>`),
			bad:    []byte(`<__serverdestination/>`),
		},
		{
			name:   "geofence",
			schema: "tak-details-__geofence",
			good:   []byte(`<__geofence elevationMonitored="true" minElevation="1" monitor="m" trigger="t" tracking="false" maxElevation="2" boundingSphere="3"/>`),
			bad:    []byte(`<__geofence elevationMonitored="true" minElevation="1" monitor="m" trigger="t" tracking="false" maxElevation="2"/>`),
		},
		{
			name:   "link",
			schema: "tak-details-link",
			good:   []byte(`<link relation="p-p" type="a-f-G" uid="u"/>`),
			bad:    []byte(`<link relation="p-p" type="a-f-G"/>`),
		},
		{
			name:   "labels_on",
			schema: "tak-details-labels_on",
			good:   []byte(`<labelson value="true"/>`),
			bad:    []byte(`<labelson/>`),
		},
		{
			name:   "archive",
			schema: "tak-details-archive",
			good:   []byte(`<archive/>`),
			bad:    []byte(`<archive foo="bar"/>`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.ValidateAgainstSchema(tt.schema, tt.good); err != nil {
				t.Fatalf("valid %s rejected: %v", tt.name, err)
			}
			if err := validator.ValidateAgainstSchema(tt.schema, tt.bad); err == nil {
				t.Fatalf("expected error for invalid %s", tt.name)
			}
		})
	}
}
