package core

import "errors"

// ErrDestroyed is returned when a function is called on or with a destroyed container.
var ErrDestroyed = errors.New("<memguard::core::ErrDestroyed> buffer is destroyed")

// ErrImmutable is returned when a function that needs to modify a LockedBuffer is given one that is immutable.
var ErrImmutable = errors.New("<memguard::core::ErrImmutable> cannot modify immutable buffer")

// ErrInvalidLength is returned when an operation would result in a container of less than one byte.
var ErrInvalidLength = errors.New("<memguard::core::ErrInvalidLength> length of buffer must be greater than zero")

// ErrInvalidConversion is returned when attempting to get a slice of a LockedBuffer that is of an inappropriate size for that slice type. For example, attempting to get a []uint16 representation of a LockedBuffer of length 9 bytes would trigger this error, since there would be a byte leftover after the conversion.
var ErrInvalidConversion = errors.New("<memguard::core::ErrInvalidConversion> length of buffer must align with target type")
