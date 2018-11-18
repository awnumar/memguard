package memguard

import (
	"bytes"
	"testing"
)

func TestArray64(t *testing.T) {
	b, _ := NewImmutableRandom(64)
	defer b.Destroy()
	a, _ := b.ByteArray64()
	buf := b.Buffer()
	if &buf[0] != &a[0] {
		t.Error("Array64 not returning address on buffer.")
	}
	if !bytes.Equal(buf[:64], a[:]) {
		t.Error("Array64 and buffer don't match in data.")
	}
	//
	b, _ = NewImmutableRandom(32)
	defer b.Destroy()
	if _, err := b.ByteArray64(); err == nil {
		t.Error("Array64 did not catch too small buffer.")
	}
}

func TestArray32(t *testing.T) {
	b, _ := NewImmutableRandom(32)
	defer b.Destroy()
	a, _ := b.ByteArray32()
	buf := b.Buffer()
	if &buf[0] != &a[0] {
		t.Error("Array32 not returning address on buffer.")
	}
	if !bytes.Equal(buf[:32], a[:]) {
		t.Error("Array32 and buffer don't match in data.")
	}
	//
	b, _ = NewImmutableRandom(16)
	defer b.Destroy()
	if _, err := b.ByteArray32(); err == nil {
		t.Error("Array32 did not catch too small buffer.")
	}
}

func TestArray16(t *testing.T) {
	b, _ := NewImmutableRandom(16)
	defer b.Destroy()
	a, _ := b.ByteArray16()
	buf := b.Buffer()
	if &buf[0] != &a[0] {
		t.Error("Array16 not returning address on buffer.")
	}
	if !bytes.Equal(buf[:16], a[:]) {
		t.Error("Array16 and buffer don't match in data.")
	}
	//
	b, _ = NewImmutableRandom(8)
	defer b.Destroy()
	if _, err := b.ByteArray16(); err == nil {
		t.Error("Array16 did not catch too small buffer.")
	}
}

/*************************************************************************/

func TestWithReadable(t *testing.T) {
	b, _ := NewMutableRandom(16)
	defer b.Destroy()
	b.MakeUnreadable()
	a, _ := b.ByteArray16()
	if err := b.WithReadable(func() { a[0] = a[1] }); err != nil {
		t.Error("WithReadable may never fail.")
	}
}
