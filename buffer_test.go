package memguard

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	mrand "math/rand"
	"os"
	"runtime"
	"testing"
	"unsafe"
)

func TestFinalizer(t *testing.T) {
	b := NewBuffer(32)
	ib := b.Buffer

	runtime.KeepAlive(b)
	// b is now unreachable

	runtime.GC()
	for {
		if !ib.Alive() {
			break
		}
		runtime.Gosched() // should collect b
	}
}

func TestPtrSafetyWithGC(t *testing.T) {
	dataToLock := []byte(`abcdefgh`)
	b := NewBufferFromBytes(dataToLock)
	dataPtr := b.Bytes()

	ib := b.Buffer
	// b is now unreachable

	runtime.GC()
	for {
		if !ib.Alive() {
			break
		}
		runtime.Gosched() // should collect b
	}

	// Check that data hasn't been garbage collected
	if !bytes.Equal(dataPtr, []byte(`abcdefgh`)) {
		t.Error("data does not have the value we set")
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
	if !b.IsMutable() {
		t.Error("buffer should be mutable")
	}
	if !b.IsAlive() {
		t.Error("buffer should not be destroyed")
	}
	b.Destroy()
	b = NewBuffer(0)
	if b.Bytes() != nil {
		t.Error("data slice should be nil")
	}
	if b.Size() != 0 {
		t.Error("size should be zero", b.Size())
	}
	if b.IsAlive() {
		t.Error("buffer should be destroyed")
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
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
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
	if !b.IsAlive() {
		t.Error("buffer should not be destroyed")
	}
	b.Destroy()
	b = NewBufferFromBytes([]byte{})
	if b.Bytes() != nil {
		t.Error("data slice should be nil")
	}
	if b.Size() != 0 {
		t.Error("size should be zero", b.Size())
	}
	if b.IsAlive() {
		t.Error("buffer should be destroyed")
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
}

func TestNewBufferFromReader(t *testing.T) {
	b, err := NewBufferFromReader(rand.Reader, 4096)
	if err != nil {
		t.Error(err)
	}
	if b.Size() != 4096 {
		t.Error("buffer of incorrect size")
	}
	if bytes.Equal(b.Bytes(), make([]byte, 4096)) {
		t.Error("didn't read from reader")
	}
	if b.IsMutable() {
		t.Error("expected buffer to be immutable")
	}
	b.Destroy()

	r := bytes.NewReader([]byte("yellow submarine"))
	b, err = NewBufferFromReader(r, 16)
	if err != nil {
		t.Error(err)
	}
	if b.Size() != 16 {
		t.Error("buffer of incorrect size")
	}
	if !bytes.Equal(b.Bytes(), []byte("yellow submarine")) {
		t.Error("incorrect data")
	}
	if b.IsMutable() {
		t.Error("expected buffer to be immutable")
	}
	b.Destroy()

	r = bytes.NewReader([]byte("yellow submarine"))
	b, err = NewBufferFromReader(r, 17)
	if err == nil {
		t.Error("expected error got nil;", err)
	}
	if b.Size() != 16 {
		t.Error("incorrect size")
	}
	if !bytes.Equal(b.Bytes(), []byte("yellow submarine")) {
		t.Error("incorrect data")
	}
	if b.IsMutable() {
		t.Error("expected buffer to be immutable")
	}
	b.Destroy()

	r = bytes.NewReader([]byte(""))
	b, err = NewBufferFromReader(r, 32)
	if err == nil {
		t.Error("expected error got nil")
	}
	if b.IsAlive() {
		t.Error("expected destroyed buffer")
	}
	if b.IsMutable() {
		t.Error("expected immutable buffer")
	}
	if b.Size() != 0 {
		t.Error("expected nul sized buffer")
	}
	r = bytes.NewReader([]byte("yellow submarine"))
	b, err = NewBufferFromReader(r, 0)
	if err != nil {
		t.Error(err)
	}
	if b.Bytes() != nil {
		t.Error("data slice should be nil")
	}
	if b.Size() != 0 {
		t.Error("size should be zero", b.Size())
	}
	if b.IsAlive() {
		t.Error("buffer should be destroyed")
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
}

type s struct {
	count int
}

func (reader *s) Read(p []byte) (n int, err error) {
	if mrand.Intn(2) == 0 {
		return 0, nil
	}
	reader.count++
	if reader.count == 5000 {
		copy(p, []byte{1})
		return 1, nil
	}
	copy(p, []byte{0})
	return 1, nil
}

func TestNewBufferFromReaderUntil(t *testing.T) {
	data := make([]byte, 5000)
	data[4999] = 1
	r := bytes.NewReader(data)
	b, err := NewBufferFromReaderUntil(r, 1)
	if err != nil {
		t.Error(err)
	}
	if b.Size() != 4999 {
		t.Error("buffer has incorrect size")
	}
	for i := range b.Bytes() {
		if b.Bytes()[i] != 0 {
			t.Error("incorrect data")
		}
	}
	if b.IsMutable() {
		t.Error("expected buffer to be immutable")
	}
	b.Destroy()

	r = bytes.NewReader(data[:32])
	b, err = NewBufferFromReaderUntil(r, 1)
	if err == nil {
		t.Error("expected error got nil")
	}
	if b.Size() != 32 {
		t.Error("invalid size")
	}
	for i := range b.Bytes() {
		if b.Bytes()[i] != 0 {
			t.Error("incorrect data")
		}
	}
	if b.IsMutable() {
		t.Error("expected buffer to be immutable")
	}
	b.Destroy()

	r = bytes.NewReader([]byte{'x'})
	b, err = NewBufferFromReaderUntil(r, 'x')
	if err != nil {
		t.Error(err)
	}
	if b.Size() != 0 {
		t.Error("expected no data")
	}
	if b.IsAlive() {
		t.Error("expected dead buffer")
	}

	r = bytes.NewReader([]byte(""))
	b, err = NewBufferFromReaderUntil(r, 1)
	if err == nil {
		t.Error("expected error got nil")
	}
	if b.IsAlive() {
		t.Error("expected destroyed buffer")
	}
	if b.IsMutable() {
		t.Error("expected immutable buffer")
	}
	if b.Size() != 0 {
		t.Error("expected nul sized buffer")
	}

	rr := new(s)
	b, err = NewBufferFromReaderUntil(rr, 1)
	if err != nil {
		t.Error(err)
	}
	if b.Size() != 4999 {
		t.Error("invalid size")
	}
	for i := range b.Bytes() {
		if b.Bytes()[i] != 0 {
			t.Error("invalid data")
		}
	}
	if b.IsMutable() {
		t.Error("expected buffer to be immutable")
	}
	b.Destroy()
}

type ss struct {
	count int
}

func (reader *ss) Read(p []byte) (n int, err error) {
	if mrand.Intn(2) == 0 {
		return 0, nil
	}
	reader.count++
	if reader.count == 5000 {
		return 0, io.EOF
	}
	copy(p, []byte{0})
	return 1, nil
}

type se struct {
	count int
}

func (reader *se) Read(p []byte) (n int, err error) {
	copy(p, []byte{0})
	reader.count++
	if reader.count == 5000 {
		return 1, errors.New("shut up bro")
	}
	return 1, nil
}

func TestNewBufferFromEntireReader(t *testing.T) {
	r := bytes.NewReader([]byte("yellow submarine"))
	b, err := NewBufferFromEntireReader(r)
	if err != nil {
		t.Error(err)
	}
	if b.Size() != 16 {
		t.Error("incorrect size", b.Size())
	}
	if !b.EqualTo([]byte("yellow submarine")) {
		t.Error("incorrect data", b.String())
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
	b.Destroy()

	data := make([]byte, 16000)
	ScrambleBytes(data)
	r = bytes.NewReader(data)
	b, err = NewBufferFromEntireReader(r)
	if err != nil {
		t.Error(err)
	}
	if b.Size() != len(data) {
		t.Error("incorrect size", b.Size())
	}
	if !b.EqualTo(data) {
		t.Error("incorrect data")
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
	b.Destroy()

	r = bytes.NewReader([]byte{})
	b, err = NewBufferFromEntireReader(r)
	if err != nil {
		t.Error(err)
	}
	if b.Size() != 0 {
		t.Error("buffer should be nil size")
	}
	if b.IsAlive() {
		t.Error("buffer should appear destroyed")
	}

	rr := new(ss)
	b, err = NewBufferFromEntireReader(rr)
	if err != nil {
		t.Error(err)
	}
	if b.Size() != 4999 {
		t.Error("incorrect size", b.Size())
	}
	if !b.EqualTo(make([]byte, 4999)) {
		t.Error("incorrect data")
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
	b.Destroy()

	re := new(se)
	b, err = NewBufferFromEntireReader(re)
	if err == nil {
		t.Error("expected error got nil")
	}
	if b.Size() != 5000 {
		t.Error(b.Size())
	}
	if !b.EqualTo(make([]byte, 5000)) {
		t.Error("incorrect data")
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
	b.Destroy()

	// real world test
	f, err := os.Open("LICENSE")
	if err != nil {
		t.Error(err)
	}
	data, err = io.ReadAll(f)
	if err != nil {
		t.Error(err)
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		t.Error(err)
	}
	b, err = NewBufferFromEntireReader(f)
	if err != nil {
		t.Error(err)
	}
	if !b.EqualTo(data) {
		t.Error("incorrect data")
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
	b.Destroy()
	f.Close()
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
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
	if !b.IsAlive() {
		t.Error("buffer should not be destroyed")
	}
	b.Destroy()
	b = NewBufferRandom(0)
	if b.Bytes() != nil {
		t.Error("data slice should be nil")
	}
	if b.Size() != 0 {
		t.Error("size should be zero", b.Size())
	}
	if b.IsAlive() {
		t.Error("buffer should be destroyed")
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
}

func TestFreeze(t *testing.T) {
	b := NewBuffer(8)
	if b == nil {
		t.Error("buffer is nil")
	}
	if !b.IsMutable() {
		t.Error("buffer isn't mutable")
	}
	b.Freeze()
	if b.IsMutable() {
		t.Error("buffer did not change to immutable")
	}
	if !bytes.Equal(b.Bytes(), make([]byte, 8)) {
		t.Error("buffer changed value") // also tests readability
	}
	b.Freeze() // Test idempotency
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
	if !bytes.Equal(b.Bytes(), make([]byte, 8)) {
		t.Error("buffer changed value") // also tests readability
	}
	b.Destroy()
	b.Freeze()
	if b.IsMutable() {
		t.Error("buffer is mutable")
	}
	if b.IsAlive() {
		t.Error("buffer should be destroyed")
	}
	b = newNullBuffer()
	b.Freeze()
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
}

func TestMelt(t *testing.T) {
	b := NewBuffer(8)
	if b == nil {
		t.Error("buffer is nil")
	}
	b.Freeze()
	if b.IsMutable() {
		t.Error("buffer is mutable")
	}
	b.Melt()
	if !b.IsMutable() {
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
	if !b.IsMutable() {
		t.Error("buffer should be mutable")
	}
	b.Bytes()[0] = 0x2
	if b.Bytes()[0] != 0x2 {
		t.Error("buffer value not changed")
	}
	b.Destroy()
	b.Melt()
	if b.IsMutable() {
		t.Error("buffer shouldn't be mutable")
	}
	if b.IsAlive() {
		t.Error("buffer should be destroyed")
	}
	b = newNullBuffer()
	b.Melt()
	if b.IsMutable() {
		t.Error("buffer should be immutable")
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
	if b.IsAlive() {
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
	b = newNullBuffer()
	b.Copy([]byte("yellow submarine"))
}

func TestCopyAt(t *testing.T) {
	b := NewBuffer(8)
	if b == nil {
		t.Error("got nil buffer")
	}
	b.CopyAt(0, []byte("1234"))
	if !bytes.Equal(b.Bytes()[:4], []byte("1234")) {
		t.Error("copy unsuccessful")
	}
	if !bytes.Equal(b.Bytes()[4:], []byte{0, 0, 0, 0}) {
		t.Error("copy overflow")
	}
	b.CopyAt(4, []byte("5678"))
	if !bytes.Equal(b.Bytes(), []byte("12345678")) {
		t.Error("copy unsuccessful")
	}
	b.Destroy()
	b.CopyAt(4, []byte("hmmm"))
	if b.Bytes() != nil {
		t.Error("buffer should be destroyed")
	}
	b = newNullBuffer()
	b.CopyAt(4, []byte("yellow submarine"))
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
	b = newNullBuffer()
	b.Move([]byte("yellow submarine"))
}

func TestMoveAt(t *testing.T) {
	b := NewBuffer(8)
	if b == nil {
		t.Error("got nil buffer")
	}
	data := []byte("12345678")
	b.MoveAt(0, data[:4])
	if !bytes.Equal(b.Bytes()[:4], []byte("1234")) {
		t.Error("copy unsuccessful")
	}
	if !bytes.Equal(b.Bytes()[4:], []byte{0, 0, 0, 0}) {
		t.Error("copy overflow")
	}
	b.MoveAt(4, data[4:])
	if !bytes.Equal(b.Bytes(), []byte("12345678")) {
		t.Error("copy unsuccessful")
	}
	if !bytes.Equal(data, make([]byte, 8)) {
		t.Error("buffer not wiped")
	}
	b.Destroy()
	b.MoveAt(4, []byte("hmmm"))
	if b.Bytes() != nil {
		t.Error("buffer should be destroyed")
	}
	b = newNullBuffer()
	b.MoveAt(4, []byte("yellow submarine"))
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
	b = newNullBuffer()
	b.Scramble()
}

func TestWipe(t *testing.T) {
	b := NewBufferRandom(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	b.Melt()
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
	b = newNullBuffer()
	b.Wipe()
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
	b = newNullBuffer()
	if b.Size() != 0 {
		t.Error("size should be zero")
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
	if !b.IsAlive() {
		t.Error("buffer should be alive")
	}
	if !b.IsMutable() {
		t.Error("buffer should be mutable")
	}
	b.Destroy()
	if b.Bytes() != nil {
		t.Error("expected buffer to be nil")
	}
	if len(b.Bytes()) != 0 || cap(b.Bytes()) != 0 {
		t.Error("buffer sizes incorrect")
	}
	if b.IsAlive() {
		t.Error("buffer should be destroyed")
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
	b.Destroy()
	if b.Bytes() != nil {
		t.Error("expected buffer to be nil")
	}
	if len(b.Bytes()) != 0 || cap(b.Bytes()) != 0 {
		t.Error("buffer sizes incorrect")
	}
	if b.IsAlive() {
		t.Error("buffer should be destroyed")
	}
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
	b = newNullBuffer()
	b.Destroy()
	if b.IsAlive() {
		t.Error("buffer should be dead")
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
	if b.IsAlive() != b.IsAlive() {
		t.Error("states don't match")
	}
	b.Destroy()
	if b.IsAlive() {
		t.Error("invalid state")
	}
	if b.IsAlive() != b.IsAlive() {
		t.Error("states don't match")
	}
	b = newNullBuffer()
	if b.IsAlive() {
		t.Error("buffer should be dead")
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
	if b.IsMutable() != b.IsMutable() {
		t.Error("states don't match")
	}
	b.Freeze()
	if b.IsMutable() {
		t.Error("invalid state")
	}
	if b.IsMutable() != b.IsMutable() {
		t.Error("states don't match")
	}
	b.Destroy()
	if b.IsMutable() {
		t.Error("invalid state")
	}
	if b.IsMutable() != b.IsMutable() {
		t.Error("states don't match")
	}
	b = newNullBuffer()
	if b.IsMutable() {
		t.Error("buffer should be immutable")
	}
}

func TestEqualTo(t *testing.T) {
	b := NewBufferFromBytes([]byte("yellow submarine"))
	if !b.EqualTo([]byte("yellow submarine")) {
		t.Error("comparison incorrect")
	}
	if b.EqualTo([]byte("yellow")) {
		t.Error("comparison incorrect")
	}
	b.Destroy()
	if b.EqualTo([]byte("yellow submarine")) {
		t.Error("comparison with destroyed should be false")
	}
	b = newNullBuffer()
	if !b.EqualTo([]byte{}) {
		t.Error("buffer should be size zero")
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
	b.Melt()
	b.Bytes()[8] = ^b.Bytes()[8]
	if !bytes.Equal(b.Buffer.Data(), b.Bytes()) {
		t.Error("methods disagree")
	}
	b.Destroy()
	if b.Bytes() != nil {
		t.Error("expected nil buffer")
	}
	b = newNullBuffer()
	if b.Bytes() != nil {
		t.Error("buffer should be nil")
	}
}

func TestReader(t *testing.T) {
	b := NewBufferRandom(32)
	c, err := NewBufferFromReader(b.Reader(), 32)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(b.Bytes(), c.Bytes()) {
		t.Error("data not equal")
	}
	b.Destroy()
	c.Destroy()
	if c.Reader().Size() != 0 {
		t.Error("expected nul reader")
	}
	b = newNullBuffer()
	if c.Reader().Size() != 0 {
		t.Error("expected nul reader")
	}
}

func TestString(t *testing.T) {
	b := NewBufferRandom(32)
	b.Melt()
	s := b.String()
	for i := range b.Bytes() {
		b.Bytes()[i] = 'x'
		if string(b.Bytes()) != s {
			t.Error("string does not map same memory")
		}
	}
	b.Destroy()
	s = b.String()
	if s != "" {
		t.Error("string should be empty")
	}
	b = newNullBuffer()
	if s != "" {
		t.Error("string should be empty")
	}
}

func TestUint16(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	u16 := b.Uint16()
	if len(u16) != 16 || cap(u16) != 16 {
		t.Error("sizes incorrect")
	}
	if uintptr(unsafe.Pointer(&u16[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(3)
	if b == nil {
		t.Error("got nil buffer")
	}
	u16 = b.Uint16()
	if len(u16) != 1 || cap(u16) != 1 {
		t.Error("sizes should be 1")
	}
	if uintptr(unsafe.Pointer(&u16[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(1)
	if b == nil {
		t.Error("got nil buffer")
	}
	u16 = b.Uint16()
	if u16 != nil {
		t.Error("expected nil slice")
	}
	b.Destroy()
	if b.Uint16() != nil {
		t.Error("expected nil slice as buffer destroyed")
	}
	b = newNullBuffer()
	if b.Uint16() != nil {
		t.Error("should be nil")
	}
}

func TestUint32(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	u32 := b.Uint32()
	if len(u32) != 8 || cap(u32) != 8 {
		t.Error("sizes incorrect")
	}
	if uintptr(unsafe.Pointer(&u32[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(5)
	if b == nil {
		t.Error("got nil buffer")
	}
	u32 = b.Uint32()
	if len(u32) != 1 || cap(u32) != 1 {
		t.Error("sizes should be 1")
	}
	if uintptr(unsafe.Pointer(&u32[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(3)
	if b == nil {
		t.Error("got nil buffer")
	}
	u32 = b.Uint32()
	if u32 != nil {
		t.Error("expected nil slice")
	}
	b.Destroy()
	if b.Uint32() != nil {
		t.Error("expected nil slice as buffer destroyed")
	}
	b = newNullBuffer()
	if b.Uint32() != nil {
		t.Error("should be nil")
	}
}

func TestUint64(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	u64 := b.Uint64()
	if len(u64) != 4 || cap(u64) != 4 {
		t.Error("sizes incorrect")
	}
	if uintptr(unsafe.Pointer(&u64[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(9)
	if b == nil {
		t.Error("got nil buffer")
	}
	u64 = b.Uint64()
	if len(u64) != 1 || cap(u64) != 1 {
		t.Error("sizes should be 1")
	}
	if uintptr(unsafe.Pointer(&u64[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(7)
	if b == nil {
		t.Error("got nil buffer")
	}
	u64 = b.Uint64()
	if u64 != nil {
		t.Error("expected nil slice")
	}
	b.Destroy()
	if b.Uint64() != nil {
		t.Error("expected nil slice as buffer destroyed")
	}
	b = newNullBuffer()
	if b.Uint64() != nil {
		t.Error("should be nil")
	}
}

func TestInt8(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	i8 := b.Int8()
	if len(i8) != 32 || cap(i8) != 32 {
		t.Error("sizes incorrect")
	}
	if uintptr(unsafe.Pointer(&i8[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	if b.Int8() != nil {
		t.Error("expected nil slice as buffer destroyed")
	}
	b = newNullBuffer()
	if b.Int8() != nil {
		t.Error("should be nil")
	}
}

func TestInt16(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	i16 := b.Int16()
	if len(i16) != 16 || cap(i16) != 16 {
		t.Error("sizes incorrect")
	}
	if uintptr(unsafe.Pointer(&i16[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(3)
	if b == nil {
		t.Error("got nil buffer")
	}
	i16 = b.Int16()
	if len(i16) != 1 || cap(i16) != 1 {
		t.Error("sizes should be 1")
	}
	if uintptr(unsafe.Pointer(&i16[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(1)
	if b == nil {
		t.Error("got nil buffer")
	}
	i16 = b.Int16()
	if i16 != nil {
		t.Error("expected nil slice")
	}
	b.Destroy()
	if b.Int16() != nil {
		t.Error("expected nil slice as buffer destroyed")
	}
	b = newNullBuffer()
	if b.Int16() != nil {
		t.Error("should be nil")
	}
}

func TestInt32(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	i32 := b.Int32()
	if len(i32) != 8 || cap(i32) != 8 {
		t.Error("sizes incorrect")
	}
	if uintptr(unsafe.Pointer(&i32[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(5)
	if b == nil {
		t.Error("got nil buffer")
	}
	i32 = b.Int32()
	if len(i32) != 1 || cap(i32) != 1 {
		t.Error("sizes should be 1")
	}
	if uintptr(unsafe.Pointer(&i32[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(3)
	if b == nil {
		t.Error("got nil buffer")
	}
	i32 = b.Int32()
	if i32 != nil {
		t.Error("expected nil slice")
	}
	b.Destroy()
	if b.Int32() != nil {
		t.Error("expected nil slice as buffer destroyed")
	}
	b = newNullBuffer()
	if b.Int32() != nil {
		t.Error("should be nil")
	}
}

func TestInt64(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	i64 := b.Int64()
	if len(i64) != 4 || cap(i64) != 4 {
		t.Error("sizes incorrect")
	}
	if uintptr(unsafe.Pointer(&i64[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(9)
	if b == nil {
		t.Error("got nil buffer")
	}
	i64 = b.Int64()
	if len(i64) != 1 || cap(i64) != 1 {
		t.Error("sizes should be 1")
	}
	if uintptr(unsafe.Pointer(&i64[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(7)
	if b == nil {
		t.Error("got nil buffer")
	}
	i64 = b.Int64()
	if i64 != nil {
		t.Error("expected nil slice")
	}
	b.Destroy()
	if b.Int64() != nil {
		t.Error("expected nil slice as buffer destroyed")
	}
	b = newNullBuffer()
	if b.Int32() != nil {
		t.Error("should be nil")
	}
}

func TestByteArray8(t *testing.T) {
	b := NewBuffer(8)
	if b == nil {
		t.Error("got nil buffer")
	}
	if uintptr(unsafe.Pointer(&b.ByteArray8()[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(7)
	if b == nil {
		t.Error("got nil buffer")
	}
	if b.ByteArray8() != nil {
		t.Error("expected nil byte array")
	}
	b.Destroy()
	if b.ByteArray8() != nil {
		t.Error("expected nil byte array from destroyed buffer")
	}
	b = newNullBuffer()
	if b.ByteArray8() != nil {
		t.Error("should be nil")
	}
}

func TestByteArray16(t *testing.T) {
	b := NewBuffer(16)
	if b == nil {
		t.Error("got nil buffer")
	}
	if uintptr(unsafe.Pointer(&b.ByteArray16()[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(15)
	if b == nil {
		t.Error("got nil buffer")
	}
	if b.ByteArray16() != nil {
		t.Error("expected nil byte array")
	}
	b.Destroy()
	if b.ByteArray16() != nil {
		t.Error("expected nil byte array from destroyed buffer")
	}
	b = newNullBuffer()
	if b.ByteArray16() != nil {
		t.Error("should be nil")
	}
}

func TestByteArray32(t *testing.T) {
	b := NewBuffer(32)
	if b == nil {
		t.Error("got nil buffer")
	}
	if uintptr(unsafe.Pointer(&b.ByteArray32()[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(31)
	if b == nil {
		t.Error("got nil buffer")
	}
	if b.ByteArray32() != nil {
		t.Error("expected nil byte array")
	}
	b.Destroy()
	if b.ByteArray32() != nil {
		t.Error("expected nil byte array from destroyed buffer")
	}
	b = newNullBuffer()
	if b.ByteArray32() != nil {
		t.Error("should be nil")
	}
}

func TestByteArray64(t *testing.T) {
	b := NewBuffer(64)
	if b == nil {
		t.Error("got nil buffer")
	}
	if uintptr(unsafe.Pointer(&b.ByteArray64()[0])) != uintptr(unsafe.Pointer(&b.Bytes()[0])) {
		t.Error("pointer locations differ")
	}
	b.Destroy()
	b = NewBuffer(63)
	if b == nil {
		t.Error("got nil buffer")
	}
	if b.ByteArray64() != nil {
		t.Error("expected nil byte array")
	}
	b.Destroy()
	if b.ByteArray64() != nil {
		t.Error("expected nil byte array from destroyed buffer")
	}
	b = newNullBuffer()
	if b.ByteArray64() != nil {
		t.Error("should be nil")
	}
}
