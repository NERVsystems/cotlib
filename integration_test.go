package cotlib_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/NERVsystems/cotlib"
)

var (
	timeRe  = regexp.MustCompile(`time="[^"]*"`)
	startRe = regexp.MustCompile(`start="[^"]*"`)
	staleRe = regexp.MustCompile(`stale="[^"]*"`)
)

func updateTimes(data []byte) []byte {
	now := time.Now().UTC().Truncate(time.Second)
	start := now.Add(-2 * time.Hour)
	stale := now.Add(2 * time.Hour)
	data = timeRe.ReplaceAll(data, []byte("time=\""+now.Format(cotlib.CotTimeFormat)+"\""))
	data = startRe.ReplaceAll(data, []byte("start=\""+start.Format(cotlib.CotTimeFormat)+"\""))
	data = staleRe.ReplaceAll(data, []byte("stale=\""+stale.Format(cotlib.CotTimeFormat)+"\""))
	return data
}

func TestChatSamples(t *testing.T) {
	entries, err := os.ReadDir("testdata/chat_samples")
	if err != nil {
		t.Fatalf("read samples: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".xml" {
			continue
		}
		t.Run(e.Name(), func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata/chat_samples", e.Name()))
			if err != nil {
				t.Fatalf("read file: %v", err)
			}
			raw = updateTimes(raw)
			evt, err := cotlib.UnmarshalXMLEvent(context.Background(), raw)
			if err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			defer cotlib.ReleaseEvent(evt)
			if evt.Detail == nil {
				t.Fatalf("missing detail")
			}
			found := false
			for _, u := range evt.Detail.Unknown {
				if bytes.Contains(u, []byte("_flow-tags_")) {
					found = true
					break
				}
			}
			if !found {
				t.Error("_flow-tags_ element not captured in Unknown")
			}
		})
	}
}
