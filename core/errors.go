package core

import "errors"

// ErrDestroyed is returned when a function is called on or with a destroyed container.
var ErrDestroyed = errors.New("<memguard::core::ErrDestroyed> buffer is destroyed")

// ErrInvalidLength is returned when an operation would result in a container of less than one byte.
var ErrInvalidLength = errors.New("<memguard::core::ErrInvalidLength> length of buffer must be greater than zero")
