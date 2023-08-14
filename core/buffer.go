package core

import (
	"errors"
	"sync"
)

var (
	allocator = NewPageAllocator()
	buffers   = new(bufferList)
)

// ErrNullBuffer is returned when attempting to construct a buffer of size less than one.
var ErrNullBuffer = errors.New("<memguard::core::ErrNullBuffer> buffer size must be greater than zero")

/*
Buffer is a structure that holds raw sensitive data.

The number of Buffers that can exist at one time is limited by how much memory your system's kernel allows each process to mlock/VirtualLock. Therefore you should call DestroyBuffer on Buffers that you no longer need, ideally defering a Destroy call after creating a new one.
*/
type Buffer struct {
	sync.RWMutex // Local mutex lock // TODO: this does not protect 'data' field

	alive   bool // Signals that destruction has not come
	mutable bool // Mutability state of underlying memory

	data []byte // Portion of memory holding the data
}

/*
NewBuffer is a raw constructor for the Buffer object.
*/
func NewBuffer(size int) (*Buffer, error) {
	if size < 1 {
		return nil, ErrNullBuffer
	}

	b := new(Buffer)

	// Allocate the total needed memory
	var err error
	b.data, err = allocator.Alloc(size)
	if err != nil {
		Panic(err)
	}

	// Set remaining properties
	b.alive = true
	b.mutable = true

	// Append the container to list of active buffers.
	buffers.add(b)

	// Return the created Buffer to the caller.
	return b, nil
}

// Data returns a byte slice representing the memory region containing the data.
func (b *Buffer) Data() []byte {
	return b.data
}

// Inner returns a byte slice representing the entire inner memory pages. This should NOT be used unless you have a specific need.
func (b *Buffer) Inner() []byte {
	return allocator.Inner(b.data)
}

// Freeze makes the underlying memory of a given buffer immutable. This will do nothing if the Buffer has been destroyed.
func (b *Buffer) Freeze() {
	if err := b.freeze(); err != nil {
		Panic(err)
	}
}

func (b *Buffer) freeze() error {
	b.Lock()
	defer b.Unlock()

	if !b.alive {
		return nil
	}

	if b.mutable {
		if err := allocator.Protect(b.data, true); err != nil {
			return err
		}
		b.mutable = false
	}

	return nil
}

// Melt makes the underlying memory of a given buffer mutable. This will do nothing if the Buffer has been destroyed.
func (b *Buffer) Melt() {
	if err := b.melt(); err != nil {
		Panic(err)
	}
}

func (b *Buffer) melt() error {
	b.Lock()
	defer b.Unlock()

	if !b.alive {
		return nil
	}

	if !b.mutable {
		if err := allocator.Protect(b.data, false); err != nil {
			return err
		}
		b.mutable = true
	}
	return nil
}

// Scramble attempts to overwrite the data with cryptographically-secure random bytes.
func (b *Buffer) Scramble() {
	if err := b.scramble(); err != nil {
		Panic(err)
	}
}

func (b *Buffer) scramble() error {
	b.Lock()
	defer b.Unlock()
	return Scramble(b.Data())
}

/*
Destroy performs some security checks, securely wipes the contents of, and then releases a Buffer's memory back to the OS. If a security check fails, the process will attempt to wipe all it can before safely panicking.

If the Buffer has already been destroyed, the function does nothing and returns nil.
*/
func (b *Buffer) Destroy() {
	if err := b.destroy(); err != nil {
		Panic(err)
	}
	// Remove this one from global slice.
	buffers.remove(b)
}

func (b *Buffer) destroy() error {
	if b == nil {
		return nil
	}

	// Attain a mutex lock on this Buffer.
	b.Lock()
	defer b.Unlock()

	// Return if it's already destroyed.
	if !b.alive {
		return nil
	}

	// Destroy the memory content and free the space
	if b.data != nil {
		if err := allocator.Free(b.data); err != nil {
			return err
		}
	}

	// Reset the fields.
	b.alive = false
	b.mutable = false
	b.data = nil
	return nil
}

// Alive returns true if the buffer has not been destroyed.
func (b *Buffer) Alive() bool {
	b.RLock()
	defer b.RUnlock()
	return b.alive
}

// Mutable returns true if the buffer is mutable.
func (b *Buffer) Mutable() bool {
	b.RLock()
	defer b.RUnlock()
	return b.mutable
}

// BufferList stores a list of buffers in a thread-safe manner.
type bufferList struct {
	sync.RWMutex
	list []*Buffer
}

// Add appends a given Buffer to the list.
func (l *bufferList) add(b ...*Buffer) {
	l.Lock()
	defer l.Unlock()

	l.list = append(l.list, b...)
}

// Copy returns an instantaneous snapshot of the list.
func (l *bufferList) copy() []*Buffer {
	l.Lock()
	defer l.Unlock()

	list := make([]*Buffer, len(l.list))
	copy(list, l.list)

	return list
}

// Remove removes a given Buffer from the list.
func (l *bufferList) remove(b *Buffer) {
	l.Lock()
	defer l.Unlock()

	for i, v := range l.list {
		if v == b {
			l.list = append(l.list[:i], l.list[i+1:]...)
			break
		}
	}
}

// Exists checks if a given buffer is in the list.
func (l *bufferList) exists(b *Buffer) bool {
	l.RLock()
	defer l.RUnlock()

	for _, v := range l.list {
		if b == v {
			return true
		}
	}

	return false
}

// Flush clears the list and returns its previous contents.
func (l *bufferList) flush() []*Buffer {
	l.Lock()
	defer l.Unlock()

	list := make([]*Buffer, len(l.list))
	copy(list, l.list)

	l.list = nil

	return list
}
