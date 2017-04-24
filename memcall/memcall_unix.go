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
		panic(fmt.Sprintf("memguard.memprot.init(): could not set rlimit [Err: %s]", err))
	}
}

// Lock is a wrapper for unix.Mlock(), with extra precautions.
func Lock(b []byte) {
	if err := unix.Mlock(b); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Lock(): could not aquire lock on %p", &b[0]))
	}
}

// Unlock is a wrapper for unix.Unlock.
func Unlock(b []byte) {
	if err := unix.Munlock(b); err != nil {
		panic(fmt.Sprintf("memguard.memcall.Unlock(): could not free lock on %p", &b[0]))
	}
}
