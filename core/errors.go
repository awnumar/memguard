package core

import "errors"

// ErrInvalidLength is returned when an operation would result in a container of less than one byte.
var ErrInvalidLength = errors.New("<memguard::core::ErrInvalidLength> length of buffer must be greater than zero")

// ErrObjectExpired is returned when attempting to use an object that has been destroyed or that relies on something that was destroyed.
var ErrObjectExpired = errors.New("<memguard::core::ErrObjectExpired> object has been destroyed and can no longer be used")
