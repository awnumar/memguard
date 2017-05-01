// +build windows

package memcall

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/alexbrainman/winapi"
)

// Placeholder variable for when we need a valid pointer to zero bytes.
var _zero uintptr

// Lock is a wrapper for windows.VirtualLock()
func Lock(b []byte) {
	if err := windows.VirtualLock(_getPtr(b), uintptr(len(b))); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Lock(): could not acquire lock on %p [Err: %s]", &b[0], err))
	}
}

// Unlock is a wrapper for windows.VirtualUnlock()
func Unlock(b []byte) {
	if err := windows.VirtualUnlock(_getPtr(b), uintptr(len(b))); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Unlock(): could not free lock on %p [Err: %s]", &b[0], err))
	}
}

// Alloc allocates a byte slice of length n and returns it.
func Alloc(n int) []byte {
	// Allocate the memory.
	ptr, err := winapi.VirtualAlloc(_zero, uintptr(n), 0x00001000|0x00002000, 0x40)
	if err != nil {
		panic(fmt.Sprintf("memguard.memcall.Alloc(): could not allocate [Err: %s]", err))
	}

	// Return the allocated memory.
	return _getBytes(ptr, n, n)
}

// Free unallocates the byte slice specified.
func Free(b []byte) {
	if err := winapi.VirtualFree(_getPtr(b), uintptr(0), 0x8000); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Free(): could not unallocate %p [Err: %s]", &b[0], err))
	}
}

// Protect modifies the Memory Protection Constants for a specified byte slice.
func Protect(b []byte, read, write bool) {
	// Ascertain protection value from arguments.
	var prot int
	if write {
		prot = 0x40 // PAGE_EXECUTE_READWRITE
	} else if read {
		prot = 0x20 // PAGE_EXECUTE_READ
	} else {
		prot = 0x01 // PAGE_NOACCESS
	}

	var oldProtect uint32
	if err := winapi.VirtualProtect(_getPtr(b), uintptr(len(b)), uint32(prot), &oldProtect); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Protect(): could not set %d on %p [Err: %s]", prot, &b[0], err))
	}
}

// DisableCoreDumps is included for compatibility reasons. On windows it is a no-op function.
func DisableCoreDumps() {}

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
