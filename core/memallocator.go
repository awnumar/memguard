package core

import (
	"errors"
	"sync/atomic"
)

// Define a memory allocator
type MemAllocator interface {
	Alloc(size int) ([]byte, error)
	Inner(buf []byte) []byte
	Protect(buf []byte, readonly bool) error
	Free(buf []byte) error
}

// AllocatorStatistics statistics about memory allocations and errors
type AllocatorStatistics struct {
	PageAllocs        atomic.Uint64
	PageAllocErrors   atomic.Uint64
	PageFrees         atomic.Uint64
	PageFreeErrors    atomic.Uint64
	ObjectAllocs      atomic.Uint64
	ObjectAllocErrors atomic.Uint64
	ObjectFrees       atomic.Uint64
	ObjectFreeErrors  atomic.Uint64
	Slabs             atomic.Uint64
}

// MemStats statistics about memory allocations and errors
var MemStats = AllocatorStatistics{}

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
