// +build windows

package memlock

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Single-word zero for use when we need a valid pointer to 0 bytes.
var _zero uintptr

// Lock is a wrapper for windows.VirtualLock()
func Lock(b []byte) error {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	return windows.VirtualLock(uintptr(_p0), uintptr(len(b)))
}

// Unlock is a wrapper for windows.VirtualUnlock()
func Unlock(b []byte) error {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	return windows.VirtualUnlock(uintptr(_p0), uintptr(len(b)))
}
