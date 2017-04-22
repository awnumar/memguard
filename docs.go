/*
Package memguard is designed to allow you to easily handle sensitive values in memory. The main functionality is to lock and watch portions of memory and wipe them on exit, but there are some supplementary functions too.

    // Declare a protected slice and copy into it.
    encryptionKey := memguard.MakeProtected(32)  // Similar to calling make([]byte, 32)
    copy(encryptionKey, generateRandomBytes(32)) // Copy secure value into the protected slice.

Please note that it is important to never use append() with sensitive values. Only ever copy() into it.

    b := memguard.MakeProtected(32)

    copy(b, []byte("some secure value")) // Correct.

    b = []byte("some secure value")            // WRONG
    b = append(b, []byte("some secure value")) // WRONG

When you do not know the length of the data in advance, you may have to allocate first and then protect, even though this is not recommended. An example is accepting user input.

    password := input() // Some arbitrary input function.
    memguard.Protect(password)

Arrays can be protected too.

    someArray := [32]byte
    memguard.Protect(someArray[:])

When you're about to exit, call cleanup first. This will wipe and then unlock all protected data.

    memguard.Cleanup()

It's useful to capture interrupts and signals and cleanup in that case too.

    c := make(chan os.Signal, 2)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        memory.SafeExit(0)
    }()
*/
package memguard
