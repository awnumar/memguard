package memguard

import (
	"bytes"
	"runtime"
	"sync"
	"testing"
	"unsafe"
)

func TestNew(t *testing.T) {
	b, err := New(8, false)
	if err != nil {
		t.Error("unexpected error")
	}
	if len(b.Buffer) != 8 || cap(b.Buffer) != 8 {
		t.Error("length or capacity != required; len, cap =", len(b.Buffer), cap(b.Buffer))
	}
	b.Destroy()

	c, err := New(0, false)
	if err != ErrInvalidLength {
		t.Error("expected err; got nil")
	}
	if c != nil {
		t.Error("expected nil, got *LockedBuffer")
	}

	a, err := New(8, true)
	if err != nil {
		t.Error("unexpected error")
	}
	if !a.IsReadOnly() {
		t.Error("unexpected state")
	}
	a.Destroy()
}

func TestNewFromBytes(t *testing.T) {
	b, err := NewFromBytes([]byte("test"), false)
	if err != nil {
		t.Error("unexpected error")
	}
	if !bytes.Equal(b.Buffer, []byte("test")) {
		t.Error("b.Buffer != required")
	}
	b.Destroy()

	c, err := NewFromBytes([]byte(""), false)
	if err != ErrInvalidLength {
		t.Error("expected err; got nil")
	}
	if c != nil {
		t.Error("expected nil, got *LockedBuffer")
	}

	a, err := NewFromBytes([]byte("test"), true)
	if err != nil {
		t.Error("unexpected error")
	}
	if !a.IsReadOnly() {
		t.Error("unexpected state")
	}
	a.Destroy()
}

func TestNewRandom(t *testing.T) {
	b, _ := NewRandom(32, false)

	if bytes.Equal(b.Buffer, make([]byte, 32)) {
		t.Error("was not filled with random data")
	}

	b.Destroy()

	c, err := NewRandom(0, false)
	if err != ErrInvalidLength {
		t.Error("expected ErrInvalidLength")
	}
	if c != nil {
		t.Error("expected nil, got *LockedBuffer")
	}

	a, err := NewRandom(8, true)
	if err != nil {
		t.Error("unexpected error")
	}
	if !a.IsReadOnly() {
		t.Error("unexpected state")
	}
	a.Destroy()
}

func TestGetMetadata(t *testing.T) {
	b, _ := New(8, false)

	if val := b.IsReadOnly(); val != false {
		t.Error("incorrect value")
	}
	if val := b.IsDestroyed(); val != false {
		t.Error("incorrect value")
	}

	b.MarkAsReadOnly()
	if val := b.IsReadOnly(); val != true {
		t.Error("incorrect value")
	}

	b.Destroy()
	if val := b.IsDestroyed(); val != true {
		t.Error("incorrect value")
	}
}

func TestEqualTo(t *testing.T) {
	a, _ := NewFromBytes([]byte("test"), false)

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
	b, _ := New(8, false)
	if b.IsReadOnly() {
		t.Error("Unexpected State")
	}

	// Test each twice for completeness.
	if err := b.MarkAsReadOnly(); err != nil {
		t.Error("unexpected error")
	}
	if !b.IsReadOnly() {
		t.Error("Unexpected State")
	}
	if err := b.MarkAsReadOnly(); err != nil {
		t.Error("unexpected error")
	}
	if !b.IsReadOnly() {
		t.Error("Unexpected State")
	}

	if err := b.MarkAsReadWrite(); err != nil {
		t.Error("unexpected error")
	}
	if b.IsReadOnly() {
		t.Error("Unexpected State")
	}
	if err := b.MarkAsReadWrite(); err != nil {
		t.Error("unexpected error")
	}
	if b.IsReadOnly() {
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
	b, _ := New(16, false)
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
	b, _ = New(16, false)
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
	b, _ = New(16, false)
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

func TestFillRandomBytes(t *testing.T) {
	a, _ := New(32, false)
	a.FillRandomBytes()

	if a.Buffer == nil {
		t.Error("not random")
	}

	WipeBytes(a.Buffer)
	a.FillRandomBytesAt(16, 16)

	if !bytes.Equal(a.Buffer[:16], make([]byte, 16)) || bytes.Equal(a.Buffer[16:], make([]byte, 16)) {
		t.Error("incorrect offset/size;", a.Buffer[:16], a.Buffer[16:])
	}

	a.MarkAsReadOnly()
	if err := a.FillRandomBytes(); err != ErrReadOnly {
		t.Error("expected ErrReadOnly")
	}

	a.Destroy()
	if err := a.FillRandomBytes(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestDestroyAll(t *testing.T) {
	b, _ := New(16, false)
	c, _ := New(16, false)

	b.Copy([]byte("yellow submarine"))
	c.Copy([]byte("yellow submarine"))

	DestroyAll()

	if b.Buffer != nil || c.Buffer != nil {
		t.Error("expected buffers to be nil")
	}

	if b.IsReadOnly() || c.IsReadOnly() {
		t.Error("expected permissions to be empty")
	}

	if !b.IsDestroyed() || !c.IsDestroyed() {
		t.Error("expected destroy flag to be set")
	}
}

func TestConcatenate(t *testing.T) {
	a, _ := NewFromBytes([]byte("xxxx"), true)
	b, _ := NewFromBytes([]byte("yyyy"), false)

	c, err := Concatenate(a, b)
	if err != nil {
		t.Error("unexpected error")
	}

	if !bytes.Equal(c.Buffer, []byte("xxxxyyyy")) {
		t.Error("unexpected output;", c.Buffer)
	}
	if !c.IsReadOnly() {
		t.Error("expected ReadOnly")
	}

	a.Destroy()
	b.Destroy()
	c.Destroy()

	if _, err := Concatenate(a, b); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestDuplicate(t *testing.T) {
	b, _ := NewFromBytes([]byte("test"), false)
	b.MarkAsReadOnly()

	c, err := Duplicate(b)
	if err != nil {
		t.Error("unexpected error")
	}
	if !bytes.Equal(b.Buffer, c.Buffer) {
		t.Error("duplicated buffer has different contents")
	}
	if !c.IsReadOnly() {
		t.Error("permissions not copied")
	}
	b.Destroy()
	c.Destroy()

	if _, err := Duplicate(b); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestEqual(t *testing.T) {
	b, _ := New(16, false)
	c, _ := New(16, false)

	equal, err := Equal(b, c)
	if err != nil {
		t.Error("unexpected error")
	}
	if !equal {
		t.Error("should be equal")
	}

	a, _ := New(8, false)
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
	a, _ := NewFromBytes([]byte("xxxxyyyy"), false)
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
	if !b.IsReadOnly() || !c.IsReadOnly() {
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
	b, _ := NewFromBytes([]byte("xxxxyyyy"), false)
	b.MarkAsReadOnly()

	c, err := Trim(b, 2, 4)
	if err != nil {
		t.Error("unexpected error")
	}

	if !bytes.Equal(c.Buffer, []byte("xxyy")) {
		t.Error("unexpected value:", c.Buffer)
	}

	if !c.IsReadOnly() {
		t.Error("unexpected state")
	}
	c.Destroy()

	if _, err := Trim(b, 4, 0); err != ErrInvalidLength {
		t.Error("expected ErrInvalidLength")
	}

	b.Destroy()

	if _, err := Trim(b, 2, 4); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestCatchInterrupt(t *testing.T) {
	CatchInterrupt(func() {
		return
	})

	var i int
	catchInterruptOnce.Do(func() {
		i++
	})
	if i != 0 {
		t.Error("sync.Once failed")
	}
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
	wg.Add(16)

	b, _ := New(4, false)
	for i := 0; i < 16; i++ {
		go func() {
			CatchInterrupt(func() {
				return
			})

			b.MarkAsReadOnly()
			b.MarkAsReadWrite()

			b.Move([]byte("Test"))
			b.Copy([]byte("test"))

			b.FillRandomBytes()

			wg.Done()
		}()
	}

	wg.Wait()
	b.Destroy()
}

func TestDisableUnixCoreDumps(t *testing.T) {
	DisableUnixCoreDumps()
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

func TestFinalizer(t *testing.T) {
	b, err := New(8, false)
	if err != nil {
		t.Error("unexpected error")
	}
	ib := b.lockedBuffer

	c, err := New(8, false)
	if err != nil {
		t.Error("unexpected error")
	}
	ic := c.lockedBuffer

	if ib.IsDestroyed() != false {
		t.Error("expected b to not be destroyed")
	}
	if ic.IsDestroyed() != false {
		t.Error("expected c to not be destroyed")
	}

	runtime.KeepAlive(b)
	// b is now unreachable

	runtime.GC() // should collect b
	for ib.IsDestroyed() != true {
		runtime.Gosched()
	}

	if ic.IsDestroyed() != false {
		t.Error("expected c to not be destroyed")
	}

	runtime.KeepAlive(c)
	// c is now unreachable

	runtime.GC() // should collect c
	for ic.IsDestroyed() != true {
		runtime.Gosched()
	}
}
