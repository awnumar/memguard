// +build !windows

package memguard

import (
	"fmt"
	"net"
	"os"
	"testing"
)

func TestCatchSignal(t *testing.T) {
	testingMode = true

	// Start a listener object
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		SafePanic(err)
	}
	defer listener.Close()

	// Spawn a handler to catch interrupts
	CatchSignal(func(s os.Signal) {
		fmt.Println("Received signal:", s.String())
		listener.Close()
	}, os.Interrupt)

	// Grab a handle on the running process
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Error(nil)
	}

	// Send it an interrupt signal
	if err := process.Signal(os.Interrupt); err != nil {
		t.Error(err)
	}

	// Todo: catch this violation
	//
	// b, err := NewBuffer(32)
	// if err != nil {
	// 	t.Error(err)
	// }
	// bA := (*[64]byte)(unsafe.Pointer(&b.Buffer.Data[0]))
	// bA[42] = 0x69
}

func TestCatchInterrupt(t *testing.T) {
	CatchInterrupt()
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Error(nil)
	}
	if err := process.Signal(os.Interrupt); err != nil {
		t.Error(err)
	}
}
