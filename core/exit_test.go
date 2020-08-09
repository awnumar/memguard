package core

import (
	"bytes"
	"testing"
)

func TestPurge(t *testing.T) {
	// Create a bunch of things to simulate a working environment.
	enclave, err := NewEnclave([]byte("yellow submarine"))
	if err != nil {
		t.Error(err)
	}
	buffer, err := NewBuffer(32)
	if err != nil {
		t.Error(err)
	}

	// Keep a reference to the old key.
	oldKey := key

	// Purge the session.
	Purge()

	// Verify that the buffers list contains only the important buffers.
	buffers.RLock()
	if len(buffers.list) != 3 {
		t.Error("buffers list was not flushed", buffers.list)
	}
	for i := range buffers.list {
		if !buffers.list[i].Alive() {
			t.Error("should not have destroyed excluded buffers")
		}
	}
	if !key.right.Alive() || !key.left.Alive() || !key.rand.Alive() {
		t.Error("buffers left in list aren't the right ones")
	}
	buffers.RUnlock()

	// Verify that the buffer was destroyed.
	if buffer.Alive() {
		t.Error("buffer was not destroyed")
	}

	// Verify that the old key was destroyed.
	if oldKey.left.Alive() || oldKey.right.Alive() {
		t.Error("old key was not destroyed")
	}

	// Verify that the key is not destroyed.
	if !key.left.Alive() || !key.right.Alive() {
		t.Error("current key is destroyed")
	}

	// Verify that the key changed by decrypting the Enclave.
	if _, err := enclave.Open(); err != ErrDecryptionFailed {
		t.Error("expected decryption failed; got", err)
	}

	// Create a buffer with invalid canary.
	b, err := NewBuffer(32)
	if err != nil {
		t.Error(err)
	}
	Scramble(b.inner)
	b.Freeze()
	if !panics(func() {
		Purge()
	}) {
		t.Error("did not panic")
	}
	if !bytes.Equal(b.data, make([]byte, 32)) {
		t.Error("data not wiped")
	}
	buffers.remove(b)
}

func TestPanic(t *testing.T) {
	// Call Panic and check if it panics.
	if !panics(func() {
		Panic("test")
	}) {
		t.Error("did not panic")
	}
}

func panics(fn func()) (panicked bool) {
	defer func() {
		panicked = (recover() != nil)
	}()
	fn()
	return
}
