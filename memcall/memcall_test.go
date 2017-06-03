package memcall

import "testing"

func TestCycle(t *testing.T) {
	DisableCoreDumps()
	buffer := Alloc(32)
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
