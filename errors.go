package memguard

import "errors"

// ErrDestroyed is returned when a function is called on a Destroyed LockedBuffer.
var ErrDestroyed = errors.New("memguard.Err: buffer is destroyed")
