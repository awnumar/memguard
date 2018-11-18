/*
Package memguard lets you easily handle sensitive values in memory.

    package main

    import (
        "fmt"

        "github.com/awnumar/memguard"
    )

    func main() {
        // Tell memguard to listen out for interrupts, and cleanup in case of one.
        memguard.CatchInterrupt(func() {
            fmt.Println("Interrupt signal received. Exiting...")
        })
        // Make sure to destroy all LockedBuffers when returning.
        defer memguard.DestroyAll()

        // Normal code continues from here.
        foo()
    }

    func foo() {
        // Create a 32 byte, immutable, random key.
        key, err := memguard.NewImmutableRandom(32)
        if err != nil {
            // Oh no, an error. Safely exit.
            fmt.Println(err)
            memguard.SafeExit(1)
        }
        // Remember to destroy this key when the function returns.
        defer key.Destroy()

        // Do something with the key.
        fmt.Printf("This is a %d byte key.\n", key.Size())
        fmt.Printf("This key starts with %x\n", key.Buffer()[0])
    }

The number of LockedBuffers that you are able to create is limited by how much memory your system kernel allows each process to mlock/VirtualLock. Therefore you should call Destroy on LockedBuffers that you no longer need, or simply defer a Destroy call after creating a new LockedBuffer.

If a function that you're using requires an array, you can cast the buffer to an array and then pass around a pointer. Make sure that you do not dereference the pointer and pass around the resulting value, as this will leave copies all over the place.

    key, err := memguard.NewImmutableRandom(16)
    if err != nil {
        fmt.Println(err)
        memguard.SafeExit(1)
    }
    defer key.Destroy()

    // Make sure the size of the array matches the size of the Buffer.
    // In this case that size is 16. This is very important.
    keyArrayPtr := (*[16]byte)(unsafe.Pointer(&key.Buffer()[0]))

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
    memguard.SafeExit(1) // 1 is the exit code.
*/
package memguard
