package memguard

import (
	"bytes"
	"testing"

	"github.com/awnumar/memguard/core"
)

func TestScrambleBytes(t *testing.T) {
	buf := make([]byte, 32)
	ScrambleBytes(buf)
	if bytes.Equal(buf, make([]byte, 32)) {
		t.Error("buffer not scrambled")
	}
}

func TestWipeBytes(t *testing.T) {
	buf := make([]byte, 32)
	ScrambleBytes(buf)
	WipeBytes(buf)
	if !bytes.Equal(buf, make([]byte, 32)) {
		t.Error("buffer not wiped")
	}
}

func TestPurge(t *testing.T) {
	key := NewEnclaveRandom(32)
	buf, err := key.Open()
	if err != nil {
		t.Error(err)
	}
	Purge()
	if buf.IsAlive() {
		t.Error("buffer not destroyed")
	}
	buf, err = key.Open()
	if !core.IsDecryptionFailed(err) {
		t.Error(buf.Bytes(), err)
	}
	if buf != nil {
		t.Error("buffer not nil:", buf)
	}
}
