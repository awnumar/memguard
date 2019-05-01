package memguard

import (
	"os"
)

func ExampleCatchSignal() {
	// Catch interrupt, doing nothing before terminating safely.
	CatchSignal(func(_ os.Signal) {}, os.Interrupt)
}
