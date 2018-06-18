package memguard

import (
	"sync"
	"time"
	"unsafe"

	"github.com/awnumar/memguard/memcall"
	"golang.org/x/crypto/blake2b"
)

var (
	// Array of all active subclaves, and associated mutex.
	subclaves      []*subclave
	subclavesMutex = &sync.Mutex{}

	// Sync object to ensure we only start a single rekey routine.
	rekeyOnce sync.Once

	// Set the interval between rekeys, in seconds.
	interval uint = 8
)

/*
SetRekeyInterval lets you decide the time interval, in seconds, between the rekeys of the subclaves.

Subclaves are special containers used only internally to protect sensitive values that are used in the protection of normal Enclaves. These subclaves are re-keyed at regular intervals, with the default being every 8 seconds.

This is the only public function exposed by the subclave implementation. Please refrain from calling this function unless you know what you're doing.
*/
func SetRekeyInterval(t uint) {
	interval = t
}

// The subclave container is similar to a normal container but it is only used internally to protect 32 byte values that are used in the protection of normal containers.
type subclave struct {
	sync.Mutex

	x []byte
	y []byte
}

// This is an immutable and ephemeral Enclave-like object that allows you to view and use the value stored inside a subclave. It holds a copy and so will not reflect any changes to the subclave upon which it's based. It should be destroyed as soon as possible after use.
type subclaveView struct {
	buffer []byte
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

	// Set the subclave object's buffers to the allocated memory.
	s.x = getBytes(uintptr(unsafe.Pointer(&x[0])), 32)
	s.y = getBytes(uintptr(unsafe.Pointer(&y[0])), 32)

	// Lock the pages into RAM.
	if err := memcall.Lock(s.x); err != nil {
		SafePanic(err)
	}
	if err := memcall.Lock(s.y); err != nil {
		SafePanic(err)
	}

	// Initialise the subclave with a random 32 byte value.
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

	// If we haven't already started a rekey routine, do it now.
	rekeyOnce.Do(func() {
		go func() {
			for {
				// Sleep for the specified interval.
				time.Sleep(time.Duration(interval) * time.Second)

				// Get a snapshot of the existing subclaves.
				subclavesMutex.Lock()
				subs := make([]*subclave, len(subclaves))
				copy(subs, subclaves)
				subclavesMutex.Unlock()

				// Rekey them all.
				for _, s := range subs {
					s.rekey()
				}
			}
		}()
	})

	// Return the created subclave object.
	return s
}

// Returns the value stored in a subclave, wrapped in a subclaveView object. The caller should destroy this object as soon as possible.
func (s *subclave) getView() *subclaveView {
	// Create a new subclaveView object.
	sv := new(subclaveView)

	// Calculate the total size of memory including the guard pages.
	roundedSize := roundToPageSize(32)
	totalSize := (2 * pageSize) + roundedSize

	// Allocate it all.
	memory, err := memcall.Alloc(totalSize)
	if err != nil {
		SafePanic(err)
	}

	// Make the guard pages inaccessible.
	if err := memcall.Protect(memory[:pageSize], false, false); err != nil {
		SafePanic(err)
	}
	if err := memcall.Protect(memory[pageSize+roundedSize:], false, false); err != nil {
		SafePanic(err)
	}

	// Lock the pages that will hold the sensitive data.
	if err := memcall.Lock(memory[pageSize : pageSize+roundedSize]); err != nil {
		SafePanic(err)
	}

	// Set Buffer to a byte slice that describes the region of memory that is protected.
	sv.buffer = getBytes(uintptr(unsafe.Pointer(&memory[pageSize+roundedSize-32])), 32)

	// Create a copy of the subclave data inside the subclaveView.
	h := h(s.y)
	for i := range sv.buffer {
		sv.buffer[i] = h[i] ^ s.x[i]
	}

	// Make the subclaveView immutable.
	if err := memcall.Protect(memory[pageSize:pageSize+roundedSize], true, false); err != nil {
		SafePanic(err)
	}

	// Return the subclaveView object.
	return sv
}

func (sv *subclaveView) destroy() {
	// Get a slice referencing all the memory associated with this subclaveView object.
	roundedSize := roundToPageSize(32)
	memLen := (pageSize * 2) + roundedSize
	memAddr := uintptr(unsafe.Pointer(&sv.buffer[0])) - uintptr((roundedSize-32)+pageSize)
	memory := getBytes(memAddr, memLen)

	// Make all of the memory readable and writable.
	if err := memcall.Protect(memory, true, true); err != nil {
		SafePanic(err)
	}

	// Wipe the pages that hold our data.
	wipeBytes(memory[pageSize : pageSize+roundedSize])

	// Unlock the pages that hold our data.
	if err := memcall.Unlock(memory[pageSize : pageSize+roundedSize]); err != nil {
		SafePanic(err)
	}

	// Free all related memory.
	if err := memcall.Free(memory); err != nil {
		SafePanic(err)
	}

	// Set the buffer to nil.
	sv.buffer = nil
}

// This method is used to update the value stored in a subclave.
func (s *subclave) update(b []byte) {
	// Check length is 32.
	if len(b) != 32 {
		SafePanic("memguard.subclave.update: input must be 32 bytes")
	}

	// Attain the mutex.
	s.Lock()
	defer s.Unlock()

	// Update the subclave with the new value, wiping the old.
	hy := h(s.y)
	for i := range hy {
		s.x[i] = hy[i] ^ b[i]
	}
}

// This method is used to reset the value stored inside a subclave to a new 32 byte random value, wiping the old.
func (s *subclave) refresh() {
	// Attain the mutex.
	s.Lock()
	defer s.Unlock()

	// Refresh the value.
	fillRandBytes(s.x)
	fillRandBytes(s.y)
	hr := h(s.y)
	for i := range hr {
		s.x[i] ^= hr[i]
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
