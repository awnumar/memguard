package core

import (
	"os"
)

/*
Purge wipes all sensitive data and keys before reinitialising the session with a fresh encryption key and secure values. Subsequent library operations will use these fresh values and the old data is assumed to be practically unrecoverable.

The creation of new Enclave objects should wait for this function to return since subsequent Enclave objects will use the newly created key.

This function should be called before the program terminates, or else the provided Exit or Panic functions should be used to terminate.
*/
func Purge() {
	// Halt the re-key cycle and prevent new enclaves.
	key.Lock()

	// Get a snapshot of existing Buffers.
	snapshot := buffers.flush()

	// Destroy them, performing the usual sanity checks.
	for _, b := range snapshot {
		b.Destroy()
	}

	// Destroy and recreate the key.
	key.Unlock()
	key.Destroy()
	key = NewCoffer()
}

/*
Exit terminates the process with a specified exit code but securely wipes and cleans up sensitive data before doing so.
*/
func Exit(c int) {
	// Wipe the encryption key used to encrypt data inside Enclaves.
	key.Destroy()

	// Get a snapshot of existing Buffers.
	snapshot := buffers.flush()

	// Destroy them, performing the usual sanity checks.
	for _, b := range snapshot {
		b.Destroy()
	}

	// Exit with the specified exit code.
	os.Exit(c)
}

/*
Panic is identical to the builtin panic except it purges the session before calling panic.
*/
func Panic(v interface{}) {
	Purge() // purge creates a new key in case the caller recovers from the panic
	panic(v)
}
