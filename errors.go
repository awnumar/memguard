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

// ErrUnsealed is returned when attempting to unseal an Enclave that is already unsealed. An error is returned in this case because the developer has assumed that a container is sealed when it isn't, and that is a dangerous assumption.
var ErrUnsealed = errors.New("memguard.ErrUnsealed: attempted to unseal container that is not sealed")
