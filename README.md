<p align="center">
  <img src="https://cdn.rawgit.com/awnumar/memguard/master/logo.svg" height="140" />
  <h3 align="center">MemGuard</h3>
  <p align="center">Easy and secure handling of sensitive memory, in pure Go.</p>
  <p align="center">
    <a href="https://travis-ci.org/awnumar/memguard"><img src="https://travis-ci.org/awnumar/memguard.svg?branch=master"></a>
    <a href="https://ci.appveyor.com/project/awnumar/memguard/branch/master"><img src="https://ci.appveyor.com/api/projects/status/nrtqmdolndm0pcac/branch/master?svg=true"></a>
    <a href="https://godoc.org/github.com/awnumar/memguard"><img src="https://godoc.org/github.com/awnumar/memguard?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/awnumar/memguard"><img src="https://goreportcard.com/badge/github.com/awnumar/memguard"></a>
  </p>
</p>

---

This is a thread-safe package, designed to allow you to easily handle sensitive values in memory. It supports all major operating systems and is written in pure Go.

## Features

* Interference from the garbage-collector is blocked by using system-calls to manually allocate memory.
* It is very difficult for another process to find or access sensitive memory as the data is sandwiched between guard-pages. This feature also acts as an immediate access alarm in case of buffer overflows.
* Buffer overflows are further protected against using a random canary value. If this value changes, the process will panic.
* We try our best to prevent the system from writing anything sensitive to the disk. The data is locked to prevent swapping, system core dumps can be disabled, and the kernel is advised (where possible) to never include the secure memory in dumps.
* True kernel-level immutability is implemented. That means that if _anything_ attempts to modify an immutable container, the kernel will throw an access violation and the process will terminate.
* All sensitive data is wiped before the associated memory is released back to the operating system.
* Side-channel attacks are mitigated against by making sure that the copying and comparison of data is done in constant-time.
* Accidental memory leaks are mitigated against by harnessing Go's own garbage-collector to automatically destroy containers that have run out of scope.

Some of these features were inspired by [libsodium](https://github.com/jedisct1/libsodium), so credits to them.

Full documentation and a complete overview of the API can be found [here](https://godoc.org/github.com/awnumar/memguard).

## Installation

Although we do recommend using a release, the simplest way to install the library is to `go get` it:

```
$ go get github.com/awnumar/memguard
```

If you would prefer a signed release that you can verify and manually compile yourself, download and extract the [latest release](https://github.com/awnumar/memguard/releases/latest). Then go ahead and run:

```
$ go install -v ./
```

The [latest release](https://github.com/awnumar/memguard/releases/latest) is guaranteed to be cryptographically signed with my most recent PGP key, which can be found on [keybase](https://keybase.io/awn). To import it directly into GPG, run:

```
$ curl https://keybase.io/awn/pgp_keys.asc | gpg --import
```

We **strongly** encourage you to vendor your dependencies for a clean and reliable build. Go's [dep](https://github.com/golang/dep) makes this task relatively frictionless.
