package core

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestNewBuffer(t *testing.T) {
	// Check the error case with zero length.
	a, err := NewBuffer(0)
	require.ErrorIs(t, err, ErrNullBuffer)
	require.Nil(t, a)

	// Check the error case with negative length.
	b, err := NewBuffer(-1)
	require.ErrorIs(t, err, ErrNullBuffer)
	require.Nil(t, b)

	// Test normal execution.
	b, err = NewBuffer(32)
	require.NoError(t, err)
	require.True(t, b.alive, "did not expect destroyed buffer")
	require.Lenf(t, b.Data(), 32, "buffer has invalid length (%d)", len(b.Data()))
	require.Equalf(t, cap(b.Data()), 32, "buffer has invalid capacity (%d)", cap(b.Data()))
	require.True(t, b.mutable, "buffer is not marked mutable")
	require.EqualValues(t, make([]byte, 32), b.Data(), "container is not zero-filled")

	// Check if the buffer was added to the buffers list.
	require.True(t, buffers.exists(b), "buffer not in buffers list")

	// Destroy the buffer.
	b.Destroy()
}

func TestLotsOfAllocs(t *testing.T) {
	for i := 1; i <= 16385; i++ {
		b, err := NewBuffer(i)
		require.NoErrorf(t, err, "creating buffer in iteration %d", i)
		require.Truef(t, b.alive, "not alive in iteration %d", i)
		require.Truef(t, b.mutable, "not mutable in iteration %d", i)
		require.Lenf(t, b.data, i, "invalid data length %d in iteration %d", len(b.data), i)
		require.Zerof(t, len(b.Inner())%pageSize, "inner length %d is not multiple of page size in iteration %d", len(b.Inner()), i)

		// Fill data
		for j := range b.data {
			b.data[j] = 1
		}
		require.Equalf(t, bytes.Repeat([]byte{1}, i), b.data, "region rw test failed in iteration %d", i)
		b.Destroy()
	}
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

	if b.Mutable() != true {
		t.Error("state mismatch: mutability")
	}

	if b.Alive() != true {
		t.Error("state mismatch: alive")
	}

	b.Freeze()

	if b.Mutable() != false {
		t.Error("state mismatch: mutability")
	}

	if b.Alive() != true {
		t.Error("state mismatch: alive")
	}

	b.Melt()

	if b.Mutable() != true {
		t.Error("state mismatch: mutability")
	}

	if b.Alive() != true {
		t.Error("state mismatch: alive")
	}

	b.Destroy()

	if b.Mutable() != false {
		t.Error("state mismatch: mutability")
	}

	if b.Alive() != false {
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
	if b.mutable || b.alive {
		t.Error("buffer should be dead and immutable")
	}
	if b.Inner() != nil {
		t.Error("inner pages slice reference not nil")
	}

	// Check if the buffer was removed from the buffers list.
	if buffers.exists(b) {
		t.Error("buffer is still in buffers list")
	}

	// Call destroy again to check idempotency.
	b.Destroy()

	// Verify that it didn't come back to life.
	if b.Data() != nil {
		t.Error("expected bytes buffer to be nil; got", b.Data())
	}
	if b.mutable || b.alive {
		t.Error("buffer should be dead and immutable")
	}
	if b.Inner() != nil {
		t.Error("inner pages slice reference not nil")
	}
}

func TestBufferList(t *testing.T) {
	// Create a new BufferList for testing with.
	l := new(bufferList)

	// Create some example buffers to test with.
	a := new(Buffer)
	b := new(Buffer)

	// Check what Exists is saying.
	if l.exists(a) || l.exists(b) {
		t.Error("list is empty yet contains buffers?!")
	}

	// Add our two buffers to the list.
	l.add(a)
	if len(l.list) != 1 || l.list[0] != a {
		t.Error("buffer was not added correctly")
	}
	l.add(b)
	if len(l.list) != 2 || l.list[1] != b {
		t.Error("buffer was not added correctly")
	}

	// Now check that they exist.
	if !l.exists(a) || !l.exists(b) {
		t.Error("expected buffers to be in list")
	}

	// Remove the buffers from the list.
	l.remove(a)
	if len(l.list) != 1 || l.list[0] != b {
		t.Error("buffer was not removed correctly")
	}
	l.remove(b)
	if len(l.list) != 0 {
		t.Error("item was not removed correctly")
	}

	// Check what exists is saying now.
	if l.exists(a) || l.exists(b) {
		t.Error("list is empty yet contains buffers?!")
	}

	// Add the buffers again to test Empty.
	l.add(a)
	l.add(b)
	bufs := l.flush()
	if l.list != nil {
		t.Error("list was not nullified")
	}
	if len(bufs) != 2 || bufs[0] != a || bufs[1] != b {
		t.Error("buffers dump incorrect")
	}

	// Try appending again.
	l.add(a)
	if !l.exists(a) || l.exists(b) {
		t.Error("list is in invalid state")
	}
	l.remove(a)
}
