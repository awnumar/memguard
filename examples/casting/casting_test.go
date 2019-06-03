package casting

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/awnumar/memguard"
)

func TestByteArray10(t *testing.T) {
	b, a := ByteArray10()
	memguard.ScrambleBytes(a[:])
	if !bytes.Equal(b.Bytes(), a[:]) {
		t.Error("array describes incorrect memory region")
	}
	b.Destroy()
}

func TestUint64Array4(t *testing.T) {
	b, a := Uint64Array4()
	if uintptr(unsafe.Pointer(&b.Bytes()[0])) != uintptr(unsafe.Pointer(&a[0])) {
		t.Error("start pointer does not match")
	}
	b.Bytes()[24] = 1
	if a[3] != 1 {
		t.Error("incorrect alignment", b.Bytes(), a)
	}
	b.Destroy()
}

func testSecureStruct(b *memguard.LockedBuffer, s *Secure, offset int, t *testing.T) {
	if uintptr(unsafe.Pointer(&b.Bytes()[offset])) != uintptr(unsafe.Pointer(s)) {
		t.Error("pointers don't match")
	}
	memguard.ScrambleBytes(b.Bytes()[offset : offset+32])
	if !bytes.Equal(b.Bytes()[offset:offset+32], s.Key[:]) {
		t.Error("key doesn't match")
	}
	b.Bytes()[offset+32] = 1
	b.Bytes()[offset+40] = 1
	if s.Salt[0] != 1 || s.Salt[1] != 1 {
		t.Error("salt doesn't match")
	}
	b.Bytes()[offset+48] = 1
	if s.Counter != 1 {
		t.Error("counter doesn't match")
	}
	b.Bytes()[offset+56] = 1
	if !s.Something {
		t.Error("bool flag Something doesn't match")
	}
}

func TestSecureStruct(t *testing.T) {
	b, s := SecureStruct()
	testSecureStruct(b, s, 0, t)
	b.Destroy()
}

func TestSecureStructArray(t *testing.T) {
	b, a := SecureStructArray()
	testSecureStruct(b, &a[0], 0, t)
	testSecureStruct(b, &a[1], 64, t)
	b.Destroy()
}

func TestSecureStructSlice(t *testing.T) {
	b, s := SecureStructSlice(3)
	testSecureStruct(b, &s[0], 0, t)
	testSecureStruct(b, &s[1], 64, t)
	testSecureStruct(b, &s[2], 128, t)
	b.Destroy()
}
