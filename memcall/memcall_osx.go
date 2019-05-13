// +build darwin

package memcall

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// Lock is a wrapper for mlock(2).
func Lock(b []byte) error {
	if err := unix.Mlock(b); err != nil {
		return fmt.Errorf("<memguard::memcall::Lock> could not acquire lock on %p, limit reached? [Err: %s]", &b[0], err)
	}

	return nil
}

// Unlock is a wrapper for munlock(2).
func Unlock(b []byte) error {
	if err := unix.Munlock(b); err != nil {
		return fmt.Errorf("<memguard::memcall::Unlock> could not free lock on %p [Err: %s]", &b[0], err)
	}

	return nil
}

// Alloc allocates a byte slice of length n and returns it.
func Alloc(n int) ([]byte, error) {
	// Allocate the memory.
	b, err := unix.Mmap(-1, 0, n, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANON)
	if err != nil {
		return nil, fmt.Errorf("<memguard::memcall::Alloc> could not allocate [Err: %s]", err)
	}

	// Wipe it just in case there is some remnant data.
	for i := range b {
		b[i] = 0
	}

	// Return the allocated memory.
	return b, nil
}

// Free deallocates the byte slice specified.
func Free(b []byte) error {
	// Make the memory region readable and writable.
	if err := Protect(b, ReadWrite); err != nil {
		return err
	}

	// Wipe the memory region in case of remnant data.
	for i := range b {
		b[i] = 0
	}

	// Free the memory back to the kernel.
	if err := unix.Munmap(b); err != nil {
		return fmt.Errorf("<memguard::memcall::Free> could not deallocate %p [Err: %s]", &b[0], err)
	}

	return nil
}

// Protect modifies the protection state for a specified byte slice.
func Protect(b []byte, mpf MemoryProtectionFlag) error {
	var prot int
	if mpf.flag == ReadWrite.flag {
		prot = unix.PROT_READ | unix.PROT_WRITE
	} else if mpf.flag == ReadOnly.flag {
		prot = unix.PROT_READ
	} else if mpf.flag == NoAccess.flag {
		prot = unix.PROT_NONE
	} else {
		return ErrInvalidFlag
	}

	// Change the protection value of the byte slice.
	if err := unix.Mprotect(b, prot); err != nil {
		return fmt.Errorf("<memguard::memcall::Protect> could not set %d on %p [Err: %s]", prot, &b[0], err)
	}

	return nil
}

// DisableCoreDumps disables core dumps on Unix systems.
func DisableCoreDumps() error {
	// Disable core dumps.
	if err := unix.Setrlimit(unix.RLIMIT_CORE, &unix.Rlimit{Cur: 0, Max: 0}); err != nil {
		return fmt.Errorf("<memguard::memcall::DisableCoreDumps> could not set rlimit [Err: %s]", err)
	}

	return nil
}
