//go:build cgo && !novalidator

package validator

import (
	"os"
	"sync"
)

// ResetForTest resets package state for testing.
func ResetForTest() {
	once = sync.Once{}
	schemas = nil
	initErr = nil
	mkTemp = os.MkdirTemp
	writeSchemasFn = writeSchemas
	eventPointXSD = defaultEventPointXSD
}

// SetMkTempForTest sets the MkdirTemp function for testing.
func SetMkTempForTest(f func(string, string) (string, error)) {
	mkTemp = f
}

// SetWriteSchemasForTest sets the schema writing function for testing.
func SetWriteSchemasForTest(f func(string) error) {
	writeSchemasFn = f
}

// SetEventPointXSDForTest sets the event point schema bytes for testing.
func SetEventPointXSDForTest(data []byte) {
	eventPointXSD = data
}
