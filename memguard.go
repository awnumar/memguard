package memguard

import (
	"fmt"
	"os"
	"sync"

	"github.com/libeclipse/memguard/memlock"
)

var (
	// Count of how many goroutines there are.
	lockersCount int

	// Let the goroutines know we're exiting.
	isExiting = make(chan bool)

	// Used to wait for goroutines to finish before exiting.
	lockers sync.WaitGroup
)

// Protect prevents memory from being paged to disk, follows it
// around until program exit, then zeros it out and unlocks it.
func Protect(data []byte) {
	// Increment counters since we're starting another goroutine.
	lockersCount++ // Normal counter.
	lockers.Add(1) // WaitGroup counter.

	// Run as a goroutine so that callers don't have to be explicit.
	go func(b []byte) {
		// Monitor if we managed to lock b.
		lockSuccess := true

		// Prevent memory from being paged to disk.
		err := memlock.Lock(b)
		if err != nil {
			lockSuccess = false
			fmt.Printf("Warning: Failed to lock %p; will still zero it out on exit. [Err: %s]\n", &b, err)
		}

		// Wait for the signal to let us know we're exiting.
		<-isExiting

		// Zero out the memory.
		Wipe(b)

		// If we managed to lock earlier, unlock.
		if lockSuccess {
			err := memlock.Unlock(b)
			if err != nil {
				fmt.Printf("Warning: Failed to unlock %p [Err: %s]\n", &b, err)
			}
		}

		// We're done. Decrement WaitGroup counter.
		lockers.Done()
	}(data)
}

// Cleanup instructs the goroutines to cleanup the
// memory they've been watching and waits for them to finish.
func Cleanup() {
	// Send the exit signal to all of the goroutines.
	for n := 0; n < lockersCount; n++ {
		isExiting <- true
	}

	// Wait for them all to finish.
	lockers.Wait()
}

// MakeProtected creates a byte slice of length l, but protects it before returning.
func MakeProtected(l int) []byte {
	// Create a byte slice of length l and protect it.
	b := make([]byte, l)
	Protect(b)

	// Return the created slice.
	return b
}

// Wipe takes a byte slice and zeroes it out.
func Wipe(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = byte(0)
	}
}

// SafeExit cleans up protected memory and then exits safely.
func SafeExit(c int) {
	// Cleanup protected memory.
	Cleanup()

	// Exit with a specified exit-code.
	os.Exit(c)
}
