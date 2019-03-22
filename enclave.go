package memguard

import (
	"github.com/awnumar/memguard/core"
	"github.com/awnumar/memguard/crypto"
)

/*
Enclave is a sealed and encrypted container for sensitive data.
*/
type Enclave struct {
	raw *core.Enclave
}

/*
NewEnclave seals up the data in a given buffer into an encrypted enclave object.

LockedBuffer objects have a Seal method which also destroy the LockedBuffers.
*/
func NewEnclave(buf []byte) *Enclave {
	if len(buf) < 1 {
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
	if err != crypto.ErrDecryptionFailed && err != nil {
		core.Panic(err)
	}
	if err != nil {
		return nil, err
	}
	return &LockedBuffer{b, new(drop)}, nil
}
