package memguard

import (
	"crypto/rand"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/libeclipse/memguard/memcall"
)

var (
	// Grab the system page size.
	pageSize = os.Getpagesize()

	// Store pointers to all of the LockedBuffers.
	allLockedBuffers []*LockedBuffer
)

// LockedBuffer implements a buffer that stores the data.
type LockedBuffer struct {
	Buffer []byte // The buffer that holds the secure data.

	//canary []byte // A slice that holds the canary we set.
	memory []byte // A slice that holds all related memory.
}

// New creates a new *LockedBuffer and returns it. The
// LockedBuffer is in an unlocked state. Length
// must be > zero.
func New(length int) *LockedBuffer {
	// Panic if length < one.
	if length < 1 {
		panic("memguard.New(): length must be > zero")
	}

	// Allocate the new LockedBuffer.
	b := new(LockedBuffer)

	// Round length to pageSize.
	roundedLength := _roundToPageSize(length)

	// Set Total Size with guard pages.
	totalSize := (2 * pageSize) + roundedLength

	// Allocate it all.
	memory := memcall.Alloc(totalSize)

	//Lock the page that will hold our data.
	memcall.Lock(memory[pageSize : pageSize+roundedLength])

	// Make the Guard Pages inaccessible.
	memcall.Protect(memory[:pageSize], false, false)
	memcall.Protect(memory[pageSize+roundedLength:], false, false)

	// Set Buffer to a byte slice that describes the reigon of memory that is protected.
	b.Buffer = _getBytes(uintptr(unsafe.Pointer(&memory[pageSize+roundedLength-length])), length, length)

	// Set memory to all related data so that we can retrieve it when destroying.
	b.memory = memory

	// Append this LockedBuffer to allLockedBuffers.
	allLockedBuffers = append(allLockedBuffers, b)

	// Return a pointer to the LockedBuffer.
	return b
}

// NewFromBytes creates a new *LockedBuffer from a byte slice,
// attempting to destroy the old value before returning. It is
// not as robust as New(), but sometimes it is necessary.
func NewFromBytes(buf []byte) *LockedBuffer {
	// Use New to create a Secured LockedBuffer.
	b := New(len(buf))

	// Copy the bytes from buf, wiping the afterwards.
	b.Move(buf)

	// Return a pointer to the LockedBuffer.
	return b
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

	// Wipe the old bytes.
	WipeBytes(buf)
}

// Destroy is self explanatory. It wipes and destroys the
// LockedBuffer. This function should be called on all secure
// values before exiting.
func (b *LockedBuffer) Destroy() {
	// Remove this one from global slice.
	for i, v := range allLockedBuffers {
		if v == b {
			allLockedBuffers = append(allLockedBuffers[:i], allLockedBuffers[i+1:]...)
			break
		}
	}

	// Get the rounded size of our data
	roundedLength := len(b.memory) - (pageSize * 2)

	// Make all the main slice readable and writable
	memcall.Protect(b.memory, true, true)

	// Wipe the pages that hold our data
	WipeBytes(b.memory[pageSize : pageSize+roundedLength])

	// Unlock the pages that hold our data
	memcall.Unlock(b.memory[pageSize : pageSize+roundedLength])

	// Free all the main_slice
	memcall.Free(b.memory)

	// Set b.Buffer to nil.
	b.Buffer = nil
}

// DestroyAll calls Destroy on all created LockedBuffers.
func DestroyAll() {
	for _, v := range allLockedBuffers {
		v.Destroy()
	}
}

// CatchInterrupt starts a goroutine that monitors for
// interrupt signals and calls Cleanup() before exiting.
func CatchInterrupt() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		SafeExit(0)
	}()
}

// SafeExit exits the program with the specified return code,
// but calls DestroyAll before doing so.
func SafeExit(c int) {
	// Cleanup protected memory.
	DestroyAll()

	// Exit with a specified exit-code.
	os.Exit(c)
}

// WipeBytes zeroes out a byte slice.
func WipeBytes(buf []byte) {
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(0)
	}
}

// DisableCoreDumps disables core dumps on Unix systems. On windows it is a no-op.
func DisableCoreDumps() {
	memcall.DisableCoreDumps()
}

func _roundToPageSize(length int) int {
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

func _csprng(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("memguard._csprng(): could not get random bytes")
	}
	return b
}
