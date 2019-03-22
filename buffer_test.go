package memguard

import (
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
