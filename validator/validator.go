package validator

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed schemas/**
var schemasFS embed.FS

var (
	schemas map[string]*Schema
	once    sync.Once
	initErr error
)

// test hooks
var (
	mkTemp         = os.MkdirTemp
	writeSchemasFn = writeSchemas
)

func writeSchemas(dir string) error {
	return fs.WalkDir(schemasFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == "." {
				return nil
			}
			return os.MkdirAll(filepath.Join(dir, path), 0o755)
		}
		data, err := schemasFS.ReadFile(path)
		if err != nil {
			return err
		}
		dest := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
}

func initSchemas() {
	schemas = make(map[string]*Schema)

	tmpDir, err := mkTemp("", "cotlib-schemas")
	if err != nil {
		initErr = fmt.Errorf("create temp dir: %w", err)
		return
	}
	defer os.RemoveAll(tmpDir)
	if err := writeSchemasFn(tmpDir); err != nil {
		initErr = fmt.Errorf("write schemas: %w", err)
		return
	}

	err = fs.WalkDir(schemasFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".xsd" {
			return nil
		}
		sc, err := CompileFile(filepath.Join(tmpDir, path))
		if err != nil {
			// Skip schemas that fail to compile
			return nil
		}
		rel := strings.TrimPrefix(path, "schemas/")
		name := strings.TrimSuffix(rel, ".xsd")
		name = strings.ReplaceAll(name, string(os.PathSeparator), "-")
		name = strings.ReplaceAll(name, " ", "_")
		schemas[name] = sc
		if strings.HasPrefix(name, "details-") {
			schemas["tak-"+name] = sc
		}
		return nil
	})
	if err != nil {
		initErr = err
	}
}

// ValidateAgainstSchema validates XML against a named schema.
func ValidateAgainstSchema(name string, xml []byte) error {
	once.Do(initSchemas)
	if initErr != nil {
		return initErr
	}
	s, ok := schemas[name]
	if !ok {
		return fmt.Errorf("unknown schema %s", name)
	}
	return s.Validate(xml)
}

// ListAvailableSchemas returns a list of all available schema names.
func ListAvailableSchemas() []string {
	once.Do(initSchemas)
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	return names
}
