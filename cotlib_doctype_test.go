package cotlib

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestTypeRegistrationsRejectDOCTYPE(t *testing.T) {
	ctx := context.Background()
	xmlData := "<?xml version=\"1.0\"?><!DOCTYPE foo><types><cot cot=\"a-f-G\"/></types>"

	t.Run("RegisterCoTTypesFromXMLContent", func(t *testing.T) {
		if err := RegisterCoTTypesFromXMLContent(ctx, xmlData); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("expected error, got %v", err)
		}
	})

	t.Run("RegisterCoTTypesFromReader", func(t *testing.T) {
		r := strings.NewReader(xmlData)
		if err := RegisterCoTTypesFromReader(ctx, r); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("expected error, got %v", err)
		}
	})

	t.Run("RegisterCoTTypesFromFile", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "types-*.xml")
		if err != nil {
			t.Fatalf("temp file: %v", err)
		}
		defer f.Close()
		if _, err := f.WriteString(xmlData); err != nil {
			t.Fatalf("write: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
		if err := RegisterCoTTypesFromFile(ctx, f.Name()); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("expected error, got %v", err)
		}
	})

	t.Run("LoadCoTTypesFromFile", func(t *testing.T) {
		// structure for LoadCoTTypesFromFile
		xmlData := "<?xml version=\"1.0\"?><!DOCTYPE foo><types><type>a-f-G</type></types>"
		f, err := os.CreateTemp(t.TempDir(), "load-*.xml")
		if err != nil {
			t.Fatalf("temp file: %v", err)
		}
		defer f.Close()
		if _, err := f.WriteString(xmlData); err != nil {
			t.Fatalf("write: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
		if err := LoadCoTTypesFromFile(ctx, f.Name()); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("expected error, got %v", err)
		}
	})
}
