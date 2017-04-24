// +build windows

package memcall

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Single-word zero for use when we need a valid pointer to 0 bytes.
var _zero uintptr

// Lock is a wrapper for windows.VirtualLock()
func Lock(b []byte) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	if err := windows.VirtualLock(uintptr(_p0), uintptr(len(b))); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Lock(): could not aquire lock on %p", &b[0]))
	}
}

// Unlock is a wrapper for windows.VirtualUnlock()
func Unlock(b []byte) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	if err := windows.VirtualUnlock(uintptr(_p0), uintptr(len(b))); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Unlock(): could not free lock on %p", &b[0]))
	}
}
