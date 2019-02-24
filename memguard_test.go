package memguard

// import (
// 	"bytes"
// 	"runtime"
// 	"sync"
// 	"testing"
// 	"unsafe"

// 	"github.com/awnumar/memguard/crypto"
// )

// func TestNewImmutable(t *testing.T) {
// 	b, err := NewImmutable(8)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	if !b.sealed {
// 		t.Error("container should be sealed")
// 	}
// 	b.unseal()
// 	for i := range b.plaintext {
// 		if b.plaintext[i] != 0 {
// 			t.Error("buffer not zero-filled", b.plaintext)
// 		}
// 	}
// 	if len(b.Bytes()) != 8 || cap(b.Bytes()) != 8 {
// 		t.Error("length or capacity != required; len, cap =", len(b.Bytes()), cap(b.Bytes()))
// 	}
// 	if b.IsMutable() {
// 		t.Error("unexpected state")
// 	}
// 	b.Destroy()

// 	c, err := NewMutable(0)
// 	if err != ErrInvalidLength {
// 		t.Error("expected err; got nil")
// 	}
// 	if c != nil {
// 		t.Error("expected nil, got *Enclave")
// 	}
// }

// func TestNewMutable(t *testing.T) {
// 	b, err := NewMutable(8)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	if !b.sealed {
// 		t.Error("container should be sealed")
// 	}
// 	b.unseal()
// 	for i := range b.plaintext {
// 		if b.plaintext[i] != 0 {
// 			t.Error("buffer not zero-filled", b.plaintext)
// 		}
// 	}
// 	if len(b.Bytes()) != 8 || cap(b.Bytes()) != 8 {
// 		t.Error("length or capacity != required; len, cap =", len(b.Bytes()), cap(b.Bytes()))
// 	}
// 	if !b.IsMutable() {
// 		t.Error("unexpected state")
// 	}
// 	b.Destroy()

// 	c, err := NewMutable(0)
// 	if err != ErrInvalidLength {
// 		t.Error("expected err; got nil")
// 	}
// 	if c != nil {
// 		t.Error("expected nil, got *Enclave")
// 	}
// }

// func TestNewImmutableFromBytes(t *testing.T) {
// 	b, err := NewImmutableFromBytes([]byte("test"))
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	if !b.sealed {
// 		t.Error("container should be sealed")
// 	}
// 	b.unseal()
// 	if !bytes.Equal(b.Bytes(), []byte("test")) {
// 		t.Error("b.Bytes() != required")
// 	}
// 	if b.IsMutable() {
// 		t.Error("unexpected state")
// 	}
// 	b.Destroy()

// 	c, err := NewImmutableFromBytes([]byte(""))
// 	if err != ErrInvalidLength {
// 		t.Error("expected err; got nil")
// 	}
// 	if c != nil {
// 		t.Error("expected nil, got *Enclave")
// 	}
// }

// func TestNewMutableFromBytes(t *testing.T) {
// 	b, err := NewMutableFromBytes([]byte("test"))
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	if !b.sealed {
// 		t.Error("container should be sealed")
// 	}
// 	b.unseal()
// 	if !bytes.Equal(b.Bytes(), []byte("test")) {
// 		t.Error("b.Bytes() != required")
// 	}
// 	if !b.IsMutable() {
// 		t.Error("unexpected state")
// 	}
// 	b.Destroy()

// 	c, err := NewMutableFromBytes([]byte(""))
// 	if err != ErrInvalidLength {
// 		t.Error("expected err; got nil")
// 	}
// 	if c != nil {
// 		t.Error("expected nil, got *Enclave")
// 	}
// }

// func TestNewImmutableRandom(t *testing.T) {
// 	b, _ := NewImmutableRandom(32)
// 	if !b.sealed {
// 		t.Error("container should be sealed")
// 	}
// 	b.unseal()
// 	if bytes.Equal(b.Bytes(), make([]byte, 32)) {
// 		t.Error("was not filled with random data")
// 	}
// 	if b.IsMutable() {
// 		t.Error("unexpected state")
// 	}
// 	b.Destroy()

// 	c, err := NewImmutableRandom(0)
// 	if err != ErrInvalidLength {
// 		t.Error("expected ErrInvalidLength")
// 	}
// 	if c != nil {
// 		t.Error("expected nil, got *Enclave")
// 	}
// }

// func TestNewMutableRandom(t *testing.T) {
// 	b, _ := NewMutableRandom(32)
// 	if !b.sealed {
// 		t.Error("container should be sealed")
// 	}
// 	b.unseal()
// 	if bytes.Equal(b.Bytes(), make([]byte, 32)) {
// 		t.Error("was not filled with random data")
// 	}
// 	if !b.IsMutable() {
// 		t.Error("unexpected state")
// 	}
// 	b.Destroy()

// 	c, err := NewMutableRandom(0)
// 	if err != ErrInvalidLength {
// 		t.Error("expected ErrInvalidLength")
// 	}
// 	if c != nil {
// 		t.Error("expected nil, got *Enclave")
// 	}
// }

// func TestSealUnseal(t *testing.T) {
// 	b, _ := NewImmutable(32)

// 	if !b.sealed {
// 		t.Error("container should be sealed")
// 	}
// 	if bytes.Equal(b.plaintext, make([]byte, 32)) {
// 		t.Error("contents should be random when sealed")
// 	}

// 	b.Unseal()

// 	if b.sealed {
// 		t.Error("container should be unsealed")
// 	}
// 	if !bytes.Equal(b.plaintext, make([]byte, 32)) {
// 		t.Error("contents should not be random when unsealed")
// 	}
// 	if b.IsMutable() {
// 		t.Error("should remain immutable")
// 	}

// 	b.Reseal()

// 	if !b.sealed {
// 		t.Error("container should be sealed")
// 	}
// 	if bytes.Equal(b.plaintext, make([]byte, 32)) {
// 		t.Error("contents should be random when sealed")
// 	}
// 	if b.IsMutable() {
// 		t.Error("should remain immutable")
// 	}
// }

// func TestBytes(t *testing.T) {
// 	b, _ := NewMutableRandom(8)
// 	b.unseal()

// 	if !bytes.Equal(b.plaintext, b.Bytes()) {
// 		t.Error("buffers inequal")
// 	}

// 	b.Destroy()

// 	if len(b.Bytes()) != 0 || cap(b.Bytes()) != 0 {
// 		t.Error("expected zero length")
// 	}
// }

// func TestUint8(t *testing.T) {
// 	b, _ := NewMutableRandom(8)

// 	conv, err := b.Uint8()
// 	if err != ErrSealed {
// 		t.Error("expected ErrSealed")
// 	}
// 	if conv != nil {
// 		t.Error("expected nil buffer")
// 	}

// 	b.unseal()

// 	x, err := b.Uint8()
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	if !bytes.Equal(b.plaintext, x) {
// 		t.Error("conversion failed")
// 	}

// 	if &b.plaintext[0] != &x[0] {
// 		t.Error("conversion points incorrectly")
// 	}
// 	if len(x) != 8 || cap(x) != 8 {
// 		t.Error("unexpected length or capacity")
// 	}

// 	b.Destroy()

// 	if _, err := b.Uint8(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestUint16(t *testing.T) {
// 	b, _ := NewMutable(8)
// 	c, _ := NewMutable(9)

// 	conv, err := b.Uint16()
// 	if err != ErrSealed {
// 		t.Error("expected ErrSealed")
// 	}
// 	if conv != nil {
// 		t.Error("expected nil buffer")
// 	}

// 	b.unseal()
// 	c.unseal()

// 	x, err := b.Uint16()
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	_, err = c.Uint16()
// 	if err != ErrInvalidConversion {
// 		t.Error("expected ErrInvalidConversion")
// 	}

// 	if unsafe.Pointer(&b.plaintext[0]) != unsafe.Pointer(&x[0]) {
// 		t.Error("conversion points incorrectly")
// 	}
// 	if len(x) != 4 || cap(x) != 4 {
// 		t.Error("unexpected length or capacity")
// 	}

// 	b.Destroy()
// 	c.Destroy()

// 	if _, err := b.Uint16(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestUint32(t *testing.T) {
// 	b, _ := NewMutable(8)
// 	c, _ := NewMutable(9)

// 	conv, err := b.Uint32()
// 	if err != ErrSealed {
// 		t.Error("expected ErrSealed")
// 	}
// 	if conv != nil {
// 		t.Error("expected nil buffer")
// 	}

// 	b.unseal()
// 	c.unseal()

// 	x, err := b.Uint32()
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	_, err = c.Uint32()
// 	if err != ErrInvalidConversion {
// 		t.Error("expected ErrInvalidConversion")
// 	}

// 	if unsafe.Pointer(&b.plaintext[0]) != unsafe.Pointer(&x[0]) {
// 		t.Error("conversion points incorrectly")
// 	}
// 	if len(x) != 2 || cap(x) != 2 {
// 		t.Error("unexpected length or capacity")
// 	}

// 	b.Destroy()
// 	c.Destroy()

// 	if _, err := b.Uint32(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestUint64(t *testing.T) {
// 	b, _ := NewMutable(8)
// 	c, _ := NewMutable(9)

// 	conv, err := b.Uint64()
// 	if err != ErrSealed {
// 		t.Error("expected ErrSealed")
// 	}
// 	if conv != nil {
// 		t.Error("expected nil buffer")
// 	}

// 	b.unseal()
// 	c.unseal()

// 	x, err := b.Uint64()
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	_, err = c.Uint64()
// 	if err != ErrInvalidConversion {
// 		t.Error("expected ErrInvalidConversion")
// 	}

// 	if unsafe.Pointer(&b.plaintext[0]) != unsafe.Pointer(&x[0]) {
// 		t.Error("conversion points incorrectly")
// 	}
// 	if len(x) != 1 || cap(x) != 1 {
// 		t.Error("unexpected length or capacity")
// 	}

// 	b.Destroy()
// 	c.Destroy()

// 	if _, err := b.Uint64(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestInt8(t *testing.T) {
// 	b, _ := NewMutable(8)
// 	c, _ := NewMutable(9)

// 	conv, err := b.Int8()
// 	if err != ErrSealed {
// 		t.Error("expected ErrSealed")
// 	}
// 	if conv != nil {
// 		t.Error("expected nil buffer")
// 	}

// 	b.unseal()
// 	c.unseal()

// 	x, err := b.Int8()
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}

// 	if unsafe.Pointer(&b.plaintext[0]) != unsafe.Pointer(&x[0]) {
// 		t.Error("conversion points incorrectly")
// 	}
// 	if len(x) != 8 || cap(x) != 8 {
// 		t.Error("unexpected length or capacity")
// 	}

// 	b.Destroy()
// 	c.Destroy()

// 	if _, err := b.Int8(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestInt16(t *testing.T) {
// 	b, _ := NewMutable(8)
// 	c, _ := NewMutable(9)

// 	conv, err := b.Int16()
// 	if err != ErrSealed {
// 		t.Error("expected ErrSealed")
// 	}
// 	if conv != nil {
// 		t.Error("expected nil buffer")
// 	}

// 	b.unseal()
// 	c.unseal()

// 	x, err := b.Int16()
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	_, err = c.Int16()
// 	if err != ErrInvalidConversion {
// 		t.Error("expected ErrInvalidConversion")
// 	}

// 	if unsafe.Pointer(&b.plaintext[0]) != unsafe.Pointer(&x[0]) {
// 		t.Error("conversion points incorrectly")
// 	}
// 	if len(x) != 4 || cap(x) != 4 {
// 		t.Error("unexpected length or capacity")
// 	}

// 	b.Destroy()
// 	c.Destroy()

// 	if _, err := b.Int16(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestInt32(t *testing.T) {
// 	b, _ := NewMutable(8)
// 	c, _ := NewMutable(9)

// 	conv, err := b.Int32()
// 	if err != ErrSealed {
// 		t.Error("expected ErrSealed")
// 	}
// 	if conv != nil {
// 		t.Error("expected nil buffer")
// 	}

// 	b.unseal()
// 	c.unseal()

// 	x, err := b.Int32()
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	_, err = c.Int32()
// 	if err != ErrInvalidConversion {
// 		t.Error("expected ErrInvalidConversion")
// 	}

// 	if unsafe.Pointer(&b.plaintext[0]) != unsafe.Pointer(&x[0]) {
// 		t.Error("conversion points incorrectly")
// 	}
// 	if len(x) != 2 || cap(x) != 2 {
// 		t.Error("unexpected length or capacity")
// 	}

// 	b.Destroy()
// 	c.Destroy()

// 	if _, err := b.Int32(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestInt64(t *testing.T) {
// 	b, _ := NewMutable(8)
// 	c, _ := NewMutable(9)

// 	conv, err := b.Int64()
// 	if err != ErrSealed {
// 		t.Error("expected ErrSealed")
// 	}
// 	if conv != nil {
// 		t.Error("expected nil buffer")
// 	}

// 	b.unseal()
// 	c.unseal()

// 	x, err := b.Int64()
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	_, err = c.Int64()
// 	if err != ErrInvalidConversion {
// 		t.Error("expected ErrInvalidConversion")
// 	}

// 	if unsafe.Pointer(&b.plaintext[0]) != unsafe.Pointer(&x[0]) {
// 		t.Error("conversion points incorrectly")
// 	}
// 	if len(x) != 1 || cap(x) != 1 {
// 		t.Error("unexpected length or capacity")
// 	}

// 	b.Destroy()
// 	c.Destroy()

// 	if _, err := b.Int64(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestIsMutable(t *testing.T) {
// 	b, _ := NewMutable(8)

// 	if b.IsMutable() != true {
// 		t.Error("incorrect value")
// 	}

// 	b.MakeImmutable()
// 	if b.IsMutable() != false {
// 		t.Error("incorrect value")
// 	}
// }

// func TestIsDestroyed(t *testing.T) {
// 	b, _ := NewMutable(8)

// 	if b.IsDestroyed() != false {
// 		t.Error("incorrect value")
// 	}

// 	b.Destroy()
// 	if b.IsDestroyed() != true {
// 		t.Error("incorrect value")
// 	}
// }

// func TestIsSealed(t *testing.T) {
// 	b, _ := NewMutable(8)
// 	if !b.IsSealed() {
// 		t.Error("should be sealed")
// 	}
// 	b.unseal()
// 	if b.IsSealed() {
// 		t.Error("should be unsealed")
// 	}
// 	b.reseal()
// 	if !b.IsSealed() {
// 		t.Error("should be sealed")
// 	}
// }

// func TestEqualBytes(t *testing.T) {
// 	a, _ := NewMutableFromBytes([]byte("test"))

// 	equal, err := a.EqualBytes([]byte("test"))
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}

// 	if !equal {
// 		t.Error("should be equal")
// 	}

// 	equal, err = a.EqualBytes([]byte("toast"))
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}

// 	if equal {
// 		t.Error("should not be equal")
// 	}

// 	a.Destroy()

// 	if equal, err := a.EqualBytes([]byte("test")); equal || err != ErrDestroyed {
// 		t.Error("unexpected return values with destroyed Enclave")
// 	}
// }

// func TestImmutable(t *testing.T) {
// 	b, _ := NewMutable(8)

// 	if err := b.MakeImmutable(); err != nil {
// 		t.Error("unexpected error")
// 	}
// 	if b.IsMutable() {
// 		t.Error("unexpected state")
// 	}
// 	if err := b.MakeMutable(); err != nil {
// 		t.Error("unexpected error")
// 	}
// 	if !b.IsMutable() {
// 		t.Error("unexpected state")
// 	}

// 	b.Destroy()

// 	if err := b.MakeImmutable(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}

// 	if err := b.MakeMutable(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestMove(t *testing.T) {
// 	// When buf is larger than Enclave.
// 	b, _ := NewMutable(16)
// 	buf := []byte("this is a very large buffer")
// 	b.Move(buf)
// 	b.unseal()
// 	if !bytes.Equal(buf, make([]byte, len(buf))) {
// 		t.Error("expected buf to be nil")
// 	}
// 	if !bytes.Equal(b.Bytes(), []byte("this is a very l")) {
// 		t.Error("bytes weren't copied properly")
// 	}
// 	b.Destroy()

// 	// When buf is smaller than Enclave.
// 	b, _ = NewMutable(16)
// 	buf = []byte("diz small buf")
// 	b.Move(buf)
// 	b.unseal()
// 	if !bytes.Equal(buf, make([]byte, len(buf))) {
// 		t.Error("expected buf to be nil")
// 	}
// 	if !bytes.Equal(b.Bytes()[:len(buf)], []byte("diz small buf")) {
// 		t.Error("bytes weren't copied properly")
// 	}
// 	if !bytes.Equal(b.Bytes()[len(buf):], make([]byte, 16-len(buf))) {
// 		t.Error("bytes weren't copied properly;", b.Bytes()[len(buf):])
// 	}
// 	b.Destroy()

// 	// When buf is equal in size to Enclave.
// 	b, _ = NewMutable(16)
// 	buf = []byte("yellow submarine")
// 	b.Move(buf)
// 	b.unseal()
// 	if !bytes.Equal(buf, make([]byte, len(buf))) {
// 		t.Error("expected buf to be nil")
// 	}
// 	if !bytes.Equal(b.Bytes(), []byte("yellow submarine")) {
// 		t.Error("bytes weren't copied properly")
// 	}

// 	b.MakeImmutable()

// 	err := b.Move([]byte("test"))
// 	if err != ErrImmutable {
// 		t.Error("expected ErrImmutable")
// 	}

// 	b.Destroy()

// 	if err := b.Move([]byte("test")); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestFillRandomBytes(t *testing.T) {
// 	a, _ := NewMutable(32)
// 	a.FillRandomBytes()

// 	a.unseal()

// 	if bytes.Equal(a.Bytes(), make([]byte, 32)) {
// 		t.Error("not random")
// 	}

// 	a.reseal()

// 	a.Wipe()
// 	a.FillRandomBytesAt(16, 16)

// 	a.unseal()

// 	if !bytes.Equal(a.Bytes()[:16], make([]byte, 16)) || bytes.Equal(a.Bytes()[16:], make([]byte, 16)) {
// 		t.Error("incorrect offset/size;", a.Bytes()[:16], a.Bytes()[16:])
// 	}

// 	a.reseal()

// 	a.MakeImmutable()
// 	if err := a.FillRandomBytes(); err != ErrImmutable {
// 		t.Error("expected ErrImmutable")
// 	}

// 	a.Destroy()
// 	if err := a.FillRandomBytes(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestDestroyAll(t *testing.T) {
// 	b, _ := NewMutable(16)
// 	c, _ := NewMutable(16)

// 	b.Copy([]byte("yellow submarine"))
// 	c.Copy([]byte("yellow submarine"))

// 	DestroyAll()

// 	if b.Bytes() != nil || c.Bytes() != nil {
// 		t.Error("expected buffers to be nil")
// 	}

// 	if b.IsMutable() || c.IsMutable() {
// 		t.Error("expected permissions to be immutable")
// 	}

// 	if !b.IsDestroyed() || !c.IsDestroyed() {
// 		t.Error("expected it to be destroyed")
// 	}
// }

// func TestSize(t *testing.T) {
// 	b, _ := NewMutable(16)

// 	if b.Size() != 16 {
// 		t.Error("unexpected size")
// 	}

// 	b.Destroy()

// 	if b.Size() != 0 {
// 		t.Error("unexpected size")
// 	}
// }

// func TestWipe(t *testing.T) {
// 	b, _ := NewMutableFromBytes([]byte("yellow submarine"))

// 	if err := b.Wipe(); err != nil {
// 		t.Error("failed to wipe:", err)
// 	}

// 	b.unseal()
// 	if !bytes.Equal(b.Bytes(), make([]byte, 16)) {
// 		t.Error("bytes not wiped; b =", b.Bytes())
// 	}
// 	b.reseal()

// 	b.FillRandomBytes()
// 	b.MakeImmutable()

// 	if err := b.Wipe(); err != ErrImmutable {
// 		t.Error("expected ErrImmutable")
// 	}

// 	b.unseal()
// 	if bytes.Equal(b.Bytes(), make([]byte, 16)) {
// 		t.Error("bytes wiped")
// 	}
// 	b.reseal()

// 	b.MakeMutable()
// 	b.FillRandomBytes()
// 	b.Destroy()

// 	if err := b.Wipe(); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestConcatenate(t *testing.T) {
// 	a, _ := NewMutableFromBytes([]byte("xxxx"))
// 	b, _ := NewMutableFromBytes([]byte("yyyy"))

// 	a.MakeImmutable()

// 	c, err := Concatenate(a, b)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}

// 	c.unseal()
// 	if !bytes.Equal(c.Bytes(), []byte("xxxxyyyy")) {
// 		t.Error("unexpected output;", c.Bytes())
// 	}
// 	c.reseal()
// 	if c.IsMutable() {
// 		t.Error("expected immutability")
// 	}

// 	a.Destroy()
// 	b.Destroy()
// 	c.Destroy()

// 	if _, err := Concatenate(a, b); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestEqual(t *testing.T) {
// 	b, _ := NewMutable(16)
// 	c, _ := NewMutable(16)

// 	equal, err := Equal(b, c)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	if !equal {
// 		t.Error("should be equal")
// 	}

// 	a, _ := NewMutable(8)
// 	equal, err = Equal(a, b)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	if equal {
// 		t.Error("should not be equal")
// 	}

// 	a.Destroy()
// 	b.Destroy()
// 	c.Destroy()

// 	if _, err := Equal(a, b); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }
// func TestDuplicate(t *testing.T) {
// 	b, _ := NewMutableFromBytes([]byte("test"))
// 	b.MakeImmutable()

// 	c, err := Duplicate(b)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}

// 	b.unseal()
// 	c.unseal()
// 	if !bytes.Equal(b.Bytes(), c.Bytes()) {
// 		t.Error("duplicated buffer has different contents")
// 	}
// 	b.reseal()
// 	c.reseal()
// 	if c.IsMutable() {
// 		t.Error("permissions not copied")
// 	}
// 	b.Destroy()
// 	c.Destroy()

// 	if _, err := Duplicate(b); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestSplit(t *testing.T) {
// 	a, _ := NewMutableFromBytes([]byte("xxxxyyyy"))
// 	a.MakeImmutable()

// 	b, c, err := Split(a, 4)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	a.unseal()
// 	b.unseal()
// 	c.unseal()
// 	if !bytes.Equal(b.Bytes(), []byte("xxxx")) {
// 		t.Error("first buffer has unexpected value")
// 	}
// 	if !bytes.Equal(c.Bytes(), []byte("yyyy")) {
// 		t.Error("second buffer has unexpected value")
// 	}
// 	if !bytes.Equal(a.Bytes(), []byte("xxxxyyyy")) {
// 		t.Error("original is not preserved")
// 	}
// 	a.reseal()
// 	b.reseal()
// 	c.reseal()
// 	if b.IsMutable() || c.IsMutable() {
// 		t.Error("permissions not preserved")
// 	}
// 	b.Destroy()
// 	c.Destroy()

// 	if _, _, err := Split(a, 0); err != ErrInvalidLength {
// 		t.Error("expected ErrInvalidLength")
// 	}
// 	if _, _, err := Split(a, 8); err != ErrInvalidLength {
// 		t.Error("expected ErrInvalidLength")
// 	}

// 	a.Destroy()

// 	if _, _, err := Split(a, 4); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestGrow(t *testing.T) {
// 	b, _ := NewImmutableFromBytes([]byte("xxyy"))

// 	c, err := Grow(b, 4)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}

// 	c.unseal()
// 	if !bytes.Equal(c.Bytes()[0:4], []byte("xxyy")) {
// 		t.Error("unexpected value:", c.Bytes())
// 	}
// 	if !bytes.Equal(c.Bytes()[4:8], make([]byte, 4)) {
// 		t.Error("unexpected value:", c.Bytes())
// 	}
// 	c.reseal()

// 	if c.IsMutable() {
// 		t.Error("unexpected state")
// 	}
// 	c.Destroy()
// 	b.Destroy()

// 	if _, err := Grow(b, 4); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestTrim(t *testing.T) {
// 	b, _ := NewImmutableFromBytes([]byte("xxxxyyyy"))

// 	c, err := Trim(b, 2, 4)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}

// 	c.unseal()
// 	if !bytes.Equal(c.Bytes(), []byte("xxyy")) {
// 		t.Error("unexpected value:", c.Bytes())
// 	}
// 	c.reseal()

// 	if c.IsMutable() {
// 		t.Error("unexpected state")
// 	}
// 	c.Destroy()

// 	if _, err := Trim(b, 4, 0); err != ErrInvalidLength {
// 		t.Error("expected ErrInvalidLength")
// 	}

// 	b.Destroy()

// 	if _, err := Trim(b, 2, 4); err != ErrDestroyed {
// 		t.Error("expected ErrDestroyed")
// 	}
// }

// func TestWipeBytes(t *testing.T) {
// 	// Create random byte slice.
// 	b := make([]byte, 32)
// 	if err := crypto.MemScr(b); err != nil {
// 		panic(err)
// 	}

// 	// Wipe it.
// 	WipeBytes(b)

// 	// Check.
// 	if !bytes.Equal(b, make([]byte, 32)) {
// 		t.Error("unsuccessful wipe")
// 	}

// 	// Try with empty list.
// 	ebuf := make([]byte, 0)
// 	WipeBytes(ebuf)
// 	if len(ebuf) != 0 || cap(ebuf) != 0 {
// 		t.Error("changes made to zero-sized slice")
// 	}
// }

// func TestCatchInterrupt(t *testing.T) {
// 	CatchInterrupt(func() {})

// 	var i int
// 	for x := 0; x < 1024; x++ {
// 		catchInterruptOnce.Do(func() {
// 			i++
// 		})
// 	}
// 	if i != 0 {
// 		t.Error("sync.Once failed")
// 	}
// }

// func TestConcurrent(t *testing.T) {
// 	var wg sync.WaitGroup

// 	b, _ := NewMutable(16)
// 	for i := 0; i < 1024; i++ {
// 		wg.Add(1)
// 		go func() {
// 			CatchInterrupt(func() {
// 				return
// 			})

// 			b.MakeImmutable()
// 			b.MakeMutable()

// 			b.Move([]byte("Test"))
// 			b.Copy([]byte("test"))

// 			b.FillRandomBytes()

// 			b.Wipe()

// 			wg.Done()
// 		}()
// 	}

// 	wg.Wait()
// 	b.Destroy()
// }

// func TestDisableUnixCoreDumps(t *testing.T) {
// 	if err := DisableUnixCoreDumps(); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestRoundPage(t *testing.T) {
// 	if roundToPageSize(pageSize) != pageSize {
// 		t.Error("incorrect rounding;", roundToPageSize(pageSize))
// 	}

// 	if roundToPageSize(pageSize+1) != 2*pageSize {
// 		t.Error("incorrect rounding;", roundToPageSize(pageSize+1))
// 	}
// }

// func TestGetBytes(t *testing.T) {
// 	b := []byte("yellow submarine")

// 	ptr := unsafe.Pointer(&b[0])
// 	length := len(b)
// 	bBytes := getBytes(uintptr(ptr), length)

// 	copy(bBytes, []byte("fellow submarine"))

// 	if !bytes.Equal(b, bBytes) {
// 		t.Error("pointer does not describe actual memory")
// 	}
// }

// func TestFinalizer(t *testing.T) {
// 	b, err := NewMutable(8)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	ib := b.container

// 	c, err := NewMutable(8)
// 	if err != nil {
// 		t.Error("unexpected error")
// 	}
// 	ic := c.container

// 	if ib.IsDestroyed() != false {
// 		t.Error("expected b to not be destroyed")
// 	}
// 	if ic.IsDestroyed() != false {
// 		t.Error("expected c to not be destroyed")
// 	}

// 	runtime.KeepAlive(b)
// 	// b is now unreachable

// 	runtime.GC() // should collect b
// 	for ib.IsDestroyed() != true {
// 		runtime.Gosched()
// 	}

// 	if ic.IsDestroyed() != false {
// 		t.Error("expected c to not be destroyed")
// 	}

// 	runtime.KeepAlive(c)
// 	// c is now unreachable

// 	runtime.GC() // should collect c
// 	for ic.IsDestroyed() != true {
// 		runtime.Gosched()
// 	}
// }
