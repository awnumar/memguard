package core

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"
)

func TestRoundToPageSize(t *testing.T) {
	fmt.Println("System page size:", pageSize, "bytes")

	if roundToPageSize(0) != 0 {
		t.Error("failed with test input 0")
	}
	if roundToPageSize(1) != pageSize {
		t.Error("failed with test input 1")
	}
	if roundToPageSize(pageSize) != pageSize {
		t.Error("failed with test input page_size")
	}
	if roundToPageSize(pageSize+1) != 2*pageSize {
		t.Error("failed with test input page_size + 1")
	}
}

func TestGetBytes(t *testing.T) {
	// Allocate an ordinary buffer.
	buffer := make([]byte, 32)

	// Get am alternate reference to it using our slice builder.
	derived := getBufferPart(buffer, 0, len(buffer))

	// Check for naive equality.
	if !bytes.Equal(buffer, derived) {
		t.Error("naive equality check failed")
	}

	// Modify and check if the change was reflected in both.
	buffer[0] = 1
	buffer[31] = 1
	if !bytes.Equal(buffer, derived) {
		t.Error("modified equality check failed")
	}

	// Do a deep comparison.
	if uintptr(unsafe.Pointer(&buffer[0])) != uintptr(unsafe.Pointer(&derived[0])) {
		t.Error("pointer values differ")
	}
	if len(buffer) != len(derived) || cap(buffer) != cap(derived) {
		t.Error("length or capacity values differ")
	}
}
