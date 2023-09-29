package core

import (
	"errors"
)

// Define a memory allocator
type MemAllocator interface {
	Alloc(size int) ([]byte, error)
	Inner(buf []byte) []byte
	Protect(buf []byte, readonly bool) error
	Free(buf []byte) error
}

var (
	// ErrBufferNotOwnedByAllocator indicating that the memory region is not owned by this allocator
	ErrBufferNotOwnedByAllocator = errors.New("<memguard::core::allocator> buffer not owned by allocator; potential double free")
	// ErrBufferOverflow indicating that the memory region was tampered with
	ErrBufferOverflow = errors.New("<memguard::core::allocator> canary verification failed; buffer overflow detected")
	// ErrNullAlloc indicating that a zero length memory region was requested
	ErrNullAlloc = errors.New("<memguard::core::allocator> zero-length allocation")
	// ErrNullPointer indicating an attempted operation on a nil buffer
	ErrNullPointer = errors.New("<memguard::core::allocator> nil buffer")
)
