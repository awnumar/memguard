package memguard

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/awnumar/memguard/core"
)

func write(t *testing.T, s *Stream, b []byte) {
	n, err := s.Write(b)
	if err != nil {
		t.Error("write should always succeed", err)
	}
	if n != len(b) {
		t.Error("not all data was written")
	}
	if !bytes.Equal(b, make([]byte, len(b))) {
		t.Error("buffer not wiped")
	}
}

func read(t *testing.T, s *Stream, ref []byte, expectedErr error) {
	b := make([]byte, len(ref))
	n, err := s.Read(b)
	if err != expectedErr {
		t.Error("Expected", expectedErr, "got;", err)
	}
	if n != len(b) {
		t.Error("not enough data read")
	}
	if !bytes.Equal(ref, b) {
		t.Error("data mismatch")
	}
}

func TestStreamReadWrite(t *testing.T) {
	// Create new stream object.
	s := NewStream()

	// Initialise some data to store.
	b := make([]byte, 1024)
	ScrambleBytes(b)
	ref := make([]byte, len(b))
	copy(ref, b)

	// Write the data to the stream.
	write(t, s, b)

	// Read the data back.
	read(t, s, ref, nil)

	// Check for end of data error
	read(t, s, nil, io.EOF)

	// Write more than the pagesize to the stream
	b = make([]byte, os.Getpagesize()*4+16)
	ScrambleBytes(b)
	copy(b[os.Getpagesize()*4:], []byte("yellow submarine"))
	ref = make([]byte, len(b))
	copy(ref, b)
	write(t, s, b)

	// Read back four pages
	for i := 0; i < 4; i++ {
		read(t, s, ref[i*os.Getpagesize():(i+1)*os.Getpagesize()], nil)
	}

	// Read back the remaining data
	read(t, s, []byte("yellow submarine"), nil)

	// Should be no data left
	read(t, s, nil, io.EOF)

	// Test reading less data than is in the chunk
	data := make([]byte, 16)
	ScrambleBytes(data)
	ref = make([]byte, len(data))
	copy(ref, data)
	write(t, s, data)
	write(t, s, data) // have two enclaves in the stream, 32 bytes total
	read(t, s, ref[:8], io.ErrShortBuffer)
	read(t, s, ref[8:], nil)
	read(t, s, make([]byte, 16), nil)
	read(t, s, nil, io.EOF)

	// Test reading after purging the session
	ScrambleBytes(data)
	write(t, s, data)
	Purge()
	read(t, s, nil, core.ErrDecryptionFailed)
}

func TestStreamingSanity(t *testing.T) {

}
