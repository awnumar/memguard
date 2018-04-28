package memguard

import (
	"bytes"
	"crypto/subtle"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/awnumar/memguard/memcall"
)

/*
NewImmutable creates a new, immutable LockedBuffer of a specified size.

The mutability can later be toggled with the MakeImmutable and MakeMutable methods.

If the given length is less than one, the call will return an ErrInvalidLength.
*/
func NewImmutable(size int) (*LockedBuffer, error) {
	return newContainer(size, false)
}

/*
NewMutable creates a new, mutable LockedBuffer of a specified length.

The mutability can later be toggled with the MakeImmutable and MakeMutable methods.

If the given length is less than one, the call will return an ErrInvalidLength.
*/
func NewMutable(size int) (*LockedBuffer, error) {
	return newContainer(size, true)
}

/*
NewImmutableFromBytes is identical to NewImmutable but for the fact that the created LockedBuffer is of the same length and has the same contents as a given slice. The slice is wiped after the bytes have been copied over.

If the size of the slice is zero, the call will return an ErrInvalidLength.
*/
func NewImmutableFromBytes(buf []byte) (*LockedBuffer, error) {
	// Create a new LockedBuffer.
	b, err := NewMutableFromBytes(buf)
	if err != nil {
		return nil, err
	}

	// Mark as immutable.
	b.MakeImmutable()

	// Return a pointer to the LockedBuffer.
	return b, nil
}

/*
NewMutableFromBytes is identical to NewMutable but for the fact that the created LockedBuffer is of the same length and has the same contents as a given slice. The slice is wiped after the bytes have been copied over.

If the size of the slice is zero, the call will return an ErrInvalidLength.
*/
func NewMutableFromBytes(buf []byte) (*LockedBuffer, error) {
	// Create a new LockedBuffer.
	b, err := newContainer(len(buf), true)
	if err != nil {
		return nil, err
	}

	// Copy the bytes from buf, wiping afterwards.
	b.Move(buf)

	// Return a pointer to the LockedBuffer.
	return b, nil
}

/*
NewImmutableRandom is identical to NewImmutable but for the fact that the created LockedBuffer is filled with cryptographically-secure pseudo-random bytes instead of zeroes. Therefore a LockedBuffer created with NewImmutableRandom can safely be used as an encryption key.
*/
func NewImmutableRandom(size int) (*LockedBuffer, error) {
	// Create a new LockedBuffer for the key.
	b, err := NewMutableRandom(size)
	if err != nil {
		return nil, err
	}

	// Mark as immutable if specified.
	b.MakeImmutable()

	// Return the LockedBuffer.
	return b, nil
}

/*
NewMutableRandom is identical to NewMutable but for the fact that the created LockedBuffer is filled with cryptographically-secure pseudo-random bytes instead of zeroes. Therefore a LockedBuffer created with NewMutableRandom can safely be used as an encryption key.
*/
func NewMutableRandom(size int) (*LockedBuffer, error) {
	// Create a new LockedBuffer for the key.
	b, err := newContainer(size, true)
	if err != nil {
		return nil, err
	}

	// Fill it with random data.
	fillRandBytes(b.buffer)

	// Return the LockedBuffer.
	return b, nil
}

/*
Buffer returns a slice that references the secure, protected portion of memory.

If the LockedBuffer that you call Buffer on has been destroyed, the returned slice will be nil (it will have a length and capacity of zero).

If a function that you're using requires an array, you can cast the buffer to an array and then pass around a pointer:

    // Make sure the size of the array matches the size of the buffer.
    // In this case that size is 16. This is *very* important.
    keyArrayPtr := (*[16]byte)(unsafe.Pointer(&b.Buffer()[0]))

Make sure that you do not dereference the pointer and pass around the resulting value, as this will leave copies all over the place.
*/
func (b *container) Buffer() []byte {
	return b.buffer
}

/*
Uint8 returns a slice (of type []uint8) that references the secure, protected portion of memory.

Uint8 is practically identical to Buffer, but it has been added for completeness' sake. Buffer will usually be the faster and easier option.
*/
func (b *container) Uint8() ([]uint8, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Return the slice.
	return []uint8(b.buffer), nil
}

/*
Uint16 returns a slice (of type []uint16) that references the secure, protected portion of memory.

The LockedBuffer must be a multiple of 2 bytes in length, or else an error will be returned.
*/
func (b *container) Uint16() ([]uint16, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Check to see if it's an appropriate length.
	if len(b.buffer)%2 != 0 {
		return nil, ErrInvalidConversion
	}

	// Perform the conversion.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.buffer[0])), b.Size() / 2, b.Size() / 2}

	// Return the new slice.
	return *(*[]uint16)(unsafe.Pointer(&sl)), nil
}

/*
Uint32 returns a slice (of type []uint32) that references the secure, protected portion of memory.

The LockedBuffer must be a multiple of 4 bytes in length, or else an error will be returned.
*/
func (b *container) Uint32() ([]uint32, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Check to see if it's an appropriate length.
	if len(b.buffer)%4 != 0 {
		return nil, ErrInvalidConversion
	}

	// Perform the conversion.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.buffer[0])), b.Size() / 4, b.Size() / 4}

	// Return the new slice.
	return *(*[]uint32)(unsafe.Pointer(&sl)), nil
}

/*
Uint64 returns a slice (of type []uint64) that references the secure, protected portion of memory.

The LockedBuffer must be a multiple of 8 bytes in length, or else an error will be returned.
*/
func (b *container) Uint64() ([]uint64, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Check to see if it's an appropriate length.
	if len(b.buffer)%8 != 0 {
		return nil, ErrInvalidConversion
	}

	// Perform the conversion.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.buffer[0])), b.Size() / 8, b.Size() / 8}

	// Return the new slice.
	return *(*[]uint64)(unsafe.Pointer(&sl)), nil
}

/*
Int8 returns a slice (of type []int8) that references the secure, protected portion of memory.
*/
func (b *container) Int8() ([]int8, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Perform the conversion.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.buffer[0])), b.Size(), b.Size()}

	// Return the new slice.
	return *(*[]int8)(unsafe.Pointer(&sl)), nil
}

/*
Int16 returns a slice (of type []int16) that references the secure, protected portion of memory.

The LockedBuffer must be a multiple of 2 bytes in length, or else an error will be returned.
*/
func (b *container) Int16() ([]int16, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Check to see if it's an appropriate length.
	if len(b.buffer)%2 != 0 {
		return nil, ErrInvalidConversion
	}

	// Perform the conversion.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.buffer[0])), b.Size() / 2, b.Size() / 2}

	// Return the new slice.
	return *(*[]int16)(unsafe.Pointer(&sl)), nil
}

/*
Int32 returns a slice (of type []int32) that references the secure, protected portion of memory.

The LockedBuffer must be a multiple of 4 bytes in length, or else an error will be returned.
*/
func (b *container) Int32() ([]int32, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Check to see if it's an appropriate length.
	if len(b.buffer)%4 != 0 {
		return nil, ErrInvalidConversion
	}

	// Perform the conversion.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.buffer[0])), b.Size() / 4, b.Size() / 4}

	// Return the new slice.
	return *(*[]int32)(unsafe.Pointer(&sl)), nil
}

/*
Int64 returns a slice (of type []int64) that references the secure, protected portion of memory.

The LockedBuffer must be a multiple of 8 bytes in length, or else an error will be returned.
*/
func (b *container) Int64() ([]int64, error) {
	// Attain the mutex lock.
	b.Lock()
	defer b.Unlock()

	// Check to see if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Check to see if it's an appropriate length.
	if len(b.buffer)%8 != 0 {
		return nil, ErrInvalidConversion
	}

	// Perform the conversion.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.buffer[0])), b.Size() / 8, b.Size() / 8}

	// Return the new slice.
	return *(*[]int64)(unsafe.Pointer(&sl)), nil
}

/*
IsMutable returns a boolean value indicating if a LockedBuffer is marked read-only.
*/
func (b *container) IsMutable() bool {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	return b.mutable
}

/*
IsDestroyed returns a boolean value indicating if a LockedBuffer has been destroyed.
*/
func (b *container) IsDestroyed() bool {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Return the appropriate value.
	return len(b.buffer) == 0
}

/*
EqualBytes compares a LockedBuffer to a byte slice in constant time.
*/
func (b *container) EqualBytes(buf []byte) (bool, error) {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return false, ErrDestroyed
	}

	// Do a time-constant comparison.
	if subtle.ConstantTimeCompare(b.buffer, buf) == 1 {
		// They're equal.
		return true, nil
	}

	// They're not equal.
	return false, nil
}

/*
MakeImmutable asks the kernel to mark the LockedBuffer's memory as immutable. Any subsequent attempts to modify this memory will result in the process crashing with a SIGSEGV memory violation.

To make the memory mutable again, MakeMutable is called.
*/
func (b *container) MakeImmutable() error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return ErrDestroyed
	}

	if b.mutable {
		// Mark the memory as mutable.
		memcall.Protect(getAllMemory(b)[pageSize:pageSize+roundToPageSize(len(b.buffer)+32)], true, false)

		// Tell everyone about the change we made.
		b.mutable = false
	}

	// Everything went well.
	return nil
}

/*
MakeMutable asks the kernel to mark the LockedBuffer's memory as mutable.

To make the memory immutable again, MakeImmutable is called.
*/
func (b *container) MakeMutable() error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return ErrDestroyed
	}

	if !b.mutable {
		// Mark the memory as mutable.
		memcall.Protect(getAllMemory(b)[pageSize:pageSize+roundToPageSize(len(b.buffer)+32)], true, true)

		// Tell everyone about the change we made.
		b.mutable = true
	}

	// Everything went well.
	return nil
}

/*
Copy copies bytes from a byte slice into a LockedBuffer in constant-time. Just like Golang's built-in copy function, Copy only copies up to the smallest of the two buffers.

It does not wipe the original slice so using Copy is less secure than using Move. Therefore Move should be favoured unless you have a good reason.

You should aim to call WipeBytes on the original slice as soon as possible.

If the LockedBuffer is marked as read-only, the call will fail and return an ErrReadOnly.
*/
func (b *container) Copy(buf []byte) error {
	// Just call CopyAt with a zero offset.
	return b.CopyAt(buf, 0)
}

/*
CopyAt is identical to Copy but it copies into the LockedBuffer at a specified offset.
*/
func (b *container) CopyAt(buf []byte, offset int) error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return ErrDestroyed
	}

	// Check if it's immutable.
	if !b.mutable {
		return ErrImmutable
	}

	// Do a time-constant copying of the bytes, copying only up to the length of the buffer.
	if len(b.buffer[offset:]) > len(buf) {
		subtle.ConstantTimeCopy(1, b.buffer[offset:offset+len(buf)], buf)
	} else if len(b.buffer[offset:]) < len(buf) {
		subtle.ConstantTimeCopy(1, b.buffer[offset:], buf[:len(b.buffer[offset:])])
	} else {
		subtle.ConstantTimeCopy(1, b.buffer[offset:], buf)
	}

	return nil
}

/*
Move moves bytes from a byte slice into a LockedBuffer in constant-time. Just like Golang's built-in copy function, Move only moves up to the smallest of the two buffers.

Unlike Copy, Move wipes the entire original slice after copying the appropriate number of bytes over, and so it should be favoured unless you have a good reason.

If the LockedBuffer is marked as read-only, the call will fail and return an ErrReadOnly.
*/
func (b *container) Move(buf []byte) error {
	// Just call MoveAt with a zero offset.
	return b.MoveAt(buf, 0)
}

/*
MoveAt is identical to Move but it copies into the LockedBuffer at a specified offset.
*/
func (b *container) MoveAt(buf []byte, offset int) error {
	// Copy buf into the LockedBuffer.
	if err := b.CopyAt(buf, offset); err != nil {
		return err
	}

	// Wipe the old bytes.
	wipeBytes(buf)

	// Everything went well.
	return nil
}

/*
FillRandomBytes fills a LockedBuffer with cryptographically-secure pseudo-random bytes.
*/
func (b *container) FillRandomBytes() error {
	// Just call FillRandomBytesAt.
	return b.FillRandomBytesAt(0, b.Size())
}

/*
FillRandomBytesAt fills a LockedBuffer with cryptographically-secure pseudo-random bytes, starting at an offset and ending after a given number of bytes.
*/
func (b *container) FillRandomBytesAt(offset, length int) error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return ErrDestroyed
	}

	// Check if it's immutable.
	if !b.mutable {
		return ErrImmutable
	}

	// Fill with random bytes.
	fillRandBytes(b.buffer[offset : offset+length])

	// Everything went well.
	return nil
}

/*
Destroy verifies that no buffer underflows occurred and then wipes, unlocks, and frees all related memory. If a buffer underflow is detected, the process panics.

This function must be called on all LockedBuffers before exiting. DestroyAll is designed for this purpose, as is CatchInterrupt and SafeExit. We recommend using all of them together.

If the LockedBuffer has already been destroyed then the call makes no changes.
*/
func (b *container) Destroy() {
	// Attain a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Return if it's already destroyed.
	if len(b.buffer) == 0 {
		return
	}

	// Remove this one from global slice.
	allLockedBuffersMutex.Lock()
	for i, v := range allLockedBuffers {
		if v == b {
			allLockedBuffers = append(allLockedBuffers[:i], allLockedBuffers[i+1:]...)
			break
		}
	}
	allLockedBuffersMutex.Unlock()

	// Get all of the memory related to this LockedBuffer.
	memory := getAllMemory(b)

	// Get the total size of all the pages between the guards.
	roundedLength := len(memory) - (pageSize * 2)

	// Verify the canary.
	if !bytes.Equal(memory[pageSize+roundedLength-len(b.buffer)-32:pageSize+roundedLength-len(b.buffer)], canary) {
		panic("memguard.Destroy(): buffer overflow detected")
	}

	// Make all of the memory readable and writable.
	memcall.Protect(memory, true, true)

	// Wipe the pages that hold our data.
	wipeBytes(memory[pageSize : pageSize+roundedLength])

	// Unlock the pages that hold our data.
	memcall.Unlock(memory[pageSize : pageSize+roundedLength])

	// Free all related memory.
	memcall.Free(memory)

	// Set the metadata appropriately.
	b.mutable = false

	// Set the buffer to nil.
	b.buffer = nil
}

/*
Size returns an integer representing the total length, in bytes, of a LockedBuffer.

If this size is zero, it is safe to assume that the LockedBuffer has been destroyed.
*/
func (b *container) Size() int {
	return len(b.buffer)
}

/*
Wipe wipes a LockedBuffer's contents by overwriting the buffer with zeroes.
*/
func (b *container) Wipe() error {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return ErrDestroyed
	}

	// Check if it's immutable.
	if !b.mutable {
		return ErrImmutable
	}

	// Wipe the buffer.
	wipeBytes(b.buffer)

	// Everything went well.
	return nil
}

/*
Concatenate takes two LockedBuffers and concatenates them.

If one of the given LockedBuffers is immutable, the resulting LockedBuffer will also be immutable. The original LockedBuffers are not destroyed.
*/
func Concatenate(a, b *LockedBuffer) (*LockedBuffer, error) {
	// Get a mutex lock on the LockedBuffers.
	a.Lock()
	b.Lock()
	defer a.Unlock()
	defer b.Unlock()

	// Check if either are destroyed.
	if len(a.buffer) == 0 || len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Create a new LockedBuffer to hold the concatenated value.
	c, _ := NewMutable(len(a.buffer) + len(b.buffer))

	// Copy the values across.
	c.Copy(a.buffer)
	c.CopyAt(b.buffer, len(a.buffer))

	// Set permissions accordingly.
	if !a.mutable || !b.mutable {
		c.MakeImmutable()
	}

	// Return the resulting LockedBuffer.
	return c, nil
}

/*
Duplicate takes a LockedBuffer and creates a new one with the same contents and mutability state as the original.
*/
func Duplicate(b *LockedBuffer) (*LockedBuffer, error) {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Create new LockedBuffer.
	newBuf, _ := NewMutable(b.Size())

	// Copy bytes into it.
	newBuf.Copy(b.buffer)

	// Set permissions accordingly.
	if !b.mutable {
		newBuf.MakeImmutable()
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
	if len(a.buffer) == 0 || len(b.buffer) == 0 {
		return false, ErrDestroyed
	}

	// Do a time-constant comparison on the two buffers.
	if subtle.ConstantTimeCompare(a.buffer, b.buffer) == 1 {
		// They're equal.
		return true, nil
	}

	// They're not equal.
	return false, nil
}

/*
Split takes a LockedBuffer, splits it at a specified offset, and then returns the two newly created LockedBuffers. The mutability state of the original is preserved in the new LockedBuffers, and the original LockedBuffer is not destroyed.
*/
func Split(b *LockedBuffer, offset int) (*LockedBuffer, *LockedBuffer, error) {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, nil, ErrDestroyed
	}

	// Create two new LockedBuffers.
	firstBuf, err := NewMutable(len(b.buffer[:offset]))
	if err != nil {
		return nil, nil, err
	}

	secondBuf, err := NewMutable(len(b.buffer[offset:]))
	if err != nil {
		firstBuf.Destroy()
		return nil, nil, err
	}

	// Copy the values into them.
	firstBuf.Copy(b.buffer[:offset])
	secondBuf.Copy(b.buffer[offset:])

	// Copy over permissions.
	if !b.mutable {
		firstBuf.MakeImmutable()
		secondBuf.MakeImmutable()
	}

	// Return the new LockedBuffers.
	return firstBuf, secondBuf, nil
}

/*
Trim shortens a LockedBuffer according to the given specifications. The mutability state of the original is preserved in the new LockedBuffer, and the original LockedBuffer is not destroyed.

Trim takes an offset and a size as arguments. The resulting LockedBuffer starts at index [offset] and ends at index [offset+size].
*/
func Trim(b *LockedBuffer, offset, size int) (*LockedBuffer, error) {
	// Get a mutex lock on this LockedBuffer.
	b.Lock()
	defer b.Unlock()

	// Check if it's destroyed.
	if len(b.buffer) == 0 {
		return nil, ErrDestroyed
	}

	// Create new LockedBuffer and copy over the old.
	newBuf, err := NewMutable(size)
	if err != nil {
		return nil, err
	}
	newBuf.Copy(b.buffer[offset : offset+size])

	// Copy over permissions.
	if !b.mutable {
		newBuf.MakeImmutable()
	}

	// Return the new LockedBuffer.
	return newBuf, nil
}

/*
WipeBytes zeroes out a given byte slice. It is recommeded that you call WipeBytes on slices after utilizing the Copy or CopyAt methods.

Due to the nature of memory allocated by the Go runtime, WipeBytes cannot guarantee that the data does not exist elsewhere in memory. Therefore, your program should aim to (when possible) store sensitive data only in LockedBuffers.
*/
func WipeBytes(b []byte) {
	wipeBytes(b)
}

/*
DestroyAll calls Destroy on all LockedBuffers that have not already been destroyed.

CatchInterrupt and SafeExit both call DestroyAll before exiting.
*/
func DestroyAll() {
	// Get a Mutex lock on allLockedBuffers, and get a copy.
	allLockedBuffersMutex.Lock()
	containers := make([]*container, len(allLockedBuffers))
	copy(containers, allLockedBuffers)
	allLockedBuffersMutex.Unlock()

	for _, b := range containers {
		b.Destroy()
	}
}

/*
CatchInterrupt starts a goroutine that monitors for interrupt signals. It accepts a function of type func() and executes that before calling SafeExit(0).

If CatchInterrupt is called multiple times, only the first call is executed and all subsequent calls are ignored.
*/
func CatchInterrupt(f func()) {
	// Only do this if it hasn't been done before.
	catchInterruptOnce.Do(func() {
		// Create a channel to listen on.
		c := make(chan os.Signal, 2)

		// Notify the channel if we receive a signal.
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// Start a goroutine to listen on the channel.
		go func() {
			<-c         // Wait for signal.
			f()         // Execute user function.
			SafeExit(0) // Exit securely.
		}()
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
DisableUnixCoreDumps disables core-dumps.

Since core-dumps are only relevant on Unix systems, if DisableUnixCoreDumps is called on any other system it will do nothing and return immediately.

This function is precautonary as core-dumps are usually disabled by default on most systems.
*/
func DisableUnixCoreDumps() {
	memcall.DisableCoreDumps()
}
