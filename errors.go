package memguard

import "errors"

// ErrDestroyed is returned when a function is called on a destroyed Enclave.
var ErrDestroyed = errors.New("memguard.ErrDestroyed: buffer is destroyed")

// ErrImmutable is returned when a function that needs to modify an Enclave is given an Enclave that is immutable.
var ErrImmutable = errors.New("memguard.ErrImmutable: cannot modify immutable buffer")

// ErrInvalidLength is returned when an Enclave of smaller than one byte is requested.
var ErrInvalidLength = errors.New("memguard.ErrInvalidLength: length of buffer must be greater than zero")

// ErrInvalidConversion is returned when attempting to get a slice of an Enclave that is of an inappropriate size for that slice type. For example, attempting to get a []uint16 representation of an Enclave of length 9 bytes would trigger this error, since there would be a byte leftover after the conversion.
var ErrInvalidConversion = errors.New("memguard.ErrInvalidConversion: length of buffer must align with target type")

// ErrSealed is returned when a function is given an Enclave that is sealed. Ideally the Unseal method should be called to unseal an Enclave, followed by Seal again soon after.
var ErrSealed = errors.New("memguard.ErrSealed: the given enclave is sealed")
