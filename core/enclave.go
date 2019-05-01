package core

import (
	"errors"
	"sync"

	"github.com/awnumar/memguard/crypto"
)

var (
	// Declare a key for use in encrypting data this session.
	key *Coffer
)

func init() {
	// Initialize the key declared above with a random value.
	key = NewCoffer()
}

/*
Enclave is a sealed and encrypted container for sensitive data.
*/
type Enclave struct {
	sync.RWMutex
	ciphertext []byte
}

/*
NewEnclave is a raw constructor for the Enclave object. The given buffer is wiped after the enclave is created.
*/
func NewEnclave(buf []byte) (*Enclave, error) {
	// Return an error if length < 1.
	if len(buf) < 1 {
		return nil, ErrNullEnclave
	}

	// Create a new Enclave.
	e := new(Enclave)

	// Get a view of the key.
	k, err := key.View()
	if err != nil {
		return nil, err
	}

	// Encrypt the plaintext.
	e.ciphertext, err = crypto.Seal(buf, k.Data)
	if err != nil {
		return nil, err // key is not 32 bytes long
	}

	// Destroy our copy of the key.
	k.Destroy()

	// Wipe the given buffer.
	crypto.MemClr(buf)

	return e, nil
}

/*
Seal consumes a given Buffer object and returns its data secured and encrypted inside an Enclave. The given Buffer is destroyed after the Enclave is created.
*/
func Seal(b *Buffer) (*Enclave, error) {
	b.RLock()

	// Check if the Buffer has been destroyed.
	if !b.alive {
		return nil, ErrBufferExpired
	}

	// Construct the Enclave from the Buffer's data.
	e, err := NewEnclave(b.Data)
	if err != nil {
		return nil, err
	}

	// Destroy the Buffer object.
	b.RUnlock()
	b.Destroy()

	// Return the newly created Enclave.
	return e, nil
}

/*
Open decrypts an Enclave and puts the contents into a Buffer object. The given Enclave is left untouched and may be reused.

The Buffer object should be destroyed after the contents are no longer needed.
*/
func Open(e *Enclave) (*Buffer, error) {
	// Allocate a secure Buffer to hold the decrypted data.
	b, err := NewBuffer(len(e.ciphertext) - crypto.Overhead)
	if err != nil {
		Panic(err) // ciphertext has invalid length
	}

	// Grab a view of the key.
	k, err := key.View()
	if err != nil {
		return nil, err
	}

	// Decrypt the enclave into the buffer we created.
	_, err = crypto.Open(e.ciphertext, k.Data, b.Data)
	if err != nil {
		return nil, err
	}

	// Destroy our copy of the key.
	k.Destroy()

	// Return the contents of the Enclave inside a Buffer.
	return b, nil
}

/* Define some errors used by these functions... */

// ErrNullEnclave is returned when attempting to construct an enclave of size less than one.
var ErrNullEnclave = errors.New("<memguard::core::ErrNullEnclave> enclave size must be greater than zero")
