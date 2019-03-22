// +build !windows

package memguard

import (
	"net"
	"os"
	"testing"
	"time"
)

func TestCatchSignal(t *testing.T) {
	// Start a listener object
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		SafePanic(err)
	}
	defer listener.Close()

	// Spawn a handler to catch interrupts
	handler := NewHandler(func(signals ...os.Signal) interface{} {
		// Close the listener
		listener.Close()

		// Return the signals we caught.
		return signals
	}, true, os.Interrupt)
	CatchSignal(handler) // Activate the handler.

	// Grab a handle on the running process
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Error(nil)
	}

	// Send it an interrupt signal
	if err := process.Signal(os.Interrupt); err != nil {
		t.Error(err)
	}
	time.Sleep(8 * time.Second)
}
