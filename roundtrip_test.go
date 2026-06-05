package cotlib

import (
	"context"
	"strings"
	"testing"
)

// TestContactEndpointRoundTrip guards INFR-207: ToXML() was a hand-written
// serializer that emitted <contact callsign> but dropped the endpoint attribute,
// even though Contact.Endpoint is set. The TAK server keys the callsign->
// connection mapping off that endpoint, so dropping it silently broke inbound
// direct-message routing to a self-SA presence (INFR-206). The endpoint must
// both appear in the serialized XML and survive a parse round-trip.
func TestContactEndpointRoundTrip(t *testing.T) {
	orig, err := NewEvent("EP1", "a-f-G-U-C", 1.0, 2.0, 0)
	if err != nil {
		t.Fatalf("new event: %v", err)
	}
	orig.Detail = &Detail{
		Contact: &Contact{Callsign: "NERVA", Endpoint: "*:-1:stcp"},
	}
	xmlData, err := orig.ToXML()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	ReleaseEvent(orig)

	if !strings.Contains(string(xmlData), `endpoint="*:-1:stcp"`) {
		t.Fatalf("ToXML dropped the contact endpoint attribute: %s", xmlData)
	}

	evt, err := UnmarshalXMLEvent(context.Background(), xmlData)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if evt.Detail == nil || evt.Detail.Contact == nil {
		t.Fatalf("no contact after round-trip: %#v", evt)
	}
	if got := evt.Detail.Contact.Endpoint; got != "*:-1:stcp" {
		t.Errorf("endpoint not preserved through round-trip: got %q want %q", got, "*:-1:stcp")
	}
	if got := evt.Detail.Contact.Callsign; got != "NERVA" {
		t.Errorf("callsign not preserved through round-trip: got %q want %q", got, "NERVA")
	}
	ReleaseEvent(evt)
}

func TestUnmarshalXMLEventRoundTrip(t *testing.T) {
	orig, err := NewEvent("RT1", "a-f-G", 10.0, 20.0, 0)
	if err != nil {
		t.Fatalf("new event: %v", err)
	}
	xmlData, err := orig.ToXML()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	ReleaseEvent(orig)

	evt, err := UnmarshalXMLEvent(context.Background(), xmlData)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if evt.Uid != "RT1" || evt.Type != "a-f-G" {
		t.Errorf("unexpected event: %#v", evt)
	}
	ReleaseEvent(evt)
}
