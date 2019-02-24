package memcall

import "testing"

func TestCycle(t *testing.T) {
	buffer, err := Alloc(32)
	if err != nil {
		t.Error(err)
	}

	if len(buffer) != 32 || cap(buffer) != 32 {
		t.Error("allocation has invalid size")
	}
	for i := range buffer {
		if buffer[i] != 0 {
			t.Error("allocated memory not zeroed:", buffer)
		}
	}

	if err := Lock(buffer); err != nil {
		t.Error(err)
	}
	if err := Unlock(buffer); err != nil {
		t.Error(err)
	}
	if err := Free(buffer); err != nil {
		t.Error(err)
	}
	if err := DisableCoreDumps(); err != nil {
		t.Error(err)
	}
}

func TestProtect(t *testing.T) {
	buffer, _ := Alloc(32)
	if err := Protect(buffer, ReadWrite); err != nil {
		t.Error(err)
	}
	if err := Protect(buffer, ReadOnly); err != nil {
		t.Error(err)
	}
	if err := Protect(buffer, NoAccess); err != nil {
		t.Error(err)
	}
	if err := Protect(buffer, MemoryProtectionFlag{4}); err != ErrInvalidFlag {
		t.Error("expected error")
	}
	Free(buffer)
}
