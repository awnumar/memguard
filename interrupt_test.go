// +build !windows

package memguard

import (
	"net"
	"os"
	"testing"
	"time"
)

func TestCatchInterrupt(t *testing.T) {
	// Start a listener object
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		SafePanic(err)
	}
	defer listener.Close()

	// Spawn a handler.
	CatchInterrupt(func() {
		listener.Close()
	})

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
