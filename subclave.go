package memguard

import (
	"sync"
	"unsafe"

	"github.com/awnumar/memguard/memcall"
	"golang.org/x/crypto/blake2b"
)

var (
	// Array of all active subclaves, and associated mutex.
	subclaves      []*subclave
	subclavesMutex = &sync.Mutex{}
)

// The subclave container is similar to a normal container but it is only used internally to protect 32 byte values that are used in the protection of normal containers.
type subclave struct {
	sync.Mutex

	x []byte
	y []byte
}

// Creates and returns a new subclave object.
func newSubclave() *subclave {
	// Allocate a new subclave object.
	s := new(subclave)

	// Allocate memory for the fields.
	roundedSize := roundToPageSize(32)
	x, err := memcall.Alloc(roundedSize)
	if err != nil {
		SafePanic(err)
	}
	y, err := memcall.Alloc(roundedSize)
	if err != nil {
		SafePanic(err)
	}

	// Lock the pages into RAM.
	if err := memcall.Lock(s.x); err != nil {
		SafePanic(err)
	}
	if err := memcall.Lock(s.y); err != nil {
		SafePanic(err)
	}

	// Set the subclave object's buffers to the allocated memory.
	s.x = getBytes(uintptr(unsafe.Pointer(&x[0])), 32)
	s.y = getBytes(uintptr(unsafe.Pointer(&y[0])), 32)

	// Initialise a subclave with a random 32 byte value.
	fillRandBytes(s.x)
	fillRandBytes(s.y)
	hr := h(s.y)
	for i := range hr {
		s.x[i] ^= hr[i]
	}

	// Store a global reference to this subclave.
	subclavesMutex.Lock()
	subclaves = append(subclaves, s)
	subclavesMutex.Unlock()

	// Return the created subclave object.
	return s
}

// Returns the value stored in a subclave, wrapped in a normal LockedBuffer. The caller should destroy this object as soon as possible.
func (s *subclave) get() *LockedBuffer {
	// Attain the mutex.
	s.Lock()
	defer s.Unlock()

	// Create a new LockedBuffer.
	b, _ := NewMutable(32)

	// Create a copy of the subclave data inside the LockedBuffer.
	h := h(s.y)
	for i := range b.buffer {
		b.buffer[i] = h[i] ^ s.x[i]
	}

	// Return the LockedBuffer.
	return b
}

// This method is used to update the value stored in a subclave.
func (s *subclave) update(b []byte) {
	// Attain the mutex.
	s.Lock()
	defer s.Unlock()

	// Update the subclave with the new value, wiping the old.
	hy := h(s.y)
	for i := range hy {
		s.x[i] = hy[i] ^ b[i]
	}
}

// This method is used to rekey a subclave. Ideally this should be done at short, regular intervals.
func (s *subclave) rekey() {
	// Attain the mutex.
	s.Lock()
	defer s.Unlock()

	// Compute the updated s.y, but don't overwrite the old value.
	r := r()
	rr := make([]byte, 32)
	for i := range s.y {
		rr[i] = s.y[i] ^ r[i]
	}

	// Update s.x with the new s.y value.
	hy := h(s.y)
	hrr := h(rr)
	for i := range r {
		s.x[i] ^= hy[i] ^ hrr[i]
	}

	// Overwrite the old s.y value with the new one.
	s.y = rr
}

// generate a random 32 byte value
func r() []byte {
	r := make([]byte, 32)
	fillRandBytes(r)
	return r
}

// Cryptographic hash function.
func h(b []byte) [32]byte {
	return blake2b.Sum256(b)
}
