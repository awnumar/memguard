package memguard

import (
	"crypto/subtle"
	"runtime"
	"sync"
	"unsafe"

	"github.com/awnumar/memguard/memcall"
)

/*
Enclave is a structure that holds secure values.

The protected memory itself can be accessed with the Bytes() method. The various states can be accessed with the IsDestroyed() and IsMutable() methods, both of which are pretty self-explanatory.

The number of Enclaves that you are able to create is limited by how much memory your system kernel allows each process to mlock/VirtualLock. Therefore you should call Destroy on Enclaves that you no longer need, or simply defer a Destroy call after creating a new Enclave.

The entire memguard API handles and passes around pointers to Enclaves, and so, for both security and convenience, you should refrain from dereferencing a Enclave.

If an API function that needs to edit a Enclave is given one that is immutable, the call will return an ErrImmutable. Similarly, if a function is given a Enclave that has been destroyed, the call will return an ErrDestroyed.
*/
type Enclave struct {
	*container  // Import all the container fields.
	*littleBird // Monitor this for auto-destruction.
}

// container implements the actual data container.
type container struct {
	sync.Mutex // Local mutex lock.

	buffer []byte // Slice that references the protected memory.

	ciphertext []byte    // Slice that references the ciphertext when sealed.
	key        *subclave // The encryption key used to protect this Enclave.

	mutable bool // Is this Enclave mutable?
	sealed  bool // Is this Enclave encrypted and sealed?
}

// littleBird is a value that we monitor instead of the Enclave
// itself. It allows us to tell the GC to auto-destroy Enclaves.
type littleBird [16]byte

// Global internal function used to create new secure containers.
func newContainer(size int, mutable bool) (*Enclave, error) {
	// Return an error if length < 1.
	if size < 1 {
		return nil, ErrInvalidLength
	}

	// Allocate a new Enclave.
	ib := new(container)
	b := &Enclave{ib, new(littleBird)}

	// Round length + 32 bytes for the canary to a multiple of the page size.
	roundedLength := roundToPageSize(size + 32)

	// Calculate the total size of memory including the guard pages.
	totalSize := (2 * pageSize) + roundedLength

	// Allocate it all.
	memory, err := memcall.Alloc(totalSize)
	if err != nil {
		SafePanic(err)
	}

	// Make the guard pages inaccessible.
	if err := memcall.Protect(memory[:pageSize], false, false); err != nil {
		SafePanic(err)
	}
	if err := memcall.Protect(memory[pageSize+roundedLength:], false, false); err != nil {
		SafePanic(err)
	}

	// Lock the pages that will hold the sensitive data.
	if err := memcall.Lock(memory[pageSize : pageSize+roundedLength]); err != nil {
		SafePanic(err)
	}

	// Set the canary.
	c := canary.getView()
	defer c.destroy()
	subtle.ConstantTimeCopy(1, memory[pageSize+roundedLength-size-32:pageSize+roundedLength-size], c.buffer)

	// Set Buffer to a byte slice that describes the region of memory that is protected.
	b.buffer = getBytes(uintptr(unsafe.Pointer(&memory[pageSize+roundedLength-size])), size)

	// Set appropriate mutability state.
	b.mutable = true
	if !mutable {
		b.MakeImmutable()
	}

	// Create and set the key subclave.
	b.key = newSubclave()

	// Use a finalizer to make sure the buffer gets destroyed if forgotten.
	runtime.SetFinalizer(b.littleBird, func(_ *littleBird) {
		go ib.Destroy()
	})

	// Append the container to enclaves. We have to add container
	// instead of Enclave so that littleBird can become unreachable.
	enclavesMutex.Lock()
	enclaves = append(enclaves, ib)
	enclavesMutex.Unlock()

	// Return a pointer to the Enclave.
	return b, nil
}

// Internal seal method encrypts the data inside an enclave.
func (b *container) seal() {

}

// Internal unseal method decrypts the data inside an enclave.
func (b *container) unseal() {

}
