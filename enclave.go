package memguard

import (
	"github.com/awnumar/memguard/core"
)

/*
Enclave is a sealed and encrypted container for sensitive data.
*/
type Enclave struct {
	raw *core.Enclave
}

/*
NewEnclave seals up some given data into an encrypted enclave object. The buffer is wiped after the data is copied.

Alternatively, a LockedBuffer may be converted into an Enclave object using the Seal method provided. This will also have the effect of destroying the LockedBuffer.
*/
func NewEnclave(buf []byte) *Enclave {
	if len(buf) == 0 {
		return nil
	}
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
	if b == nil {
		return nil
	}
	return b.Seal()
}

/*
Open decrypts an Enclave object and places its contents into a LockedBuffer. An error will be returned if decryption failed.
*/
func (e *Enclave) Open() (*LockedBuffer, error) {
	b, err := core.Open(e.raw)
	if err != nil {
		if err != core.ErrDecryptionFailed {
			core.Panic(err)
		}
		return nil, err
	}
	return &LockedBuffer{b, new(drop)}, nil
}
