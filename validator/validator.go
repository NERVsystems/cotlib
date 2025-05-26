package validator

import (
	_ "embed"
	"fmt"
	"sync"
)

//go:embed schemas/chat.xsd
var chatXSD []byte

//go:embed schemas/chatReceipt.xsd
var chatReceiptXSD []byte

var (
	schemas map[string]*Schema
	once    sync.Once
)

func initSchemas() {
	schemas = make(map[string]*Schema)
	chat, err := Compile(chatXSD)
	if err != nil {
		panic(fmt.Errorf("compile chat schema: %w", err))
	}
	schemas["chat"] = chat

	receipt, err := Compile(chatReceiptXSD)
	if err != nil {
		panic(fmt.Errorf("compile chatReceipt schema: %w", err))
	}
	schemas["chatReceipt"] = receipt
}

// ValidateAgainstSchema validates XML against a named schema.
func ValidateAgainstSchema(name string, xml []byte) error {
	once.Do(initSchemas)
	s, ok := schemas[name]
	if !ok {
		return fmt.Errorf("unknown schema %s", name)
	}
	return s.Validate(xml)
}
