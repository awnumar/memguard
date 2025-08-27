package core

import (
	"fmt"
	"testing"
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
