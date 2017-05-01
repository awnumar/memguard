package memguard

import (
	"bytes"
	"crypto/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unsafe"

	"github.com/libeclipse/memguard/memcall"
)

var (
	// Store pointers to all of the LockedBuffers.
	allLockedBuffers      []*LockedBuffer
	allLockedBuffersMutex = &sync.Mutex{}

	// Mutex for the DestroyAll function.
	destroyAllMutex = &sync.Mutex{}

	// A slice that holds the canary we set.
	canary = _csprng(32)

	// Grab the system page size.
	pageSize = os.Getpagesize()
)

// LockedBuffer implements a buffer that stores the data.
type LockedBuffer struct {
	// Mutex for access to this struct.
	sync.Mutex

	// The buffer that holds the secure data.
	Buffer []byte

	// Holds the current protection value of Buffer.
	// Possible values are `ReadWrite` and `ReadOnly`.
	State string
}

// ExitFunc is passed to CatchInterrupt and is executed by
// CatchInterrupt before cleaning up memory and exiting.
type ExitFunc func()

// New creates a new *LockedBuffer and returns it. The
// LockedBuffer is in an unlocked state. Length
// must be > zero.
func New(length int) *LockedBuffer {
	// Panic if length < one.
	if length < 1 {
		panic("memguard.New(): length must be > zero")
	}

	// Allocate a new LockedBuffer.
	b := new(LockedBuffer)

	// Round length + 32 bytes for the canary to a multiple of the page size..
	roundedLength := _roundToPageSize(length + 32)

	// Calculate the total size of memory including the guard pages.
	totalSize := (2 * pageSize) + roundedLength

	// Allocate it all.
	memory := memcall.Alloc(totalSize)

	// Lock the pages that will hold the sensitive data.
	memcall.Lock(memory[pageSize : pageSize+roundedLength])

	// Make the guard pages inaccessible.
	memcall.Protect(memory[:pageSize], false, false)
	memcall.Protect(memory[pageSize+roundedLength:], false, false)

	// Generate and set the canary.
	copy(memory[pageSize+roundedLength-length-32:pageSize+roundedLength-length], canary)

	// Set Buffer to a byte slice that describes the reigon of memory that is protected.
	b.Buffer = _getBytes(uintptr(unsafe.Pointer(&memory[pageSize+roundedLength-length])), length)

	// Set the correct protection value to exported State field.
	b.State = "ReadWrite"

	// Append this LockedBuffer to allLockedBuffers.
	allLockedBuffersMutex.Lock()
	allLockedBuffers = append(allLockedBuffers, b)
	allLockedBuffersMutex.Unlock()

	// Return a pointer to the LockedBuffer.
	return b
}

// NewFromBytes creates a new *LockedBuffer from a byte slice,
// attempting to destroy the old value before returning. It is
// not as robust as New(), but sometimes it is necessary.
func NewFromBytes(buf []byte) *LockedBuffer {
	// Use New to create a Secured LockedBuffer.
	b := New(len(buf))

	// Copy the bytes from buf, wiping afterwards.
	b.Move(buf)

	// Return a pointer to the LockedBuffer.
	return b
}

// ReadWrite makes the buffer readable and writable.
// This is the default state of new LockedBuffers.
func (b *LockedBuffer) ReadWrite() {
	b.Lock()
	defer b.Unlock()

	memory := _getAllMemory(b)
	memcall.Protect(memory[pageSize:pageSize+_roundToPageSize(len(b.Buffer)+32)], true, true)
	b.State = "ReadWrite"
}

// ReadOnly makes the buffer read-only.
// Anything else triggers a SIGSEGV violation.
func (b *LockedBuffer) ReadOnly() {
	b.Lock()
	defer b.Unlock()

	memory := _getAllMemory(b)
	memcall.Protect(memory[pageSize:pageSize+_roundToPageSize(len(b.Buffer)+32)], true, false)
	b.State = "ReadOnly"
}

// Copy copies bytes from a byte slice into a LockedBuffer,
// preserving the original slice. This is insecure and so
// Move() should be favoured generally.
func (b *LockedBuffer) Copy(buf []byte) {
	b.Lock()
	defer b.Unlock()

	copy(b.Buffer, buf)
}

// Move copies bytes from a byte slice into a LockedBuffer,
// destroying the original slice.
func (b *LockedBuffer) Move(buf []byte) {
	// Copy buf into the LockedBuffer.
	b.Copy(buf)

	// Wipe the old bytes.
	WipeBytes(buf)
}

// Destroy is self explanatory. It wipes and destroys the
// LockedBuffer. This function should be called on all secure
// values before exiting. If the LockedBuffer has already been
// destroyed, then nothing happens and the function returns.
func (b *LockedBuffer) Destroy() {
	// Remove this one from global slice.
	var exists bool
	allLockedBuffersMutex.Lock()
	for i, v := range allLockedBuffers {
		if v == b {
			allLockedBuffers = append(allLockedBuffers[:i], allLockedBuffers[i+1:]...)
			exists = true
			break
		}
	}
	allLockedBuffersMutex.Unlock()

	if exists {
		// Attain a Mutex lock to this LockedBuffer first.
		b.Lock()
		defer b.Unlock()

		// Get all of the memory related to this LockedBuffer.
		memory := _getAllMemory(b)

		// Get the total size of all the pages between the guards.
		roundedLength := len(memory) - (pageSize * 2)

		// Make all of the memory readable and writable.
		memcall.Protect(memory, true, true)

		// Verify the canary.
		if !bytes.Equal(memory[pageSize+roundedLength-len(b.Buffer)-32:pageSize+roundedLength-len(b.Buffer)], canary) {
			panic("memguard.Destroy(): buffer underflow detected; canary has changed")
		}

		// Wipe the pages that hold our data.
		WipeBytes(memory[pageSize : pageSize+roundedLength])

		// Unlock the pages that hold our data.
		memcall.Unlock(memory[pageSize : pageSize+roundedLength])

		// Free all related memory.
		memcall.Free(memory)

		// Set the State back to an empty string.
		b.State = ""

		// Set b.Buffer to nil.
		b.Buffer = nil
	}
}

// DestroyAll calls Destroy on all created LockedBuffers.
// This function can be called even if no LockedBuffers exist.
func DestroyAll() {
	// Only allow one routine to DestroyAll at a time.
	destroyAllMutex.Lock()
	defer destroyAllMutex.Unlock()

	// Get a Mutex lock on allLockedBuffers, and get the length.
	allLockedBuffersMutex.Lock()
	toDestroy := make([]*LockedBuffer, len(allLockedBuffers))
	copy(toDestroy, allLockedBuffers)
	allLockedBuffersMutex.Unlock()

	// Call destroy on each LockedBuffer.
	for _, v := range toDestroy {
		v.Destroy()
	}
}

// CatchInterrupt starts a goroutine that monitors for
// interrupt signals. It accepts a function of type ExitFunc
// and executes that before calling SafeExit(0).
func CatchInterrupt(f ExitFunc) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c         // Wait for signal.
		f()         // Execute user function.
		SafeExit(0) // Exit securely.
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

// Round a length to a multiple of the system page size.
func _roundToPageSize(length int) int {
	return (length + (pageSize - 1)) & (^(pageSize - 1))
}

// Get a slice that describes all memory related to a LockedBuffer.
func _getAllMemory(b *LockedBuffer) []byte {
	bufLen, roundedBufLen := len(b.Buffer), _roundToPageSize(len(b.Buffer)+32)
	memAddr := uintptr(unsafe.Pointer(&b.Buffer[0])) - uintptr((roundedBufLen-bufLen)+pageSize)
	memLen := (pageSize * 2) + roundedBufLen
	return _getBytes(memAddr, memLen)
}

// Convert a pointer and length to a byte slice that describes that memory.
func _getBytes(ptr uintptr, len int) []byte {
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{ptr, len, len}
	return *(*[]byte)(unsafe.Pointer(&sl))
}

// Cryptographically Secure Pseudo-Random Number Generator.
func _csprng(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("memguard._csprng(): could not get random bytes")
	}
	return b
}
