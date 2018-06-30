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

        // Why not also disable core dumps on Unix hosts.
        memguard.DisableUnixCoreDumps()

        // Normal code continues from here.
        foo()
    }

    func foo() {
        // Create a new and immutable 16 byte Enclave filled with cryptographically-secure random bytes.
        key, err := memguard.NewImmutableRandom(16)
        if err != nil {
            // Oh no, an error. Safely exit.
            fmt.Println(err)
            memguard.SafeExit(1)
        }
        // Remember to destroy the key when the function returns.
        defer key.Destroy()

        // Unseal the key so that the contents can be accessed externally.
        key.Unseal()

        // Do something with the key.
        fmt.Printf("This is a %d byte key.\n", key.Size())
        fmt.Printf("This key starts with %x\n", key.Bytes()[0])

        // Remember to Reseal it, or defer the Reseal call.
        key.Reseal()

        // Make the memory mutable before editing it. No need to unseal
        // here since the memguard API handles this automatically.
        key.MakeMutable()

        // Move a slice's data into the buffer. The slice's memory is wiped.
        key.Move([]byte("yellow submarine"))

        // Make the buffer immutable again if we don't anticipate more editing.
        key.MakeImmutable()

        // ...
    }

The number of Enclaves that you are able to create is limited by how much memory your system kernel allows each process to mlock/VirtualLock. Therefore you should call Destroy on Enclaves that you no longer need, or simply defer a Destroy call after creating each new Enclave.

If a function that you're using requires an array, you can cast the buffer to an array (without making a copy) and then pass around a pointer. Make sure that you do not dereference the pointer and pass around the resulting value, as this will leave copies all over the place.

    key, err := memguard.NewImmutableRandom(16)
    if err != nil {
        fmt.Println(err)
        memguard.SafeExit(1)
    }
    defer key.Destroy()

    // Get a reference to the buffer's underlying array without making a copy.
    // Also make sure the size of the array matches the size of the buffer. In
    // this case that size is 16. This is very important.
    keyArrayPtr := (*[16]byte)(unsafe.Pointer(&key.Bytes()[0]))

    // Unseal the key so that the buffer holds the actual data instead of random bytes.
    key.Unseal()

    // Do something with the key, passing the pointer without dereferencing.
    Encrypt(plaintext, keyArrayPtr)

    // RESEAL the Enclave afterwards, or defer the Reseal call.
    key.Reseal()

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
