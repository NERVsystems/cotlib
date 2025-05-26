//go:build ignore

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/NERVsystems/cotlib"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cotlib.SetLogger(logger)
	ctx := cotlib.WithLogger(context.Background(), logger)

	evt, err := cotlib.NewEvent("EXAMPLE-1", "u-d-p", 34.0, -117.0, 0)
	if err != nil {
		logger.Error("new event", "err", err)
		return
	}
	evt.StrokeColor = "#ff0000ff"
	evt.UserIcon = "http://example.com/icon.png"

	xmlData, err := evt.ToXML()
	if err != nil {
		logger.Error("marshal", "err", err)
		return
	}
	fmt.Println(string(xmlData))

	out, err := cotlib.UnmarshalXMLEvent(ctx, xmlData)
	if err != nil {
		logger.Error("unmarshal", "err", err)
		return
	}
	fmt.Printf("StrokeColor: %s\n", out.StrokeColor)
	fmt.Printf("UserIcon: %s\n", out.UserIcon)
	cotlib.ReleaseEvent(evt)
	cotlib.ReleaseEvent(out)
}
