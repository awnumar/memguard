package memguard

import "github.com/libeclipse/memguard/memcall"

// LockedBuffer implements a buffer that stores the data.
type LockedBuffer struct {
	Buffer []byte
}

// New creates a new *LockedBuffer and returns it.
func New(length int) *LockedBuffer {
	// Allocate the new LockedBuffer.
	b := new(LockedBuffer)

	// Initialise the environment.
	memcall.Init()

	// Allocate and lock the buffer.
	b.Buffer = memcall.Alloc(length)
	memcall.Lock(b.Buffer)

	// Return a pointer to the LockedBuffer.
	return b
}

// NewFromBytes creates a new *LockedBuffer from a byte slice,
// attempting to destroy the old value before returning. It is
// not as robust as New(), but sometimes it is necessary.
func NewFromBytes(buf []byte) *LockedBuffer {
	// Allocate the new LockedBuffer.
	b := new(LockedBuffer)

	// Initialise the environment.
	memcall.Init()

	// Allocate and lock the buffer.
	b.Buffer = memcall.Alloc(len(buf))
	memcall.Lock(b.Buffer)

	// Unlock, copy over bytes, and lock again.
	memcall.Protect(b.Buffer, false, true)
	copy(b.Buffer, buf)
	memcall.Protect(b.Buffer, false, false)

	// Wipe the old bytes and set to nil.
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(0)
	}

	// Return a pointer to the LockedBuffer.
	return b
}

// AllowRead unlocks the LockedBuffer for reading. Care
// should be taken to call Lock() after use.
func (b *LockedBuffer) AllowRead() {
	memcall.Protect(b.Buffer, true, false)
}

// AllowWrite unlocks the LockedBuffer for writing. Care
// should be taken to call Lock() after use.
func (b *LockedBuffer) AllowWrite() {
	memcall.Protect(b.Buffer, false, true)
}

// AllowReadWrite unlocks the LockedBuffer for reading and
// writing. Care should be taken to call Lock() after use.
func (b *LockedBuffer) AllowReadWrite() {
	memcall.Protect(b.Buffer, true, true)
}

// Lock locks the LockedBuffer. Subsequent reading or writing
// attempts will trigger a SIGSEGV access violation and the
// program will crash.
func (b *LockedBuffer) Lock() {
	memcall.Protect(b.Buffer, false, false)
}

// Copy copies bytes from a byte slice into a LockedBuffer,
// preserving the original slice. This is insecure and so
// Move() should be favoured generally.
func (b *LockedBuffer) Copy(buf []byte) {
	// Unlock, copy over bytes, and lock again.
	memcall.Protect(b.Buffer, false, true)
	copy(b.Buffer, buf)
	memcall.Protect(b.Buffer, false, false)
}

// Move copies bytes from a byte slice into a LockedBuffer,
// destroying the original slice.
func (b *LockedBuffer) Move(buf []byte) {
	// Copy buf into the LockedBuffer.
	b.Copy(buf)

	// Wipe the old bytes and set to nil.
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(0)
	}
}

// Destroy is self explanatory. It wipes and destroys the
// LockedBuffer. This function should be called on all secure
// values before exiting.
func (b *LockedBuffer) Destroy() {
	// Allow write permissions on Buffer.
	memcall.Protect(b.Buffer, false, true)

	// Wipe and unallocate.
	memcall.Free(b.Buffer)

	// Unlock the slice.
	memcall.Unlock(b.Buffer)
}

// WipeBytes zeroes out a byte slice.
func WipeBytes(buf []byte) {
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(0)
	}
}
