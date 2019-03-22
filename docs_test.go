package memguard

import "os"

func ExampleCatchSignal() {
	// Catches interrupt and kill, cleanly wipe memory and terminate, returning caught signals.
	handler := NewHandler(func(signals ...os.Signal) interface{} {
		return signals
	}, true, os.Interrupt, os.Kill)
	CatchSignal(handler)
}
