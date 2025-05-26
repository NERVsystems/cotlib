package cotlib

import (
	"bytes"
	"encoding/xml"
	"github.com/NERVsystems/cotlib/validator"
	"io"
)

// RawMessage represents raw XML data preserved during decoding.
type RawMessage []byte

// Chat represents the TAK __chat extension with ID, Message, and Sender attributes.
type Chat struct {
	XMLName xml.Name `xml:"__chat"`
	ID      string   `xml:"id,attr,omitempty"`
	Message string   `xml:"message,attr,omitempty"`
	Sender  string   `xml:"sender,attr,omitempty"`
}

// ChatReceipt represents the TAK __chatReceipt extension acknowledging chat messages.
type ChatReceipt struct {
	XMLName xml.Name `xml:"__chatReceipt"`
	Ack     string   `xml:"ack,attr"`
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

// StrokeColor represents the TAK strokecolor extension.
type StrokeColor struct {
	Raw RawMessage
}

// StrokeWeight represents the TAK strokeweight extension.
type StrokeWeight struct {
	Raw RawMessage
}

// FillColor represents the TAK fillcolor extension.
type FillColor struct {
	Raw RawMessage
}

// LabelsOn represents the TAK labelson extension.
type LabelsOn struct {
	Raw RawMessage
}

// ColorExtension represents the TAK color extension.
type ColorExtension struct {
	Raw RawMessage
}

// UserIcon represents the TAK usericon extension.
type UserIcon struct {
	Raw RawMessage
}

// UID represents the TAK uid extension.
type UID struct {
	Raw RawMessage
}

// Emergency represents the TAK emergency extension.
type Emergency struct {
	Raw RawMessage
}

// Height represents the TAK height extension.
type Height struct {
	Raw RawMessage
}

// HeightUnit represents the TAK height_unit extension.
type HeightUnit struct {
	Raw RawMessage
}

// Hierarchy represents the TAK hierarchy extension.
type Hierarchy struct {
	Raw RawMessage
}

// DetailLink represents the TAK link detail extension.
type DetailLink struct {
	Raw RawMessage
}

// Bullseye represents the TAK bullseye extension.
type Bullseye struct {
	Raw RawMessage
}

// RouteInfo represents the TAK routeInfo extension.
type RouteInfo struct {
	Raw RawMessage
}

// Remarks represents the TAK remarks extension.
type Remarks struct {
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
	if err := validator.ValidateAgainstSchema("chat", raw); err != nil {
		return err
	}
	type alias Chat
	return xml.Unmarshal(raw, (*alias)(c))
}

func (c *ChatReceipt) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	if err := validator.ValidateAgainstSchema("chatReceipt", raw); err != nil {
		return err
	}
	type alias ChatReceipt
	return xml.Unmarshal(raw, (*alias)(c))
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

func (sc *StrokeColor) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	sc.Raw = raw
	return nil
}

func (sc StrokeColor) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, sc.Raw)
}

func (sw *StrokeWeight) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	sw.Raw = raw
	return nil
}

func (sw StrokeWeight) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, sw.Raw)
}

func (fc *FillColor) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	fc.Raw = raw
	return nil
}

func (fc FillColor) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, fc.Raw)
}

func (lo *LabelsOn) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	lo.Raw = raw
	return nil
}

func (lo LabelsOn) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, lo.Raw)
}

func (c *ColorExtension) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	c.Raw = raw
	return nil
}

func (c ColorExtension) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, c.Raw)
}

func (ui *UserIcon) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	ui.Raw = raw
	return nil
}

func (ui UserIcon) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, ui.Raw)
}

func (u *UID) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	u.Raw = raw
	return nil
}

func (u UID) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, u.Raw)
}

func (e *Emergency) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	e.Raw = raw
	return nil
}

func (e Emergency) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, e.Raw)
}

func (h *Height) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	h.Raw = raw
	return nil
}

func (h Height) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, h.Raw)
}

func (hu *HeightUnit) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	hu.Raw = raw
	return nil
}

func (hu HeightUnit) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, hu.Raw)
}

func (h *Hierarchy) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	h.Raw = raw
	return nil
}

func (h Hierarchy) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, h.Raw)
}

func (dl *DetailLink) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	dl.Raw = raw
	return nil
}

func (dl DetailLink) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, dl.Raw)
}

func (b *Bullseye) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	b.Raw = raw
	return nil
}

func (b Bullseye) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, b.Raw)
}

func (ri *RouteInfo) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	ri.Raw = raw
	return nil
}

func (ri RouteInfo) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, ri.Raw)
}

func (r *Remarks) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	r.Raw = raw
	return nil
}

func (r Remarks) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeRaw(enc, r.Raw)
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
