package memguard

import "errors"

// ErrInvalidLength is returned when a LockedBuffer of smaller than one byte is requested.
var ErrInvalidLength = errors.New("memguard.Err: length of buffer must be greater than zero")

// ErrReadOnly is returned when a function that needs to modify a LockedBuffer
// is given a LockedBuffer that is marked as being read-only.
var ErrReadOnly = errors.New("memguard.Err: buffer is marked read-only")

// ErrDestroyed is returned when a function is called on a destroyed LockedBuffer.
var ErrDestroyed = errors.New("memguard.Err: buffer is destroyed")
