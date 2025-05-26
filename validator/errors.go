package validator

import "errors"

// ErrInvalidChat indicates the chat extension does not conform to the schema.
var ErrInvalidChat = errors.New("invalid chat")
