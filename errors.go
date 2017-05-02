package memguard

import "errors"

// ErrZeroLength is returned when a LockedBuffer of smaller than one bytes is requested.
var ErrZeroLength = errors.New("memguard.Err: length of buffer must be non-zero")

// ErrDestroyed is returned when a function is called on a Destroyed LockedBuffer.
var ErrDestroyed = errors.New("memguard.Err: buffer is destroyed")
