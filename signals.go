package memguard

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/awnumar/memguard/core"
)

var (
	create   sync.Once
	handlers chan *Handler
	listener chan os.Signal
)

/*
Handler is a mutexed container for a handler function.
*/
type Handler struct {
	sync.RWMutex
	handler   func(...os.Signal) interface{}
	signals   []os.Signal
	terminate bool
}

/*
NewHandler constructs a handler object out of a function and a list of signals that will trigger it. This object can be passed as an updatable "config" to CatchSignal. The arguments are

  var handler func(...os.Signal) interface{} // Function that is run on catching a signal. Return value is written to stdout.
  var signals []os.Signal                    // List of signals to listen out for.
  var terminate bool                         // Whether to purge the session and terminate after running handler(<-signals).
*/
func NewHandler(handler func(...os.Signal) interface{}, terminate bool, signals ...os.Signal) *Handler {
	h := new(Handler)
	h.handler = handler
	h.signals = signals
	h.terminate = terminate
	return h
}

/*
CatchSignal assigns a given function to be run in the event of a signal being received by the process. If no signals are provided all signals will be caught.

  i.   Signal is received by the process and caught by the handler routine.
  ii.  Interrupt handler is called and return value written to stdout.
  iii. If handler is terminating, memory is wiped and process terminates.

This function can be called multiple times with the effect that only the last handler to be specified will have any effect.
*/
func CatchSignal(handler *Handler) {
	create.Do(func() {
		// Create the channels the goroutine will listen on.
		handlers = make(chan *Handler, 1)
		listener = make(chan os.Signal, 4*len(handler.signals))

		// Start a goroutine to listen on the channels.
		go func() {
			var f *Handler
			for {
				select {
				case signals := <-listener:
					f.RLock()
					if out := f.handler(signals); out != nil {
						fmt.Printf("Signals caught ::%b::\n", out)
					}
					if f.terminate {
						core.Exit(0)
					}
					f.RUnlock()
				case handler := <-handlers:
					f = handler
				}
			}
		}()

		// Send the handler to the channel to initialise it.
		handlers <- handler
	})

	// Update the handler
	handlers <- handler

	// Notify the channel if we receive a signal.
	signal.Reset()
	signal.Notify(listener, handler.signals...)
}