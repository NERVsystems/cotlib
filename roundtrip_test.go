package cotlib

import "testing"

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

	evt, err := UnmarshalXMLEvent(xmlData)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if evt.Uid != "RT1" || evt.Type != "a-f-G" {
		t.Errorf("unexpected event: %#v", evt)
	}
	ReleaseEvent(evt)
}
