package crypto

import (
	"crypto/rand"
	"sync"
)

var mutex = &sync.Mutex{}

// MemClr takes a buffer and wipes it with zeroes.
func MemClr(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}

// MemScr takes a buffer and overwrites it with random bytes.
func MemScr(buf []byte) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, err := rand.Read(buf); err != nil {
		return err
	}
	return nil
}
