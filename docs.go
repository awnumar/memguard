/*
Package memguard is designed to allow you to easily handle sensitive values in memory. The main functionality is to lock and watch portions of memory and wipe them on exit, but there are some supplementary function too.

Multiple operating systems are supported, using `mlock` on Unix and `VirtualLock` on Windows.

    // Declare a protected slice and copy into it.
    encryptionKey := memguard.MakeProtected(32)
    encryptionKey = _generateRandomBytes(32)

It is important to never use append() with sensitive values. Assign directly or call copy() instead.

Although this is less recommended, sometimes it is necessary when the length of the slice is not known in advance. For example, when accepting user input.

    password := input() // Some arbitrary input function.
    memguard.Protect(password)

Arrays can be protected too.

    someArray := [32]byte
    memguard.Protect(someArray[:])

When you're about to exit, call cleanup first. This will wipe and then unlock all protected data.

    memguard.Cleanup()

// It's useful to capture interrupts and signals and cleanup in that case too.

    c := make(chan os.Signal, 2)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        memory.SafeExit(0)
    }()
*/
package memguard
