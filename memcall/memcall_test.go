package memcall

import "testing"

func TestCycle(t *testing.T) {
	buffer, err := Alloc(32)
	if err != nil {
		t.Error(err)
	}

	// Test if the whole memory is filled with 0xdb.
	for i := 0; i < 32; i++ {
		if buffer[i] != byte(0xdb) {
			t.Error("unexpected byte:", buffer[i])
		}
	}

	if err := Protect(buffer, true, true); err != nil {
		t.Error(err)
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
	if err := Protect(buffer, true, true); err != nil {
		t.Error(err)
	}
	if err := Protect(buffer, true, false); err != nil {
		t.Error(err)
	}
	if err := Protect(buffer, false, true); err != nil {
		t.Error(err)
	}
	if err := Protect(buffer, false, false); err != nil {
		t.Error(err)
	}
	Free(buffer)
}
