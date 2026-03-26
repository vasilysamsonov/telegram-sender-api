package message

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrExternal     = errors.New("external error")
)
