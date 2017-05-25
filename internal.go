package memguard

import (
	"crypto/rand"
	"unsafe"
)

// Round a length to a multiple of the system page size.
func roundToPageSize(length int) int {
	return (length + (pageSize - 1)) & (^(pageSize - 1))
}

// Get a slice that describes all memory related to a LockedBuffer.
func getAllMemory(b *LockedBuffer) []byte {
	bufLen, roundedBufLen := len(b.Buffer), roundToPageSize(len(b.Buffer)+32)
	memAddr := uintptr(unsafe.Pointer(&b.Buffer[0])) - uintptr((roundedBufLen-bufLen)+pageSize)
	memLen := (pageSize * 2) + roundedBufLen
	return getBytes(memAddr, memLen)
}

// Convert a pointer and length to a byte slice that describes that memory.
func getBytes(ptr uintptr, len int) []byte {
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{ptr, len, len}
	return *(*[]byte)(unsafe.Pointer(&sl))
}

// Cryptographically Secure Pseudo-Random Number Generator.
func csprng(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("memguard.csprng(): could not get random bytes")
	}
	return b
}
