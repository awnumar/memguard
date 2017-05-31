/*
Package memguard is designed to allow you to easily handle sensitive values in memory.

The general working cycle is easy to follow:

    // Declare a protected slice and move bytes into it.
    encryptionKey, _ := memguard.New(16, false) // Size 16; Not read-only.
    encryptionKey.Move([]byte("yellow submarine"))

    // Use the buffer wherever you need it.
    Encrypt(encryptionKey.Buffer, plaintext)

    // Destroy it after you're done.
    encryptionKey.Destroy()

As you'll have noted, the example above does not append or assign the key to the buffer, but rather it uses the built-in API function Move.

    b := memguard.New(32)

    b.Move([]byte("...")) // Correct.
    b.Copy([]byte("...")) // Less correct; original buffer isn't wiped.

    b.Buffer = []byte("...")                   // WRONG
    b.Buffer = append(b.Buffer, []byte("...")) // WRONG

If a function that you're using requires an array, you can cast the Buffer to an array and then pass around a pointer. Make sure that you do not dereference the pointer and pass around the resulting value, as this will leave copies all over the place.

    key, _ := memguard.NewFromBytes([]byte("yellow submarine"), false)

    // Make sure the size of the array matches the size of the Buffer.
    // In this case that size is 16. This is very important.
    keyArrayPtr := (*[16]byte)(unsafe.Pointer(&key.Buffer[0]))

The MemGuard API is thread-safe. You can extend this thread-safety to outside of the API functions by using the Mutex that each LockedBuffer exposes. Do not use the mutex when calling a function that is part of the MemGuard API. For example:

    b := memguard.New(4, false)
    b.Lock()
    copy(b.Buffer, []byte("test"))
    b.Unlock()

    c := memguard.New(4, false)
    c.Lock()
    c.Copy([]byte("test")) // This will deadlock.
    c.Unlock()

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
