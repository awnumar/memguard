// +build !windows

package memcall

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// Init initialises the environment. It must be called before anything esle.
func Init() {
	// Disable core dumps.
	if err := unix.Setrlimit(unix.RLIMIT_CORE, &unix.Rlimit{Cur: 0, Max: 0}); err != nil {
		panic(fmt.Sprintf("memguard.memprot.Init(): could not set rlimit [Err: %s]", err))
	}
}

// Lock is a wrapper for unix.Mlock(), with extra precautions.
func Lock(b []byte) {
	if err := unix.Mlock(b); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Lock(): could not aquire lock on %p [Err: %s]", &b[0], err))
	}
}

// Unlock is a wrapper for unix.Unlock.
func Unlock(b []byte) {
	if err := unix.Munlock(b); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Unlock(): could not free lock on %p [Err: %s]", &b[0], err))
	}
}

// Alloc allocates a byte slice of length n and returns it.
func Alloc(n int) []byte {
	// Allocate the memory.
	b, err := unix.Mmap(-1, 0, n, unix.PROT_NONE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		panic(fmt.Sprintf("memguard.memcall.Alloc(): could not allocate [Err: %s]", err))
	}

	// Advise the kernel not to dump. Ignore failure.
	unix.Madvise(b, 0x10)

	// Return the allocated memory.
	return b
}

// Free wipes and unallocates the byte slice specified.
func Free(b []byte) {
	// Wipe it first.
	for i := 0; i < len(b); i++ {
		b[i] = byte(0)
	}

	// Unallocate it.
	if err := unix.Munmap(b); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Free(): could not unallocate %p [Err: %s]", &b[0], err))
	}

	// Set it to nil.
	b = nil
	_ = b
}

// Protect modifies the PROT_ flags for a specified byte slice.
func Protect(b []byte, prot int) {
	if err := unix.Mprotect(b, prot); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Protect(): could not set %d on %p [Err: %s]", prot, &b[0], err))
	}
}
