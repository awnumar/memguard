package memguard

import (
	"os"
	"os/signal"
	"sync"

	"github.com/awnumar/memguard/core"
)

var (
	// Ensure we only start a single signal handling instance
	create sync.Once

	// Channel for updating the signal handler
	sigfunc = make(chan func(os.Signal), 1)

	// Channel that caught signals are sent to by the runtime
	listener = make(chan os.Signal, 4)
)

/*
CatchSignal assigns a given function to be run in the event of a signal being received by the process. If no signals are provided all signals will be caught.

  i.   Signal is received by the process and caught by the handler routine.
  ii.  Interrupt handler is called and return value written to stdout.
  iii. Secure session state is wiped and process terminates.

This function can be called multiple times with the effect that only the call will have any effect. The arguments are

  var handler func(os.Signal) // Function that is run on catching a signal. Signal is passed to function.
  var signals os.Signal...    // List of signals to listen out for. If none provided it will default to all.
*/
func CatchSignal(f func(os.Signal), signals ...os.Signal) {
	create.Do(func() {
		// Start a goroutine to listen on the channels.
		go func() {
			var handler func(os.Signal)
			for {
				select {
				case signal := <-listener:
					handler(signal)
					core.Exit(1)
				case handler = <-sigfunc:
				}
			}
		}()
	})

	// Update the handler function.
	sigfunc <- f

	// Notify the channel if we receive a signal.
	signal.Reset()
	signal.Notify(listener, signals...)
}

/*
CatchInterrupt is a wrapper around CatchSignal that makes it easy to safely handle receiving interrupt signals. If an interrupt is received, the process will wipe sensitive data in memory before terminating.

A subsequent call to CatchSignal will override this call.
*/
func CatchInterrupt() {
	CatchSignal(func(_ os.Signal) {}, os.Interrupt)
}
