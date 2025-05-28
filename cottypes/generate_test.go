package cottypes_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}

func TestGeneratedTypesUpToDate(t *testing.T) {
	tmp := t.TempDir()

	// Setup minimal module in tmp
	if err := copyFile(filepath.Join("..", "go.mod"), filepath.Join(tmp, "go.mod")); err != nil {
		t.Fatalf("copy go.mod: %v", err)
	}

	// Copy generator source
	genDst := filepath.Join(tmp, "cmd", "cotgen")
	if err := os.MkdirAll(genDst, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := copyFile(filepath.Join("..", "cmd", "cotgen", "main.go"), filepath.Join(genDst, "main.go")); err != nil {
		t.Fatalf("copy generator: %v", err)
	}

	// Copy XML definitions
	xmlDst := filepath.Join(tmp, "cottypes")
	if err := os.MkdirAll(xmlDst, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	files, err := filepath.Glob(filepath.Join(".", "*.xml"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	for _, f := range files {
		if err := copyFile(f, filepath.Join(xmlDst, filepath.Base(f))); err != nil {
			t.Fatalf("copy xml: %v", err)
		}
	}

	// Run the generator
	cmd := exec.Command("go", "run", "./cmd/cotgen")
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generator failed: %v\n%s", err, out)
	}

	// Compare with repository version
	generated := filepath.Join(xmlDst, "generated_types.go")
	expected := filepath.Join(".", "generated_types.go")
	want, err := os.ReadFile(expected)
	if err != nil {
		t.Fatalf("read expected: %v", err)
	}
	got, err := os.ReadFile(generated)
	if err != nil {
		t.Fatalf("read generated: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("generated_types.go is out of date; run 'go generate ./cottypes'")
	}
}
