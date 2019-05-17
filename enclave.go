package memguard

import (
	"github.com/awnumar/memguard/core"
)

/*
Enclave is a sealed and encrypted container for sensitive data.
*/
type Enclave struct {
	*core.Enclave
}

/*
NewEnclave seals up some given data into an encrypted enclave object. The buffer is wiped after the data is copied. The length of the buffer must be strictly positive or else the function will panic.

Alternatively, a LockedBuffer may be converted into an Enclave object using the Seal method provided. This will also have the effect of destroying the LockedBuffer.
*/
func NewEnclave(buf []byte) *Enclave {
	e, err := core.NewEnclave(buf)
	if err != nil {
		core.Panic(err)
	}
	return &Enclave{e}
}

/*
NewEnclaveRandom generates and seals arbitrary amounts of cryptographically-secure random bytes into an encrypted enclave object.
*/
func NewEnclaveRandom(size int) *Enclave {
	// todo: stream data into enclave
	b := NewBufferRandom(size)
	return b.Seal()
}

/*
Open decrypts an Enclave object and places its contents into an immutable LockedBuffer. An error will be returned if decryption failed.
*/
func (e *Enclave) Open() (*LockedBuffer, error) {
	b, err := core.Open(e.Enclave)
	if err != nil {
		if err != core.ErrDecryptionFailed {
			core.Panic(err)
		}
		return nil, err
	}
	b.Freeze()
	return &LockedBuffer{b, new(drop)}, nil
}
