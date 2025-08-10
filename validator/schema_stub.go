//go:build !cgo || novalidator

package validator

import "errors"

// Schema is a stub type when CGO is disabled
type Schema struct{}

// Compile is a stub function when CGO is disabled
func Compile(data []byte) (*Schema, error) {
	return nil, errors.New("XML validation disabled: built without CGO support")
}

// CompileFile is a stub function when CGO is disabled
func CompileFile(filename string) (*Schema, error) {
	return nil, errors.New("XML validation disabled: built without CGO support")
}

// Validate is a stub method when CGO is disabled
func (s *Schema) Validate(xml []byte) error {
	return errors.New("XML validation disabled: built without CGO support")
}

// Free is a stub method when CGO is disabled
func (s *Schema) Free() {
	// No-op when CGO is disabled
}
