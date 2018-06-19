package memguard

import (
	"os"
	"sync"
	"time"
	"unsafe"
)

var (
	// Ascertain and store the system memory page size.
	pageSize = os.Getpagesize()

	// Canary value that acts as an alarm in case of disallowed memory access.
	canary = createCanary()

	// Create a dedicated sync object for the CatchInterrupt function.
	catchInterruptOnce sync.Once

	// Array of all active containers, and associated mutex.
	enclaves      []*container
	enclavesMutex = &sync.Mutex{}
)

func createCanary() *subclave {
	// Create the canary.
	c := newSubclave()

	// Create a goroutine to rekey it regularly.
	go func(c *subclave) {
		for {
			// Sleep for the specified interval.
			time.Sleep(time.Duration(interval) * time.Second)

			// Rekey it.
			c.rekey()
		}
	}(c)

	// Return the canary we just created.
	return c
}

// Round a length to a multiple of the system page size.
func roundToPageSize(length int) int {
	return (length + (pageSize - 1)) & (^(pageSize - 1))
}

// Get a slice that describes all memory related to an Enclave.
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
