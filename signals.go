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
