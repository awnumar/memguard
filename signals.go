package memguard

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/awnumar/memguard/core"
)

var (
	interruptHandler = func() {}
)

func init() {
	// Create channel to listen on.
	s := make(chan os.Signal, 2)

	// Notify the channel if we receive a signal.
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to listen on the channel.
	go func() {
		<-s
		interruptHandler()
		core.Exit(0)
	}()
}

/*
CatchInterrupt assigns a given function to be run in the event of an exit signal being received by the process.

i.   <- Signal received
ii.  Interrupt handler f() is called
iii. Memory is securely wiped
iv.  Process terminates

This function can be called multiple times with the effect that the last handler to be specified will be executed.
*/
func CatchInterrupt(f func()) {
	interruptHandler = f
}
