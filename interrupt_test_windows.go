// +build windows

package memguard

import (
	"net"
	"os"
	"syscall"
	"testing"
	"time"
)

func sendCtrlBreak(pid int) error {
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		return e
	}

	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		return e
	}

	r, _, e := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r == 0 {
		return e
	}
	return nil
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

	// Send ourselves an interrupt
	if err := sendCtrlBreak(os.Getpid()); err != nil {
		t.Error(err)
	}

	time.Sleep(8 * time.Second)
}
