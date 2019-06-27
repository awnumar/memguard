package memguard

import (
	"bytes"
	"testing"

	"github.com/awnumar/memguard/core"
)

func TestNewEnclave(t *testing.T) {
	e := NewEnclave([]byte("yellow submarine"))
	if e == nil {
		t.Error("got nil enclave")
	}
	data, err := e.Open()
	if err != nil {
		t.Error("unexpected error:", err)
	}
	if !bytes.Equal(data.Bytes(), []byte("yellow submarine")) {
		t.Error("data doesn't match input")
	}
	data.Destroy()
}

func TestNewEnclaveRandom(t *testing.T) {
	e := NewEnclaveRandom(32)
	if e == nil {
		t.Error("got nil enclave")
	}
	data, err := e.Open()
	if err != nil {
		t.Error("unexpected error:", err)
	}
	if len(data.Bytes()) != 32 || cap(data.Bytes()) != 32 {
		t.Error("buffer sizes incorrect")
	}
	if bytes.Equal(data.Bytes(), make([]byte, 32)) {
		t.Error("buffer not randomised")
	}
	data.Destroy()
}

func TestOpen(t *testing.T) {
	e := NewEnclave([]byte("yellow submarine"))
	if e == nil {
		t.Error("got nil enclave")
	}
	b, err := e.Open()
	if err != nil {
		t.Error("unexpected error;", err)
	}
	if b == nil {
		t.Error("buffer should not be nil")
	}
	if !bytes.Equal(b.Bytes(), []byte("yellow submarine")) {
		t.Error("data does not match")
	}
	Purge() // reset the session
	b, err = e.Open()
	if err != core.ErrDecryptionFailed {
		t.Error("expected decryption error; got", err)
	}
	if b != nil {
		t.Error("buffer should be nil")
	}
}
