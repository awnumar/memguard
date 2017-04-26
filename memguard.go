package memguard

import (
	"os"
	"unsafe"

	"github.com/libeclipse/memguard/memcall"
)

var (
	// Grab the system page size.
	pageSize = os.Getpagesize()

	// Placeholder variable for when we need a valid pointer to zero bytes.
	_zero uintptr
)

// LockedBuffer implements a buffer that stores the data.
type LockedBuffer struct {
	Buffer    []byte
	mainSlice []byte
}

// New creates a new *LockedBuffer and returns it. The
// LockedBuffer is in the locked state.
func New(length int) *LockedBuffer {
	// Allocate the new LockedBuffer.
	b := new(LockedBuffer)

	// Round length to pageSize.
	roundedLength := _roundPage(length)

	// Set Total Size with guard pages.
	totalSize := (2 * pageSize) + roundedLength

	// Allocate it all.
	mainSlice := memcall.Alloc(totalSize)

	//Lock the page that will hold our data.
	memcall.Lock(mainSlice[pageSize : pageSize+roundedLength])

	// Make the Guard Pages inaccessible.
	memcall.Protect(mainSlice[:pageSize], false, false)
	memcall.Protect(mainSlice[pageSize+roundedLength:totalSize], false, false)

	// Set Buffer to a byte slice that describes the reigon of memory that is protected.
	b.Buffer = _getBytes(uintptr(unsafe.Pointer(&mainSlice[pageSize+roundedLength-length])), length, length)

	// Lock this down yo.
	b.Lock()

	// Save the address (needed when freeing).
	b.mainSlice = mainSlice[:]

	// Return a pointer to the LockedBuffer.
	return b
}

// NewFromBytes creates a new *LockedBuffer from a byte slice,
// attempting to destroy the old value before returning. It is
// not as robust as New(), but sometimes it is necessary.
func NewFromBytes(buf []byte) *LockedBuffer {
	// Use New to create a Secured LockedBuffer
	b := New(len(buf))

	// Copy the bytes from buf, wiping the afterwards.
	b.Move(buf)

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
// Move() should be favoured generally. The LockedBuffer
// should be unlocked first, and relocked afterwards.
func (b *LockedBuffer) Copy(buf []byte) {
	copy(b.Buffer, buf)
}

// Move copies bytes from a byte slice into a LockedBuffer,
// destroying the original slice. The LockedBuffer should be
// unlocked first, and relocked afterwards.
func (b *LockedBuffer) Move(buf []byte) {
	// Copy buf into the LockedBuffer.
	b.Copy(buf)

	// Wipe the old bytes and set to nil.
	WipeBytes(buf)
}

// Destroy is self explanatory. It wipes and destroys the
// LockedBuffer. This function should be called on all secure
// values before exiting.
func (b *LockedBuffer) Destroy() {
	// Get the rounded size of our data
	roundedSize := len(b.mainSlice) - (pageSize * 2)

	// Make all the main slice readable and writable
	memcall.Protect(b.mainSlice, true, true)

	// Wipe the pages that hold our data
	WipeBytes(b.mainSlice[pageSize : pageSize+roundedSize])

	// Unlock the pages that hold our data
	memcall.Unlock(b.mainSlice[pageSize : pageSize+roundedSize])

	// Free all the mainSlice
	memcall.Free(b.mainSlice)
}

// WipeBytes zeroes out a byte slice.
func WipeBytes(buf []byte) {
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(0)
	}
}

func _roundPage(length int) int {
	return (length + (pageSize - 1)) & (^(pageSize - 1))
}

func _getBytes(ptr uintptr, len int, cap int) []byte {
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{ptr, len, cap}
	return *(*[]byte)(unsafe.Pointer(&sl))
}
