package memguard

import (
	"runtime"

	"github.com/awnumar/memguard/core"
)

// NewBuffer is a generic constructor for the LockedBuffer object.
func NewBuffer(size int) (*LockedBuffer, error) {
	// Construct a Buffer of the specified size.
	buf, err := core.NewBuffer(size)
	if err != nil {
		return nil, err
	}

	// Initialise a LockedBuffer object around it.
	b := &LockedBuffer{buf, new(drop)}

	// Use a finalizer to destroy the Buffer if it falls out of scope.
	runtime.SetFinalizer(b.drop, func(_ *drop) {
		go core.DestroyBuffer(buf)
	})

	// Return the created buffer to the caller.
	return b, nil
}
