// +build freebsd

package memcall

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// Lock is a wrapper for unix.Mlock(), with extra precautions.
func Lock(b []byte) error {
	// Advise the kernel not to dump. Ignore failure.
	unix.Madvise(b, unix.MADV_DONTDUMP)

	// Call mlock.
	if err := unix.Mlock(b); err != nil {
		return fmt.Errorf("memguard.memcall.Lock(): could not acquire lock on %p, limit reached? [Err: %s]", &b[0], err)
	}

	return nil
}

// Unlock is a wrapper for unix.Munlock().
func Unlock(b []byte) error {
	if err := unix.Munlock(b); err != nil {
		return fmt.Errorf("memguard.memcall.Unlock(): could not free lock on %p [Err: %s]", &b[0], err)
	}

	return nil
}

// Alloc allocates a byte slice of length n and returns it.
func Alloc(n int) ([]byte, error) {
	// Allocate the memory.
	b, err := unix.Mmap(-1, 0, n, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS|unix.MAP_NOCORE)
	if err != nil {
		return nil, fmt.Errorf("memguard.memcall.Alloc(): could not allocate [Err: %s]", err)
	}

	// Return the allocated memory.
	return b, nil
}

// Free unallocates the byte slice specified.
func Free(b []byte) error {
	if err := unix.Munmap(b); err != nil {
		return fmt.Errorf("memguard.memcall.Free(): could not unallocate %p [Err: %s]", &b[0], err)
	}

	return nil
}

// Protect modifies the PROT_ flags for a specified byte slice.
func Protect(b []byte, read, write bool) error {
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
		return fmt.Errorf("memguard.memcall.Protect(): could not set %d on %p [Err: %s]", prot, &b[0], err)
	}

	return nil
}

// DisableCoreDumps disables core dumps on Unix systems.
func DisableCoreDumps() error {
	// Disable core dumps.
	if err := unix.Setrlimit(unix.RLIMIT_CORE, &unix.Rlimit{Cur: 0, Max: 0}); err != nil {
		return fmt.Errorf("memguard.memcall.DisableCoreDumps(): could not set rlimit [Err: %s]", err)
	}

	return nil
}
