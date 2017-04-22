// +build !windows

package memprot

import "golang.org/x/sys/unix"

// Protect is a wrapper for unix.Mprotect()
func Protect(b []byte, prot int) error {
	return unix.Mprotect(b, prot)
}
