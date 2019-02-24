package memguard

import "github.com/awnumar/memguard/core"

/*
Enclave is a sealed and encrypted container for sensitive data.
*/
type Enclave struct {
	*core.Enclave
}

/*
LockedBuffer is a structure that holds raw sensitive data.

The number of LockedBuffers that you are able to create is limited by how much memory your system's kernel allows each process to mlock/VirtualLock. Therefore you should call Destroy on LockedBuffers that you no longer need or defer a Destroy call after creating a new LockedBuffer.
*/
type LockedBuffer struct {
	*core.Buffer
	*drop
}

// This is a value that is monitored by a finalizer so that
// we can clean up LockedBuffers that have gone out of scope.
type drop [16]byte
