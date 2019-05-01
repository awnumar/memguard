package memguard

import (
	"github.com/awnumar/memguard/core"
	"github.com/awnumar/memguard/crypto"
	"github.com/awnumar/memguard/memcall"
)

/*
DisableUnixCoreDumps disables core dumps in he event of a fault. Since core-dumps are only relevant on Unix systems on windows it will do nothing and return immediately.

This function is precautionary as core-dumps are usually disabled by default on most systems.
*/
func DisableUnixCoreDumps() {
	memcall.DisableCoreDumps()
}

/*
ScrambleBytes overwrites an arbitrary buffer with cryptographically-secure random bytes.
*/
func ScrambleBytes(buf []byte) {
	if err := crypto.MemScr(buf); err != nil {
		core.Panic(err)
	}
}

/*
WipeBytes overwrites an arbitrary buffer with zeroes.
*/
func WipeBytes(buf []byte) {
	crypto.MemClr(buf)
}

/*
Purge resets the session key to a fresh value and destroys all existing LockedBuffers. Existing Enclave objects will no longer be decryptable.
*/
func Purge() {
	core.Purge()
}

/*
SafePanic wipes all it can before calling panic(v).
*/
func SafePanic(v interface{}) {
	core.Panic(v)
}

/*
SafeExit destroys everything sensitive before exiting with a specified status code.
*/
func SafeExit(c int) {
	core.Exit(c)
}
