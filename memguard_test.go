package memguard

import (
	"bytes"
	"runtime"
	"sync"
	"testing"
	"unsafe"
)

func TestNew(t *testing.T) {
	b, err := NewImmutable(8)
	if err != nil {
		t.Error("unexpected error")
	}
	for i := range b.buffer {
		if b.buffer[i] != 0 {
			t.Error("buffer not zero-filled", b.buffer)
		}
	}
	if len(b.Bytes()) != 8 || cap(b.Bytes()) != 8 {
		t.Error("length or capacity != required; len, cap =", len(b.Bytes()), cap(b.Bytes()))
	}
	if b.IsMutable() {
		t.Error("unexpected state")
	}
	b.Destroy()

	c, err := NewImmutable(0)
	if err != ErrInvalidLength {
		t.Error("expected err; got nil")
	}
	if c != nil {
		t.Error("expected nil, got *Enclave")
	}

	a, err := NewMutable(8)
	if err != nil {
		t.Error("unexpected error")
	}
	for i := range b.buffer {
		if b.buffer[i] != 0 {
			t.Error("buffer not zero-filled", b.buffer)
		}
	}
	if !a.IsMutable() {
		t.Error("unexpected state")
	}
	a.Destroy()
}

func TestNewFromBytes(t *testing.T) {
	b, err := NewImmutableFromBytes([]byte("test"))
	if err != nil {
		t.Error("unexpected error")
	}
	if !bytes.Equal(b.Bytes(), []byte("test")) {
		t.Error("b.Bytes() != required")
	}
	if b.IsMutable() {
		t.Error("unexpected state")
	}
	b.Destroy()

	c, err := NewImmutableFromBytes([]byte(""))
	if err != ErrInvalidLength {
		t.Error("expected err; got nil")
	}
	if c != nil {
		t.Error("expected nil, got *Enclave")
	}

	a, err := NewMutableFromBytes([]byte("test"))
	if err != nil {
		t.Error("unexpected error")
	}
	if !a.IsMutable() {
		t.Error("unexpected state")
	}
	a.Destroy()
}

func TestNewRandom(t *testing.T) {
	b, _ := NewImmutableRandom(32)
	if bytes.Equal(b.Bytes(), make([]byte, 32)) {
		t.Error("was not filled with random data")
	}
	if b.IsMutable() {
		t.Error("unexpected state")
	}

	b.Destroy()

	c, err := NewImmutableRandom(0)
	if err != ErrInvalidLength {
		t.Error("expected ErrInvalidLength")
	}
	if c != nil {
		t.Error("expected nil, got *Enclave")
	}

	a, err := NewMutableRandom(8)
	if err != nil {
		t.Error("unexpected error")
	}
	if !a.IsMutable() {
		t.Error("unexpected state")
	}
	a.Destroy()
}

func TestBytes(t *testing.T) {
	b, _ := NewImmutableRandom(8)

	if !bytes.Equal(b.buffer, b.Bytes()) {
		t.Error("buffers inequal")
	}

	b.Destroy()

	if len(b.Bytes()) != 0 || cap(b.Bytes()) != 0 {
		t.Error("expected zero length")
	}
}

func TestUint8(t *testing.T) {
	b, _ := NewImmutableRandom(8)

	x, err := b.Uint8()
	if err != nil {
		t.Error("unexpected error")
	}
	if !bytes.Equal(b.buffer, x) {
		t.Error("conversion failed")
	}

	if &b.buffer[0] != &x[0] {
		t.Error("conversion points incorrectly")
	}
	if len(x) != 8 || cap(x) != 8 {
		t.Error("unexpected length or capacity")
	}

	b.Destroy()

	if _, err := b.Uint8(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestUint16(t *testing.T) {
	b, _ := NewImmutable(8)
	c, _ := NewImmutable(9)

	x, err := b.Uint16()
	if err != nil {
		t.Error("unexpected error")
	}
	_, err = c.Uint16()
	if err != ErrInvalidConversion {
		t.Error("expected ErrInvalidConversion")
	}

	if unsafe.Pointer(&b.buffer[0]) != unsafe.Pointer(&x[0]) {
		t.Error("conversion points incorrectly")
	}
	if len(x) != 4 || cap(x) != 4 {
		t.Error("unexpected length or capacity")
	}

	b.Destroy()
	c.Destroy()

	if _, err := b.Uint16(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestUint32(t *testing.T) {
	b, _ := NewImmutable(8)
	c, _ := NewImmutable(9)

	x, err := b.Uint32()
	if err != nil {
		t.Error("unexpected error")
	}
	_, err = c.Uint32()
	if err != ErrInvalidConversion {
		t.Error("expected ErrInvalidConversion")
	}

	if unsafe.Pointer(&b.buffer[0]) != unsafe.Pointer(&x[0]) {
		t.Error("conversion points incorrectly")
	}
	if len(x) != 2 || cap(x) != 2 {
		t.Error("unexpected length or capacity")
	}

	b.Destroy()
	c.Destroy()

	if _, err := b.Uint32(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestUint64(t *testing.T) {
	b, _ := NewImmutable(8)
	c, _ := NewImmutable(9)

	x, err := b.Uint64()
	if err != nil {
		t.Error("unexpected error")
	}
	_, err = c.Uint64()
	if err != ErrInvalidConversion {
		t.Error("expected ErrInvalidConversion")
	}

	if unsafe.Pointer(&b.buffer[0]) != unsafe.Pointer(&x[0]) {
		t.Error("conversion points incorrectly")
	}
	if len(x) != 1 || cap(x) != 1 {
		t.Error("unexpected length or capacity")
	}

	b.Destroy()
	c.Destroy()

	if _, err := b.Uint64(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestInt8(t *testing.T) {
	b, _ := NewImmutable(8)
	c, _ := NewImmutable(9)

	x, err := b.Int8()
	if err != nil {
		t.Error("unexpected error")
	}

	if unsafe.Pointer(&b.buffer[0]) != unsafe.Pointer(&x[0]) {
		t.Error("conversion points incorrectly")
	}
	if len(x) != 8 || cap(x) != 8 {
		t.Error("unexpected length or capacity")
	}

	b.Destroy()
	c.Destroy()

	if _, err := b.Int8(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestInt16(t *testing.T) {
	b, _ := NewImmutable(8)
	c, _ := NewImmutable(9)

	x, err := b.Int16()
	if err != nil {
		t.Error("unexpected error")
	}
	_, err = c.Int16()
	if err != ErrInvalidConversion {
		t.Error("expected ErrInvalidConversion")
	}

	if unsafe.Pointer(&b.buffer[0]) != unsafe.Pointer(&x[0]) {
		t.Error("conversion points incorrectly")
	}
	if len(x) != 4 || cap(x) != 4 {
		t.Error("unexpected length or capacity")
	}

	b.Destroy()
	c.Destroy()

	if _, err := b.Int16(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestInt32(t *testing.T) {
	b, _ := NewImmutable(8)
	c, _ := NewImmutable(9)

	x, err := b.Int32()
	if err != nil {
		t.Error("unexpected error")
	}
	_, err = c.Int32()
	if err != ErrInvalidConversion {
		t.Error("expected ErrInvalidConversion")
	}

	if unsafe.Pointer(&b.buffer[0]) != unsafe.Pointer(&x[0]) {
		t.Error("conversion points incorrectly")
	}
	if len(x) != 2 || cap(x) != 2 {
		t.Error("unexpected length or capacity")
	}

	b.Destroy()
	c.Destroy()

	if _, err := b.Int32(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestInt64(t *testing.T) {
	b, _ := NewImmutable(8)
	c, _ := NewImmutable(9)

	x, err := b.Int64()
	if err != nil {
		t.Error("unexpected error")
	}
	_, err = c.Int64()
	if err != ErrInvalidConversion {
		t.Error("expected ErrInvalidConversion")
	}

	if unsafe.Pointer(&b.buffer[0]) != unsafe.Pointer(&x[0]) {
		t.Error("conversion points incorrectly")
	}
	if len(x) != 1 || cap(x) != 1 {
		t.Error("unexpected length or capacity")
	}

	b.Destroy()
	c.Destroy()

	if _, err := b.Int64(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestGetMetadata(t *testing.T) {
	b, _ := NewMutable(8)

	if b.IsMutable() != true {
		t.Error("incorrect value")
	}
	if b.IsDestroyed() != false {
		t.Error("incorrect value")
	}

	b.MakeImmutable()
	if b.IsMutable() != false {
		t.Error("incorrect value")
	}

	b.Destroy()
	if b.IsDestroyed() != true {
		t.Error("incorrect value")
	}
}

func TestEqualTo(t *testing.T) {
	a, _ := NewImmutableFromBytes([]byte("test"))

	equal, err := a.EqualBytes([]byte("test"))
	if err != nil {
		t.Error("unexpected error")
	}

	if !equal {
		t.Error("should be equal")
	}

	equal, err = a.EqualBytes([]byte("toast"))
	if err != nil {
		t.Error("unexpected error")
	}

	if equal {
		t.Error("should not be equal")
	}

	a.Destroy()

	if equal, err := a.EqualBytes([]byte("test")); equal || err != ErrDestroyed {
		t.Error("unexpected return values with destroyed Enclave")
	}
}

func TestReadOnly(t *testing.T) {
	b, _ := NewMutable(8)

	if err := b.MakeImmutable(); err != nil {
		t.Error("unexpected error")
	}
	if b.IsMutable() {
		t.Error("unexpected state")
	}
	if err := b.MakeMutable(); err != nil {
		t.Error("unexpected error")
	}
	if !b.IsMutable() {
		t.Error("unexpected state")
	}

	b.Destroy()

	if err := b.MakeImmutable(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}

	if err := b.MakeMutable(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestMove(t *testing.T) {
	// When buf is larger than Enclave.
	b, _ := NewMutable(16)
	buf := []byte("this is a very large buffer")
	b.Move(buf)
	if !bytes.Equal(buf, make([]byte, len(buf))) {
		t.Error("expected buf to be nil")
	}
	if !bytes.Equal(b.Bytes(), []byte("this is a very l")) {
		t.Error("bytes were't copied properly")
	}
	b.Destroy()

	// When buf is smaller than Enclave.
	b, _ = NewMutable(16)
	buf = []byte("diz small buf")
	b.Move(buf)
	if !bytes.Equal(buf, make([]byte, len(buf))) {
		t.Error("expected buf to be nil")
	}
	if !bytes.Equal(b.Bytes()[:len(buf)], []byte("diz small buf")) {
		t.Error("bytes weren't copied properly")
	}
	if !bytes.Equal(b.Bytes()[len(buf):], make([]byte, 16-len(buf))) {
		t.Error("bytes were't copied properly;", b.Bytes()[len(buf):])
	}
	b.Destroy()

	// When buf is equal in size to Enclave.
	b, _ = NewMutable(16)
	buf = []byte("yellow submarine")
	b.Move(buf)
	if !bytes.Equal(buf, make([]byte, len(buf))) {
		t.Error("expected buf to be nil")
	}
	if !bytes.Equal(b.Bytes(), []byte("yellow submarine")) {
		t.Error("bytes were't copied properly")
	}

	b.MakeImmutable()

	err := b.Move([]byte("test"))
	if err != ErrImmutable {
		t.Error("expected ErrImmutable")
	}

	b.Destroy()

	if err := b.Move([]byte("test")); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestFillRandomBytes(t *testing.T) {
	a, _ := NewMutable(32)
	a.FillRandomBytes()

	if bytes.Equal(a.Bytes(), make([]byte, 32)) {
		t.Error("not random")
	}

	a.Wipe()
	a.FillRandomBytesAt(16, 16)

	if !bytes.Equal(a.Bytes()[:16], make([]byte, 16)) || bytes.Equal(a.Bytes()[16:], make([]byte, 16)) {
		t.Error("incorrect offset/size;", a.Bytes()[:16], a.Bytes()[16:])
	}

	a.MakeImmutable()
	if err := a.FillRandomBytes(); err != ErrImmutable {
		t.Error("expected ErrImmutable")
	}

	a.Destroy()
	if err := a.FillRandomBytes(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestDestroyAll(t *testing.T) {
	oldCanary := canary.getView()
	defer oldCanary.destroy()

	b, _ := NewMutable(16)
	c, _ := NewMutable(16)

	b.Copy([]byte("yellow submarine"))
	c.Copy([]byte("yellow submarine"))

	DestroyAll()

	if b.Bytes() != nil || c.Bytes() != nil {
		t.Error("expected buffers to be nil")
	}

	if b.IsMutable() || c.IsMutable() {
		t.Error("expected permissions to be immutable")
	}

	if !b.IsDestroyed() || !c.IsDestroyed() {
		t.Error("expected it to be destroyed")
	}

	if b.key.x != nil || c.key.x != nil {
		t.Error("keys not destroyed")
	}

	newCanary := canary.getView()
	defer newCanary.destroy()

	if bytes.Equal(oldCanary.buffer, newCanary.buffer) {
		t.Error("canary didn't refresh")
	}
}

func TestSize(t *testing.T) {
	b, _ := NewMutable(16)

	if b.Size() != 16 {
		t.Error("unexpected size")
	}

	b.Destroy()

	if b.Size() != 0 {
		t.Error("unexpected size")
	}
}

func TestWipe(t *testing.T) {
	b, _ := NewMutableFromBytes([]byte("yellow submarine"))

	if err := b.Wipe(); err != nil {
		t.Error("failed to wipe:", err)
	}

	if !bytes.Equal(b.Bytes(), make([]byte, 16)) {
		t.Error("bytes not wiped; b =", b.Bytes())
	}

	b.FillRandomBytes()
	b.MakeImmutable()

	if err := b.Wipe(); err != ErrImmutable {
		t.Error("expected ErrImmutable")
	}

	if bytes.Equal(b.Bytes(), make([]byte, 16)) {
		t.Error("bytes wiped")
	}

	b.MakeMutable()
	b.FillRandomBytes()
	b.Destroy()

	if err := b.Wipe(); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestConcatenate(t *testing.T) {
	a, _ := NewImmutableFromBytes([]byte("xxxx"))
	b, _ := NewMutableFromBytes([]byte("yyyy"))

	c, err := Concatenate(a, b)
	if err != nil {
		t.Error("unexpected error")
	}

	if !bytes.Equal(c.Bytes(), []byte("xxxxyyyy")) {
		t.Error("unexpected output;", c.Bytes())
	}
	if c.IsMutable() {
		t.Error("expected immutability")
	}

	a.Destroy()
	b.Destroy()
	c.Destroy()

	if _, err := Concatenate(a, b); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestDuplicate(t *testing.T) {
	b, _ := NewImmutableFromBytes([]byte("test"))

	c, err := Duplicate(b)
	if err != nil {
		t.Error("unexpected error")
	}
	if !bytes.Equal(b.Bytes(), c.Bytes()) {
		t.Error("duplicated buffer has different contents")
	}
	if c.IsMutable() {
		t.Error("permissions not copied")
	}
	b.Destroy()
	c.Destroy()

	if _, err := Duplicate(b); err != ErrDestroyed {
		t.Error("expected ErrDestroyed")
	}
}

func TestEqual(t *testing.T) {
	b, _ := NewMutable(16)
	c, _ := NewMutable(16)

	equal, err := Equal(b, c)
	if err != nil {
		t.Error("unexpected error")
	}
	if !equal {
		t.Error("should be equal")
	}

	a, _ := NewMutable(8)
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
	a, _ := NewImmutableFromBytes([]byte("xxxxyyyy"))

	b, c, err := Split(a, 4)
	if err != nil {
		t.Error("unexpected error")
	}
	if !bytes.Equal(b.Bytes(), []byte("xxxx")) {
		t.Error("first buffer has unexpected value")
	}
	if !bytes.Equal(c.Bytes(), []byte("yyyy")) {
		t.Error("second buffer has unexpected value")
	}
	if b.IsMutable() || c.IsMutable() {
		t.Error("permissions not preserved")
	}
	if !bytes.Equal(a.Bytes(), []byte("xxxxyyyy")) {
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
	b, _ := NewImmutableFromBytes([]byte("xxxxyyyy"))

	c, err := Trim(b, 2, 4)
	if err != nil {
		t.Error("unexpected error")
	}

	if !bytes.Equal(c.Bytes(), []byte("xxyy")) {
		t.Error("unexpected value:", c.Bytes())
	}

	if c.IsMutable() {
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

func TestWipeBytes(t *testing.T) {
	// Create random byte slice.
	b := make([]byte, 32)
	fillRandBytes(b)

	// Wipe it.
	WipeBytes(b)

	// Check.
	if !bytes.Equal(b, make([]byte, 32)) {
		t.Error("unsuccessful wipe")
	}

	// Try with empty list.
	ebuf := make([]byte, 0)
	WipeBytes(ebuf)
	if len(ebuf) != 0 || cap(ebuf) != 0 {
		t.Error("changes made to zero-sized slice")
	}
}

func TestCatchInterrupt(t *testing.T) {
	CatchInterrupt(func() {})

	var i int
	for x := 0; x < 1024; x++ {
		catchInterruptOnce.Do(func() {
			i++
		})
	}
	if i != 0 {
		t.Error("sync.Once failed")
	}
}

func TestConcurrent(t *testing.T) {
	var wg sync.WaitGroup

	b, _ := NewMutable(16)
	for i := 0; i < 1024; i++ {
		wg.Add(1)
		go func() {
			CatchInterrupt(func() {
				return
			})

			b.MakeImmutable()
			b.MakeMutable()

			b.Move([]byte("Test"))
			b.Copy([]byte("test"))

			b.FillRandomBytes()

			b.Wipe()

			wg.Done()
		}()
	}

	wg.Wait()
	b.Destroy()
}

func TestDisableUnixCoreDumps(t *testing.T) {
	if err := DisableUnixCoreDumps(); err != nil {
		t.Error(err)
	}
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
	b, err := NewMutable(8)
	if err != nil {
		t.Error("unexpected error")
	}
	ib := b.container

	c, err := NewImmutable(8)
	if err != nil {
		t.Error("unexpected error")
	}
	ic := c.container

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

func TestSetRekeyInterval(t *testing.T) {
	oldValue := interval
	SetRekeyInterval(oldValue + 1)
	if interval != oldValue+1 {
		t.Error("unexpected interval value")
	}
}

func TestNewSubclave(t *testing.T) {
	s := newSubclave()
	defer s.destroy()
	sv := s.getView()
	defer sv.destroy()

	if len(s.x) != 32 || len(s.y) != 32 {
		t.Error("unexpected subclave length")
	}
	if cap(s.x) != 32 || cap(s.y) != 32 {
		t.Error("unexpected subclave capacity")
	}

	if bytes.Equal(sv.buffer, make([]byte, 32)) {
		t.Error("subclave is zero")
	}
}

func TestSubclaveIO(t *testing.T) {
	s := newSubclave()
	defer s.destroy()
	sv := s.getView()
	defer sv.destroy()

	randVal := r()
	s.update(randVal)
	nsv := s.getView()
	defer nsv.destroy()

	if bytes.Equal(sv.buffer, nsv.buffer) {
		t.Error("update subclave val didn't work")
	}
}

func TestSubclaveViewDestroy(t *testing.T) {
	s := newSubclave()
	defer s.destroy()
	sv := s.getView()
	val := sv.buffer
	sv.destroy()

	if sv.buffer != nil {
		t.Error("could not properly destroy subclave")
	}

	sv = s.getView()
	if !bytes.Equal(sv.buffer, val) {
		t.Error("unexpectedly changed subclave value")
	}
}

func TestSubclaveRefresh(t *testing.T) {
	s := newSubclave()
	defer s.destroy()
	oldValue := s.getView()
	defer oldValue.destroy()
	s.refresh()
	newValue := s.getView()
	defer newValue.destroy()
	if bytes.Equal(oldValue.buffer, newValue.buffer) {
		t.Error("subclave refresh unsuccessful")
	}
}

func TestSubclaveRekey(t *testing.T) {
	s := newSubclave()
	defer s.destroy()
	oldValue := s.getView()
	defer oldValue.destroy()
	s.rekey()
	newValue := s.getView()
	defer newValue.destroy()
	if !bytes.Equal(oldValue.buffer, newValue.buffer) {
		t.Error("subclave rekey changed value")
	}
}

func TestSubclaveDestroy(t *testing.T) {
	s := newSubclave()
	s.destroy()
	if len(s.x) != 0 || len(s.y) != 0 {
		t.Error("could not destroy")
	}
}
