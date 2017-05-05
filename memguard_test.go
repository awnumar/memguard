package memguard

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"unsafe"
)

func TestNew(t *testing.T) {
	b, err := New(8)
	if err != nil {
		t.Error("unexpected error")
	}
	if len(b.Buffer) != 8 || cap(b.Buffer) != 8 {
		t.Error("length or capacity != required; len, cap =", len(b.Buffer), cap(b.Buffer))
	}
	b.Destroy()

	c, err := New(0)
	if err == nil {
		t.Error("expected err; got nil")
	}
	c.Destroy()
}

func TestNewFromBytes(t *testing.T) {
	b, err := NewFromBytes([]byte("test"))
	if err != nil {
		t.Error("unexpected error")
	}
	if !bytes.Equal(b.Buffer, []byte("test")) {
		t.Error("b.Buffer != required")
	}
	b.Destroy()

	c, err := NewFromBytes([]byte(""))
	if err == nil {
		t.Error("expected err; got nil")
	}
	c.Destroy()
}

func TestEqualTo(t *testing.T) {
	a, _ := NewFromBytes([]byte("test"))

	equal, err := a.EqualTo([]byte("test"))
	if err != nil {
		t.Error("unexpected error")
	}

	if !equal {
		t.Error("should be equal")
	}

	equal, err = a.EqualTo([]byte("toast"))
	if err != nil {
		t.Error("unexpected error")
	}

	if equal {
		t.Error("should not be equal")
	}
}

func TestPermissions(t *testing.T) {
	b, _ := New(8)
	if b.ReadOnly {
		t.Error("Unexpected State")
	}

	b.MarkAsReadOnly()
	if !b.ReadOnly {
		t.Error("Unexpected State")
	}

	b.MarkAsReadWrite()
	if b.ReadOnly {
		t.Error("Unexpected State")
	}

	b.Destroy()
}

func TestMove(t *testing.T) {
	b, _ := New(16)
	buf := []byte("yellow submarine")

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
	b, _ := New(16)
	c, _ := New(16)

	b.Copy([]byte("yellow submarine"))
	c.Copy([]byte("yellow submarine"))

	DestroyAll()

	if b.Buffer != nil || c.Buffer != nil {
		t.Error("expected buffers to be nil")
	}

	if b.ReadOnly || c.ReadOnly {
		t.Error("expected permissions to be empty")
	}

	if !b.Destroyed || !c.Destroyed {
		t.Error("expected destroy flag to be set")
	}
}

func TestDestroyedFlag(t *testing.T) {
	b, _ := New(4)
	b.Destroy()

	if err := b.Copy([]byte("test")); err == nil {
		t.Error("expected ErrDestroyed; got nil")
	}

	if err := b.Move([]byte("test")); err == nil {
		t.Error("expected ErrDestroyed; got nil")
	}

	if err := b.MarkAsReadOnly(); err == nil {
		t.Error("expected ErrDestroyed; got nil")
	}

	if err := b.MarkAsReadWrite(); err == nil {
		t.Error("expected ErrDestroyed; got nil")
	}

	if _, err := b.EqualTo([]byte("test")); err == nil {
		t.Error("expected ErrDestroyed; got nil")
	}

	if err := b.Trim(10); err == nil {
		t.Error("expected ErrDestroyed; got nil")
	}

	if _, err := Duplicate(b); err == nil {
		t.Error("expected ErrDestroyed; got nil")
	}

	if _, err := Equal(b, new(LockedBuffer)); err == nil {
		t.Error("expected ErrDestroyed; got nil")
	}

	if _, _, err := Split(b, 10); err == nil {
		t.Error("expected ErrDestroyed; got nil")
	}
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

func TestConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(4)

	b, _ := New(4)
	for i := 0; i < 4; i++ {
		go func() {
			CatchInterrupt(func() {
				return
			})

			b.MarkAsReadOnly()
			b.MarkAsReadWrite()

			b.Move([]byte("Test"))
			b.Copy([]byte("test"))

			WipeBytes(b.Buffer)

			wg.Done()
		}()
	}

	wg.Wait()
	b.Destroy()
}

func TestDisableCoreDumps(t *testing.T) {
	DisableCoreDumps()
}

func TestRoundPage(t *testing.T) {
	if roundToPageSize(pageSize) != pageSize {
		t.Error("incorrect rounding;", roundToPageSize(pageSize))
	}

	if roundToPageSize(pageSize+1) != 2*pageSize {
		t.Error("incorrect rounding;", roundToPageSize(pageSize+1))
	}
}

func TestGetBytes(t *testing.T) {
	b := []byte("yellow submarine")

	ptr := unsafe.Pointer(&b[0])
	length := len(b)
	bBytes := getBytes(uintptr(ptr), length)

	copy(bBytes, []byte("fellow submarine"))

	if !bytes.Equal(b, bBytes) {
		t.Error("pointer does not describe actual memory")
	}
}
