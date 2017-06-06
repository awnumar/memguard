package memguard

import (
	"bytes"
	"crypto/subtle"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unsafe"

	"github.com/awnumar/memguard/memcall"
)

var (
	// A slice that holds the canary we set.
	canary = getRandBytes(32)
)

/*
LockedBuffer is a structure that holds secure values. It
exposes a Mutex, various metadata flags, and a slice that
maps to the protected memory.

The number of LockedBuffers that you are able to create is
limited by how much memory your system kernel allows each
process to mlock/VirtualLock. Therefore you should call
Destroy on LockedBuffers that you no longer need, or simply
defer a Destroy call after creating a new LockedBuffer.

The entire memguard API handles and passes around pointers
to LockedBuffers, and so, for both security and convenience,
you should refrain from dereferencing a LockedBuffer.

If an API function that needs to edit a LockedBuffer is given
one marked as read-only, the call will return an ErrReadOnly.
Similarly, if a function is given a LockedBuffer that has been
destroyed, the call will return an ErrDestroyed.

For obvious reasons, you should never edit these metadata values
yourself. Doing so will result in undefined behaviour.
*/
type LockedBuffer struct {
	sync.Mutex
	Buffer    []byte
	ReadOnly  bool
	Destroyed bool
}

/*
New creates a new LockedBuffer of a specified length and
permissions.

If the given length is less than one, the call will return
an ErrInvalidLength.
*/
func New(length int, readOnly bool) (*LockedBuffer, error) {
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
	copy(memory[pageSize+roundedLength-length-32:pageSize+roundedLength-length], canary)

	// Set Buffer to a byte slice that describes the reigon of memory that is protected.
	b.Buffer = getBytes(uintptr(unsafe.Pointer(&memory[pageSize+roundedLength-length])), length)

	// Mark as read-only if requested.
	if readOnly {
		b.MarkAsReadOnly()
	}

	// Append this LockedBuffer to allLockedBuffers.
	allLockedBuffersMutex.Lock()
	allLockedBuffers = append(allLockedBuffers, b)
	allLockedBuffersMutex.Unlock()

	// Return a pointer to the LockedBuffer.
	return b, nil
}

/*
NewFromBytes is identical to New but for the fact that the created
LockedBuffer is of the same length and has the same contents as a
given slice. The slice is wiped after the bytes have been copied over.

If the size of the slice is zero, the call will return an ErrInvalidLength.
*/
func NewFromBytes(buf []byte, readOnly bool) (*LockedBuffer, error) {
	// Create a new LockedBuffer.
	b, err := New(len(buf), false)
	if err != nil {
		return nil, err
	}

	// Copy the bytes from buf, wiping afterwards.
	b.Move(buf)

	// Make it read-only if requested.
	if readOnly {
		b.MarkAsReadOnly()
	}

	// Return a pointer to the LockedBuffer.
	return b, nil
}

/*
NewRandom is identical to New but for the fact that the created
LockedBuffer is filled with cryptographically-secure pseudo-random
bytes instead of zeroes. Therefore a LockedBuffer created with
NewRandom can safely be used as an encryption key.
*/
func NewRandom(length int, readOnly bool) (*LockedBuffer, error) {
	// Create a new LockedBuffer for the key.
	b, err := New(length, false)
	if err != nil {
		return nil, err
	}

	// Fill it with random data.
	fillRandBytes(b.Buffer)

	// Mark as read-only if requested.
	if readOnly {
		b.MarkAsReadOnly()
	}

	// Return the LockedBuffer.
	return b, nil
}

/*
EqualTo compares a LockedBuffer to a byte slice in constant time.
*/
func (b *LockedBuffer) EqualTo(buf []byte) (bool, error) {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if b.Destroyed {
		return false, ErrDestroyed
	}

	// Do a time-constant comparison.
	if equal := subtle.ConstantTimeCompare(b.Buffer, buf); equal == 1 {
		// They're equal.
		return true, nil
	}

	// They're not equal.
	return false, nil
}

/*
MarkAsReadOnly asks the kernel to mark the LockedBuffer's
memory as read-only. Any subsequent attempts to write to
this memory will result in the process crashing with a
SIGSEGV memory violation.

To make the memory writable again, MarkAsReadWrite is called.
*/
func (b *LockedBuffer) MarkAsReadOnly() error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if b.Destroyed {
		return ErrDestroyed
	}

	// Check if it's already read-only.
	if b.ReadOnly {
		return nil
	}

	// Mark the memory as read-only.
	memoryToMark := getAllMemory(b)[pageSize : pageSize+roundToPageSize(len(b.Buffer)+32)]
	memcall.Protect(memoryToMark, true, false)

	// Tell everyone about the change we made.
	b.ReadOnly = true

	// Everything went well.
	return nil
}

/*
MarkAsReadWrite asks the kernel to mark the LockedBuffer's
memory as readable and writable.

This method is the counterpart of MarkAsReadOnly.
*/
func (b *LockedBuffer) MarkAsReadWrite() error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if b.Destroyed {
		return ErrDestroyed
	}

	// Check if it's already readable and writable.
	if !b.ReadOnly {
		return nil
	}

	// Mark the memory as readable and writable.
	memoryToMark := getAllMemory(b)[pageSize : pageSize+roundToPageSize(len(b.Buffer)+32)]
	memcall.Protect(memoryToMark, true, true)

	// Tell everyone about the change we made.
	b.ReadOnly = false

	// Everything went well.
	return nil
}

/*
Copy copies bytes from a byte slice into a LockedBuffer in
constant-time. Just like Golang's built-in copy function,
Copy only copies up to the smallest of the two buffers.

It does not wipe the original slice so using Copy is less
secure than using Move. Therefore Move should be favoured
unless you have a good reason.

You should aim to call WipeBytes on the original slice as
soon as possible.

If the LockedBuffer is marked as read-only, the call will
fail and return an ErrReadOnly.
*/
func (b *LockedBuffer) Copy(buf []byte) error {
	// Just call CopyAt with a zero offset.
	return b.CopyAt(buf, 0)
}

/*
CopyAt is identical to Copy but it copies into the LockedBuffer
at a specified offset.
*/
func (b *LockedBuffer) CopyAt(buf []byte, offset int) error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if b.Destroyed {
		return ErrDestroyed
	}

	// Check if it's marked as ReadOnly.
	if b.ReadOnly {
		return ErrReadOnly
	}

	// Do a time-constant copying of the bytes, copying only up to the length of the buffer.
	if len(b.Buffer[offset:]) > len(buf) {
		subtle.ConstantTimeCopy(1, b.Buffer[offset:len(buf)], buf)
	} else if len(b.Buffer[offset:]) < len(buf) {
		subtle.ConstantTimeCopy(1, b.Buffer[offset:], buf[:len(b.Buffer[offset:])])
	} else {
		subtle.ConstantTimeCopy(1, b.Buffer[offset:], buf)
	}

	return nil
}

/*
Move moves bytes from a byte slice into a LockedBuffer in
constant-time. Just like Golang's built-in copy function,
Move only moves up to the smallest of the two buffers.

Unlike Copy, Move wipes the entire original slice after
copying the appropriate number of bytes over, and so it
should be favoured unless you have a good reason.

If the LockedBuffer is marked as read-only, the call will
fail and return an ErrReadOnly.
*/
func (b *LockedBuffer) Move(buf []byte) error {
	// Just call MoveAt with a zero offset.
	return b.MoveAt(buf, 0)
}

/*
MoveAt is identical to Move but it copies into the LockedBuffer
at a specified offset.
*/
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
FillRandomBytes fills a LockedBuffer with cryptographically-secure
pseudo-random bytes.
*/
func (b *LockedBuffer) FillRandomBytes() error {
	// Just call FillRandomBytesAt.
	return b.FillRandomBytesAt(0, len(b.Buffer))
}

/*
FillRandomBytesAt fills a LockedBuffer with cryptographically-secure
pseudo-random bytes, starting at an offset and ending after a given
number of bytes.
*/
func (b *LockedBuffer) FillRandomBytesAt(offset, length int) error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if b.Destroyed {
		return ErrDestroyed
	}

	// Check if it's marked as ReadOnly.
	if b.ReadOnly {
		return ErrReadOnly
	}

	// Fill with random bytes.
	fillRandBytes(b.Buffer[offset : offset+length])

	// Everything went well.
	return nil
}

/*
Destroy verifies that no buffer underflows occurred and then wipes,
unlocks, and frees all related memory. If a buffer underflow is
detected, the process panics.

This function must be called on all LockedBuffers before exiting.
DestroyAll is designed for this purpose, as is CatchInterrupt and
SafeExit. We recommend using all of them together.

If the LockedBuffer has already been destroyed then the call
makes no changes.
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

		// Verify the canary.
		if !bytes.Equal(memory[pageSize+roundedLength-len(b.Buffer)-32:pageSize+roundedLength-len(b.Buffer)], canary) {
			panic("memguard.Destroy(): buffer underflow detected")
		}

		// Make all of the memory readable and writable.
		memcall.Protect(memory, true, true)

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

/*
DestroyAll calls Destroy on all LockedBuffers that have not already
been destroyed.

CatchInterrupt and SafeExit both call DestroyAll before exiting.
*/
func DestroyAll() {
	// Get a copy of allLockedBuffers.
	toDestroy := LockedBuffers()

	// Call destroy on each LockedBuffer.
	for _, v := range toDestroy {
		v.Destroy()
	}
}

/*
Concatenate takes two LockedBuffers and concatenates them.

If one of the given LockedBuffers is read-only, the resulting
LockedBuffer will also be read-only. The original LockedBuffers
are not destroyed.
*/
func Concatenate(a, b *LockedBuffer) (*LockedBuffer, error) {
	// Get a mutex lock on the LockedBuffers.
	a.Lock()
	b.Lock()
	defer a.Unlock()
	defer b.Unlock()

	// Check if either are destroyed.
	if a.Destroyed || b.Destroyed {
		return nil, ErrDestroyed
	}

	// Create a new LockedBuffer to hold the concatenated value.
	c, _ := New(len(a.Buffer)+len(b.Buffer), false)

	// Copy the values across.
	c.Copy(a.Buffer)
	c.CopyAt(b.Buffer, len(a.Buffer))

	// Set permissions accordingly.
	if a.ReadOnly || b.ReadOnly {
		c.MarkAsReadOnly()
	}

	// Return the resulting LockedBuffer.
	return c, nil
}

/*
Duplicate takes a LockedBuffer and creates a new one with
the same contents and permissions as the original.
*/
func Duplicate(b *LockedBuffer) (*LockedBuffer, error) {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if b.Destroyed {
		return nil, ErrDestroyed
	}

	// Create new LockedBuffer.
	newBuf, _ := New(len(b.Buffer), false)

	// Copy bytes into it.
	newBuf.Copy(b.Buffer)

	// Set permissions accordingly.
	if b.ReadOnly {
		newBuf.MarkAsReadOnly()
	}

	// Return duplicated.
	return newBuf, nil
}

/*
Equal compares the contents of two LockedBuffers in constant time.
*/
func Equal(a, b *LockedBuffer) (bool, error) {
	// Get a mutex lock on the LockedBuffers.
	a.Lock()
	b.Lock()
	defer a.Unlock()
	defer b.Unlock()

	// Check if either are destroyed.
	if a.Destroyed || b.Destroyed {
		return false, ErrDestroyed
	}

	// Do a time-constant comparison on the two buffers.
	if equal := subtle.ConstantTimeCompare(a.Buffer, b.Buffer); equal == 1 {
		// They're equal.
		return true, nil
	}

	// They're not equal.
	return false, nil
}

/*
Split takes a LockedBuffer, splits it at a specified offset, and
then returns the two newly created LockedBuffers. The permissions
of the original are preserved and the original LockedBuffer is not
destroyed.
*/
func Split(b *LockedBuffer, offset int) (*LockedBuffer, *LockedBuffer, error) {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if b.Destroyed {
		return nil, nil, ErrDestroyed
	}

	// Create two new LockedBuffers.
	firstBuf, err := New(len(b.Buffer[:offset]), false)
	if err != nil {
		return nil, nil, err
	}

	secondBuf, err := New(len(b.Buffer[offset:]), false)
	if err != nil {
		firstBuf.Destroy()
		return nil, nil, err
	}

	// Copy the values into them.
	firstBuf.Copy(b.Buffer[:offset])
	secondBuf.Copy(b.Buffer[offset:])

	// Copy over permissions.
	if b.ReadOnly {
		firstBuf.MarkAsReadOnly()
		secondBuf.MarkAsReadOnly()
	}

	// Return the new LockedBuffers.
	return firstBuf, secondBuf, nil
}

/*
Trim shortens a LockedBuffer according to the given specifications.
The permissions of the original are preserved and the original
LockedBuffer is not destroyed.

Trim takes an offset and a size as arguments. The resulting LockedBuffer
starts at index [offset] and ends at index [offset+size].
*/
func Trim(b *LockedBuffer, offset, size int) (*LockedBuffer, error) {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if b.Destroyed {
		return nil, ErrDestroyed
	}

	// Create new LockedBuffer and copy over the old.
	newBuf, err := New(size, false)
	if err != nil {
		return nil, err
	}
	newBuf.Copy(b.Buffer[offset : offset+size])

	// Copy over permissions.
	if b.ReadOnly {
		newBuf.MarkAsReadOnly()
	}

	// Return the new LockedBuffer.
	return newBuf, nil
}

/*
LockedBuffers returns a slice containing a pointer to
each LockedBuffer that has not been destroyed.
*/
func LockedBuffers() []*LockedBuffer {
	// Get a Mutex lock on allLockedBuffers, and get a copy.
	allLockedBuffersMutex.Lock()
	lockedBuffers := make([]*LockedBuffer, len(allLockedBuffers))
	copy(lockedBuffers, allLockedBuffers)
	allLockedBuffersMutex.Unlock()

	// Return this copy.
	return lockedBuffers
}

/*
CatchInterrupt starts a goroutine that monitors for
interrupt signals. It accepts a function of type func()
and executes that before calling SafeExit(0).

If CatchInterrupt is called multiple times, only the first
call is executed and all subsequent calls are ignored.
*/
func CatchInterrupt(f func()) {
	// Only do this if it hasn't been done before.
	catchInterruptOnce.Do(func() {
		// Create a channel to listen on.
		c := make(chan os.Signal, 2)

		// Start a goroutine to listen on the channel.
		go func() {
			<-c         // Wait for signal.
			f()         // Execute user function.
			SafeExit(0) // Exit securely.
		}()

		// Notify the channel if we receive a signal.
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	})
}

/*
SafeExit exits the program with a specified exit-code, but calls DestroyAll first.
*/
func SafeExit(c int) {
	// Cleanup protected memory.
	DestroyAll()

	// Exit with a specified exit-code.
	os.Exit(c)
}

/*
WipeBytes wipes a byte slice with zeroes.
*/
func WipeBytes(buf []byte) {
	// Iterate over the slice...
	for i := 0; i < len(buf); i++ {
		// ... setting each element to zero.
		buf[i] = byte(0)
	}
}

/*
DisableUnixCoreDumps disables core-dumps.

Since core-dumps are only relevant on Unix systems,
if DisableUnixCoreDumps is called on any other system it
will do nothing and return immediately.

This function is precautonary as core-dumps are usually
disabled by default on most systems.
*/
func DisableUnixCoreDumps() {
	memcall.DisableCoreDumps()
}
