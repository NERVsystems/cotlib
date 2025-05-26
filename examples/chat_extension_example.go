//go:build ignore

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/NERVsystems/cotlib"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cotlib.SetLogger(logger)

	xmlInput := `<?xml version="1.0" encoding="UTF-8"?>
<event version="2.0" uid="CHAT-1" type="t-x-c" time="2023-05-15T18:30:22Z" start="2023-05-15T18:30:22Z" stale="2023-05-15T18:30:32Z">
  <point lat="0" lon="0" ce="9999999.0" le="9999999.0"/>
  <detail>
    <__chat message="Hello world!" sender="Alpha"/>
    <__video url="http://example/video"/>
  </detail>
</event>`

	evt, err := cotlib.UnmarshalXMLEvent([]byte(xmlInput))
	if err != nil {
		logger.Error("failed to parse event", "error", err)
		return
	}

	if evt.Detail.Chat != nil {
		fmt.Printf("Chat from %s: %s\n", evt.Detail.Chat.Sender, evt.Detail.Chat.Message)
	}
	if evt.Detail.Video != nil {
		fmt.Printf("Video extension: %s\n", string(evt.Detail.Video.Raw))
	}

	xmlOut, err := evt.ToXML()
	if err != nil {
		logger.Error("failed to serialize", "error", err)
		return
	}

	fmt.Println(string(xmlOut))
	cotlib.ReleaseEvent(evt)
}
