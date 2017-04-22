package memguard

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

// Protect spawns a goroutine that prevents the data from being swapped out to disk,
// and then waits around for the signal from Cleanup(). When this signal arrives,
// the goroutine zeroes out the memory that it was protecting, and then unlocks it
// before returning. Protect can be called multiple times with different pieces of data,
// but the caller should be aware that the underlying kernel may impose its own limits
// on the amount of memory that can be locked. For this reason, it is recommended to only
// call this function on small, highly sensitive structures that contain, for example,
// encryption keys. In the event of a limit being reached and attaining the lock fails,
// a warning will be written to stdout and the goroutine will still wipe the memory on exit.
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

// Make creates a byte slice of specified length, but protects it before returning.
// You can also specify an optional capacity, just like with the native make()
// function. Note that the returned array is only properly protected up until
// its length, and not its capacity.
func Make(length int, capacity ...int) (b []byte) {
	// Check if arguments are valid.
	if len(capacity) > 1 {
		panic("memguard.Make: too many arguments")
	} else if len(capacity) > 0 {
		if length > capacity[0] {
			panic("memguard.Make: length larger than capacity")
		}
	}

	// Create a byte slice of length l and protect it.
	if len(capacity) != 0 {
		b = make([]byte, length, capacity[0])
	} else {
		b = make([]byte, length)
	}

	// Protect the byte slice.
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

// CatchInterrupt starts a goroutine that monitors for interrupt
// signals and calls Cleanup() before exiting.
func CatchInterrupt() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		SafeExit(0)
	}()
}

// SafeExit cleans up protected memory and then exits safely.
func SafeExit(c int) {
	// Cleanup protected memory.
	Cleanup()

	// Exit with a specified exit-code.
	os.Exit(c)
}
