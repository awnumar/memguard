// +build !windows

package memguard

import (
	"fmt"
	"net"
	"os"
	"testing"
)

// TODO: run these tests in a subroutine
// https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859
// trick 5: subprocessing
// this will allow removing the dirty testing flag and just checking for exit code

func TestCatchSignal(t *testing.T) {
	testingModeLock.Lock()
	testingMode = true
	testingModeLock.Unlock()

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

	testingModeLock.Lock()
	testingMode = false
	testingModeLock.Unlock()

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
	testingModeLock.Lock()
	testingMode = true
	testingModeLock.Unlock()

	CatchInterrupt()
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Error(nil)
	}
	if err := process.Signal(os.Interrupt); err != nil {
		t.Error(err)
	}

	testingModeLock.Lock()
	testingMode = false
	testingModeLock.Unlock()
}
