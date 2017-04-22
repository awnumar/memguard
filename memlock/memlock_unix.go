// +build !windows

package memlock

import "golang.org/x/sys/unix"

// Lock is a wrapper for unix.Mlock()
func Lock(b []byte) error {
	return unix.Mlock(b)
}

// Unlock is a wrapper for unix.Munlock()
func Unlock(b []byte) error {
	return unix.Munlock(b)
}
