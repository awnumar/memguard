/*
Package memguard is designed to allow you to easily handle sensitive values in memory. The main functionality is to lock and watch portions of memory and wipe them on exit, but there are some supplementary functions too.

    // Declare a protected slice and copy into it.
    encryptionKey := memguard.Make(32)  // Similar to calling make([]byte, 32)
    copy(encryptionKey, generateRandomBytes(32))

Please note that it is important to never use append() with sensitive values. Only ever copy() into it.

    b := memguard.Make(32)

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

In order to handle most exit cases, do the following:

    // In the main function.
    memguard.CatchInterrupt()
    defer memguard.Cleanup()

    // Anywhere where you would terminate.
    memguard.SafeExit(0) // 0 is the status code.
*/
package memguard
