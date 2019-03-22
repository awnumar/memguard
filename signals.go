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
CatchSignal assigns a given function to be run in the event of a signal being received by the process.

i.   <- Signal received
ii.  Interrupt handler f() is called
iii. If handler is terminating, memory is wiped and process terminates.

This function can be called multiple times with the effect that the last handler to be specified will be executed.
*/
func CatchSignal(handler *Handler) {
	create.Do(func() {
		// Create the channels the goroutine will listen on.
		handlers = make(chan *Handler, 1)
		listener = make(chan os.Signal, 2*len(handler.signals))

		// Start a goroutine to listen on the channels.
		go func() {
			var f *Handler
			for {
				select {
				case signals := <-listener:
					f.RLock()
					fmt.Printf("Signal caught ::%b::\n", f.handler(signals))
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

/*
NewHandler constructs a handler object out of a function and a list of signals that will trigger it. This object can be passed as an updatable "config" to CatchInterrupt. The arguments are

	var handler func(...os.Signal) []byte // Function that is run on catching a signal. Return value is written to stdout.
	var signals []os.Signal               // List of signals to listen out for.
	var terminate bool					  // Whether to purge the session and terminate after running handler(<-signals).
*/
func NewHandler(handler func(...os.Signal) interface{}, terminate bool, signals ...os.Signal) *Handler {
	h := new(Handler)
	h.handler = handler
	h.signals = signals
	h.terminate = terminate
	return h
}
