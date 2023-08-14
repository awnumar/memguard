package core

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPageAllocAllocInvalidSize(t *testing.T) {
	alloc := NewPageAllocator()

	a, err := alloc.Alloc(0)
	require.Nil(t, a)
	require.ErrorIs(t, err, ErrNullAlloc)

	b, err := alloc.Alloc(-1)
	require.Nil(t, b)
	require.ErrorIs(t, err, ErrNullAlloc)
}

func TestPageAllocAlloc(t *testing.T) {
	alloc := NewPageAllocator()

	b, err := alloc.Alloc(32)
	require.NoError(t, err)
	require.Lenf(t, b, 32, "invalid buffer len %d", len(b))

	o, found := alloc.(*pageAllocator).lookup(b)
	require.True(t, found)
	require.Lenf(t, o.data, 32, "invalid data len %d", len(o.data))
	require.Equalf(t, cap(o.data), 32, "invalid data capacity %d", cap(o.data))
	require.Len(t, o.memory, 3*pageSize)
	require.EqualValues(t, make([]byte, 32), o.data, "container is not zero-filled")

	// Destroy the buffer.
	require.NoError(t, alloc.Free(b))
}

func TestPageAllocLotsOfAllocs(t *testing.T) {
	// Attain a lock to halt the verify & rekey cycle.
	s := NewCoffer()
	s.Lock()

	// Create a local allocator instance
	alloc := NewPageAllocator()
	palloc := alloc.(*pageAllocator)

	for i := 1; i <= 16385; i++ {
		b, err := alloc.Alloc(i)
		require.NoErrorf(t, err, "size: %d", i)

		o, found := palloc.lookup(b)
		require.True(t, found)

		require.Lenf(t, o.data, i, "size: %d", i)
		require.Lenf(t, o.memory, roundToPageSize(i)+2*pageSize, "memory length invalid size: %d", i)
		require.Lenf(t, o.preguard, pageSize, "pre-guard length invalid size: %d", i)
		require.Lenf(t, o.postguard, pageSize, "pre-guard length invalid size: %d", i)
		require.Lenf(t, o.canary, len(o.inner)-i, "canary length invalid size: %d", i)
		require.Zerof(t, len(o.inner)%pageSize, "inner length is not multiple of page size size: %d", i)

		// Fill the data
		for j := range o.data {
			o.data[j] = 1
		}
		require.EqualValuesf(t, bytes.Repeat([]byte{1}, i), o.data, "region rw test failed", "size: %d", i)
		require.NoErrorf(t, alloc.Free(b), "size: %d", i)
	}
}

func TestPageAllocDestroy(t *testing.T) {
	alloc := NewPageAllocator()

	// Allocate a new buffer.
	b, err := alloc.Alloc(32)
	require.NoError(t, err)

	o, found := alloc.(*pageAllocator).lookup(b)
	require.True(t, found)

	// Destroy it and check it is gone...
	require.NoError(t, o.wipe())

	// Pick apart the destruction.
	require.Nil(t, o.data, "data not nil")
	require.Nil(t, o.inner, "inner not nil")
	require.Nil(t, o.preguard, "preguard not nil")
	require.Nil(t, o.postguard, "postguard not nil")
	require.Nil(t, o.canary, "canary not nil")
	require.EqualValues(t, make([]byte, len(o.memory)), o.memory, "memory not zero'ed")

	// Call destroy again to check idempotency.
	require.NoError(t, alloc.Free(b))
}

func TestPageAllocOverflow(t *testing.T) {
	alloc := NewPageAllocator()

	// Allocate a new buffer.
	b, err := alloc.Alloc(32)
	require.NoError(t, err)

	o, found := alloc.(*pageAllocator).lookup(b)
	require.True(t, found)

	// Modify the canary as if we overflow
	o.canary[0] = ^o.canary[0]

	// Destroy it and check it is gone...
	require.ErrorIs(t, alloc.Free(b), ErrBufferOverflow)
}
