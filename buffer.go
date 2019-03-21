package memguard

import (
	"runtime"
	"unsafe"

	"github.com/awnumar/memguard/core"
	"github.com/awnumar/memguard/crypto"
)

/*
LockedBuffer is a structure that holds raw sensitive data.

The number of LockedBuffers that you are able to create is limited by how much memory your system's kernel allows each process to mlock/VirtualLock. Therefore you should call Destroy on LockedBuffers that you no longer need or defer a Destroy call after creating a new LockedBuffer.
*/
type LockedBuffer struct {
	*core.Buffer
	*drop
}

/*
This is a value that is monitored by a finalizer so that we can clean up LockedBuffers that have gone out of scope.
*/
type drop [16]byte

/*
NewBuffer is a generic constructor for the LockedBuffer object. Valid values for the size argument are positive integers strictly greater than zero.

The returned buffer will be mutable but this can be changed with the Freeze and Melt methods.
*/
func NewBuffer(size int) (*LockedBuffer, error) {
	// Construct a Buffer of the specified size.
	buf, err := core.NewBuffer(size)
	if err != nil {
		return nil, err
	}

	// Initialise a LockedBuffer object around it.
	b := &LockedBuffer{buf, new(drop)}

	// Use a finalizer to destroy the Buffer if it falls out of scope.
	runtime.SetFinalizer(b.drop, func(_ *drop) {
		go buf.Destroy()
	})

	// Return the created buffer to the caller.
	return b, nil
}

/*
NewBufferFromBytes constructs a buffer from a byte slice. The given slice must have a length of at least 1 byte and it is wiped after being copied over to the LockedBuffer.

The returned buffer will be mutable but this can be changed with the Freeze and Melt methods.
*/
func NewBufferFromBytes(buf []byte) (*LockedBuffer, error) {
	// Construct a buffer of the correct size.
	b, err := NewBuffer(len(buf))
	if err != nil {
		return nil, err
	}

	// Move the data over.
	crypto.Move(b.Data, buf)

	// Return the created Buffer object.
	return b, nil
}

/*
NewBufferRandom constructs a buffer filled with cryptographically-secure random bytes. Valid values for the size argument are positive integers strictly greater than zero.

The returned buffer will be mutable but this can be changed with the Freeze and Melt methods.
*/
func NewBufferRandom(size int) (*LockedBuffer, error) {
	// Construct a buffer of the specified size.
	b, err := NewBuffer(size)
	if err != nil {
		return nil, err
	}

	// Fill the buffer with random bytes.
	if err := crypto.MemScr(b.Data); err != nil {
		core.Panic(err)
	}

	// Return the created Buffer object.
	return b, nil
}

// Freeze makes a LockedBuffer's memory immutable. The call can be reversed with Melt.
func (b *LockedBuffer) Freeze() {
	b.Buffer.Freeze()
}

// Melt makes a LockedBuffer's memory mutable. The call can be reversed with Freeze.
func (b *LockedBuffer) Melt() {
	b.Buffer.Melt()
}

/*
Seal takes a LockedBuffer object and returns its contents encrypted inside a sealed Enclave object. The LockedBuffer is subsequently destroyed and its contents wiped.
*/
func (b *LockedBuffer) Seal() (*Enclave, error) {
	e, err := core.Seal(b.Buffer)
	if err != nil {
		return nil, err
	}
	return &Enclave{e}, nil
}

/*
Copy performs a time-constant copy into a LockedBuffer. Move is preferred if the source is not a LockedBuffer or if the source is no longer needed.
*/
func (b *LockedBuffer) Copy(buf []byte) {
	if !b.IsAlive() {
		return
	}

	b.Lock()
	defer b.Unlock()

	crypto.Copy(b.Buffer.Data, buf)
}

/*
Move performs a time-constant move into a LockedBuffer. The source is wiped after the bytes are copied.
*/
func (b *LockedBuffer) Move(buf []byte) {
	b.Copy(buf)
	crypto.MemClr(buf)
}

/*
Scramble attempts to overwrite the data with cryptographically-secure random bytes.
*/
func (b *LockedBuffer) Scramble() {
	if !b.IsAlive() {
		return
	}

	b.Lock()
	defer b.Unlock()

	if err := crypto.MemScr(b.Buffer.Data); err != nil {
		core.Panic(err)
	}
}

/*
Wipe attempts to overwrite the data with zeros.
*/
func (b *LockedBuffer) Wipe() {
	if !b.IsAlive() {
		return
	}

	b.Lock()
	defer b.Unlock()

	crypto.MemClr(b.Buffer.Data)
}

/*
Size gives you the length of a given LockedBuffer's data segment. A destroyed LockedBuffer will have a size of zero.
*/
func (b *LockedBuffer) Size() int {
	return len(b.Buffer.Data)
}

/*
Resize allocates a new buffer of a positive integer size strictly greater than zero, and copies the data and mutability attribute over from the old one before destroying it.
*/
func (b *LockedBuffer) Resize(size int) (*LockedBuffer, error) {
	if !b.IsAlive() {
		return nil, core.ErrObjectExpired
	}

	b.RLock()

	new, err := NewBuffer(size)
	if err != nil {
		return nil, err
	}

	crypto.Move(new.Buffer.Data, b.Buffer.Data)

	if !b.IsMutable() {
		new.Freeze()
	}

	b.RUnlock()
	b.Destroy()

	return new, nil
}

/*
Destroy wipes and frees the underlying memory of a LockedBuffer. The LockedBuffer will not be accessible or usable after this calls is made.
*/
func (b *LockedBuffer) Destroy() {
	b.Buffer.Destroy()
}

/*
IsAlive returns a boolean value indicating if a LockedBuffer is alive, i.e. that it has not been destroyed.
*/
func (b *LockedBuffer) IsAlive() bool {
	return core.GetBufferState(b.Buffer).IsAlive
}

/*
IsMutable returns a boolean value indicating if a LockedBuffer is mutable.
*/
func (b *LockedBuffer) IsMutable() bool {
	return core.GetBufferState(b.Buffer).IsMutable
}

/*
	Functions for representing the memory region as various data types.
*/

/*
Bytes returns a byte slice referencing the protected region of memory within which you are able to store and view sensitive data.
*/
func (b *LockedBuffer) Bytes() []byte {
	return b.Buffer.Data
}

/*
Uint16 returns a slice pointing to the protected region of memory with the data represented as []uint16.

The length of the buffer must be a multiple of two bytes in size and the LockedBuffer should not be destroyed. In either of these cases a nil value is returned.
*/
func (b *LockedBuffer) Uint16() []uint16 {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Check if data size a multiple of two.
	if len(b.Buffer.Data)%2 != 0 {
		return nil
	}

	// Construct the new slice representation.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Data[0])), len(b.Buffer.Data) / 2, len(b.Buffer.Data) / 2}

	// Cast the representation to the correct type and return it.
	return *(*[]uint16)(unsafe.Pointer(&sl))
}

/*
Uint32 returns a slice pointing to the protected region of memory with the data represented as []uint32.

The length of the buffer must be a multiple of four bytes in size and the LockedBuffer should not be destroyed. In either of these cases a nil value is returned.
*/
func (b *LockedBuffer) Uint32() []uint32 {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Check if data size a multiple of two.
	if len(b.Buffer.Data)%4 != 0 {
		return nil
	}

	// Construct the new slice representation.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Data[0])), len(b.Buffer.Data) / 4, len(b.Buffer.Data) / 4}

	// Cast the representation to the correct type and return it.
	return *(*[]uint32)(unsafe.Pointer(&sl))
}

/*
Uint64 returns a slice pointing to the protected region of memory with the data represented as []uint64.

The length of the buffer must be a multiple of eight bytes in size and the LockedBuffer should not be destroyed. In either of these cases a nil value is returned.
*/
func (b *LockedBuffer) Uint64() []uint64 {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Check if data size a multiple of two.
	if len(b.Buffer.Data)%8 != 0 {
		return nil
	}

	// Construct the new slice representation.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Data[0])), len(b.Buffer.Data) / 8, len(b.Buffer.Data) / 8}

	// Cast the representation to the correct type and return it.
	return *(*[]uint64)(unsafe.Pointer(&sl))
}

/*
Int8 returns a slice pointing to the protected region of memory with the data represented as []int8.

The LockedBuffer should not be destroyed or else a nil value is returned.
*/
func (b *LockedBuffer) Int8() []int8 {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Construct the new slice representation.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Data[0])), len(b.Buffer.Data), len(b.Buffer.Data)}

	// Cast the representation to the correct type and return it.
	return *(*[]int8)(unsafe.Pointer(&sl))
}

/*
Int16 returns a slice pointing to the protected region of memory with the data represented as []int16.

The length of the buffer must be a multiple of two bytes in size and the LockedBuffer should not be destroyed. In either of these cases a nil value is returned.
*/
func (b *LockedBuffer) Int16() []int16 {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Check if data size a multiple of two.
	if len(b.Buffer.Data)%2 != 0 {
		return nil
	}

	// Construct the new slice representation.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Data[0])), len(b.Buffer.Data) / 2, len(b.Buffer.Data) / 2}

	// Cast the representation to the correct type and return it.
	return *(*[]int16)(unsafe.Pointer(&sl))
}

/*
Int32 returns a slice pointing to the protected region of memory with the data represented as []int32.

The length of the buffer must be a multiple of four bytes in size and the LockedBuffer should not be destroyed. In either of these cases a nil value is returned.
*/
func (b *LockedBuffer) Int32() []int32 {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Check if data size a multiple of two.
	if len(b.Buffer.Data)%4 != 0 {
		return nil
	}

	// Construct the new slice representation.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Data[0])), len(b.Buffer.Data) / 4, len(b.Buffer.Data) / 4}

	// Cast the representation to the correct type and return it.
	return *(*[]int32)(unsafe.Pointer(&sl))
}

/*
Int64 returns a slice pointing to the protected region of memory with the data represented as []int64.

The length of the buffer must be a multiple of eight bytes in size and the LockedBuffer should not be destroyed. In either of these cases a nil value is returned.
*/
func (b *LockedBuffer) Int64() []int64 {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Check if data size a multiple of two.
	if len(b.Buffer.Data)%8 != 0 {
		return nil
	}

	// Construct the new slice representation.
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Data[0])), len(b.Buffer.Data) / 8, len(b.Buffer.Data) / 8}

	// Cast the representation to the correct type and return it.
	return *(*[]int64)(unsafe.Pointer(&sl))
}

/*
ByteArray8 takes a start byte and returns a pointer to the start of some 16 byte array.

The length of the buffer must be at least 8 bytes in size and the LockedBuffer should not be destroyed. In either of these cases a nil value is returned.
*/
func (b *LockedBuffer) ByteArray8(start *byte) *[8]byte {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Check if the length is large enough.
	if len(b.Buffer.Data) < 8 {
		return nil
	}

	// Cast the representation to the correct type.
	return (*[8]byte)(unsafe.Pointer(&b.Buffer.Data[0]))
}

/*
ByteArray16 takes a start byte and returns a pointer to the start of some 16 byte array.

The length of the buffer must be at least 16 bytes in size and the LockedBuffer should not be destroyed. In either of these cases a nil value is returned.
*/
func (b *LockedBuffer) ByteArray16(start *byte) *[16]byte {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Check if the length is large enough.
	if len(b.Buffer.Data) < 16 {
		return nil
	}

	// Cast the representation to the correct type.
	return (*[16]byte)(unsafe.Pointer(start))
}

/*
ByteArray32 takes a start byte and returns a pointer to the start of some 32 byte array.

The length of the buffer must be at least 32 bytes in size and the LockedBuffer should not be destroyed. In either of these cases a nil value is returned.
*/
func (b *LockedBuffer) ByteArray32(start *byte) *[32]byte {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Check if the length is large enough.
	if len(b.Buffer.Data) < 32 {
		return nil
	}

	// Cast the representation to the correct type.
	return (*[32]byte)(unsafe.Pointer(start))
}

/*
ByteArray64 takes a start byte and returns a pointer to the start of some 64 byte array.

The length of the buffer must be at least 64 bytes in size and the LockedBuffer should not be destroyed. In either of these cases a nil value is returned.
*/
func (b *LockedBuffer) ByteArray64(start *byte) *[64]byte {
	// Check if still alive.
	if !core.GetBufferState(b.Buffer).IsAlive {
		return nil
	}

	// Check if the length is large enough.
	if len(b.Buffer.Data) < 64 {
		return nil
	}

	// Cast the representation to the correct type.
	return (*[64]byte)(unsafe.Pointer(start))
}
