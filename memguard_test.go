package memguard

import (
	"bytes"
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
	if err != ErrInvalidLength {
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
	if err != ErrInvalidLength {
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

	a.Destroy()

	if equal, err := a.EqualTo([]byte("test")); equal || err != ErrDestroyed {
		t.Error("unexpected return values with destroyed LockedBuffer")
	}
}

func TestReadOnly(t *testing.T) {
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

	if err := b.MarkAsReadOnly(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}

	if err := b.MarkAsReadWrite(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestMove(t *testing.T) {
	// When buf is larger than LockedBuffer.
	b, _ := New(16)
	buf := []byte("this is a very large buffer")
	b.Move(buf)
	if !bytes.Equal(buf, make([]byte, len(buf))) {
		t.Error("expected buf to be nil")
	}
	if !bytes.Equal(b.Buffer, []byte("this is a very l")) {
		t.Error("bytes were't copied properly")
	}
	b.Destroy()

	// When buf is smaller than LockedBuffer.
	b, _ = New(16)
	buf = []byte("diz small buf")
	b.Move(buf)
	if !bytes.Equal(buf, make([]byte, len(buf))) {
		t.Error("expected buf to be nil")
	}
	if !bytes.Equal(b.Buffer[:len(buf)], []byte("diz small buf")) {
		t.Error("bytes weren't copied properly")
	}
	if !bytes.Equal(b.Buffer[len(buf):], make([]byte, 16-len(buf))) {
		t.Error("bytes were't copied properly;", b.Buffer[len(buf):])
	}
	b.Destroy()

	// When buf is equal in size to LockedBuffer.
	b, _ = New(16)
	buf = []byte("yellow submarine")
	b.Move(buf)
	if !bytes.Equal(buf, make([]byte, len(buf))) {
		t.Error("expected buf to be nil")
	}
	if !bytes.Equal(b.Buffer, []byte("yellow submarine")) {
		t.Error("bytes were't copied properly")
	}

	b.MarkAsReadOnly()

	err := b.Move([]byte("test"))
	if err != ErrReadOnly {
		t.Error("expected ErrReadOnly")
	}

	b.Destroy()

	if err := b.Move([]byte("test")); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
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

func TestDuplicate(t *testing.T) {
	b, _ := NewFromBytes([]byte("test"))
	b.MarkAsReadOnly()

	c, err := Duplicate(b)
	if err != nil {
		t.Error("unexpected error")
	}
	if !bytes.Equal(b.Buffer, c.Buffer) {
		t.Error("duplicated buffer has different contents")
	}
	if !c.ReadOnly {
		t.Error("permissions not copied")
	}
	b.Destroy()
	c.Destroy()

	if _, err := Duplicate(b); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestEqual(t *testing.T) {
	b, _ := New(16)
	c, _ := New(16)

	equal, err := Equal(b, c)
	if err != nil {
		t.Error("unexpected error")
	}
	if !equal {
		t.Error("should be equal")
	}

	a, _ := New(8)
	equal, err = Equal(a, b)
	if err != nil {
		t.Error("unexpected error")
	}
	if equal {
		t.Error("should not be equal")
	}

	a.Destroy()
	b.Destroy()
	c.Destroy()

	if _, err := Equal(a, b); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestSplit(t *testing.T) {
	a, _ := NewFromBytes([]byte("xxxxyyyy"))
	a.MarkAsReadOnly()

	b, c, err := Split(a, 4)
	if err != nil {
		t.Error("unexpected error")
	}
	if !bytes.Equal(b.Buffer, []byte("xxxx")) {
		t.Error("first buffer has unexpected value")
	}
	if !bytes.Equal(c.Buffer, []byte("yyyy")) {
		t.Error("second buffer has unexpected value")
	}
	if !b.ReadOnly || !c.ReadOnly {
		t.Error("permissions not preserved")
	}
	if !bytes.Equal(a.Buffer, []byte("xxxxyyyy")) {
		t.Error("original is not preserved")
	}

	b.Destroy()
	c.Destroy()

	if _, _, err := Split(a, 0); err != ErrInvalidLength {
		t.Error("expected ErrInvalidLength")
	}
	if _, _, err := Split(a, 8); err != ErrInvalidLength {
		t.Error("expected ErrInvalidLength")
	}

	a.Destroy()

	if _, _, err := Split(a, 4); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestTrim(t *testing.T) {
	b, _ := NewFromBytes([]byte("xxxxyyyy"))
	b.MarkAsReadOnly()

	c, err := Trim(b, 2, 4)
	if err != nil {
		t.Error("unexpected error")
	}

	if !bytes.Equal(c.Buffer, []byte("xxyy")) {
		t.Error("unexpected value:", c.Buffer)
	}

	if !c.ReadOnly {
		t.Error("unexpected state")
	}

	c.Destroy()
	b.Destroy()

	if _, err := Trim(b, 2, 4); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
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
