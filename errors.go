package memguard

import "errors"

// ErrInvalidLength is returned when a LockedBuffer of smaller than one byte is requested.
var ErrInvalidLength = errors.New("memguard.Err: length of buffer must be greater than zero")

// ErrDestroyed is returned when a function is called on a destroyed LockedBuffer.
var ErrDestroyed = errors.New("memguard.Err: buffer is destroyed")
