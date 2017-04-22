package memguard

import (
	"bytes"
	"testing"
)

func TestLocking(t *testing.T) {
	// Declare two slices to test on.
	dataOne := []byte("yellow submarine")
	dataTwo := []byte("yellow submarine")

	// Lock them.
	Protect(dataOne)
	Protect(dataTwo)

	// Check if they're zeroed out. They shouldn't be.
	if bytes.Equal(dataOne, make([]byte, 16)) || bytes.Equal(dataOne, make([]byte, 16)) {
		t.Error("Ctitical error: memory zeroed out early")
	}

	// Cleanup.
	Cleanup()

	// Check if data is zeroed out.
	for _, v := range dataOne {
		if v != 0 {
			t.Error("Didn't zero out memory; dataOne =", dataOne)
		}
	}
	for _, v := range dataTwo {
		if v != 0 {
			t.Error("Didn't zero out memory; dataTwo =", dataTwo)
		}
	}
}

func TestMake(t *testing.T) {
	b := Make(32)

	// Test if its length is really 32.
	if len(b) != 32 {
		t.Error("len(b) != 32")
	}

	c := Make(32, 64)

	// Test length and capacities.
	if len(c) != 32 || cap(c) != 64 {
		t.Error("length or capacity incorrect")
	}
}

func TestWipe(t *testing.T) {
	// Declare specimen byte slice.
	b := []byte("yellow submarine")

	// Call wipe.
	Wipe(b)

	// Check if it's wiped.
	for _, v := range b {
		if v != 0 {
			t.Error("Didn't zero out memory; b =", b)
		}
	}
}
