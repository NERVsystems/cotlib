package cotlib

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/NERVsystems/cotlib/validator"
)

// RawMessage represents raw XML data preserved during decoding.
type RawMessage []byte

// ChatGrp represents a chat group entry within a chat message.
type ChatGrp struct {
	XMLName xml.Name `xml:"chatgrp"`
	ID      string   `xml:"id,attr,omitempty"`
	UID0    string   `xml:"uid0,attr,omitempty"`
	UID1    string   `xml:"uid1,attr,omitempty"`
	UID2    string   `xml:"uid2,attr,omitempty"`
}

// Chat represents the TAK __chat extension including group information.
type Chat struct {
	XMLName        xml.Name   `xml:"__chat"`
	ID             string     `xml:"id,attr,omitempty"`
	Message        string     `xml:"message,attr,omitempty"`
	Sender         string     `xml:"sender,attr,omitempty"`
	Chatroom       string     `xml:"chatroom,attr,omitempty"`
	GroupOwner     string     `xml:"groupOwner,attr,omitempty"`
	SenderCallsign string     `xml:"senderCallsign,attr,omitempty"`
	Parent         string     `xml:"parent,attr,omitempty"`
	MessageID      string     `xml:"messageId,attr,omitempty"`
	DeleteChild    string     `xml:"deleteChild,attr,omitempty"`
	ChatGrps       []ChatGrp  `xml:"chatgrp,omitempty"`
	Hierarchy      *Hierarchy `xml:"hierarchy,omitempty"`
	Raw            RawMessage `xml:"-"`
}

// ChatReceipt represents the TAK chat receipt extensions.
type ChatReceipt struct {
	XMLName        xml.Name   `xml:""`
	Ack            string     `xml:"ack,attr,omitempty"`
	ID             string     `xml:"id,attr,omitempty"`
	Chatroom       string     `xml:"chatroom,attr,omitempty"`
	GroupOwner     string     `xml:"groupOwner,attr,omitempty"`
	SenderCallsign string     `xml:"senderCallsign,attr,omitempty"`
	MessageID      string     `xml:"messageId,attr,omitempty"`
	Parent         string     `xml:"parent,attr,omitempty"`
	ChatGrp        *ChatGrp   `xml:"chatgrp,omitempty"`
	Raw            RawMessage `xml:"-"`
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

// MartiDest represents a destination callsign within a Marti extension.
type MartiDest struct {
	Callsign string `xml:"callsign,attr,omitempty"`
}

// Marti represents the TAK marti extension containing destination callsigns.
type Marti struct {
	XMLName xml.Name    `xml:"marti"`
	Dest    []MartiDest `xml:"dest"`
}

// Remarks represents the TAK remarks extension.
// Remarks represents the TAK remarks extension.
// It preserves the original XML while also allowing
// convenient access to common attributes and the text
// payload.
type Remarks struct {
	XMLName  xml.Name   `xml:"remarks"`
	Source   string     `xml:"source,attr,omitempty"`
	SourceID string     `xml:"sourceID,attr,omitempty"`
	To       string     `xml:"to,attr,omitempty"`
	Time     CoTTime    `xml:"time,attr,omitempty"`
	Text     string     `xml:",chardata"`
	Raw      RawMessage `xml:"-"`
}

// Parse fills the Remarks fields from Raw if present.
// It is safe to call multiple times.
func (r *Remarks) Parse() error {
	if len(r.Raw) == 0 {
		return nil
	}
	var helper struct {
		XMLName  xml.Name `xml:"remarks"`
		Source   string   `xml:"source,attr,omitempty"`
		SourceID string   `xml:"sourceID,attr,omitempty"`
		To       string   `xml:"to,attr,omitempty"`
		Time     CoTTime  `xml:"time,attr,omitempty"`
		Text     string   `xml:",chardata"`
	}
	if err := xml.Unmarshal(r.Raw, &helper); err != nil {
		return err
	}
	r.XMLName = helper.XMLName
	r.Source = helper.Source
	r.SourceID = helper.SourceID
	r.To = helper.To
	r.Time = helper.Time
	r.Text = helper.Text
	return nil
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
	if err := validator.ValidateAgainstSchema("chat", raw); err != nil {
		if err2 := validator.ValidateAgainstSchema("tak-details-__chat", raw); err2 != nil {
			return err
		}
	} else {
		if err := validator.ValidateChat(raw); err != nil {
			return err
		}
	}

	var helper struct {
		XMLName        xml.Name   `xml:"__chat"`
		ID             string     `xml:"id,attr,omitempty"`
		Message        string     `xml:"message,attr,omitempty"`
		Sender         string     `xml:"sender,attr,omitempty"`
		Chatroom       string     `xml:"chatroom,attr,omitempty"`
		GroupOwner     string     `xml:"groupOwner,attr,omitempty"`
		SenderCallsign string     `xml:"senderCallsign,attr,omitempty"`
		Parent         string     `xml:"parent,attr,omitempty"`
		MessageID      string     `xml:"messageId,attr,omitempty"`
		DeleteChild    string     `xml:"deleteChild,attr,omitempty"`
		ChatGrps       []ChatGrp  `xml:"chatgrp"`
		Hierarchy      *Hierarchy `xml:"hierarchy"`
		_              string
	}

	if err := xml.Unmarshal(raw, &helper); err != nil {
		return err
	}

	c.XMLName = helper.XMLName
	c.ID = helper.ID
	c.Message = helper.Message
	c.Sender = helper.Sender
	c.Chatroom = helper.Chatroom
	c.GroupOwner = helper.GroupOwner
	c.SenderCallsign = helper.SenderCallsign
	c.Parent = helper.Parent
	c.MessageID = helper.MessageID
	c.DeleteChild = helper.DeleteChild
	c.ChatGrps = helper.ChatGrps
	c.Hierarchy = helper.Hierarchy
	return nil
}

func (c Chat) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if len(c.Raw) > 0 && c.Message == "" {
		return encodeRaw(enc, c.Raw)
	}
	type alias Chat
	start.Name.Local = "__chat"
	return enc.EncodeElement(alias(c), start)
}

func (c *ChatReceipt) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	raw, err := captureRaw(dec, start)
	if err != nil {
		return err
	}
	c.Raw = raw
	switch start.Name.Local {
	case "__chatReceipt":
		if err := validator.ValidateAgainstSchema("chatReceipt", raw); err != nil {
			return err
		}
		type alias ChatReceipt
		return xml.Unmarshal(raw, (*alias)(c))
	case "__chatreceipt":
		if err := validator.ValidateAgainstSchema("tak-details-__chatreceipt", raw); err != nil {
			return err
		}
		var helper struct {
			XMLName        xml.Name `xml:"__chatreceipt"`
			Ack            string   `xml:"ack,attr,omitempty"`
			ID             string   `xml:"id,attr,omitempty"`
			Chatroom       string   `xml:"chatroom,attr,omitempty"`
			GroupOwner     string   `xml:"groupOwner,attr,omitempty"`
			SenderCallsign string   `xml:"senderCallsign,attr,omitempty"`
			MessageID      string   `xml:"messageId,attr,omitempty"`
			Parent         string   `xml:"parent,attr,omitempty"`
			ChatGrp        *ChatGrp `xml:"chatgrp,omitempty"`
			_              string
		}
		if err := xml.Unmarshal(raw, &helper); err != nil {
			return err
		}
		c.XMLName = helper.XMLName
		c.Ack = helper.Ack
		c.ID = helper.ID
		c.Chatroom = helper.Chatroom
		c.GroupOwner = helper.GroupOwner
		c.SenderCallsign = helper.SenderCallsign
		c.MessageID = helper.MessageID
		c.Parent = helper.Parent
		c.ChatGrp = helper.ChatGrp
		return nil
	default:
		return fmt.Errorf("unknown element %s", start.Name.Local)
	}
}

func (c ChatReceipt) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if len(c.Raw) > 0 && c.Ack == "" && c.ID == "" && c.Chatroom == "" && c.GroupOwner == "" && c.SenderCallsign == "" && c.MessageID == "" && c.Parent == "" && c.ChatGrp == nil {
		return encodeRaw(enc, c.Raw)
	}
	if c.XMLName.Local != "" {
		start.Name = c.XMLName
	}
	type alias ChatReceipt
	return enc.EncodeElement(alias(c), start)
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
	if err := r.Parse(); err != nil {
		return err
	}
	return nil
}

func (r Remarks) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if len(r.Raw) > 0 && r.Source == "" && r.SourceID == "" &&
		r.To == "" && r.Time.Time().IsZero() && r.Text == "" {
		return encodeRaw(enc, r.Raw)
	}
	type alias Remarks
	return enc.EncodeElement(alias(r), start)
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
