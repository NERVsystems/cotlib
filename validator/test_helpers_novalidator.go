//go:build !cgo || novalidator

package validator

// ResetForTest is a no-op when built without CGO/with novalidator.
func ResetForTest() {}

// SetMkTempForTest is a no-op when built without CGO/with novalidator.
func SetMkTempForTest(f func(string, string) (string, error)) {}

// SetWriteSchemasForTest is a no-op when built without CGO/with novalidator.
func SetWriteSchemasForTest(f func(string) error) {}

// SetEventPointXSDForTest is a no-op when built without CGO/with novalidator.
func SetEventPointXSDForTest(data []byte) {}
