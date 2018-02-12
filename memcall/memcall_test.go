package memcall

import "testing"

func TestCycle(t *testing.T) {
	DisableCoreDumps()
	buffer := Alloc(32)

	// Test if the whole memory is filled with 0xdb.
	for i := 0; i < 32; i++ {
		if buffer[i] != byte(0xdb) {
			t.Error("unexpected byte:", buffer[i])
		}
	}

	Protect(buffer, true, true)
	Lock(buffer)
	Unlock(buffer)
	Free(buffer)
}

func TestProtect(t *testing.T) {
	buffer := Alloc(32)
	Protect(buffer, true, true)
	Protect(buffer, true, false)
	Protect(buffer, false, true)
	Protect(buffer, false, false)
	Free(buffer)
}
