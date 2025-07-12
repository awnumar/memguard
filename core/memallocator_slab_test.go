package core

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlabAllocAllocInvalidSize(t *testing.T) {
	alloc := NewSlabAllocator()

	a, err := alloc.Alloc(0)
	require.Nil(t, a)
	require.ErrorIs(t, err, ErrNullAlloc)

	b, err := alloc.Alloc(-1)
	require.Nil(t, b)
	require.ErrorIs(t, err, ErrNullAlloc)
}

func TestSlabAllocAlloc(t *testing.T) {
	alloc := NewSlabAllocator()

	b, err := alloc.Alloc(32)
	require.NoError(t, err)
	require.Lenf(t, b, 32, "invalid buffer len %d", len(b))

	require.Lenf(t, b, 32, "invalid data len %d", len(b))
	require.Equalf(t, cap(b), 32, "invalid data capacity %d", cap(b))
	// require.Len(t, o.memory, 3*pageSize)
	// require.EqualValues(t, make([]byte, 32), o.data, "container is not zero-filled")

	// Destroy the buffer.
	require.NoError(t, alloc.Free(b))
}

func TestSlabAllocLotsOfAllocs(t *testing.T) {
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

func TestSlabAllocDestroy(t *testing.T) {
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

func TestSlabAllocOverflow(t *testing.T) {
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
