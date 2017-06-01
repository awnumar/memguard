/*
Package memguard lets you easily handle sensitive values in memory.

The general working cycle is as follows:

    // Create a new writable LockedBuffer of length 16.
    encryptionKey, err := memguard.New(16, false)
    if err != nil {
        panic(err.Error())
    }
    defer encryptionKey.Destroy()

    // Move bytes into the buffer.
    encryptionKey.Move([]byte("yellow submarine"))

    // Use the buffer wherever you need it.
    Encrypt(encryptionKey.Buffer, plaintext)

As you'll have noted, the example above does not append or assign the key to the buffer, but rather it uses the built-in API function Move.

    b, err := memguard.New(32, false)
    if err != nil {
        panic(err.Error())
    }
    defer b.Destroy()

    b.Move([]byte("...")) // Correct.
    b.Copy([]byte("...")) // Less correct; original buffer isn't wiped.

    b.Buffer = []byte("...")                   // WRONG
    b.Buffer = append(b.Buffer, []byte("...")) // WRONG

The number of LockedBuffers that you are able to create is limited by how much memory your system kernel allows each process to mlock/VirtualLock. Therefore we recommend deferring a Destroy call after creating a new LockedBuffer and calling Destroy on LockedBuffers that you no longer need.

If a function that you're using requires an array, you can cast the Buffer to an array and then pass around a pointer. Make sure that you do not dereference the pointer and pass around the resulting value, as this will leave copies all over the place.

    key, _ := memguard.NewFromBytes([]byte("yellow submarine"), false)

    // Make sure the size of the array matches the size of the Buffer.
    // In this case that size is 16. This is very important.
    keyArrayPtr := (*[16]byte)(unsafe.Pointer(&key.Buffer[0]))

The MemGuard API is thread-safe. You can extend this thread-safety to outside of the API functions by using the Mutex that each LockedBuffer exposes. Don't use the mutex when calling a function that is part of the MemGuard API though, or the process will deadlock.

When terminating your application, care should be taken to securely cleanup everything.

    // Start a listener that will wait for interrupt signals and catch them.
    memguard.CatchInterrupt(func() {
        // Over here put anything you want executing before program exit.
        fmt.Println("Interrupt signal received. Exiting...")
    })

    // Defer a DestroyAll() in your main() function.
    defer memguard.DestroyAll()

    // Use memguard.SafeExit() instead of os.Exit().
    memguard.SafeExit(0) // 0 is the status code.
*/
package memguard
