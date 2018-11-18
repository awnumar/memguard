package memguard

import (
	"unsafe"

	"github.com/awnumar/memguard/memcall"
)

/*
ByteArray16 returns an array (of type *[16]byte) that references the secure, protected portion of memory.

The LockedBuffer must be a multiple of 2 bytes in length, and at least 16 byte longs, or else an error will be returned.
*/
func (b *container) ByteArray16() (*[16]byte, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Check to see if it's an appropriate length.
	if len(b.buffer)%2 != 0 {
		return nil, ErrInvalidConversion
	}

	if len(b.buffer) < 16 {
		return nil, ErrInvalidConversion
	}

	// Return the array.
	return (*[16]byte)(unsafe.Pointer(&b.buffer[0])), nil
}

/*
ByteArray32 returns an array (of type *[32]byte) that references the secure, protected portion of memory.

The LockedBuffer must be a multiple of 2 bytes in length, and at least 32 byte longs, or else an error will be returned.
*/
func (b *container) ByteArray32() (*[32]byte, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Check to see if it's an appropriate length.
	if len(b.buffer)%2 != 0 {
		return nil, ErrInvalidConversion
	}

	if len(b.buffer) < 32 {
		return nil, ErrInvalidConversion
	}

	// Return the array.
	return (*[32]byte)(unsafe.Pointer(&b.buffer[0])), nil
}

/*
ByteArray64 returns an array (of type *[64]byte) that references the secure, protected portion of memory.

The LockedBuffer must be a multiple of 2 bytes in length, and at least 64 byte longs, or else an error will be returned.
*/
func (b *container) ByteArray64() (*[64]byte, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Check to see if it's an appropriate length.
	if len(b.buffer)%2 != 0 {
		return nil, ErrInvalidConversion
	}

	if len(b.buffer) < 64 {
		return nil, ErrInvalidConversion
	}

	// Return the array.
	return (*[64]byte)(unsafe.Pointer(&b.buffer[0])), nil
}

/*********************************************************************************************************/

/*
MakeUnreadable asks the kernel to mark the LockedBuffer's memory as unreadable. Any subsequent attempts to modify this memory will result in the process crashing with a SIGSEGV memory violation.

To make the memory readable again, MakeReadable is called.
*/
func (b *container) MakeUnreadable() error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()
	return b.makeUnreadable()
}

func (b *container) makeUnreadable() error {
	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return ErrDestroyed
	}

	if b.readable {
		// Mark the memory as unreadable.
		memcall.Protect(getAllMemory(b)[pageSize:pageSize+roundToPageSize(len(b.buffer)+32)], false, b.mutable)

		// Tell everyone about the change we made.
		b.readable = false
	}

	// Everything went well.
	return nil
}

/*
MakeMutable asks the kernel to mark the LockedBuffer's memory as readable.

To make the memory unreadable again, MakeUnreadable is called.
*/
func (b *container) MakeReadable() error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	return b.makeReadable()
}

func (b *container) makeReadable() error {
	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return ErrDestroyed
	}

	if !b.readable {
		// Mark the memory as readable.
		memcall.Protect(getAllMemory(b)[pageSize:pageSize+roundToPageSize(len(b.buffer)+32)], true, b.mutable)

		// Tell everyone about the change we made.
		b.readable = true
	}

	// Everything went well.
	return nil
}

/*
IsReadable returns a boolean value indicating if a LockedBuffer is marked un-readable.
*/
func (b *container) IsReadable() bool {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	return b.readable
}

/*
WithReadable executes the function "fun" after making a LockedBuffer read-able. After
the function returns, the LockedBuffer is marked un-readable again.

The LockedBuffer will remain locked during the function execution, preventing other method
calls on the container. Therefor all calls to LockedBuffer must be made BEFORE running WithReadable.
*/
func (b *container) WithReadable(fun func()) error {
	b.Lock()
	defer b.Unlock()
	// make the buffer readable.
	if err := b.makeReadable(); err != nil {
		return err
	}
	// execute the function.
	fun()
	// make the buffer unreadable again.
	if err := b.makeUnreadable(); err != nil {
		return err
	}
	return nil
}
