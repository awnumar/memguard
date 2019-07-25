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
		if !GetBufferState(buffers.list[i]).IsAlive {
			t.Error("should not have destroyed excluded buffers")
		}
	}
	if !GetBufferState(key.right).IsAlive || !GetBufferState(key.left).IsAlive || !GetBufferState(buf32).IsAlive {
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
	key.Lock()

	b, _ := NewBuffer(32)
	Scramble(b.Data())

	c, _ := NewBuffer(32)
	Scramble(c.Data())
	c.Freeze()

	if !panics(func() {
		Panic("test")
	}) {
		t.Error("should panic")
	}

	for i := range key.left.Data() {
		if key.left.Data()[i] != 0 || key.right.Data()[i] != 0 {
			t.Error("key not wiped")
		}
		if b.Data()[i] != 0 {
			t.Error("mutable buffer not wiped")
		}
		if c.Data()[i] != 0 {
			t.Error("immutable buffer not wiped")
		}
	}

	b.Destroy()
	c.Destroy()

	key.Unlock()
	key.Destroy()
	key = NewCoffer()
}

func panics(fn func()) (panicked bool) {
	defer func() {
		panicked = (recover() != nil)
	}()
	fn()
	return
}
