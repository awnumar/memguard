package crypto

import (
	"errors"
	"unsafe"

	"golang.org/x/crypto/nacl/secretbox"
)

// Seal takes a plaintext message and a key and returns an authenticated ciphertext.
func Seal(m []byte, key []byte) ([]byte, error) {
	// Check the length of the key is correct.
	if len(key) != 32 {
		return nil, errors.New("crypto.Seal: key must be exactly 32 bytes")
	}

	// Get a reference to the key's underlying array without making a copy.
	k := (*[32]byte)(unsafe.Pointer(&key[0]))

	// Allocate space for and generate a nonce value.
	var nonce [24]byte
	if err := MemScr(nonce[:]); err != nil {
		return nil, err
	}

	// Encrypt m and return the result.
	return secretbox.Seal(nonce[:], m, &nonce, k), nil
}

// Open takes an authenticated ciphertext and a key, and returns the plaintext.
func Open(x []byte, key []byte) ([]byte, error) {
	// Check the length of the key is correct.
	if len(key) != 32 {
		return nil, errors.New("crypto.Open: key must be exactly 32 bytes")
	}

	// Get a reference to the key's underlying array without making a copy.
	k := (*[32]byte)(unsafe.Pointer(&key[0]))

	// Retrieve and store the nonce value.
	var nonce [24]byte
	copy(nonce[:], x[:24])

	// Decrypt and return the result.
	m, ok := secretbox.Open(nil, x[24:], &nonce, k)
	if ok {
		// Decryption successful.
		return m, nil
	}

	// Decryption unsuccessful. Either the key was wrong or the authentication failed.
	return nil, errors.New("crypto.Open: decryption failed")
}
