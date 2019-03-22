package memguard

import "os"

func ExampleCatchSignal() {
	// Catches interrupt signals, outputs them, and exits.
	handler := NewHandler(func(signals ...os.Signal) []byte {
		var s []byte
		for _, signal := range signals {
			s = append(s, []byte(signal.String())...)
		}
		return s
	}, true, os.Interrupt)
	CatchSignal(handler)
}
