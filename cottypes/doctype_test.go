package cottypes_test

import (
	"context"
	"testing"

	"github.com/NERVsystems/cotlib/cottypes"
)

func TestRegisterXMLRejectsDOCTYPE(t *testing.T) {
	data := []byte("<?xml version=\"1.0\"?><!DOCTYPE foo><types><cot cot=\"a-f-G\"/></types>")
	if err := cottypes.RegisterXML(context.Background(), data); err == nil {
		t.Fatal("expected error but got nil")
	}
}
