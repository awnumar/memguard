package core

import (
	"bytes"
	"testing"
)

func TestEnclaveInit(t *testing.T) {
	if key == nil {
		t.Error("key is nil")
	}

	view, err := getKey().View()
	if err != nil {
		t.Error("unexpected error;", err)
	}

	if view.Data() == nil || len(view.Data()) != 32 {
		t.Error("key is invalid")
	}

	if bytes.Equal(view.Data(), make([]byte, 32)) {
		t.Error("key is zero")
	}

	view.Destroy()
}

func TestNewEnclave(t *testing.T) {
	// Initialise some sample plaintext.
	data := []byte("yellow submarine")

	// Create the Enclave object from this data.
	e, err := NewEnclave(data)
	if err != nil {
		t.Error(err)
	}

	// Check that the buffer has been wiped.
	if !bytes.Equal(data, make([]byte, 16)) {
		t.Error("data buffer was not wiped")
	}

	// Verify the length of the ciphertext is correct.
	if len(e.Ciphertext) != len(data)+Overhead {
		t.Error("ciphertext has unexpected length;", len(e.Ciphertext))
	}

	// Attempt with an empty data slice.
	data = make([]byte, 0)
	_, err = NewEnclave(data)
	if err != ErrNullEnclave {
		t.Error("expected ErrNullEnclave; got", err)
	}
}

func TestSeal(t *testing.T) {
	// Create a new buffer for testing with.
	b, err := NewBuffer(32)
	if err != nil {
		t.Error(err)
	}

	// Encrypt it into an Enclave.
	e, err := Seal(b)
	if err != nil {
		t.Error(err)
	}

	// Do a sanity check on the length of the ciphertext.
	if len(e.Ciphertext) != 32+Overhead {
		t.Error("ciphertext has unexpected length:", len(e.Ciphertext))
	}

	// Check that the buffer was destroyed.
	if b.alive {
		t.Error("buffer was not consumed")
	}

	// Decrypt the enclave into a new buffer.
	buf, err := Open(e)
	if err != nil {
		t.Error(err)
	}

	// Check that the decrypted data is correct.
	if !bytes.Equal(buf.Data(), make([]byte, 32)) {
		t.Error("decrypted data does not match original")
	}

	// Attempt sealing the destroyed buffer.
	e, err = Seal(b)
	if err != ErrBufferExpired {
		t.Error("expected ErrBufferExpired; got", err)
	}
	if e != nil {
		t.Error("expected nil enclave in error case")
	}

	// Destroy the hanging buffer.
	buf.Destroy()
}

func TestOpen(t *testing.T) {
	// Initialise an enclave to test on.
	data := []byte("yellow submarine")
	e, err := NewEnclave(data)
	if err != nil {
		t.Error(err)
	}

	// Open it.
	buf, err := Open(e)
	if err != nil {
		t.Error(err)
	}

	// Sanity check the output.
	if !bytes.Equal(buf.Data(), []byte("yellow submarine")) {
		t.Error("decrypted data does not match original")
	}
	buf.Destroy()

	// Modify the ciphertext to trigger an error case.
	for i := range e.Ciphertext {
		e.Ciphertext[i] = 0xdb
	}

	// Check for the error.
	buf, err = Open(e)
	if err != ErrDecryptionFailed {
		t.Error("expected decryption error; got", err)
	}
	if buf != nil {
		t.Error("expected nil buffer in error case")
	}
}

func TestEnclaveSize(t *testing.T) {
	if EnclaveSize(&Enclave{make([]byte, 1234)}) != 1234-Overhead {
		t.Error("invalid enclave size")
	}
}
