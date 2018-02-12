// +build !windows,!darwin,!openbsd,!freebsd

package memcall

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// Lock is a wrapper for unix.Mlock(), with extra precautions.
func Lock(b []byte) {
	// Advise the kernel not to dump. Ignore failure.
	unix.Madvise(b, unix.MADV_DONTDUMP)

	// Call mlock.
	if err := unix.Mlock(b); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Lock(): could not acquire lock on %p, limit reached? [Err: %s]", &b[0], err))
	}
}

// Unlock is a wrapper for unix.Munlock().
func Unlock(b []byte) {
	if err := unix.Munlock(b); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Unlock(): could not free lock on %p [Err: %s]", &b[0], err))
	}
}

// Alloc allocates a byte slice of length n and returns it.
func Alloc(n int) []byte {
	// Allocate the memory.
	b, err := unix.Mmap(-1, 0, n, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		panic(fmt.Sprintf("memguard.memcall.Alloc(): could not allocate [Err: %s]", err))
	}

	// Fill memory with weird bytes in order to help catch bugs due to uninitialized data.
	for i := 0; i < n; i++ {
		b[i] = byte(0xdb)
	}

	// Return the allocated memory.
	return b
}

// Free unallocates the byte slice specified.
func Free(b []byte) {
	if err := unix.Munmap(b); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Free(): could not unallocate %p [Err: %s]", &b[0], err))
	}
}

// Protect modifies the PROT_ flags for a specified byte slice.
func Protect(b []byte, read, write bool) {
	// Ascertain protection value from arguments.
	var prot int
	if read && write {
		prot = unix.PROT_READ | unix.PROT_WRITE
	} else if read {
		prot = unix.PROT_READ
	} else if write {
		prot = unix.PROT_WRITE
	} else {
		prot = unix.PROT_NONE
	}

	// Change the protection value of the byte slice.
	if err := unix.Mprotect(b, prot); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Protect(): could not set %d on %p [Err: %s]", prot, &b[0], err))
	}
}

// DisableCoreDumps disables core dumps on Unix systems.
func DisableCoreDumps() {
	// Disable core dumps.
	if err := unix.Setrlimit(unix.RLIMIT_CORE, &unix.Rlimit{Cur: 0, Max: 0}); err != nil {
		panic(fmt.Sprintf("memguard.memcall.DisableCoreDumps(): could not set rlimit [Err: %s]", err))
	}
}
