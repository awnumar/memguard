package memguard

import (
	"crypto/rand"
	"os"
	"sync"
	"unsafe"
)

var (
	// Ascertain and store the system memory page size.
	pageSize = os.Getpagesize()

	// Canary value that acts as an alarm in case of disallowed memory access.
	canary = newSubclave()

	// Create a dedicated sync object for the CatchInterrupt function.
	catchInterruptOnce sync.Once

	// Array of all active containers, and associated mutex.
	allLockedBuffers      []*container
	allLockedBuffersMutex = &sync.Mutex{}
)

// Round a length to a multiple of the system page size.
func roundToPageSize(length int) int {
	return (length + (pageSize - 1)) & (^(pageSize - 1))
}

// Get a slice that describes all memory related to a LockedBuffer.
func getAllMemory(b *container) []byte {
	// Calculate the size of the entire container's memory.
	roundedBufLen := roundToPageSize(len(b.buffer) + 32)

	// Calculate the address of the start of the memory.
	memAddr := uintptr(unsafe.Pointer(&b.buffer[0])) - uintptr((roundedBufLen-len(b.buffer))+pageSize)

	// Calculate the size of the entire memory.
	memLen := (pageSize * 2) + roundedBufLen

	// Use this information to generate a slice and return it.
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

// Takes a byte slice and fills it with random data.
func fillRandBytes(b []byte) {
	// Read len(b) bytes into the buffer.
	if _, err := rand.Read(b); err != nil {
		panic("memguard.csprng(): could not get random bytes")
	}
}

// Wipes a byte slice with zeroes.
func wipeBytes(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}
