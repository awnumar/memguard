package memguard

import "errors"

// ErrDestroyed is returned when a function is called on a destroyed Enclave.
var ErrDestroyed = errors.New("memguard.ErrDestroyed: buffer is destroyed")

// ErrImmutable is returned when a function that needs to modify a Enclave is given a Enclave that is immutable.
var ErrImmutable = errors.New("memguard.ErrImmutable: cannot modify immutable buffer")

// ErrInvalidLength is returned when a Enclave of smaller than one byte is requested.
var ErrInvalidLength = errors.New("memguard.ErrInvalidLength: length of buffer must be greater than zero")

// ErrInvalidConversion is returned when attempting to get a slice of a Enclave that is of an inappropriate size for that slice type. For example, attempting to get a []uint16 representation of a Enclave of length 9 bytes would trigger this error, since there would be a byte leftover after the conversion.
var ErrInvalidConversion = errors.New("memguard.ErrInvalidConversion: length of buffer must align with target type")
