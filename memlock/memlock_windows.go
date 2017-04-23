// +build windows

package memlock

/*
#cgo LDFLAGS: -lsodium

#include <stdlib.h>

#include <sodium.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

var (
	// Single-word zero for use when we need a valid pointer to 0 bytes.
	_zero uintptr

	// Stores the initialisation state of libsodium.
	initialised bool
)

// Lock is a wrapper for sodium_mlock()
func Lock(b []byte) {
	if !initialised {
		_init()
	}

	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}

	if err := C.sodium_mlock(_p0, C.size_t(len(b))); err == -1 {
		panic(fmt.Sprintf("memguard.memprot.Lock(): could not aquire lock on %p", &b[0]))
	}
}

// Unlock is a wrapper for sodium_munlock()
func Unlock(b []byte) {
	var _p0 unsafe.Pointer
	if len(b) > 0 {
		_p0 = unsafe.Pointer(&b[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}

	if err := C.sodium_munlock(_p0, C.size_t(len(b))); err == -1 {
		panic(fmt.Sprintf("memguard.memprot.Unlock(): could not free lock on %p", &b[0]))
	}
}

func _init() {
	// Initialise libsodium.
	if int(C.sodium_init()) == -1 {
		panic("memguard.memprot.init(): could not initialise libsodium")
	}
}
