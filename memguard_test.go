package memguard

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"
)

func TestNew(t *testing.T) {
	b := New(8)
	if len(b.Buffer) != 8 || cap(b.Buffer) != 8 {
		t.Error("length or capacity != required; len, cap =", len(b.Buffer), cap(b.Buffer))
	}
	b.Destroy()
}

func TestNewFromBytes(t *testing.T) {
	b := NewFromBytes([]byte("test"))
	if !bytes.Equal(b.Buffer, []byte("test")) {
		t.Error("b.Buffer != required")
	}
	b.Destroy()
}

func TestPermissions(t *testing.T) {
	b := New(8)
	if b.State != "ReadWrite" {
		t.Error("Unexpected State")
	}

	b.MakeReadOnly()
	if b.State != "ReadOnly" {
		t.Error("Unexpected State")
	}

	b.MakeWriteOnly()
	if b.State != "WriteOnly" {
		t.Error("Unexpected State")
	}

	b.Unlock()
	if b.State != "ReadWrite" {
		t.Error("Unexpected State")
	}

	b.Destroy()
}

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
	b.Destroy()
}

func TestDestroyAll(t *testing.T) {
	b := New(16)
	c := New(16)

	b.Copy([]byte("yellow submarine"))
	c.Copy([]byte("yellow submarine"))

	DestroyAll()
}

func TestCatchInterrupt(t *testing.T) {
	CatchInterrupt(func() {
		return
	})
}

func TestWipeBytes(t *testing.T) {
	b := []byte("yellow submarine")
	WipeBytes(b)
	if !bytes.Equal(b, make([]byte, 16)) {
		t.Error("bytes not wiped; b =", b)
	}
}

func TestDisableCoreDumps(t *testing.T) {
	DisableCoreDumps()
}

func TestRoundPage(t *testing.T) {
	if _roundToPageSize(pageSize) != pageSize {
		t.Error("incorrect rounding;", _roundToPageSize(pageSize))
	}

	if _roundToPageSize(pageSize+1) != 2*pageSize {
		t.Error("incorrect rounding;", _roundToPageSize(pageSize+1))
	}
}

func TestGetBytes(t *testing.T) {
	b := []byte("yellow submarine")

	ptr := unsafe.Pointer(&b[0])
	length := len(b)
	bBytes := _getBytes(uintptr(ptr), length)

	copy(bBytes, []byte("fellow submarine"))

	if !bytes.Equal(b, bBytes) {
		t.Error("pointer does not describe actual memory")
	}
}
