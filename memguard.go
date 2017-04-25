package memguard

import (
	"os"
	"github.com/libeclipse/memguard/memcall"
)

// LockedBuffer implements a buffer that stores the data.
type LockedBuffer struct {
	Buffer []byte
	main_slice []byte
}

var page_size int

func Init() {
	page_size = os.Getpagesize()
}

func round_page(length int) int{
	return (length + (page_size - 1)) & (^(page_size - 1)) 
}

// New creates a new *LockedBuffer and returns it.
func New(length int) *LockedBuffer {
	// Allocate the new LockedBuffer
	b := new(LockedBuffer)

	// Round length to page_size
	rounded_length := round_page(length)

	// Set Total Size with guard pages
	total_size := page_size + rounded_length + page_size

	// Allocate it all
	main_slice := memcall.Alloc(total_size)

	//Lock the page that will hold our data
	memcall.Lock(main_slice[page_size:page_size + rounded_length])

	// Make the Guard Pages inaccessible
	memcall.Protect(main_slice[:page_size], false, false)
	memcall.Protect(main_slice[page_size + rounded_length: total_size], false, false)

	// Set the user pointer
	b.Buffer = main_slice[page_size + rounded_length - length:page_size + rounded_length]

	// Save the address (needed when freeing)
	b.main_slice = main_slice[:]

	// Return a pointer to the LockedBuffer.
	return b
}

// NewFromBytes creates a new *LockedBuffer from a byte slice,
// attempting to destroy the old value before returning. It is
// not as robust as New(), but sometimes it is necessary.
func NewFromBytes(buf []byte) *LockedBuffer {
	// Use New to create a Secured LockedBuffer
	b := New(len(buf))

	// Copy the bytes from the old slice
	copy(b.Buffer, buf)

	// Wipe the old bytes and set to nil
	WipeBytes(buf)

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
/*func (b *LockedBuffer) Copy(buf []byte) {
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
}*/

// Destroy is self explanatory. It wipes and destroys the
// LockedBuffer. This function should be called on all secure
// values before exiting.
func (b *LockedBuffer) Destroy() {
	// Get the rounded size of our data
	rounded_size := len(b.main_slice) - (page_size * 2)

	// Make all the main slice readable and writable
	memcall.Protect(b.main_slice, true, true)

	// Wipe the pages that hold our data
	WipeBytes(b.main_slice[page_size : page_size + rounded_size])

	// Unlock the pages that hold our data
	memcall.Unlock(b.main_slice[page_size : page_size + rounded_size])

	// Free all the main_slice
	memcall.Free(b.main_slice)
}

// WipeBytes zeroes out a byte slice.
func WipeBytes(buf []byte) {
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(0)
	}
}
