package core

import (
	"errors"
)

var (
	// Declare a key for use in encrypting data this session.
	key *Coffer
)

func init() {
	// Initialize the key declared above with a random value
	key = NewCoffer()
}

/*
Enclave is a sealed and encrypted container for sensitive data.
*/
type Enclave struct {
	ciphertext []byte
}

/*
NewEnclave is a raw constructor for the Enclave object. The given buffer is wiped after the enclave is created.
*/
func NewEnclave(buf []byte) (Enclave, error) {
	if len(buf) < 1 {
		return Enclave{}, ErrNullEnclave
	}

	e := Enclave{}

	k, err := key.View()
	if err != nil {
		return Enclave{}, err
	}

	e.ciphertext, err = Encrypt(buf, k.Data())
	if err != nil {
		Panic(err) // key is not 32 bytes long
	}

	k.Destroy()

	Wipe(buf)

	return e, nil
}

/*
Seal consumes a given Buffer object and returns its data secured and encrypted inside an Enclave. The given Buffer is destroyed after the Enclave is created.
*/
func Seal(b *Buffer) (Enclave, error) {
	b.Lock()
	defer b.Unlock()

	if !b.Alive() {
		return Enclave{}, ErrBufferExpired
	}

	if err := b.melt(); err != nil {
		Panic(err)
	}

	e, err := NewEnclave(b.Data())
	if err != nil {
		return Enclave{}, err
	}

	if err := b.destroy(); err != nil {
		Panic(err)
	}

	return e, nil
}

/*
Open decrypts an Enclave and puts the contents into a Buffer object. The given Enclave is left untouched and may be reused.

The Buffer object should be destroyed after the contents are no longer needed.
*/
func (e Enclave) Open() (*Buffer, error) {
	b, err := NewBuffer(e.Size())
	if err != nil {
		Panic("<memguard:core> ciphertext has invalid length") // ciphertext has invalid length
	}

	k, err := key.View()
	if err != nil {
		return nil, err
	}

	_, err = Decrypt(e.ciphertext, k.Data(), b.Data())
	if err != nil {
		return nil, err
	}

	k.Destroy()

	return b, nil
}

/*
Ciphertext returns the encrypted data in byte form.
*/
func (e Enclave) Ciphertext() []byte {
	return e.ciphertext
}

/*
Size returns the number of bytes of plaintext data stored inside an Enclave.
*/
func (e Enclave) Size() int {
	return len(e.ciphertext) - Overhead
}

// ErrNullEnclave is returned when attempting to construct an enclave of size less than one.
var ErrNullEnclave = errors.New("<memguard::core::ErrNullEnclave> enclave size must be greater than zero")
