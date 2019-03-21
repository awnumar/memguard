package crypto

import "crypto/rand"

// MemClr takes a buffer and wipes it with zeroes.
func MemClr(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}

// MemScr takes a buffer and overwrites it with random bytes.
func MemScr(buf []byte) error {
	if _, err := rand.Read(buf); err != nil {
		return err
	}
	return nil
}
