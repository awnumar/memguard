package memguard

import (
	"runtime"
	"testing"

	"github.com/awnumar/memguard/core"
)

func TestFinalizer(t *testing.T) {
	b, err := NewBuffer(32)
	if err != nil {
		t.Error("expected nil err; got", err)
	}
	ib := b.Buffer

	runtime.KeepAlive(b)
	// b is now unreachable

	runtime.GC()
	for {
		state := core.GetBufferState(ib)
		if !state.IsAlive {
			break
		}
		runtime.Gosched() // should collect b
	}
}
