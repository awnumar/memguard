<p align="center">
  <img src="https://cdn.rawgit.com/libeclipse/memguard/master/logo.svg" height="140" />
  <h3 align="center">MemGuard</h3>
  <p align="center">A library that handles the secure handling of sensitive values in memory.</p>
  <p align="center">
    <a href="https://travis-ci.org/libeclipse/memguard"><img src="https://travis-ci.org/libeclipse/memguard.svg?branch=master"></a>
    <a href="https://ci.appveyor.com/project/libeclipse/memguard/branch/master"><img src="https://ci.appveyor.com/api/projects/status/g6cg347cam7lli5m/branch/master?svg=true"></a>
    <a href="https://godoc.org/github.com/libeclipse/memguard"><img src="https://godoc.org/github.com/libeclipse/memguard?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/libeclipse/memguard"><img src="https://goreportcard.com/badge/github.com/libeclipse/memguard"></a>
  </p>
</p>

---

This library is designed to allow you to easily handle sensitive values in memory. The main functionality is to lock and watch portions of memory and wipe them on exit, but there are some supplementary function too.

Multiple operating systems are supported, using `mlock` on Unix and `VirtualLock` on Windows.

## Installation

This library can be retrieved with `go get`.

`go get github.com/libeclipse/memguard`

## Usage

```
// Declare a protected slice and copy into it.
encryptionKey := memguard.MakeProtected(32)
encryptionKey = _generateRandomBytes(32)

// It is important to never use append() with sensitive
// values. Assign directly or call copy() instead.

// Although this is less recommended, sometimes it
// is necessary when the length of the slice is not
// known in advance. For example, when accepting user
// input.

password := input() // Some arbitrary input function.
memguard.Protect(password)

// Arrays can be protected too.
someArray := [32]byte
memguard.Protect(someArray[:])

// When you're about to exit, call cleanup first. This
// will wipe and then unlock all protected data.
memguard.Cleanup()

// It's useful to capture interrupts and signals
// and cleanup in that case too.
c := make(chan os.Signal, 2)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)
go func() {
    <-c
    memory.SafeExit(0)
}()

```
