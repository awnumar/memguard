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
NewEnclave seals up the data in a given buffer into an encrypted Enclave object.

LockedBuffer objects have a Seal method which also destroy the LockedBuffers.
*/
func NewEnclave(buf []byte) (*Enclave, error) {
	e, err := core.NewEnclave(buf)
	if err != nil {
		return nil, err
	}

	return &Enclave{e}, err
}

/*
Open decrypts an Enclave object and places its contents into a LockedBuffer.
*/
func (e *Enclave) Open() (*LockedBuffer, error) {
	b, err := core.Open(e.raw)
	if err != nil {
		return nil, err
	}
	return &LockedBuffer{b, new(drop)}, nil
}
