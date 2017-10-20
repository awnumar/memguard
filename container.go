package memguard

import (
	"crypto/subtle"
	"runtime"
	"sync"
	"unsafe"

	"github.com/awnumar/memguard/memcall"
)

/*
LockedBuffer is a structure that holds secure values.

The protected memory itself can be accessed with the Buffer() method. The various states can be accessed with the IsDestroyed() and IsMutable() methods, both of which are pretty self-explanatory.

The number of LockedBuffers that you are able to create is limited by how much memory your system kernel allows each process to mlock/VirtualLock. Therefore you should call Destroy on LockedBuffers that you no longer need, or simply defer a Destroy call after creating a new LockedBuffer.

The entire memguard API handles and passes around pointers to LockedBuffers, and so, for both security and convenience, you should refrain from dereferencing a LockedBuffer.

If an API function that needs to edit a LockedBuffer is given one that is immutable, the call will return an ErrImmutable. Similarly, if a function is given a LockedBuffer that has been destroyed, the call will return an ErrDestroyed.
*/
type LockedBuffer struct {
	*container  // Import all the container fields.
	*littleBird // Monitor this for auto-destruction.
}

// container implements the actual data container.
type container struct {
	sync.Mutex // Local mutex lock.

	buffer  []byte // Slice that references the protected memory.
	mutable bool   // Is this LockedBuffer mutable?
}

// littleBird is a value that we monitor instead of the LockedBuffer
// itself. It allows us to tell the GC to auto-destroy LockedBuffers.
type littleBird [16]byte

// Global internal function used to create new secure containers.
func newContainer(size int, mutable bool) (*LockedBuffer, error) {
	// Return an error if length < 1.
	if size < 1 {
		return nil, ErrInvalidLength
	}

	// Allocate a new LockedBuffer.
	ib := new(container)
	b := &LockedBuffer{ib, new(littleBird)}

	// Round length + 32 bytes for the canary to a multiple of the page size..
	roundedLength := roundToPageSize(size + 32)

	// Calculate the total size of memory including the guard pages.
	totalSize := (2 * pageSize) + roundedLength

	// Allocate it all.
	memory := memcall.Alloc(totalSize)

	// Make the guard pages inaccessible.
	memcall.Protect(memory[:pageSize], false, false)
	memcall.Protect(memory[pageSize+roundedLength:], false, false)

	// Lock the pages that will hold the sensitive data.
	memcall.Lock(memory[pageSize : pageSize+roundedLength])

	// Set the canary.
	subtle.ConstantTimeCopy(1, memory[pageSize+roundedLength-size-32:pageSize+roundedLength-size], canary)

	// Set Buffer to a byte slice that describes the reigon of memory that is protected.
	b.buffer = getBytes(uintptr(unsafe.Pointer(&memory[pageSize+roundedLength-size])), size)

	// Set appropriate mutability state.
	b.mutable = true
	if !mutable {
		b.MakeImmutable()
	}

	// Use a finalizer to make sure the buffer gets destroyed if forgotten.
	runtime.SetFinalizer(b.littleBird, func(_ *littleBird) {
		go ib.Destroy()
	})

	// Append the container to allLockedBuffers. We have to add container
	// instead of LockedBuffer so that littleBird can become unreachable.
	allLockedBuffersMutex.Lock()
	allLockedBuffers = append(allLockedBuffers, ib)
	allLockedBuffersMutex.Unlock()

	// Return a pointer to the LockedBuffer.
	return b, nil
}
