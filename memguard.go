package memguard

import (
	"bytes"
	"crypto/subtle"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unsafe"

	"github.com/libeclipse/memguard/memcall"
)

var (
	// Are we listening for interrupts?
	monInterrupt bool

	// Store pointers to all of the LockedBuffers.
	allLockedBuffers      []*LockedBuffer
	allLockedBuffersMutex = &sync.Mutex{}

	// Mutex for the DestroyAll function.
	destroyAllMutex = &sync.Mutex{}

	// A slice that holds the canary we set.
	canary = csprng(32)

	// Grab the system page size.
	pageSize = os.Getpagesize()
)

// LockedBuffer implements a structure that holds protected values.
type LockedBuffer struct {
	// Exposed mutex for implementing thread-safety
	// both within and outside of the API.
	sync.Mutex

	// Buffer holds the secure values themselves.
	Buffer []byte

	// A boolean flag indicating if this memory has
	// been marked as ReadOnly by a call to b.ReadOnly()
	ReadOnly bool

	// A boolean flag indicating whether this
	// LockedBuffer has been destroyed. No API
	// calls succeed on a destroyed buffer.
	Destroyed bool
}

/*
ExitFunc is a function type that takes no arguments and returns
no values. It is passed to CatchInterrupt which executes it before
terminating the application securely.
*/
type ExitFunc func()

// New creates a new *LockedBuffer and returns it. The
// LockedBuffer's state is `ReadWrite`. Length
// must be greater than zero.
func New(length int) (*LockedBuffer, error) {
	// Panic if length < one.
	if length < 1 {
		return nil, ErrInvalidLength
	}

	// Allocate a new LockedBuffer.
	b := new(LockedBuffer)

	// Round length + 32 bytes for the canary to a multiple of the page size..
	roundedLength := roundToPageSize(length + 32)

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
	subtle.ConstantTimeCopy(1, memory[pageSize+roundedLength-length-32:pageSize+roundedLength-length], canary)

	// Set Buffer to a byte slice that describes the reigon of memory that is protected.
	b.Buffer = getBytes(uintptr(unsafe.Pointer(&memory[pageSize+roundedLength-length])), length)

	// Append this LockedBuffer to allLockedBuffers.
	allLockedBuffersMutex.Lock()
	allLockedBuffers = append(allLockedBuffers, b)
	allLockedBuffersMutex.Unlock()

	// Return a pointer to the LockedBuffer.
	return b, nil
}

// NewFromBytes creates a new *LockedBuffer from a byte slice,
// attempting to destroy the old value before returning. It is
// identicle to calling New() followed by Move().
func NewFromBytes(buf []byte) (*LockedBuffer, error) {
	// Use New to create a Secured LockedBuffer.
	b, err := New(len(buf))
	if err != nil {
		return nil, err
	}

	// Copy the bytes from buf, wiping afterwards.
	b.Move(buf)

	// Return a pointer to the LockedBuffer.
	return b, nil
}

// EqualTo compares a LockedBuffer to a byte slice in constant time.
func (b *LockedBuffer) EqualTo(buf []byte) (bool, error) {
	b.Lock()
	defer b.Unlock()

	if b.Destroyed {
		return false, ErrDestroyed
	}

	if equal := subtle.ConstantTimeCompare(b.Buffer, buf); equal == 1 {
		return true, nil
	}

	return false, nil
}

// MarkAsReadWrite makes the buffer readable and writable.
// This is the default state of new LockedBuffers.
func (b *LockedBuffer) MarkAsReadWrite() error {
	b.Lock()
	defer b.Unlock()

	if b.Destroyed {
		return ErrDestroyed
	}

	// Mark the memory as readable and writable.
	memoryToMark := getAllMemory(b)[pageSize : pageSize+roundToPageSize(len(b.Buffer)+32)]
	memcall.Protect(memoryToMark, true, true)

	// Tell everyone about the change we made.
	b.ReadOnly = false

	return nil
}

// MarkAsReadOnly makes the buffer read-only. After setting
// this, any other action will trigger a SIGSEGV violation.
func (b *LockedBuffer) MarkAsReadOnly() error {
	b.Lock()
	defer b.Unlock()

	if b.Destroyed {
		return ErrDestroyed
	}

	// Mark the memory as read-only.
	memoryToMark := getAllMemory(b)[pageSize : pageSize+roundToPageSize(len(b.Buffer)+32)]
	memcall.Protect(memoryToMark, true, false)

	// Tell everyone about the change we made.
	b.ReadOnly = true

	return nil
}

// Copy copies bytes from a byte slice into a LockedBuffer,
// preserving the original slice. This is insecure, and so
// Move() should be favoured unless you have a specific need.
// You should aim to call WipeBytes(buf) as soon as possible.
func (b *LockedBuffer) Copy(buf []byte) error {
	return b.CopyAt(buf, 0)
}

// CopyAt copies bytes from a byte slice into a LockedBuffer,
// preserving the original slice. This is insecure, and so
// Move() should be favoured unless you have a specific need.
// It also takes an offset, and starts copying at that index.
// You should aim to call WipeBytes(buf) as soon as possible.
func (b *LockedBuffer) CopyAt(buf []byte, offset int) error {
	b.Lock()
	defer b.Unlock()

	if b.Destroyed {
		return ErrDestroyed
	}

	if b.ReadOnly {
		return ErrReadOnly
	}

	subtle.ConstantTimeCopy(1, b.Buffer[offset:], buf[:len(b.Buffer[offset:])])

	return nil
}

// Move copies bytes from a byte slice into a LockedBuffer,
// wiping the original slice afterwards.
func (b *LockedBuffer) Move(buf []byte) error {
	return b.MoveAt(buf, 0)
}

// MoveAt copies bytes from a byte slice into a LockedBuffer,
// wiping the original slice afterwards. It also takes an
// offset, and starts copying at that index.
func (b *LockedBuffer) MoveAt(buf []byte, offset int) error {
	// Copy buf into the LockedBuffer.
	if err := b.CopyAt(buf, offset); err != nil {
		return err
	}

	// Wipe the old bytes.
	WipeBytes(buf)

	// Everything went well.
	return nil
}

/*
Destroy verifies that everything went well, wipes the Buffer,
and then unlocks and frees all related memory. This function
should be called on all LockedBuffers before exiting. If the
LockedBuffer has already been destroyed, then nothing happens
and the function returns.
*/
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
		memory := getAllMemory(b)

		// Get the total size of all the pages between the guards.
		roundedLength := len(memory) - (pageSize * 2)

		// Make all of the memory readable and writable.
		memcall.Protect(memory, true, true)

		// Verify the canary.
		if !bytes.Equal(memory[pageSize+roundedLength-len(b.Buffer)-32:pageSize+roundedLength-len(b.Buffer)], canary) {
			panic("memguard.Destroy(): buffer underflow detected")
		}

		// Wipe the pages that hold our data.
		WipeBytes(memory[pageSize : pageSize+roundedLength])

		// Unlock the pages that hold our data.
		memcall.Unlock(memory[pageSize : pageSize+roundedLength])

		// Free all related memory.
		memcall.Free(memory)

		// Set the metadata appropriately.
		b.ReadOnly = false
		b.Destroyed = true

		// Set the buffer to nil.
		b.Buffer = nil
	}
}

// DestroyAll calls Destroy on all LockedBuffers. This
// function can be called even if no LockedBuffers exist.
func DestroyAll() {
	// Only allow one routine to DestroyAll at a time.
	destroyAllMutex.Lock()
	defer destroyAllMutex.Unlock()

	// Get a Mutex lock on allLockedBuffers, and get a copy.
	allLockedBuffersMutex.Lock()
	toDestroy := make([]*LockedBuffer, len(allLockedBuffers))
	copy(toDestroy, allLockedBuffers)
	allLockedBuffersMutex.Unlock()

	// Call destroy on each LockedBuffer.
	for _, v := range toDestroy {
		v.Destroy()
	}
}

// Duplicate takes a LockedBuffer as an argument and creates
// a new one with the same contents and permissions.
func Duplicate(b *LockedBuffer) (*LockedBuffer, error) {
	b.Lock()
	defer b.Unlock()

	if b.Destroyed {
		return nil, ErrDestroyed
	}

	// Create new LockedBuffer.
	newBuf, _ := New(len(b.Buffer))

	// Copy bytes into it.
	newBuf.Copy(b.Buffer)

	// Set permissions accordingly.
	if b.ReadOnly {
		memoryToMark := getAllMemory(newBuf)[pageSize : pageSize+roundToPageSize(len(newBuf.Buffer)+32)]
		memcall.Protect(memoryToMark, true, false)
		newBuf.ReadOnly = true
	}

	// Return duplicated.
	return newBuf, nil
}

// Equal compares the contents of two LockedBuffers in constant time.
// The LockedBuffers' respective permissions are ignored.
func Equal(a, b *LockedBuffer) (bool, error) {
	a.Lock()
	defer a.Unlock()
	b.Lock()
	defer b.Unlock()

	if a.Destroyed || b.Destroyed {
		return false, ErrDestroyed
	}

	if equal := subtle.ConstantTimeCompare(a.Buffer, b.Buffer); equal == 1 {
		return true, nil
	}

	return false, nil
}

// Split takes a LockedBuffer and splits it at a specified offset,
// then returning the two created LockedBuffers. The permissions
// of the original are copied over, and the original is preserved.
// This can be called with a LockedBuffer that is marked ReadOnly.
func Split(b *LockedBuffer, offset int) (*LockedBuffer, *LockedBuffer, error) {
	b.Lock()
	defer b.Unlock()

	if b.Destroyed {
		return nil, nil, ErrDestroyed
	}

	firstBuf, _ := NewFromBytes(b.Buffer[:offset])
	secondBuf, _ := NewFromBytes(b.Buffer[offset:])

	if b.ReadOnly {
		memoryToMark := getAllMemory(firstBuf)[pageSize : pageSize+roundToPageSize(len(firstBuf.Buffer)+32)]
		memcall.Protect(memoryToMark, true, false)
		firstBuf.ReadOnly = true

		memoryToMark = getAllMemory(secondBuf)[pageSize : pageSize+roundToPageSize(len(secondBuf.Buffer)+32)]
		memcall.Protect(memoryToMark, true, false)
		secondBuf.ReadOnly = true
	}

	return firstBuf, secondBuf, nil
}

// Trim shortens a LockedBuffer to a specified size. The returned
// `LockedBuffer.Buffer` is equal to `b.Buffer[offset:offset+size]`.
// This can be called with a LockedBuffer that is marked ReadOnly.
// The permissions of the original LockedBuffer are also copied over.
func Trim(b *LockedBuffer, offset, size int) (*LockedBuffer, error) {
	b.Lock()
	defer b.Unlock()

	if b.Destroyed {
		return nil, ErrDestroyed
	}

	// Create new LockedBuffer and copy over the old.
	newBuf, _ := New(size)
	newBuf.Copy(b.Buffer[offset : offset+size])

	// Copy over permissions.
	if b.ReadOnly {
		memoryToMark := getAllMemory(newBuf)[pageSize : pageSize+roundToPageSize(len(newBuf.Buffer)+32)]
		memcall.Protect(memoryToMark, true, false)
		newBuf.ReadOnly = true
	}

	// Return the new LockedBuffer.
	return newBuf, nil
}

/*
CatchInterrupt starts a goroutine that monitors for
interrupt signals. It accepts a function of type ExitFunc
and executes that before calling SafeExit(0).

	memguard.CatchInterrupt(func() {
		fmt.Println("Interrupt signal received. Exiting...")
	})

If CatchInterrupt is called multiple times, only the first
call is executed and all subsequent calls are ignored.
*/
func CatchInterrupt(f ExitFunc) {
	if !monInterrupt {
		monInterrupt = true
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c         // Wait for signal.
			f()         // Execute user function.
			SafeExit(0) // Exit securely.
		}()
	}
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
