package memguard

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/awnumar/memguard/core"
)

func TestFinalizer(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("nil object")
	}
	ib := b.Buffer

	runtime.KeepAlive(b)
	// b is now unreachable

	runtime.GC()
	for {
		if !core.GetBufferState(ib).IsAlive {
			break
		}
		runtime.Gosched() // should collect b
	}
}

func TestNewBuffer(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("buffer should not be nil")
	}
	if len(b.Bytes()) != 32 || cap(b.Bytes()) != 32 {
		t.Error("buffer sizes incorrect")
	}
	if !bytes.Equal(b.Bytes(), make([]byte, 32)) {
		t.Error("buffer is not zeroed")
	}
	if !core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer should be mutable")
	}
	if !core.GetBufferState(b.Buffer).IsAlive {
		t.Error("buffer should not be destroyed")
	}
	b.Destroy()
}

func TestNewBufferFromBytes(t *testing.T) {
	data := []byte("yellow submarine")
	b := NewBufferFromBytes(data)
	if b == nil {
		t.Error("buffer should not be nil")
	}
	if len(b.Bytes()) != 16 || cap(b.Bytes()) != 16 {
		t.Error("buffer sizes invalid")
	}
	if !bytes.Equal(b.Bytes(), []byte("yellow submarine")) {
		t.Error("data does not match\n", b.Bytes(), "\n", data)
	}
	if !bytes.Equal(data, make([]byte, 16)) {
		t.Error("source buffer not wiped")
	}
	if !core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer should be mutable")
	}
	if !core.GetBufferState(b.Buffer).IsAlive {
		t.Error("buffer should not be destroyed")
	}
	b.Destroy()
}

func TestNewBufferRandom(t *testing.T) {
	b := NewBufferRandom(32)
	if b == nil {
		t.Error("buffer is nil")
	}
	if len(b.Bytes()) != 32 || cap(b.Bytes()) != 32 {
		t.Error("buffer sizes incorrect")
	}
	if bytes.Equal(b.Bytes(), make([]byte, 32)) {
		t.Error("buffer is zeroed")
	}
	if !core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer should be mutable")
	}
	if !core.GetBufferState(b.Buffer).IsAlive {
		t.Error("buffer should not be destroyed")
	}
	b.Destroy()
}

func TestFreeze(t *testing.T) {
	b := NewBuffer(8)
	if b == nil {
		t.Error("buffer is nil")
	}
	if !core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer isn't mutable")
	}
	b.Freeze()
	if core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer did not change to immutable")
	}
	if !bytes.Equal(b.Bytes(), make([]byte, 8)) {
		t.Error("buffer changed value") // also tests readability
	}
	b.Freeze() // Test idempotency
	if core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer should be immutable")
	}
	if !bytes.Equal(b.Bytes(), make([]byte, 8)) {
		t.Error("buffer changed value") // also tests readability
	}
	b.Destroy()
	b.Freeze()
	if core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer is mutable")
	}
	if core.GetBufferState(b.Buffer).IsAlive {
		t.Error("buffer should be destroyed")
	}
}

func TestMelt(t *testing.T) {
	b := NewBuffer(8)
	if b == nil {
		t.Error("buffer is nil")
	}
	b.Freeze()
	if core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer is mutable")
	}
	b.Melt()
	if !core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer did not become mutable")
	}
	if !bytes.Equal(b.Bytes(), make([]byte, 8)) {
		t.Error("buffer changed value") // also tests readability
	}
	b.Bytes()[0] = 0x1 // test writability
	if b.Bytes()[0] != 0x1 {
		t.Error("buffer value not changed")
	}
	b.Melt() // Test idempotency
	if !core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer should be mutable")
	}
	b.Bytes()[0] = 0x2
	if b.Bytes()[0] != 0x2 {
		t.Error("buffer value not changed")
	}
	b.Destroy()
	b.Melt()
	if core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer shouldn't be mutable")
	}
	if core.GetBufferState(b.Buffer).IsAlive {
		t.Error("buffer should be destroyed")
	}
}

func TestSeal(t *testing.T) {
	b := NewBufferRandom(32)
	if b == nil {
		t.Error("buffer is nil")
	}
	data := make([]byte, 32)
	copy(data, b.Bytes())
	e := b.Seal()
	if e == nil {
		t.Error("got nil enclave")
	}
	if core.GetBufferState(b.Buffer).IsAlive {
		t.Error("buffer should be destroyed")
	}
	b, err := e.Open()
	if err != nil {
		t.Error("unexpected error;", err)
	}
	if !bytes.Equal(b.Bytes(), data) {
		t.Error("data does not match")
	}
	b.Destroy()
	e = b.Seal() // call on destroyed buffer
	if e != nil {
		t.Error("expected nil enclave")
	}
}

func TestCopy(t *testing.T) {
	b := NewBuffer(16)
	if b == nil {
		t.Error("buffer is nil")
	}
	b.Copy([]byte("yellow submarine"))
	if !bytes.Equal(b.Bytes(), []byte("yellow submarine")) {
		t.Error("copy unsuccessful")
	}
	b.Destroy()
	b.Copy([]byte("yellow submarine"))
	if b.Bytes() != nil {
		t.Error("buffer should be destroyed")
	}
}

func TestMove(t *testing.T) {
	b := NewBuffer(16)
	if b == nil {
		t.Error("buffer is nil")
	}
	b.Move([]byte("yellow submarine"))
	if !bytes.Equal(b.Bytes(), []byte("yellow submarine")) {
		t.Error("copy unsuccessful")
	}
	data := []byte("yellow submarine")
	b.Move(data)
	for b := range data {
		if data[b] != 0x0 {
			t.Error("buffer was not wiped", b)
		}
	}
	b.Destroy()
	b.Move(data)
	if b.Bytes() != nil {
		t.Error("buffer should be destroyed")
	}
}

func TestScramble(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("buffer is nil")
	}
	b.Scramble()
	if bytes.Equal(b.Bytes(), make([]byte, 32)) {
		t.Error("buffer was not randomised")
	}
	one := make([]byte, 32)
	copy(one, b.Bytes())
	b.Scramble()
	if bytes.Equal(b.Bytes(), make([]byte, 32)) {
		t.Error("buffer was not randomised")
	}
	if bytes.Equal(b.Bytes(), one) {
		t.Error("buffer did not change")
	}
	b.Destroy()
	b.Scramble()
	if b.Bytes() != nil {
		t.Error("buffer should be destroyed")
	}
}

func TestWipe(t *testing.T) {
	b := NewBufferRandom(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	if bytes.Equal(b.Bytes(), make([]byte, 32)) {
		t.Error("buffer was not randomised")
	}
	b.Wipe()
	for i := range b.Bytes() {
		if b.Bytes()[i] != 0 {
			t.Error("buffer was not wiped; index", i)
		}
	}
	b.Destroy()
	b.Wipe()
	if b.Bytes() != nil {
		t.Error("buffer should be destroyed")
	}
}

func TestSize(t *testing.T) {
	b := NewBuffer(1234)
	if b == nil {
		t.Error("got nil buffer")
	}
	if b.Size() != 1234 {
		t.Error("size does not match expected")
	}
	b.Destroy()
	if b.Size() != 0 {
		t.Error("destroyed buffer size should be zero")
	}
}

func TestDestroy(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	if b.Bytes() == nil {
		t.Error("expected buffer to not be nil")
	}
	if len(b.Bytes()) != 32 || cap(b.Bytes()) != 32 {
		t.Error("buffer sizes incorrect")
	}
	if !core.GetBufferState(b.Buffer).IsAlive {
		t.Error("buffer should be alive")
	}
	if !core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer should be mutable")
	}
	b.Destroy()
	if b.Bytes() != nil {
		t.Error("expected buffer to be nil")
	}
	if len(b.Bytes()) != 0 || cap(b.Bytes()) != 0 {
		t.Error("buffer sizes incorrect")
	}
	if core.GetBufferState(b.Buffer).IsAlive {
		t.Error("buffer should be destroyed")
	}
	if core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer should be immutable")
	}
	b.Destroy()
	if b.Bytes() != nil {
		t.Error("expected buffer to be nil")
	}
	if len(b.Bytes()) != 0 || cap(b.Bytes()) != 0 {
		t.Error("buffer sizes incorrect")
	}
	if core.GetBufferState(b.Buffer).IsAlive {
		t.Error("buffer should be destroyed")
	}
	if core.GetBufferState(b.Buffer).IsMutable {
		t.Error("buffer should be immutable")
	}
}

func TestIsAlive(t *testing.T) {
	b := NewBuffer(8)
	if b == nil {
		t.Error("got nil buffer")
	}
	if !b.IsAlive() {
		t.Error("invalid state")
	}
	if b.IsAlive() != core.GetBufferState(b.Buffer).IsAlive {
		t.Error("states don't match")
	}
	b.Destroy()
	if b.IsAlive() {
		t.Error("invalid state")
	}
	if b.IsAlive() != core.GetBufferState(b.Buffer).IsAlive {
		t.Error("states don't match")
	}
}

func TestIsMutable(t *testing.T) {
	b := NewBuffer(8)
	if b == nil {
		t.Error("got nil buffer")
	}
	if !b.IsMutable() {
		t.Error("invalid state")
	}
	if b.IsMutable() != core.GetBufferState(b.Buffer).IsMutable {
		t.Error("states don't match")
	}
	b.Freeze()
	if b.IsMutable() {
		t.Error("invalid state")
	}
	if b.IsMutable() != core.GetBufferState(b.Buffer).IsMutable {
		t.Error("states don't match")
	}
	b.Destroy()
	if b.IsMutable() {
		t.Error("invalid state")
	}
	if b.IsMutable() != core.GetBufferState(b.Buffer).IsMutable {
		t.Error("states don't match")
	}
}

func TestBytes(t *testing.T) {
	b := NewBufferFromBytes([]byte("yellow submarine"))
	if b == nil {
		t.Error("got nil buffer")
	}
	if !bytes.Equal(b.Bytes(), []byte("yellow submarine")) {
		t.Error("not equal contents")
	}
	b.Bytes()[8] = ^b.Bytes()[8]
	if !bytes.Equal(b.Buffer.Data(), b.Bytes()) {
		t.Error("methods disagree")
	}
	b.Destroy()
	if b.Bytes() != nil {
		t.Error("expected nil buffer")
	}
}

func TestUint16(t *testing.T) {

}

func TestUint32(t *testing.T) {

}

func TestUint64(t *testing.T) {

}

func TestInt8(t *testing.T) {

}

func TestInt16(t *testing.T) {

}

func TestInt32(t *testing.T) {

}

func TestInt64(t *testing.T) {

}

func TestByteArray8(t *testing.T) {

}

func TestByteArray16(t *testing.T) {

}

func TestByteArray32(t *testing.T) {

}

func TestByteArray64(t *testing.T) {

}
