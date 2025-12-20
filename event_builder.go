package cotlib

import "time"

// EventBuilder is a helper for constructing Event objects.
type EventBuilder struct {
	evt *Event
	err error
}

// NewEventBuilder creates a new EventBuilder with the basic event fields set.
func NewEventBuilder(uid, typ string, lat, lon, hae float64) *EventBuilder {
	now := time.Now().UTC().Truncate(time.Second)
	e := getEvent()
	*e = Event{
		Version: "2.0",
		Uid:     uid,
		Type:    typ,
		How:     "m-g",
		Time:    CoTTime(now),
		Start:   CoTTime(now),
		Stale:   CoTTime(now.Add(6 * time.Second)),
		Point: Point{
			Lat: lat,
			Lon: lon,
			Hae: hae,
			Ce:  9999999.0,
			Le:  9999999.0,
		},
	}
	return &EventBuilder{evt: e}
}

// WithContact sets the contact detail on the event.
func (b *EventBuilder) WithContact(c *Contact) *EventBuilder {
	if b.err != nil {
		return b
	}
	if b.evt.Detail == nil {
		b.evt.Detail = &Detail{}
	}
	if c != nil {
		tmp := *c
		b.evt.Detail.Contact = &tmp
	}
	return b
}

// WithGroup sets the group detail on the event.
func (b *EventBuilder) WithGroup(g *Group) *EventBuilder {
	if b.err != nil {
		return b
	}
	if b.evt.Detail == nil {
		b.evt.Detail = &Detail{}
	}
	if g != nil {
		tmp := *g
		b.evt.Detail.Group = &tmp
	}
	return b
}

// WithStaleTime sets a custom stale time for the event.
func (b *EventBuilder) WithStaleTime(t time.Time) *EventBuilder {
	if b.err != nil {
		return b
	}
	b.evt.Stale = CoTTime(t)
	return b
}

// WithHow sets the how field on the event.
// Common values: "m-g" (machine generated), "h-e" (human entry).
func (b *EventBuilder) WithHow(how string) *EventBuilder {
	if b.err != nil {
		return b
	}
	b.evt.How = how
	return b
}

// WithGroupExtension sets the TAK __group extension on the event.
func (b *EventBuilder) WithGroupExtension(g *GroupExtension) *EventBuilder {
	if b.err != nil {
		return b
	}
	if b.evt.Detail == nil {
		b.evt.Detail = &Detail{}
	}
	if g != nil {
		tmp := *g
		b.evt.Detail.GroupExtension = &tmp
	}
	return b
}

// WithTakv sets the TAK client version extension on the event.
func (b *EventBuilder) WithTakv(t *Takv) *EventBuilder {
	if b.err != nil {
		return b
	}
	if b.evt.Detail == nil {
		b.evt.Detail = &Detail{}
	}
	if t != nil {
		tmp := *t
		b.evt.Detail.Takv = &tmp
	}
	return b
}

// WithTrack sets the track extension on the event.
func (b *EventBuilder) WithTrack(t *Track) *EventBuilder {
	if b.err != nil {
		return b
	}
	if b.evt.Detail == nil {
		b.evt.Detail = &Detail{}
	}
	if t != nil {
		tmp := *t
		b.evt.Detail.Track = &tmp
	}
	return b
}

// WithUID sets the TAK uid extension on the event.
func (b *EventBuilder) WithUID(u *UID) *EventBuilder {
	if b.err != nil {
		return b
	}
	if b.evt.Detail == nil {
		b.evt.Detail = &Detail{}
	}
	if u != nil {
		tmp := *u
		b.evt.Detail.UID = &tmp
	}
	return b
}

// Build validates and returns the constructed Event.
func (b *EventBuilder) Build() (*Event, error) {
	if b.err != nil {
		ReleaseEvent(b.evt)
		return nil, b.err
	}
	if err := b.evt.ValidateAt(time.Now().UTC()); err != nil {
		ReleaseEvent(b.evt)
		return nil, err
	}
	e := b.evt
	b.evt = nil
	return e, nil
}
