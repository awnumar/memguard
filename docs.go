/*
Package memguard facilitates the easy and secure handling of sensitive memory, in pure Go.

    package main

    import (
        "fmt"

        "github.com/awnumar/memguard"
    )

    func main() {
        // Tell memguard to listen out for interrupts, and clean-up in case of one.
        memguard.CatchInterrupt(func() {
            fmt.Println("Interrupt signal received. Exiting...")
        })
        // Make sure to destroy all Enclaves when returning.
        defer memguard.DestroyAll()

        // Maybe even disable core dumps on unix hosts.
        memguard.DisableUnixCoreDumps()

        // Normal code continues from here.
        foo()
    }

    func foo() {
        // Create a random, cryptographically-secure, 32 byte key.
        key, err := memguard.NewRandom(32)
        if err != nil {
            // Oh no, an error. Safely exit.
            fmt.Println(err)
            memguard.SafeExit(1)
        }
        // Remember to destroy this key when the function returns.
        defer key.Destroy()

        // Unseal the key so that we can view its contents.
        key.Unseal()

        // Do something with the key.
        fmt.Printf("This is a %d byte key.\n", key.Size())
        fmt.Printf("This key starts with %x\n", key.Bytes()[0])

        // Remember to Reseal it, or defer the Reseal call.
        key.Reseal()
    }

The number of Enclaves that you are able to create is limited by how much memory your system kernel allows each process to mlock/VirtualLock. Therefore you should call Destroy on Enclaves that you no longer need, or simply defer a Destroy call after creating each new Enclave.

If a function that you're using requires an array, you can cast the buffer to an array (without making a copy) and then pass around a pointer. Make sure that you do not dereference the pointer and pass around the resulting value, as this will leave copies all over the place.

    key, err := memguard.NewRandom(16)
    if err != nil {
        fmt.Println(err)
        memguard.SafeExit(1)
    }
    defer key.Destroy()

    // Unseal the key so that it's viewable.
    key.Unseal()
    defer key.Reseal()

    // Make sure the size of the array matches the size of the Buffer.
    // In this case that size is 16. This is very important.
    keyArrayPtr := (*[16]byte)(unsafe.Pointer(&key.Bytes()[0]))

The MemGuard API is thread-safe. You can extend this thread-safety to outside of the API functions by using the Mutex that each Enclave exposes. Don't use the mutex when calling a function that is part of the MemGuard API though, or the process will deadlock.

When terminating your application, care should be taken to securely clean-up everything.

    // Start a listener that will wait for interrupt signals and catch them.
    memguard.CatchInterrupt(func() {
        // Over here put anything you want executing before program exit.
        fmt.Println("Interrupt signal received. Exiting...")
    })

    // Defer a DestroyAll() in your main() function.
    defer memguard.DestroyAll()

    // Use memguard.SafeExit() instead of os.Exit().
    memguard.SafeExit(1) // 1 is the exit code.

    // If you must panic, use SafePanic instead.
    memguard.SafePanic(err)
*/
package memguard
