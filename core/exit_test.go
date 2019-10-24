package core

import (
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
	if buffer.alive {
		t.Error("buffer was not destroyed")
	}

	// Verify that the key is not destroyed.
	if !key.left.alive || !key.right.alive {
		t.Error("key was destroyed")
	}

	// Verify that the key changed by decrypting the Enclave.
	if _, err := Open(enclave); err != ErrDecryptionFailed {
		t.Error("expected decryption failed; got", err)
	}
}

func TestPanic(t *testing.T) {
	// Call Panic and check if it panics.
	if !panics(func() {
		Panic("test")
	}) {
		t.Error("should panic")
	}
}

func panics(fn func()) (panicked bool) {
	defer func() {
		panicked = (recover() != nil)
	}()
	fn()
	return
}
