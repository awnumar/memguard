package memguard

import "os"

func ExampleCatchSignal() {
	// Catches interrupt and kill, cleanly wipe memory and terminate, returning caught signals.
	CatchSignal(NewHandler(func(signals ...os.Signal) interface{} {
		var caught []string
		for _, signal := range signals {
			caught = append(caught, signal.String())
		}
		return caught
	}, true, os.Interrupt, os.Kill)) //, terminating, signals...

	// Catches all signals, cleanly wipe memory and terminate, returning caught signals.
	CatchSignal(NewHandler(func(signals ...os.Signal) interface{} {
		var caught []string
		for _, signal := range signals {
			caught = append(caught, signal.String())
		}
		return caught
	}, true)) //, terminating
}
