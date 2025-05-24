package cotlib

import (
	"bytes"
	"encoding/xml"
	"io"
)

// RawMessage represents raw XML data preserved during decoding.
type RawMessage []byte

// Chat represents the TAK __chat extension.
type Chat struct {
	Raw RawMessage
}

// ChatReceipt represents the TAK __chatReceipt extension.
type ChatReceipt struct {
	Raw RawMessage
}

// Geofence represents the TAK __geofence extension.
type Geofence struct {
	Raw RawMessage
}

// ServerDestination represents the TAK __serverdestination extension.
type ServerDestination struct {
	Raw RawMessage
}

// Video represents the TAK __video extension.
type Video struct {
	Raw RawMessage
}

// GroupExtension represents the TAK __group extension.
type GroupExtension struct {
	Raw RawMessage
}

// captureRaw reads an element starting from start and returns its raw XML
// representation.
func captureRaw(dec *xml.Decoder, start xml.StartElement) (RawMessage, error) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	if err := enc.EncodeToken(start); err != nil {
		return nil, err
	}
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
		if err := enc.EncodeToken(tok); err != nil {
			return nil, err
		}
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Chat) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	c.Raw = raw
	return nil
}

func (c Chat) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, c.Raw)
}

func (c *ChatReceipt) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	c.Raw = raw
	return nil
}

func (c ChatReceipt) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, c.Raw)
}

func (g *Geofence) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	g.Raw = raw
	return nil
}

func (g Geofence) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, g.Raw)
}

func (sd *ServerDestination) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	sd.Raw = raw
	return nil
}

func (sd ServerDestination) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, sd.Raw)
}

func (v *Video) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	v.Raw = raw
	return nil
}

func (v Video) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, v.Raw)
}

func (g *GroupExtension) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	g.Raw = raw
	return nil
}

func (g GroupExtension) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, g.Raw)
}

// encodeRaw writes pre-encoded XML directly to the encoder.
func encodeRaw(enc *xml.Encoder, raw RawMessage) error {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := enc.EncodeToken(tok); err != nil {
			return err
		}
	}
	return nil
}
