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

// Archive represents the TAK archive extension.
type Archive struct {
	Raw RawMessage
}

// AttachmentList represents the TAK attachmentList extension.
type AttachmentList struct {
	Raw RawMessage
}

// Environment represents the TAK environment extension.
type Environment struct {
	Raw RawMessage
}

// FileShare represents the TAK fileshare extension.
type FileShare struct {
	Raw RawMessage
}

// PrecisionLocation represents the TAK precisionlocation extension.
type PrecisionLocation struct {
	Raw RawMessage
}

// Takv represents the TAK takv extension.
type Takv struct {
	Raw RawMessage
}

// Track represents the TAK track extension.
type Track struct {
	Raw RawMessage
}

// Mission represents the TAK mission extension.
type Mission struct {
	Raw RawMessage
}

// Status represents the TAK status extension.
type Status struct {
	Raw RawMessage
}

// Shape represents the TAK shape extension.
type Shape struct {
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

func (a *Archive) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	a.Raw = raw
	return nil
}

func (a Archive) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, a.Raw)
}

func (a *AttachmentList) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	a.Raw = raw
	return nil
}

func (a AttachmentList) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, a.Raw)
}

func (e *Environment) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	e.Raw = raw
	return nil
}

func (e Environment) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, e.Raw)
}

func (f *FileShare) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	f.Raw = raw
	return nil
}

func (f FileShare) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, f.Raw)
}

func (p *PrecisionLocation) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	p.Raw = raw
	return nil
}

func (p PrecisionLocation) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, p.Raw)
}

func (t *Takv) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	t.Raw = raw
	return nil
}

func (t Takv) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, t.Raw)
}

func (t *Track) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	t.Raw = raw
	return nil
}

func (t Track) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, t.Raw)
}

func (m *Mission) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	m.Raw = raw
	return nil
}

func (m Mission) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, m.Raw)
}

func (s *Status) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	s.Raw = raw
	return nil
}

func (s Status) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, s.Raw)
}

func (s *Shape) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	s.Raw = raw
	return nil
}

func (s Shape) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, s.Raw)
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
