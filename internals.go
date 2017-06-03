package memguard

import (
	"crypto/rand"
	"os"
	"sync"
	"unsafe"
)

var (
	// Once object to ensure CatchInterrupt is only executed once.
	catchInterruptOnce sync.Once

	// Store pointers to all of the LockedBuffers.
	allLockedBuffers      []*LockedBuffer
	allLockedBuffersMutex = &sync.Mutex{}

	// Mutex for getting random data from the csprng.
	csprngMutex = &sync.Mutex{}

	// Grab the system page size.
	pageSize = os.Getpagesize()
)

// Round a length to a multiple of the system page size.
func roundToPageSize(length int) int {
	return (length + (pageSize - 1)) & (^(pageSize - 1))
}

// Get a slice that describes all memory related to a LockedBuffer.
func getAllMemory(b *LockedBuffer) []byte {
	// Calculate the length of the buffer and the associated rounded value.
	bufLen, roundedBufLen := len(b.Buffer), roundToPageSize(len(b.Buffer)+32)

	// Calculate the address of the start of the memory.
	memAddr := uintptr(unsafe.Pointer(&b.Buffer[0])) - uintptr((roundedBufLen-bufLen)+pageSize)

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
	// Get a mutex lock on the csprng.
	csprngMutex.Lock()
	defer csprngMutex.Unlock()

	// Read len(b) bytes into the buffer.
	if _, err := rand.Read(b); err != nil {
		panic("memguard.csprng(): could not get random bytes")
	}
}

// Create and return a slice of length n, filled with random data.
func getRandBytes(n int) []byte {
	// Create a buffer to hold this data.
	b := make([]byte, n)

	// Read len(b) bytes into the created buffer.
	fillRandBytes(b)

	// Return the buffer.
	return b
}
