package core

import (
	"bytes"
	"testing"
	"unsafe"
)

func TestNewBuffer(t *testing.T) {
	// Check the error case with zero length.
	b, err := NewBuffer(0)
	if err != ErrNullBuffer {
		t.Error("expected ErrNullBuffer; got", err)
	}
	if b != nil {
		t.Error("expected nil buffer; got", b)
	}

	// Check the error case with negative length.
	b, err = NewBuffer(-1)
	if err != ErrNullBuffer {
		t.Error("expected ErrNullBuffer; got", err)
	}
	if b != nil {
		t.Error("expected nil buffer; got", b)
	}

	// Test normal execution.
	b, err = NewBuffer(32)
	if err != nil {
		t.Error("expected nil err; got", err)
	}
	if !b.alive {
		t.Error("did not expect destroyed buffer")
	}
	if len(b.Data()) != 32 || cap(b.Data()) != 32 {
		t.Errorf("buffer has invalid length (%d) or capacity (%d)", len(b.Data()), cap(b.Data()))
	}
	if !b.mutable {
		t.Error("buffer is not marked mutable")
	}
	if len(b.memory) != roundToPageSize(32)+(2*pageSize) {
		t.Error("allocated incorrect length of memory")
	}
	if !bytes.Equal(b.Data(), make([]byte, 32)) {
		t.Error("container is not zero-filled")
	}

	// Check if the buffer was added to the buffers list.
	if !buffers.Exists(b) {
		t.Error("buffer not in buffers list")
	}

	// Destroy the buffer.
	b.Destroy()
}

func TestData(t *testing.T) {
	b, err := NewBuffer(32)
	if err != nil {
		t.Error(err)
	}
	datasegm := b.data
	datameth := b.Data()

	// Check for naive equality.
	if !bytes.Equal(datasegm, datameth) {
		t.Error("naive equality check failed")
	}

	// Modify and check if the change was reflected in both.
	datameth[0] = 1
	datameth[31] = 1
	if !bytes.Equal(datasegm, datameth) {
		t.Error("modified equality check failed")
	}

	// Do a deep comparison.
	if uintptr(unsafe.Pointer(&datameth[0])) != uintptr(unsafe.Pointer(&datasegm[0])) {
		t.Error("pointer values differ")
	}
	if len(datameth) != len(datasegm) || cap(datameth) != cap(datasegm) {
		t.Error("length or capacity values differ")
	}

	b.Destroy()
	if b.Data() != nil {
		t.Error("expected nil data slice for destroyed buffer")
	}
}

func TestBufferState(t *testing.T) {
	b, err := NewBuffer(32)
	if err != nil {
		t.Error("expected nil err; got", err)
	}

	state := GetBufferState(b)

	if state.IsMutable != true {
		t.Error("state mismatch: mutability")
	}

	if state.IsAlive != true {
		t.Error("state mismatch: alive")
	}

	b.Freeze()

	state = GetBufferState(b)

	if state.IsMutable != false {
		t.Error("state mismatch: mutability")
	}

	if state.IsAlive != true {
		t.Error("state mismatch: alive")
	}

	b.Melt()

	state = GetBufferState(b)

	if state.IsMutable != true {
		t.Error("state mismatch: mutability")
	}

	if state.IsAlive != true {
		t.Error("state mismatch: alive")
	}

	b.Destroy()

	state = GetBufferState(b)

	if state.IsMutable != false {
		t.Error("state mismatch: mutability")
	}

	if state.IsAlive != false {
		t.Error("state mismatch: alive")
	}
}

func TestDestroy(t *testing.T) {
	// Allocate a new buffer.
	b, err := NewBuffer(32)
	if err != nil {
		t.Error("expected nil err; got", err)
	}

	// Destroy it.
	b.Destroy()

	// Pick apart the destruction.
	if b.Data() != nil {
		t.Error("expected bytes buffer to be nil; got", b.Data())
	}
	if b.memory != nil {
		t.Error("expected memory to be nil; got", b.memory)
	}
	if b.mutable || b.alive {
		t.Error("buffer should be dead and immutable")
	}
	if b.preguard != nil || b.postguard != nil {
		t.Error("guard page slice references are not nil")
	}
	if b.canary != nil {
		t.Error("canary slice reference not nil")
	}

	// Check if the buffer was removed from the buffers list.
	if buffers.Exists(b) {
		t.Error("buffer is still in buffers list")
	}

	// Call destroy again to check idempotency.
	b.Destroy()

	// Verify that it didn't come back to life.
	if b.Data() != nil {
		t.Error("expected bytes buffer to be nil; got", b.Data())
	}
	if b.memory != nil {
		t.Error("expected memory to be nil; got", b.memory)
	}
	if b.mutable || b.alive {
		t.Error("buffer should be dead and immutable")
	}
	if b.preguard != nil || b.postguard != nil {
		t.Error("guard page slice references are not nil")
	}
	if b.canary != nil {
		t.Error("canary slice reference not nil")
	}
}

func TestBufferList(t *testing.T) {
	// Create a new BufferList for testing with.
	l := new(BufferList)

	// Create some example buffers to test with.
	a := new(Buffer)
	b := new(Buffer)

	// Check what Exists is saying.
	if l.Exists(a) || l.Exists(b) {
		t.Error("list is empty yet contains buffers?!")
	}

	// Add our two buffers to the list.
	l.Add(a)
	if len(l.list) != 1 || l.list[0] != a {
		t.Error("buffer was not added correctly")
	}
	l.Add(b)
	if len(l.list) != 2 || l.list[1] != b {
		t.Error("buffer was not added correctly")
	}

	// Now check that they exist.
	if !l.Exists(a) || !l.Exists(b) {
		t.Error("expected buffers to be in list")
	}

	// Remove the buffers from the list.
	l.Remove(a)
	if len(l.list) != 1 || l.list[0] != b {
		t.Error("buffer was not removed correctly")
	}
	l.Remove(b)
	if len(l.list) != 0 {
		t.Error("item was not removed correctly")
	}

	// Check what exists is saying now.
	if l.Exists(a) || l.Exists(b) {
		t.Error("list is empty yet contains buffers?!")
	}

	// Add the buffers again to test Empty.
	l.Add(a)
	l.Add(b)
	bufs := l.Flush()
	if l.list != nil {
		t.Error("list was not nullified")
	}
	if len(bufs) != 2 || bufs[0] != a || bufs[1] != b {
		t.Error("buffers dump incorrect")
	}

	// Try appending again.
	l.Add(a)
	if !l.Exists(a) || l.Exists(b) {
		t.Error("list is in invalid state")
	}
	l.Remove(a)
}
