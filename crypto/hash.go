package crypto

import "golang.org/x/crypto/blake2b"

// Hash implements a cryptographic hash function using Blake2b.
func Hash(b []byte) []byte {
	h := blake2b.Sum256(b)
	return h[:]
}
