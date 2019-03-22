// +build windows

package memguard

import (
	"net"
	"os"
	"syscall"
	"testing"
	"time"
)

func sendCtrlBreak(pid int) {
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		panic("LoadDLL: %v\n", e)
	}

	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		panic("FindProc: %v\n", e)
	}

	r, _, e := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r == 0 {
		panic("GenerateConsoleCtrlEvent: %v\n", e)
	}
}

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
	process, err := sendCtrlBreak(os.Getpid())
	if err != nil {
		t.Error(nil)
	}

	// Send it an interrupt signal
	if err := process.Signal(os.Interrupt); err != nil {
		t.Error(err)
	}
	time.Sleep(8 * time.Second)
}
