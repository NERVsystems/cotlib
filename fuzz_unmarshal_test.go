package cotlib

import (
	"context"
	"testing"
)

func FuzzUnmarshalXMLEvent(f *testing.F) {
	seed := []string{
		`<event version="2.0" uid="1" type="a-f-G" time="2020-01-01T00:00:00.000Z" start="2020-01-01T00:00:00.000Z" stale="2020-01-01T01:00:00.000Z"><point lat="0" lon="0" hae="0" ce="0" le="0"/></event>`,
	}
	for _, s := range seed {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		evt, err := UnmarshalXMLEvent(context.Background(), data)
		if err == nil {
			ReleaseEvent(evt)
		}
	})
}
