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

	// Verify that the buffers list is empty.
	buffers.RLock()
	if len(buffers.list) != 0 {
		t.Error("buffers list was not flushed")
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
