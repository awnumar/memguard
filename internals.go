package memguard

import (
	"crypto/rand"
	"crypto/subtle"
	"os"
	"sync"
	"unsafe"

	"github.com/awnumar/memguard/memcall"
)

var (
	// Grab the system page size.
	pageSize = os.Getpagesize()

	// Once object to ensure CatchInterrupt is only executed once.
	catchInterruptOnce sync.Once

	// Store pointers to all of the LockedBuffers.
	allLockedBuffers      []*container
	allLockedBuffersMutex = &sync.Mutex{}
)

// container implements the actual data container.
type container struct {
	sync.Mutex // Local mutex lock.

	buffer []byte // Slice that references protected memory.

	readOnly  bool // Is this memory read-only?
	destroyed bool // Is this LockedBuffer destroyed?
}

// finaliserHint is a value that we monitor instead of the LockedBuffer
// itself. It allows us to tell the GC to auto-destroy LockedBuffers.
type finaliserHint [16]byte

// Create and allocate a canary value. Return to caller.
func createCanary() []byte {
	// Canary length rounded to page size.
	roundedLen := roundToPageSize(32)

	// Therefore the total length is...
	totalLen := (2 * pageSize) + roundedLen

	// Allocate it.
	memory := memcall.Alloc(totalLen)

	// Lock the pages that will hold the canary.
	memcall.Lock(memory[pageSize : pageSize+roundedLen])

	// Make the guard pages inaccessible.
	memcall.Protect(memory[:pageSize], false, false)
	memcall.Protect(memory[pageSize+roundedLen:], false, false)

	// Generate and copy the canary to the correct location.
	c := getRandBytes(32)
	subtle.ConstantTimeCopy(1, memory[pageSize+roundedLen-32:pageSize+roundedLen], c)

	// Mark the middle page as read-only.
	memcall.Protect(memory[pageSize:pageSize+roundedLen], true, false)

	// Return a slice that describes the correct portion of memory.
	return getBytes(uintptr(unsafe.Pointer(&memory[pageSize+roundedLen-32])), 32)
}

// Round a length to a multiple of the system page size.
func roundToPageSize(length int) int {
	return (length + (pageSize - 1)) & (^(pageSize - 1))
}

// Get a slice that describes all memory related to a LockedBuffer.
func getAllMemory(b *container) []byte {
	// Calculate the length of the buffer and the associated rounded value.
	bufLen, roundedBufLen := len(b.buffer), roundToPageSize(len(b.buffer)+32)

	// Calculate the address of the start of the memory.
	memAddr := uintptr(unsafe.Pointer(&b.buffer[0])) - uintptr((roundedBufLen-bufLen)+pageSize)

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

// Create and return a slice of length n, filled with random data.
func getRandBytes(n int) []byte {
	// Create a buffer to hold this data.
	b := make([]byte, n)

	// Read len(b) bytes into the created buffer.
	fillRandBytes(b)

	// Return the buffer.
	return b
}
