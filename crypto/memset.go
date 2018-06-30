package crypto

import "crypto/rand"

// MemClr takes a buffer and wipes it with zeroes.
func MemClr(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}

// MemSet takes a buffer and overwrites it with a given byte.
func MemSet(buf []byte, b byte) {
	for i := range buf {
		buf[i] = b
	}
}

// MemScr takes a buffer and overwrites it with random bytes.
func MemScr(buf []byte) error {
	if _, err := rand.Read(buf); err != nil {
		return err
	}
	return nil
}
