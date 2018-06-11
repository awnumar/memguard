// +build windows

package memcall

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Placeholder variable for when we need a valid pointer to zero bytes.
var _zero uintptr

// Lock is a wrapper for windows.VirtualLock()
func Lock(b []byte) error {
	if err := windows.VirtualLock(_getPtr(b), uintptr(len(b))); err != nil {
		return fmt.Errorf("memguard.memcall.Lock(): could not acquire lock on %p, limit reached? [Err: %s]", &b[0], err)
	}

	return nil
}

// Unlock is a wrapper for windows.VirtualUnlock()
func Unlock(b []byte) error {
	if err := windows.VirtualUnlock(_getPtr(b), uintptr(len(b))); err != nil {
		return fmt.Errorf("memguard.memcall.Unlock(): could not free lock on %p [Err: %s]", &b[0], err)
	}

	return nil
}

// Alloc allocates a byte slice of length n and returns it.
func Alloc(n int) ([]byte, error) {
	// Allocate the memory.
	ptr, err := windows.VirtualAlloc(_zero, uintptr(n), 0x1000|0x2000, 0x4)
	if err != nil {
		return nil, fmt.Errorf("memguard.memcall.Alloc(): could not allocate [Err: %s]", err)
	}

	// Convert into a byte slice.
	b := _getBytes(ptr, n, n)

	// Fill memory with weird bytes in order to help catch bugs due to uninitialized data.
	for i := 0; i < n; i++ {
		b[i] = byte(0xdb)
	}

	// Return the allocated memory.
	return b, nil
}

// Free unallocates the byte slice specified.
func Free(b []byte) error {
	if err := windows.VirtualFree(_getPtr(b), uintptr(0), 0x8000); err != nil {
		return fmt.Errorf("memguard.memcall.Free(): could not unallocate %p [Err: %s]", &b[0], err)s
	}

	return nil
}

// Protect modifies the Memory Protection Constants for a specified byte slice.
func Protect(b []byte, read, write bool) error {
	// Ascertain protection value from arguments.
	var prot int
	if write {
		prot = 0x4 // PAGE_READWRITE
	} else if read {
		prot = 0x2 // PAGE_READ
	} else {
		prot = 0x1 // PAGE_NOACCESS
	}

	var oldProtect uint32
	if err := windows.VirtualProtect(_getPtr(b), uintptr(len(b)), uint32(prot), &oldProtect); err != nil {
		return fmt.Errorf("memguard.memcall.Protect(): could not set %d on %p [Err: %s]", prot, &b[0], err)
	}

	return nil
}

// DisableCoreDumps is included for compatibility reasons. On windows it is a no-op function.
func DisableCoreDumps() error { return nil }

// Auxiliary functions.
func _getPtr(b []byte) uintptr {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	return uintptr(_p0)
}

func _getBytes(ptr uintptr, len int, cap int) []byte {
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{ptr, len, cap}
	return *(*[]byte)(unsafe.Pointer(&sl))
}
