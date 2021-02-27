package core

import (
	"errors"
	"sync"

	"github.com/awnumar/memcall"
)

var (
	buffers = new(bufferList)
)

// ErrNullBuffer is returned when attempting to construct a buffer of size less than one.
var ErrNullBuffer = errors.New("<memguard::core::ErrNullBuffer> buffer size must be greater than zero")

// ErrBufferExpired is returned when attempting to perform an operation on or with a buffer that has been destroyed.
var ErrBufferExpired = errors.New("<memguard::core::ErrBufferExpired> buffer has been purged from memory and can no longer be used")

/*
Buffer is a structure that holds raw sensitive data.

The number of Buffers that can exist at one time is limited by how much memory your system's kernel allows each process to mlock/VirtualLock. Therefore you should call DestroyBuffer on Buffers that you no longer need, ideally defering a Destroy call after creating a new one.
*/
type Buffer struct {
	mu sync.RWMutex // caller's responsibility to acquire

	alive   bool // Signals that destruction has not come
	mutable bool // Mutability state of underlying memory

	data   []byte // Portion of memory holding the data
	memory []byte // Entire allocated memory region

	preguard  []byte // Guard page addressed before the data
	inner     []byte // Inner region between the guard pages
	postguard []byte // Guard page addressed after the data

	canary []byte // Value written behind data to detect spillage
}

/*
NewBuffer is a raw constructor for the Buffer object.
*/
func NewBuffer(size int) (b *Buffer, err error) {
	b = &Buffer{}

	if size < 1 {
		err = ErrNullBuffer
		return
	}

	innerLen := roundToPageSize(size)
	b.memory, err = memcall.Alloc((2 * pageSize) + innerLen)
	if err != nil {
		return
	}

	b.data = getBytes(&b.memory[pageSize+innerLen-size], size)

	b.preguard = getBytes(&b.memory[0], pageSize)
	b.inner = getBytes(&b.memory[pageSize], innerLen)
	b.postguard = getBytes(&b.memory[pageSize+innerLen], pageSize)

	b.canary = getBytes(&b.memory[pageSize], len(b.inner)-len(b.data))

	if err = memcall.Lock(b.inner); err != nil {
		return
	}

	if err = Scramble(b.canary); err != nil {
		return
	}
	Copy(b.preguard, b.canary)
	Copy(b.postguard, b.canary)

	if err = memcall.Protect(b.preguard, memcall.NoAccess()); err != nil {
		return
	}
	if err = memcall.Protect(b.postguard, memcall.NoAccess()); err != nil {
		return
	}

	b.alive = true
	b.mutable = true

	buffers.add(b)

	return b, nil
}

// Data returns a byte slice representing the memory region containing the data.
func (b *Buffer) Data() []byte {
	return b.data
}

// Inner returns a byte slice representing the entire inner memory pages. This should NOT be used unless you have a specific need.
func (b *Buffer) Inner() []byte {
	return b.inner
}

// Freeze makes the underlying memory of a given buffer immutable. This will do nothing if the Buffer has been destroyed.
func (b *Buffer) Freeze() error {
	if !b.alive {
		return ErrBufferExpired
	}

	// Only do anything if currently mutable.
	if b.mutable {
		// Make the memory immutable.
		if err := memcall.Protect(b.inner, memcall.ReadOnly()); err != nil {
			return err
		}
		b.mutable = false
	}

	return nil
}

// Melt makes the underlying memory of a given buffer mutable. This will do nothing if the Buffer has been destroyed.
func (b *Buffer) Melt() error {
	if !b.alive {
		return ErrBufferExpired
	}

	if !b.mutable {
		if err := memcall.Protect(b.inner, memcall.ReadWrite()); err != nil {
			return err
		}
		b.mutable = true
	}

	return nil
}

// Scramble attempts to overwrite the data with cryptographically-secure random bytes.
func (b *Buffer) Scramble() error {
	if !b.alive {
		return ErrBufferExpired
	}

	return Scramble(b.Data())
}

/*
Destroy performs some security checks, securely wipes the contents of, and then releases a Buffer's memory back to the OS. If a security check fails, the process will attempt to wipe all it can before safely panicking.

If the Buffer has already been destroyed, the function does nothing and returns nil.
*/
func (b *Buffer) Destroy() (err error) {
	if !b.alive {
		return
	}

	// Make all of the memory readable and writable.
	if err = memcall.Protect(b.memory, memcall.ReadWrite()); err != nil {
		return
	}
	b.mutable = true

	Wipe(b.data)

	if !Equal(b.preguard, b.postguard) || !Equal(b.preguard[:len(b.canary)], b.canary) {
		return errors.New("<memguard::core::buffer> canary verification failed; buffer overflow detected")
	}

	Wipe(b.memory)

	if err = memcall.Unlock(b.inner); err != nil {
		return
	}

	if err = memcall.Free(b.memory); err != nil {
		return
	}

	buffers.remove(b)

	b.alive = false
	b.mutable = false
	b.data = nil
	b.memory = nil
	b.preguard = nil
	b.inner = nil
	b.postguard = nil
	b.canary = nil

	return nil
}

// Alive returns true if the buffer has not been destroyed.
func (b Buffer) Alive() bool {
	return b.alive
}

// Mutable returns true if the buffer is mutable.
func (b Buffer) Mutable() bool {
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
