package memguard

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/awnumar/memguard/core"
	"github.com/awnumar/memguard/crypto"
	"github.com/awnumar/memguard/memcall"
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

/*
DisableUnixCoreDumps disables core dumps in he event of a fault. Since core-dumps are only relevant on Unix systems on windows it will do nothing and return immediately.

This function is precautionary as core-dumps are usually disabled by default on most systems.
*/
func DisableUnixCoreDumps() {
	memcall.DisableCoreDumps()
}

/*
ScrambleBytes overwrites an arbitrary buffer with cryptographically-secure random bytes.
*/
func ScrambleBytes(buf []byte) {
	if err := crypto.MemScr(buf); err != nil {
		core.Panic(err)
	}
}

/*
WipeBytes overwrites an arbitrary buffer with zeroes.
*/
func WipeBytes(buf []byte) {
	crypto.MemClr(buf)
}

/*
Purge resets the session key to a fresh value and destroys all existing LockedBuffers. Any Enclave objects will no longer be decryptable.
*/
func Purge() {
	core.Purge()
}

/*
SafePanic wipes all it can before calling panic(v) for you.
*/
func SafePanic(v interface{}) {
	core.Panic(v)
}

/*
SafeExit destroys everything sensitive before exiting with a specified status code.
*/
func SafeExit(c int) {
	core.Exit(c)
}
