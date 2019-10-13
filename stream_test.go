package memguard

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/awnumar/memguard/core"
)

func TestStreamReadWrite(t *testing.T) {
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
	if !bytes.Equal(b, make([]byte, len(b))) {
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

	// Check for end of data error
	n, err = s.Read(b)
	if err != io.EOF {
		t.Error("expected EOF")
	}
	if n != 0 {
		t.Error("expected no data")
	}

	// Write more than the pagesize to the stream
	b = make([]byte, os.Getpagesize()*4+16)
	ScrambleBytes(b)
	copy(b[os.Getpagesize()*4:], []byte("yellow submarine"))
	ref = make([]byte, len(b))
	copy(ref, b)
	n, err = s.Write(b)
	if err != nil {
		t.Error(err)
	}
	if n != len(b) {
		t.Error("not all the data written")
	}
	if !bytes.Equal(b, make([]byte, len(b))) {
		t.Error("buffer not wiped")
	}

	// Read back four pages
	for i := 0; i < 4; i++ {
		c := make([]byte, os.Getpagesize())
		n, err = s.Read(c)
		if err != nil {
			t.Error(err)
		}
		if n != os.Getpagesize() {
			t.Error("incorrect amount of data read")
		}
		if !bytes.Equal(c, ref[i*os.Getpagesize():(i+1)*os.Getpagesize()]) {
			t.Error("data mismatch")
		}
	}

	// Read back the remaining data
	data := make([]byte, 16)
	n, err = s.Read(data)
	if err != nil {
		t.Error(err)
	}
	if n != 16 {
		t.Error("not enough data read")
	}
	if !bytes.Equal(data, []byte("yellow submarine")) {
		t.Error("data mismatch")
	}

	// Should be no data left
	n, err = s.Read(data)
	if err != io.EOF {
		t.Error("expected end of file error")
	}
	if n != 0 {
		t.Error("expected no data left")
	}

	// Test reading less data than is in the chunk
	ScrambleBytes(data)
	ref = make([]byte, len(data))
	copy(ref, data)
	n, err = s.Write(data)
	if err != nil {
		t.Error(err)
	}
	if n != len(data) {
		t.Error("not enough data written")
	}
	if !bytes.Equal(data, make([]byte, len(data))) {
		t.Error("buffer not wiped")
	}
	b = make([]byte, 8)
	n, err = s.Read(b)
	if err != nil {
		t.Error(err)
	}
	if n != len(b) {
		t.Error("not enough data read")
	}
	if !bytes.Equal(b, ref[:8]) {
		t.Error("data mismatch")
	}
	n, err = s.Read(b)
	if err != nil {
		t.Error(err)
	}
	if n != len(b) {
		t.Error("not enough data read")
	}
	if !bytes.Equal(b, ref[8:]) {
		t.Error("data mismatch")
	}
	n, err = s.Read(data)
	if err != io.EOF {
		t.Error("expected end of file error")
	}
	if n != 0 {
		t.Error("expected no data left")
	}

	// Test reading after purging the session
	ScrambleBytes(data)
	n, err = s.Write(data)
	if err != nil {
		t.Error(err)
	}
	if n != len(data) {
		t.Error("not enough data written")
	}
	Purge()
	n, err = s.Read(data)
	if err != core.ErrDecryptionFailed {
		t.Error("expected decryption failed error")
	}
	if n != 0 {
		t.Error("expected zero bytes read")
	}
	if !bytes.Equal(data, make([]byte, len(data))) {
		t.Error("expected data buffer empty")
	}
}
