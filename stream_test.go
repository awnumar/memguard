package memguard

import (
	"bytes"
	"testing"
)

func TestReadWrite(t *testing.T) {
	// Create new stream object.
	s := NewStream()

	// Initialise some data to store.
	b := make([]byte, 1024)
	ScrambleBytes(b)
	ref := make([]byte, len(b))
	copy(ref, b)

	// Write the data to the stream.
	n, err := s.Write(b)
	if err != nil {
		t.Error("write should always succeed", err)
	}
	if n != len(b) {
		t.Error("not all data was written")
	}
	if bytes.Equal(ref, b) {
		t.Error("buffer not wiped")
	}

	// Read the data back.
	n, err = s.Read(b)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(ref, b) {
		t.Error("data mismatch")
	}
}
