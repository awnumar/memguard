package memguard

import (
	"os"
	"sync"
	"unsafe"
)

var (
	// Array of all active containers, and associated mutex.
	enclaves      []*container
	enclavesMutex = &sync.RWMutex{}

	// Ascertain and store the system memory page size.
	pageSize = os.Getpagesize()

	// Global reference to subclaves.
	subclaves *globalProtVals

	// Create a dedicated sync object for the CatchInterrupt function.
	catchInterruptOnce sync.Once
)

// A global struct of which there is a single instance.
// Will hold the subclaves that are needed to protect normal containers.
type globalProtVals struct {
	canary *subclave
	enckey *subclave
}

// Initialise the global subclaves.
func init() {
	// Create a new globalProtVals struct.
	gpvs := new(globalProtVals)

	// Allocate and create them.
	gpvs.canary = newSubclave()
	gpvs.enckey = newSubclave()

	// Make global.
	subclaves = gpvs
}

// Round a length to a multiple of the system page size.
func roundToPageSize(length int) int {
	return (length + (pageSize - 1)) & (^(pageSize - 1))
}

// Get a slice that describes all memory related to an Enclave.
func getAllMemory(b *container) []byte {
	// Calculate the size of the entire container's memory.
	roundedBufLen := roundToPageSize(len(b.plaintext) + 32)

	// Calculate the address of the start of the memory.
	memAddr := uintptr(unsafe.Pointer(&b.plaintext[0])) - uintptr((roundedBufLen-len(b.plaintext))+pageSize)

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
