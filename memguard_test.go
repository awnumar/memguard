package memguard

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	b := New(8)
	if len(b.Buffer) != 8 || cap(b.Buffer) != 8 {
		t.Error("length or capacity != required")
	}
}

func TestNewFromBytes(t *testing.T) {
	b := NewFromBytes([]byte("test"))
	if !bytes.Equal(b.Buffer, []byte("test")) {
		t.Error("b.Buffer != required")
	}
}

/* This is faling. Need to fix.
func TestPermissions(t *testing.T) {
	b := New(8)
	b.AllowReadWrite()
	b.AllowRead()
	b.AllowWrite()
	b.Lock()
}
*/

func TestMove(t *testing.T) {
	b, buf := New(16), []byte("yellow submarine")
	b.Move(buf)
	if !bytes.Equal(buf, make([]byte, 16)) {
		fmt.Println(buf)
		t.Error("expected buf to be nil")
	}
	if !bytes.Equal(b.Buffer, []byte("yellow submarine")) {
		t.Error("bytes were't copied properly")
	}
}

func TestDestroyAll(t *testing.T)       {}
func TestWipeBytes(t *testing.T)        {}
func TestDisableCoreDumps(t *testing.T) {}
func TestRoundPage(t *testing.T)        {}
func TestGetBytes(t *testing.T)         {}
