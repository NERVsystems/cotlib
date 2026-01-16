//go:build ignore

package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"

	"github.com/NERVsystems/cotlib"
	"github.com/NERVsystems/cotlib/validator"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cotlib.SetLogger(logger)
	ctx := cotlib.WithLogger(context.Background(), logger)

	xmlInput := `<?xml version="1.0" encoding="UTF-8"?>
<event version="2.0" uid="CHAT-2" type="b-t-f" time="2020-01-01T00:00:00Z" start="2020-01-01T00:00:00Z" stale="2020-01-01T01:00:00Z">
  <point lat="10" lon="20" ce="1" le="1"/>
  <detail>
    <__chat chatroom="room" groupOwner="false" id="1" senderCallsign="Bravo">
      <chatgrp id="room" uid0="Bravo"/>
    </__chat>
  </detail>
</event>`

	evt, err := cotlib.UnmarshalXMLEvent(ctx, []byte(xmlInput))
	if err != nil {
		logger.Error("failed to parse event", "err", err)
		return
	}

	c := evt.Detail.Chat
	fmt.Printf("Chatroom: %s\n", c.Chatroom)
	fmt.Printf("GroupOwner: %s\n", c.GroupOwner)
	fmt.Printf("SenderCallsign: %s\n", c.SenderCallsign)

	if err := evt.Validate(); err != nil {
		logger.Error("Event validation failed", "err", err)
		return
	}
	fmt.Println("Event validated via Event.Validate")

	data, err := xml.Marshal(c)
	if err != nil {
		logger.Error("marshal chat", "err", err)
		return
	}
	if err := validator.ValidateAgainstSchema("tak-details-__chat", data); err != nil {
		logger.Error("schema validation failed", "err", err)
		return
	}
	fmt.Println("Chat validated with tak-details-__chat schema")

	cotlib.ReleaseEvent(evt)
}
