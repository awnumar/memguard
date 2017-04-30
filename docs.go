/*
Package memguard is designed to allow you to easily handle sensitive values in memory.

    // Declare a protected slice and move bytes into it.
    encryptionKey := memguard.New(32)
    encryptionKey.Move(generateRandomBytes(32))

Please note that it is important to never use append() or to assign values directly. Only ever copy() values into protected slices.

    b := memguard.New(32)

    b.Move([]byte("...")) // Correct.
    b.Copy([]byte("...")) // Less correct; original buffer isn't wiped.

    b = []byte("some secure value")            // WRONG
    b = append(b, []byte("some secure value")) // WRONG

When you do not know the length of the data in advance, you may have to allocate first and then protect, even though this is not generally the best way of doing things. An example is accepting user input.

    password := input() // Some arbitrary input function.
    lockedPassword := memguard.NewFromBytes(password)

When you're about to exit, call DestroyAll() first. This will wipe and then unlock all protected data.

    memguard.DestroyAll()

In order to handle most exit cases, do the following:

    // In your main() function.
    memguard.CatchInterrupt(func() {
        // Over here put anything you want executing before
        // program exit. (In case of an interrupt signal.)
    })
    defer memguard.DestroyAll()

    // Anywhere you would terminate.
    memguard.SafeExit(0) // 0 is the status code.

If you would like to disable Unix core dumps for your application, simply do:

    memguard.DisableCoreDumps()
*/
package memguard
