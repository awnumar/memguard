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
Open decrypts an Enclave object and places its contents into a LockedBuffer.
*/
func (e *Enclave) Open() (*LockedBuffer, error) {
	b, err := core.Open(e.raw)
	if err != nil {
		return nil, err
	}
	return &LockedBuffer{b, new(drop)}, nil
}
