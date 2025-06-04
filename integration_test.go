package cotlib_test

import (
	"bytes"
	"context"
	"io/fs"
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
	var files []string
	err := filepath.WalkDir("testdata/chat_samples", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(d.Name()) != ".xml" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk samples: %v", err)
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
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
