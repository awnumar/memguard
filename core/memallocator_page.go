package core

import (
	"sync"
	"unsafe"

	"github.com/awnumar/memcall"
)

type pageAllocator struct {
	objects map[int]*pageObject
	sync.Mutex
}

func NewPageAllocator() MemAllocator {
	a := &pageAllocator{
		objects: make(map[int]*pageObject),
	}
	return a
}

func (a *pageAllocator) Alloc(size int) ([]byte, error) {
	MemStats.ObjectAllocs.Add(1)
	if size < 1 {
		return nil, ErrNullAlloc
	}
	o, err := newPageObject(size)
	if err != nil {
		MemStats.ObjectAllocErrors.Add(1)
		return nil, err
	}

	// Store the allocated object with the lookup key of the inner
	// buffers address. This allows to efficiently free the buffer
	// later
	addr := int(uintptr(unsafe.Pointer(&o.data[0])))
	a.Lock()
	a.objects[addr] = o
	a.Unlock()

	return o.data, nil
}

func (a *pageAllocator) Protect(buf []byte, readonly bool) error {
	if len(buf) == 0 {
		return ErrNullPointer
	}

	// Determine the object belonging to the buffer
	o, found := a.lookup(buf)
	if !found {
		Panic(ErrBufferNotOwnedByAllocator)
	}

	var flag memcall.MemoryProtectionFlag
	if readonly {
		flag = memcall.ReadOnly()
	} else {
		flag = memcall.ReadWrite()
	}

	return memcall.Protect(o.inner, flag)
}

func (a *pageAllocator) Inner(buf []byte) []byte {
	if len(buf) == 0 {
		return nil
	}

	// Determine the object belonging to the buffer
	o, found := a.lookup(buf)
	if !found {
		Panic(ErrBufferNotOwnedByAllocator)
	}

	return o.inner
}

func (a *pageAllocator) Free(buf []byte) error {
	MemStats.ObjectFrees.Add(1)

	// Determine the address of the buffer we should free
	o, found := a.pop(buf)
	if !found {
		MemStats.ObjectFreeErrors.Add(1)
		return ErrBufferNotOwnedByAllocator
	}

	// Destroy the object's content
	if err := o.wipe(); err != nil {
		return err
	}

	// Free the related memory
	MemStats.PageFrees.Add(uint64(len(o.memory) / pageSize))
	if err := memcall.Free(o.memory); err != nil {
		MemStats.PageFreeErrors.Add(1)
		return err
	}

	return nil
}

// *** INTERNAL FUNCTIONS *** //
func (a *pageAllocator) lookup(buf []byte) (*pageObject, bool) {
	if len(buf) == 0 {
		return nil, false
	}

	// Determine the address of the buffer we should free
	addr := int(uintptr(unsafe.Pointer(&buf[0])))

	a.Lock()
	defer a.Unlock()
	o, found := a.objects[addr]
	return o, found
}

func (a *pageAllocator) pop(buf []byte) (*pageObject, bool) {
	if len(buf) == 0 {
		return nil, false
	}

	addr := int(uintptr(unsafe.Pointer(&buf[0])))

	a.Lock()
	defer a.Unlock()
	o, found := a.objects[addr]
	if !found {
		return nil, false
	}
	delete(a.objects, addr)

	return o, true
}

// object holding each allocation
type pageObject struct {
	data   []byte // Portion of memory holding the data
	memory []byte // Entire allocated memory region

	preguard  []byte // Guard page addressed before the data
	inner     []byte // Inner region between the guard pages
	postguard []byte // Guard page addressed after the data

	canary []byte // Value written behind data to detect spillage
}

func newPageObject(size int) (*pageObject, error) {
	// Round a length to a multiple of the system page size for page locking
	// and protection
	innerLen := roundToPageSize(size)

	// Allocate the total needed memory
	MemStats.PageAllocs.Add(uint64(2 + innerLen/pageSize))
	memory, err := memcall.Alloc((2 * pageSize) + innerLen)
	if err != nil {
		MemStats.PageAllocErrors.Add(1)
		return nil, err
	}

	o := &pageObject{
		memory: memory,
		// Construct slice reference for data buffer.
		data: getBufferPart(memory, pageSize+innerLen-size, size),
		// Construct slice references for page sectors.
		preguard:  getBufferPart(memory, 0, pageSize),
		inner:     getBufferPart(memory, pageSize, innerLen),
		postguard: getBufferPart(memory, pageSize+innerLen, pageSize),
	}
	// Construct slice reference for canary portion of inner page.
	o.canary = getBufferPart(memory, pageSize, len(o.inner)-len(o.data))

	// Lock the pages that will hold sensitive data.
	if err := memcall.Lock(o.inner); err != nil {
		return nil, err
	}

	// Create a random signature for the protection pages and reuse the
	// fitting part for the canary
	if err := Scramble(o.preguard); err != nil {
		return nil, err
	}
	Copy(o.postguard, o.preguard)
	Copy(o.canary, o.preguard)

	// Make the guard pages inaccessible.
	if err := memcall.Protect(o.preguard, memcall.NoAccess()); err != nil {
		return nil, err
	}
	if err := memcall.Protect(o.postguard, memcall.NoAccess()); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *pageObject) wipe() error {
	// Make all of the memory readable and writable.
	if err := memcall.Protect(o.memory, memcall.ReadWrite()); err != nil {
		return err
	}

	// Wipe data field.
	Wipe(o.data)
	o.data = nil

	// Verify the canary
	if !Equal(o.preguard, o.postguard) || !Equal(o.preguard[:len(o.canary)], o.canary) {
		return ErrBufferOverflow
	}

	// Wipe the memory.
	Wipe(o.memory)

	// Unlock pages locked into memory.
	if err := memcall.Unlock(o.inner); err != nil {
		return err
	}

	// Reset the fields.
	o.data = nil
	o.inner = nil
	o.preguard = nil
	o.postguard = nil
	o.canary = nil

	return nil
}
