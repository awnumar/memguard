package memcall

import "testing"

func TestInit(t *testing.T) {
	//Init()
}

func TestCycle(t *testing.T) {
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
}
